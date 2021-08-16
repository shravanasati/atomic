package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Shravan-1908/bench/internal"
	"github.com/thatisuday/commando"
)

const (
	// NAME is the executable name.
	NAME = "bench"
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
		internal.Log("red", fmt.Sprintf("The command `%s` couldn't be called!", strings.Join(command, " ")))
		internal.Log("white", e.Error())
		return 0, e
	}
	if e := cmd.Wait(); e != nil {
		internal.Log("red", fmt.Sprintf("The command `%s` failed to execute!", strings.Join(command, " ")))
		internal.Log("white", e.Error())
		return 0, e
	}
	duration := time.Since(init)

	return duration, nil
}

func main() {
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
		SetShortDescription("Benchmark a command for given number of iterations.").
		SetDescription("Benchmark a command for given number of iterations.").
		AddArgument("command...", "The command to run for benchmarking.", "").
		AddFlag("iterations,i", "The number of iterations to perform", commando.Int, 10).
		AddFlag("export,e", "Export the benchmarking summary in a json, csv, or text format.", commando.String, "none").
		AddFlag("verbose,V", "Enable verbose output.", commando.Bool, false).
		AddFlag("no-color", "Disable colored output.", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {

			// * getting args and flag values
			if strings.TrimSpace(args["command"].Value) == "" {
				fmt.Println("Error: not enough arguments.")
				return
			}

			command := strings.Split(args["command"].Value, ",")

			iterations, e := flags["iterations"].GetInt()
			if e != nil {
				internal.Log("red", "The number of iterations must be an integer!")
				internal.Log("white", e.Error())
				return
			}

			verbose, e := flags["verbose"].GetBool()
			if e != nil {
				internal.Log("red", "Application error: cannot parse flag values.")
			}
			NO_COLOR, e = (flags["color"].GetBool())
			if e != nil {
				internal.Log("red", "Application error: cannot parse flag values.")
			}
			internal.NO_COLOR = !NO_COLOR

			var sum time.Duration
			started := time.Now().Format("02-01-2006 15:04:05")

			// * looping for given iterations
			for i := 1; i <= iterations; i++ {
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
				Command:    strings.Join(command, " "),
				Iterations: iterations,
				Average:    (sum / time.Duration(iterations)).String(),
			}

			result.Consolify()

			// * getting export values
			exportFormat, ierr := flags["export"].GetString()
			if ierr != nil {
				internal.Log("red", "Application error: cannot parse flag values.")
			}
			result.Export(exportFormat)

		})

	// * the update command
	commando.
		Register("up").
		SetShortDescription("Update bench.").
		SetDescription("Update bench to the latest version.").
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
