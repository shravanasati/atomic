package main

import (
	"encoding/json"
	"os"
	"text/template"
)

// consolify prints the benchmark summary of the Result struct to the console, with color codes.
func consolify(result *Result) {
	// * result text
	text := format(`
${blue}Benchmarking Summary ${reset}
${blue}-------------------- ${reset}

${yellow}Started: ${green}{{ .Started }} ${reset}
${yellow}Ended: ${green}{{ .Ended }} ${reset}
${yellow}Executed Command: ${green}{{ .Command }} ${reset}
${yellow}Total iterations: ${green}{{ .Iterations }} ${reset}
${yellow}Average time taken: ${green}{{ .Average }} ${reset}
`,
		map[string]string{"blue": BLUE, "yellow": YELLOW, "green": GREEN, "reset": RESET})

	// * parsing the template
	tmpl, err := template.New("result").Parse(text)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, result)
	if err != nil {
		panic(err)
	}
}

// textify writes the benchmark summary of the Result struct to a text file.
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
		log("red", "Failed to create the file.")
	}
	defer f.Close()
	if terr := tmpl.Execute(f, r); terr != nil {
		log("red", "Failed to write to the file.")
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
		log("red", "Failed to create the file.")
	}
	defer f.Close()
	if terr := tmpl.Execute(f, r); terr != nil {
		log("red", "Failed to write to the file.")
	}
}

// jsonify converts the Result struct to JSON.
func jsonify(r *Result) ([]byte, error) {
	return json.MarshalIndent(r, "", "    ")
}

// csvify converts the Result struct to CSV.
// func csvify(r *Result) ([]byte, error) {
// }

func export(exportFormat string, result *Result) {
	// * exporting the results
	if exportFormat == "json" {
		jsonText, e := jsonify(result)
		if e != nil {
			log("red", "Failed to export the results to json.")
			return
		}
		writeToFile(string(jsonText), "bench-summary.json")

	} else if exportFormat == "csv" {
		// TODO write to csv

	} else if exportFormat == "text" {
		textify(result)

	} else if exportFormat == "markdown" {
		markdownify(result)

	} else if exportFormat != "none" {
		log("red", "Invalid export format.")
	}
}
