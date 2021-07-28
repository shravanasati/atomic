package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/thatisuday/commando"
)

const (
	NAME    = "bench"
	VERSION = "0.3.0"
)

var NO_COLOR bool = false

// Result struct which is shown at the end as benchmarking summary and is written to a file.
type Result struct {
	Started    string
	Ended      string
	Command    string
	Iterations int
	Average    string
}

func main() {
	log("white", fmt.Sprintf("%v %v\n", NAME, VERSION))

	go deletePreviousInstallation()

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
		AddFlag("export,e", "Export the benchmarking summary in a json, csv, or text format.", commando.String, "none").
		AddFlag("verbose,V", "Enable verbose output.", commando.Bool, false).
		AddFlag("no-color", "Disable colored output.", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {

			// * initialising some variables
			executable := args["command"].Value
			command := strings.Fields(executable)
			verbose, e := flags["verbose"].GetBool()
			if e != nil {
				log("red", "Application error: cannot parse flag values.")
			}
			NO_COLOR, e = (flags["color"].GetBool())
			if e != nil {
				log("red", "Application error: cannot parse flag values.")
			}
			NO_COLOR = !NO_COLOR

			x, e := strconv.Atoi(args["iterations"].Value)
			if e != nil {
				log("red", "Invalid input for iterations value.")
				return
			}

			var sum time.Duration
			started := time.Now().Format("02-01-2006 15:04:05")

			// * looping for given iterations
			for i := 1; i <= x; i++ {
				log("purple", fmt.Sprintf("***********\nRunning iteration %d\n***********", i))

				cmd := exec.Command(command[0], command[1:]...)
				_, e := cmd.StdoutPipe()
				if e != nil {
					panic(e)
				}

				if verbose {
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
				}

				init := time.Now()
				if e := cmd.Start(); e != nil {
					log("red", "The command couldnt be called!")
					log("white", e.Error())
					return
				}
				if e := cmd.Wait(); e != nil {
					log("red", "The command failed to execute!")
					log("white", e.Error())
					return
				}
				sum += time.Since(init)
			}

			ended := time.Now().Format("02-01-2006 15:04:05")

			// * intialising the template struct
			result := Result{
				Started:    started,
				Ended:      ended,
				Command:    executable,
				Iterations: x,
				Average:    (sum / time.Duration(x)).String(),
			}

			consolify(&result)

			// * getting export values
			exportFormat, ierr := flags["export"].GetString()
			if ierr != nil {
				log("red", "Invalid export format.")
			}
			export(exportFormat, &result)

		})

	// * the update command
	commando.
		Register("up").
		AddFlag("no-color", "Disable colored output.", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			_no_color, e := flags["no-color"].GetBool()
			if e != nil {
				log("red", "Application error: cannot parse flag values.")
			}
			NO_COLOR = _no_color
			update()
		})

	commando.Parse(nil)
}
