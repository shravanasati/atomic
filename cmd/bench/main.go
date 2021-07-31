package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Shravan-1908/bench/internal"
	"github.com/thatisuday/commando"
)

const (
	// NAME is the executable name.
	NAME    = "bench"
	// VERSION is the executable version.
	VERSION = "0.3.0"
)

// NO_COLOR is a global variable that is used to determine whether or not to enable color output.
var NO_COLOR bool = false

func run(command []string, verbose bool) (time.Duration, error) {
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
		internal.Log("red", "The command couldnt be called!")
		internal.Log("white", e.Error())
		return 0, e
	}
	if e := cmd.Wait(); e != nil {
		internal.Log("red", "The command failed to execute!")
		internal.Log("white", e.Error())
		return 0, e
	}
	duration := time.Since(init)

	return duration, nil
}

func main() {
	internal.NO_COLOR = NO_COLOR
	internal.Log("white", fmt.Sprintf("%v %v\n", NAME, VERSION))

	go internal.DeletePreviousInstallation()

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
				internal.Log("red", "Application error: cannot parse flag values.")
			}
			NO_COLOR, e = (flags["color"].GetBool())
			if e != nil {
				internal.Log("red", "Application error: cannot parse flag values.")
			}
			internal.NO_COLOR = !NO_COLOR

			x, e := strconv.Atoi(args["iterations"].Value)
			if e != nil {
				internal.Log("red", "Invalid input for iterations value.")
				return
			}

			var sum time.Duration
			started := time.Now().Format("02-01-2006 15:04:05")

			// * looping for given iterations
			for i := 1; i <= x; i++ {
				internal.Log("purple", fmt.Sprintf("***********\nRunning iteration %d\n***********", i))

				dur, e := run(command, verbose)
				if e != nil {
					return
				}
				sum += dur
			}

			ended := time.Now().Format("02-01-2006 15:04:05")

			// * intialising the template struct
			result := internal.Result{
				Started:    started,
				Ended:      ended,
				Command:    executable,
				Iterations: x,
				Average:    (sum / time.Duration(x)).String(),
			}

			result.Consolify()

			// * getting export values
			exportFormat, ierr := flags["export"].GetString()
			if ierr != nil {
				internal.Log("red", "Invalid export format.")
			}
			result.Export(exportFormat)

		})

	// * the update command
	commando.
		Register("up").
		AddFlag("no-color", "Disable colored output.", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			_noColor, e := flags["color"].GetBool()
			if e != nil {
				internal.Log("red", "Application error: cannot parse flag values.")
			}
			internal.NO_COLOR = !_noColor
			internal.Update()
		})

	commando.Parse(nil)
}
