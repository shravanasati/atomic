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
	VERSION = "0.1.0"
)

// Result struct which is shown at the end as benchmarking summary and is written to a file.
type Result struct {
	Started    string
	Ended      string
	Command    string
	Iterations int
	Average    time.Duration
}

func main() {
	fmt.Println(NAME, VERSION)

	// * basic configuration
	commando.
		SetExecutableName(NAME).
		SetVersion(VERSION).
		SetDescription("bench is a simple CLI tool to make benchmarking easy.")

	// * root command
	commando.
		Register(nil).
		AddArgument("command", "The command to run for benchmarking.", "").
		AddArgument("iterations", "The number of iterations.", "10").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			// * initialising some variables
			executable := args["command"].Value
			command := strings.Fields(executable)
			x, e := strconv.Atoi(args["iterations"].Value)
			if e != nil {
				fmt.Println("Invalid input for iterations value.")
				return
			}
			var sum time.Duration

			started := time.Now().Format("02-01-2006 15:04:05")
			for i := 1; i <= x; i++ {
				fmt.Printf("***********\nRunning iteration %v\n***********\n", i)
				cmd := exec.Command(command[0], command[1:]...)
				_, e := cmd.StdoutPipe()
				if e != nil {
					panic(e)
				}

				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				init := time.Now()
				if e := cmd.Start(); e != nil {
					fmt.Println("The command couldnt be started!")
					fmt.Println(e)
					return
				}
				if e := cmd.Wait(); e != nil {
					fmt.Println("The command failed to execute!")
					fmt.Println(e)
					return
				}
				sum += time.Since(init)
			}
			ended := time.Now().Format("02-01-2006 15:04:05")

			// * result text
			text := `
Benchmarking Summary
--------------------

Started: {{ .Started }}
Ended: {{ .Ended }}
Executed Command: {{ .Command }}
Total iterations: {{ .Iterations }}
Average time taken: {{ .Average }}
`
			// * intialising the template struct
			result := Result{
				Started: started,
				Ended: ended,
				Command: executable,
				Iterations: x,
				Average: sum / time.Duration(x),
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
		})

	commando.Parse(nil)
}
