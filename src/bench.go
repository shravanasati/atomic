package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/thatisuday/commando"
)

const (
	NAME    = "bench"
	VERSION = "0.2.0"
)

// colors
const (
	RED    = "\033[31m"
	GREEN  = "\033[32m"
	YELLOW = "\033[33m"
	BLUE   = "\033[34m"
	PURPLE = "\033[35m"
	CYAN   = "\033[36m"
	RESET  = "\033[0m"
)

// Result struct which is shown at the end as benchmarking summary and is written to a file.
type Result struct {
	Started    string
	Ended      string
	Command    string
	Iterations int
	Average    string
}

// formats the text in a javascript like syntax.
func format(text string, params map[string]string) string {
	for key, val := range params {
		text = strings.Replace(text, fmt.Sprintf("${%v}", key), val, -1)
	}
	return text
}

func main() {
	fmt.Println(CYAN+NAME, VERSION+RESET)
	deletePreviousInstallation()

	// * basic configuration
	commando.
		SetExecutableName(NAME).
		SetVersion(VERSION).
		SetDescription("bench is a simple CLI tool to make benchmarking easy. \nFor more info visit https://github.com/Shravan-1908/bench.")

	// * root command
	commando.
		Register(nil).
		AddArgument("command", "The command to run for benchmarking.", "").
		AddArgument("iterations", "The number of iterations.", "10").
		AddFlag("export", "Export the benchmarking summary in a json, csv, or text format.", commando.String, "").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			// * initialising some variables
			executable := args["command"].Value
			command := strings.Fields(executable)
			x, e := strconv.Atoi(args["iterations"].Value)

			if e != nil {
				fmt.Println(RED + "Invalid input for iterations value." + RESET)
				return
			}
			var sum time.Duration
			started := time.Now().Format("02-01-2006 15:04:05")

			// * looping for given iterations
			for i := 1; i <= x; i++ {
				fmt.Printf(PURPLE+"***********\nRunning iteration %v\n***********\n"+RESET, i)
				cmd := exec.Command(command[0], command[1:]...)
				_, e := cmd.StdoutPipe()
				if e != nil {
					panic(e)
				}

				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				init := time.Now()
				if e := cmd.Start(); e != nil {
					fmt.Println(RED + "The command couldnt be started!" + RESET)
					fmt.Println(e)
					return
				}
				if e := cmd.Wait(); e != nil {
					fmt.Println(RED + "The command failed to execute!" + RESET)
					fmt.Println(e)
					return
				}
				sum += time.Since(init)
			}
			ended := time.Now().Format("02-01-2006 15:04:05")

			// * result text
			text := format(`
${blue}Benchmarking Summary ${reset}
--------------------

${yellow}Started: ${green}{{ .Started }} ${reset}
${yellow}Ended: ${green}{{ .Ended }} ${reset}
${yellow}Executed Command: ${green}{{ .Command }} ${reset}
${yellow}Total iterations: ${green}{{ .Iterations }} ${reset}
${yellow}Average time taken: ${green}{{ .Average }} ${reset}
`,
				map[string]string{"blue": BLUE, "yellow": YELLOW, "green": GREEN, "reset": RESET})

			// * intialising the template struct
			result := Result{
				Started:    started,
				Ended:      ended,
				Command:    executable,
				Iterations: x,
				Average:    (sum / time.Duration(x)).String(),
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

			// * getting export values
			exportFormat, ierr := flags["export"].GetString()
			if ierr != nil {
				fmt.Println(RED + "Invalid export format." + RESET)
			}

			// * exporting the results
			if exportFormat == "json" {
				jsonText, e := jsonify(&result)
				if e != nil {
					fmt.Println(RED + "Failed to export the results to json." + RESET)
					return
				}
				writeToFile(string(jsonText), "bench-summary.json")

			} else if exportFormat == "csv" {
				// TODO write to csv

			} else if exportFormat == "text" {
				textify(&result)

			} else if exportFormat == "markdown" {
				markdownify(&result)
			}
		})

	// * the update command
	commando.
		Register("up").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			update()
		})

	commando.Parse(nil)
}
