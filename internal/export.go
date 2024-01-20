package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"text/template"
	"time"
)

var summaryNoColor = `
Executed Command:   {{ .Command }} 
Total runs:         {{ .Runs }} 
Average time taken: {{ .AverageElapsed }} ± {{ .StandardDeviation }} [User: {{ .AverageUser }}, System: {{ .AverageSystem }}]
Range:              {{ .Min }} ... {{ .Max }}
`

var summaryColor = `
${yellow}Executed Command:   ${green}{{ .Command }} ${reset}
${yellow}Total runs:         ${green}{{ .Runs }} ${reset}
${yellow}Average time taken: ${green}{{ .AverageElapsed }} ± {{ .StandardDeviation }} ${reset} [User: ${blue}{{ .AverageUser }}${reset}, System: ${blue}{{ .AverageSystem }}${reset}]
${yellow}Range:              ${green}{{ .Min }} ... {{ .Max }} ${reset}
`

// Consolify prints the benchmark summary of the Result struct to the console, with color codes.
func (result *PrintableResult) String() string {
	// * result text
	var text string

	if NO_COLOR {
		text = summaryNoColor
	} else {
		text = format(summaryColor,
			map[string]string{"blue": BLUE, "yellow": YELLOW, "green": GREEN, "cyan": CYAN, "reset": RESET})
	}

	var bobTheBuilder strings.Builder
	// * parsing the template
	tmpl, err := template.New("result").Parse(text)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(&bobTheBuilder, result)
	if err != nil {
		panic(err)
	}
	return bobTheBuilder.String()
}

// textify writes the benchmark summary of the Result struct to a text file.
func textify(results []*PrintableResult) {
	// temporarily turn off colors so that [PrintableResult.String] used non-colored summary
	origVal := NO_COLOR
	NO_COLOR = true
	defer func() {
		NO_COLOR = origVal
	}()

	f, ferr := os.Create("atomic-summary.txt")
	if ferr != nil {
		Log("red", "Failed to create the file.")
	}
	defer f.Close()

	for _, r := range results {
		f.WriteString(r.String() + "\n")
	}

	absPath, err := filepath.Abs("atomic-summary.txt")
	if err != nil {
		Log("red", "unable to get the absolute path for text file: "+err.Error())
		return
	} else {
		Log("green", "Successfully wrote benchmark summary to `"+absPath+"`.")
	}
}

func markdownify(results []*SpeedResult) {
	text := `
# atomic-summary

| Command | Runs | Average | User | System | Min | Max | Relative |
| ------- | ---- | ------- | ---- | ------ | --- | --- | -------- |
`
	f, ferr := os.Create("atomic-summary.md")
	if ferr != nil {
		Log("red", "Failed to create the file.")
	}
	defer f.Close()
	
	for _, r = range results {
		text += fmt.Sprintf("")
	}

	absPath, err := filepath.Abs("atomic-summary.md")
	if err != nil {
		Log("red", "unable to get the absolute path for markdown file: "+err.Error())
	} else {
		Log("green", "Successfully wrote benchmark summary to `"+absPath+"`.")
	}
}

// jsonify converts the Result struct to JSON.
func jsonify(r *PrintableResult) ([]byte, error) {
	return json.MarshalIndent(r, "", "    ")
}

// csvify converts the Result struct to CSV.
func csvify(r *PrintableResult) {
	text := `
Executed Command,Total runs,Average time taken,Range
{{.Command}}, {{.Runs}}, {{.Average}} ± {{ .StandardDeviation }}, {{.Min}} ... {{.Max}}
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
func (result *PrintableResult) Export(exportFormats string) {
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

func VerifyExportFormats(formats string) ([]string, error) {
	validFormats := []string{"csv", "markdown", "txt", "json"}
	formatList := strings.Split(strings.ToLower(formats), ",")
	for _, f := range formatList {
		if !slices.Contains(validFormats, f) {
			return nil, fmt.Errorf("invalid export format: %s", f)
		}
	}
	return formatList, nil
}

func convertToTimeUnit(given float64, unit time.Duration) float64 {
	// first get duration from microseconds
	duration := DurationFromNumber(given, time.Microsecond)
	switch unit {
	case time.Nanosecond:
		return float64(duration.Nanoseconds())
	case time.Microsecond:
		return float64(duration.Microseconds())
	case time.Millisecond:
		return float64(duration.Milliseconds())
	case time.Second:
		return duration.Seconds()
	case time.Minute:
		return duration.Minutes()
	case time.Hour:
		return duration.Hours()
	default:
		panic("convertToTimeUnit: unknown time unit: " + unit.String())
	}
}

func Export(formats []string, results []*SpeedResult, timeUnit time.Duration) {
	// first convert all speed results to the given time unit
	// except for microseconds, because that's what used internally
	if timeUnit != time.Microsecond {
		var wg sync.WaitGroup
		for _, sr := range results {
			wg.Add(1)
			go func(sr *SpeedResult) {
				sr.AverageElapsed = convertToTimeUnit(sr.AverageElapsed, timeUnit)
				sr.AverageUser = convertToTimeUnit(sr.AverageUser, timeUnit)
				sr.AverageSystem = convertToTimeUnit(sr.AverageSystem, timeUnit)
				sr.StandardDeviation = convertToTimeUnit(sr.StandardDeviation, timeUnit)
				sr.Max = convertToTimeUnit(sr.Max, timeUnit)
				sr.Min = convertToTimeUnit(sr.Min, timeUnit)
				for i, t := range sr.Times {
					sr.Times[i] = convertToTimeUnit(t, timeUnit)
				}
				wg.Done()
			}(sr)
		}
		wg.Wait()
	}

	for _, format := range formats {
		switch format {
		case "json":
			jsonify()
		case "csv":
			csvify()
		case "markdown":
			markdownify()
		case "txt":
			printables := MapFunc[[]*SpeedResult, []*PrintableResult](func(r *SpeedResult) *PrintableResult { return NewPrintableResult().FromSpeedResult(*r) }, results)
			textify(printables)
		}
	}
}
