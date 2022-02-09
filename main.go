package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func main() {
	// define handlers
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/analysis", func(w http.ResponseWriter, r *http.Request) {
		// check http method
		if r.Method != http.MethodPost {
			WriteAPIResp(w, NewErrorResp(NewErrMethodNotAllowed()))
			return
		}
		// parse request body
		var reqBody analyzeReqBody
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			WriteAPIResp(w, NewErrorResp(NewErrBadRequest(err.Error())))
			return
		}
		// validate request body
		err = reqBody.Validate()
		if err != nil {
			WriteAPIResp(w, NewErrorResp(err))
			return
		}
		// do analysis
		matches := doAnalysis(reqBody.InputText, reqBody.RefText)
		// output success response
		WriteAPIResp(w, NewSuccessResp(map[string]interface{}{
			"matches": matches,
		}))
	})
	// define port, we need to set it as env for Heroku deployment
	port := os.Getenv("PORT")
	if port == "" {
		port = "9056"
	}
	// run server
	log.Printf("server is listening on :%v", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("unable to run server due: %v", err)
	}
}

func doAnalysis(input, ref string) []match {
	// Declare variable to return the result of analysis
	var slice_match []match

	// Convert String become lower case to make sure the document is in the same pattern
	new_ref := strings.ToLower(ref)
	new_input := strings.ToLower(input)

	// Remove some punctuation to clearing the string in input string
	regex_char, _ := regexp.Compile(`[;—:!,'-]`)
	new_input = regex_char.ReplaceAllString(new_input, " ")

	// Clearing double whitespace that comes from the previous step
	new_input = strings.ReplaceAll(new_input, "  ", " ")

	// Split sentence using unique character . and ?
	regex_char_split, _ := regexp.Compile(`[.?]`)
	slice_input := regex_char_split.Split(new_input, -1)

	// Removing whitespace in front of each string
	for i := 0; i<len(slice_input); i++{
		slice_input[i] = strings.TrimSpace(slice_input[i])
	}

	// Remove all the punctuation to clearing the string in refference
	regex_char_ref, _ := regexp.Compile(`[[:punct:]—\t\n\f\r ]`)
	new_ref = regex_char_ref.ReplaceAllString(new_ref, " ")

	// Remove double whitespacing that comes from regex string replacement
	new_ref = strings.ReplaceAll(new_ref, "  ", " ")

	// Declare variable to get the similar index on input string
	count_indeks_input := 0

	// Process the analysis for each sentences compare to the refference
	for i := 0; i < len(slice_input); i++ {
		check_string := slice_input[i]

		// Ignore meaningless sentence
		if len(check_string) < 3{
			continue
		}

		// Debug sentences in console
		print(check_string, "\n")

		// Checking substring of the refference string
		idx := strings.Index(new_ref, check_string)
		if idx == -1 {
			continue
		}

		// Fill variable data with the content of similar string
		data := match{
			Input: matchDetails{
				Text:     slice_input[i],
				StartIdx: count_indeks_input,
				EndIdx:   len(slice_input[i]) - 1,
			},
			Reference: matchDetails{
				Text:     new_ref[idx : idx+len(slice_input[i])],
				StartIdx: idx,
				EndIdx:   idx + len(slice_input[i]) - 1,
			},
		}

		// Append data into slice
		slice_match = append(slice_match, data)

		// Count input index to get the key index on the input string
		count_indeks_input += len(check_string) - 1

	}

	return slice_match

}