package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
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
func textify(results []*PrintableResult, filename string) {
	// temporarily turn off colors so that [PrintableResult.String] used non-colored summary
	origVal := NO_COLOR
	NO_COLOR = true
	var ferr error
	var f *os.File
	defer func() {
		NO_COLOR = origVal
		absPath, err := filepath.Abs(filename)
		if err != nil {
			Log("red", "unable to get the absolute path for text file: "+err.Error())
			return
		} else {
			if ferr == nil {
				Log("green", "Successfully wrote benchmark summary to `"+absPath+"`.")
			}
		}
	}()

	f, ferr = os.Create(filename)
	if ferr != nil {
		Log("red", "Failed to create the file.")
	}
	defer f.Close()

	for _, r := range results {
		f.WriteString(r.String() + "\n")
	}

}

func markdownify(results []*SpeedResult, filename, timeUnit string) {
	text := `
# atomic-summary

| Command | Runs | Average [${timeUnit}] | User [${timeUnit}] | System [${timeUnit}] | Min [${timeUnit}] | Max [${timeUnit}] | Relative |
| ------- | ---- | ------- | ---- | ------ | --- | --- | -------- |
`
	text = format(text, map[string]string{"timeUnit": timeUnit})
	for _, r := range results {
		text += fmt.Sprintf("`%s` | %d | %.2f ± %.2f | %.2f | %.2f | %.2f | %.2f | %.2f ± %.2f \n", r.Command, len(r.Times), r.AverageElapsed, r.StandardDeviation, r.AverageUser, r.AverageSystem, r.Min, r.Max, r.RelativeMean, r.RelativeStddev)
	}

	err := writeToFile(text, filename)
	if err != nil {
		Log("red", "error in writing to file: "+filename+"\nerror: "+err.Error())
		return
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		Log("red", "unable to get the absolute path for markdown file: "+err.Error())
	} else {
		Log("green", "Successfully wrote benchmark summary to `"+absPath+"`.")
	}
}

// jsonify converts the Result struct to JSON.
func jsonify(data any) ([]byte, error) {
	return json.MarshalIndent(data, "", "    ")
}

// csvify converts the Result struct to CSV.
func csvify(results []*SpeedResult, filename string) {
	text := "command,runs,average_elapsed,stddev,average_user,average_system,min,max,relative_average,relative_stddev\n"

	for _, r := range results {
		text += fmt.Sprintf("%s,%d,%f,%f,%f,%f,%f,%f,%f,%f\n", r.Command, len(r.Times), r.AverageElapsed, r.StandardDeviation, r.AverageUser, r.AverageSystem, r.Min, r.Max, r.RelativeMean, r.RelativeStddev)
	}

	err := writeToFile(text, filename)
	if err != nil {
		Log("red", "error in writing to file: "+filename+"\nerror: "+err.Error())
		return
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		Log("red", "unable to get the absolute path for csv file: "+err.Error())
	} else {
		Log("green", "Successfully wrote benchmark summary to `"+absPath+"`.")
	}
}

func VerifyExportFormats(formats string) ([]string, error) {
	validFormats := []string{"csv", "markdown", "md", "txt", "json"}
	formatList := strings.Split(strings.ToLower(formats), ",")
	for _, f := range formatList {
		if !slices.Contains(validFormats, f) {
			return nil, fmt.Errorf("invalid export format: %s", f)
		}
	}
	return formatList, nil
}

func Export(formats []string, filename string, results []*SpeedResult, timeUnit time.Duration) {
	for _, format := range formats {
		switch format {
		case "json":
			jsonMap := map[string]any{"time_unit": timeUnit.String()[1:], "results": results}
			jsonData, err := jsonify(jsonMap)
			if err != nil {
				panic("unable to convert to json: " + err.Error())
			}
			filename := addExtension(filename, "json")
			err = writeToFile(string(jsonData), filename)
			if err != nil {
				Log("red", "an unknown error occured in writing to file: "+err.Error())
				return
			}
			absPath, err := filepath.Abs(filename)
			if err != nil {
				Log("red", "an unknown error occured in getting the full path to the file: "+filename+"\nerror: "+err.Error())
				return
			}
			Log("green", fmt.Sprintf("Successfully wrote benchmark summary to `%s`.", absPath))

		case "csv":
			filename := addExtension(filename, "csv")
			csvify(results, filename)

		case "markdown", "md":
			filename := addExtension(filename, "md")
			markdownify(results, filename, timeUnit.String()[1:])

		case "txt":
			printables := MapFunc[[]*SpeedResult, []*PrintableResult](func(r *SpeedResult) *PrintableResult { return NewPrintableResult().FromSpeedResult(*r) }, results)
			filename := addExtension(filename, "txt")
			textify(printables, filename)
		}
	}
}
