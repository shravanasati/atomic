package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"
)

// consolify prints the benchmark summary of the Result struct to the console, with color codes.
func textify(r *Result) {
	text := `
Benchmarking Summary
--------------------

Started: {{ .Started }} 
Ended: {{ .Ended }} 
Executed Command: {{ .Command }} 
Total iterations: {{ .Iterations }} 
Average time taken: {{ .Average }} 
`
	tmpl, err := template.New("summary").Parse(text)
	if err != nil {
		panic(err)
	}

	f, ferr := os.Create("bench-summary.txt")
	if ferr != nil {
		fmt.Println(RED + "Failed to create the file." + RESET)
	}
	defer f.Close()
	if terr := tmpl.Execute(f, r); terr != nil {
		fmt.Println(RED + "Failed to write to the file." + RESET)
	}
}

func markdownify(r *Result) {
	text := `
# bench-summary

| Fields             | Values          |
| -----------        | -----------     |
| Started            | {{.Started}}    |
| Ended              | {{.Ended}}      |
| Executed Command   | {{.Command}}    |
| Total iterations   | {{.Iterations}} |
| Average time taken | {{.Average}}    |
`
	tmpl, err := template.New("summary").Parse(text)
	if err != nil {
		panic(err)
	}

	f, ferr := os.Create("bench-summary.md")
	if ferr != nil {
		fmt.Println(RED + "Failed to create the file." + RESET)
	}
	defer f.Close()
	if terr := tmpl.Execute(f, r); terr != nil {
		fmt.Println(RED + "Failed to write to the file." + RESET)
	}
}

// jsonify converts the Result struct to JSON.
func jsonify(r *Result) ([]byte, error) {
	return json.MarshalIndent(r, "", "    ")
}

// csvify converts the Result struct to CSV.
// func csvify(r *Result) ([]byte, error) {
// }

// writeToFile writes text string to the given filename.
func writeToFile(text, filename string) (err error) {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(text)
	return err
}
