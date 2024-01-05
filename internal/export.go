package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Result struct which is shown at the end as benchmarking summary and is written to a file.
type Result struct {
	Started           string
	Ended             string
	Command           string
	Iterations        int
	Average           string
	StandardDeviation string
	Min               string
	Max               string
}

// todo in all exports, include individual run details

var summaryNoColor = `
Benchmarking Summary
--------------------

Started:            {{ .Started }} 
Ended:              {{ .Ended }} 
Executed Command:   {{ .Command }} 
Total iterations:   {{ .Iterations }} 
Average time taken: {{ .Average }} ± {{ .StandardDeviation }}
Range:              {{ .Min }} ... {{ .Max }}
`

var summaryColor = `
${blue}Benchmarking Summary ${reset}
${blue}-------------------- ${reset}

${yellow}Started:            ${green}{{ .Started }} ${reset}
${yellow}Ended:              ${green}{{ .Ended }} ${reset}
${yellow}Executed Command:   ${green}{{ .Command }} ${reset}
${yellow}Total iterations:   ${green}{{ .Iterations }} ${reset}
${yellow}Average time taken: ${green}{{ .Average }} ± {{ .StandardDeviation }} ${reset}
${yellow}Range:              ${green}{{ .Min }} ... {{ .Max }} ${reset}
`

// Consolify prints the benchmark summary of the Result struct to the console, with color codes.
func (result *Result) Consolify() {
	// * result text
	text := format(summaryColor,
		map[string]string{"blue": CYAN, "yellow": YELLOW, "green": GREEN, "reset": RESET})

	if NO_COLOR {
		text = summaryNoColor
	}

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
	text := summaryNoColor

	tmpl, err := template.New("summary").Parse(text)
	if err != nil {
		panic(err)
	}

	f, ferr := os.Create("atomic-summary.txt")
	if ferr != nil {
		Log("red", "Failed to create the file.")
	}
	defer f.Close()
	if terr := tmpl.Execute(f, r); terr != nil {
		Log("red", "Failed to write to the file.")
	} else {
		absPath, err := filepath.Abs("atomic-summary.txt")
		if err != nil {
			Log("red", "unable to get the absolute path for text file: "+err.Error())
		} else {
			Log("green", "Successfully wrote benchmark summary to `"+absPath+"`.")
		}
	}

}

func markdownify(r *Result) {
	text := `
# atomic-summary

| Fields             | Values          					       |
| -----------        | -----------     						   |
| Started            | {{.Started}}    						   |
| Ended              | {{.Ended}}      						   |
| Executed Command   | {{.Command}}   						   |
| Total iterations   | {{.Iterations}} 						   |
| Average time taken | {{.Average}} ± {{ .StandardDeviation }} |
| Range				 | {{.Min}} ... {{ .Max }}   			   |
`
	tmpl, err := template.New("summary").Parse(text)
	if err != nil {
		panic(err)
	}

	f, ferr := os.Create("atomic-summary.md")
	if ferr != nil {
		Log("red", "Failed to create the file.")
	}
	defer f.Close()
	if terr := tmpl.Execute(f, r); terr != nil {
		Log("red", "Failed to write to the file.")
	} else {
		absPath, err := filepath.Abs("atomic-summary.md")
		if err != nil {
			Log("red", "unable to get the absolute path for markdown file: "+err.Error())
		} else {
			Log("green", "Successfully wrote benchmark summary to `"+absPath+"`.")
		}
	}

}

// jsonify converts the Result struct to JSON.
func jsonify(r *Result) ([]byte, error) {
	return json.MarshalIndent(r, "", "    ")
}

// csvify converts the Result struct to CSV.
func csvify(r *Result) {
	text := `
Started,Ended,Executed Command,Total iterations,Average time taken,Range
{{.Started}}, {{.Ended}}, {{.Command}}, {{.Iterations}}, {{.Average}} ± {{ .StandardDeviation }}, {{.Min}} ... {{.Max}}
`
	tmpl, err := template.New("summary").Parse(text)
	if err != nil {
		panic(err)
	}

	f, ferr := os.Create("atomic-summary.csv")
	if ferr != nil {
		Log("red", "Failed to create the file.")
	}
	defer f.Close()
	if terr := tmpl.Execute(f, r); terr != nil {
		Log("red", "Failed to write to the file.")
	} else {
		absPath, err := filepath.Abs("atomic-summary.csv")
		if err != nil {
			Log("red", "unable to get the absolute path for csv file: "+err.Error())
		} else {
			Log("green", "Successfully wrote benchmark summary to `"+absPath+"`.")
		}
	}
}

// Export writes the benchmark summary of the Result struct to a file in the specified format.
func (result *Result) Export(exportFormats string) {
	// * exporting the results

	for _, exportFormat := range strings.Split(exportFormats, ",") {
		exportFormat = strings.ToLower(exportFormat)
		if exportFormat == "json" {
			jsonText, e := jsonify(result)
			if e != nil {
				Log("red", "Failed to export the results to json.")
				return
			}
			e = writeToFile(string(jsonText), "atomic-summary.json")
			if e == nil {
				absPath, err := filepath.Abs("atomic-summary.json")
				if err != nil {
					Log("red", "unable to get the absolute path for json file: "+err.Error())
				} else {
					Log("green", "Successfully wrote benchmark summary to `"+absPath+"`.")
				}
			} else {
				Log("red", "Unable to write to file ./atomic-summary.json: "+e.Error())
			}

		} else if exportFormat == "csv" {
			csvify(result)

		} else if exportFormat == "text" {
			textify(result)

		} else if exportFormat == "markdown" {
			markdownify(result)

		} else if exportFormat != "none" {
			Log("red", "Invalid export format: "+exportFormat+".")
		}

	}

}
