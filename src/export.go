package main

import (
	"encoding/json"
	"os"
)

// jsonify converts the Result struct to JSON.
func jsonify(r *Result) ([]byte, error) {
	return json.MarshalIndent(r, "", "    ")
}

// csvify converts the Result struct to CSV.
// func csvify(r *Result) ([]byte, error) {
// }


// writeToFile writes jsonText string to the given filename.
func writeToFile(jsonText, filename string) (err error) {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(jsonText)
	return err
}
