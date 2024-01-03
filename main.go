package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Shravan-1908/bench/internal"
	"github.com/google/shlex"
	"github.com/thatisuday/commando"
)

const (
	// NAME is the executable name.
	NAME = "bench"
	// VERSION is the executable version.
	VERSION = "v0.2.0"
)

// NO_COLOR is a global variable that is used to determine whether or not to enable color output.
var NO_COLOR bool = false

// todo add range too

func run(command []string, verbose bool, ignoreError bool) (time.Duration, error) {
	// todo add shell support
	// todo measure shell spawn time too and deduct it from runs
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
	if e := cmd.Wait(); e != nil && !ignoreError {
		internal.Log("red", fmt.Sprintf("The command `%s` failed to execute!", strings.Join(command, " ")))
		internal.Log("white", e.Error())
		return 0, e
	}
	duration := time.Since(init)

	return duration, nil
}

func main() {
	internal.Log("white", fmt.Sprintf("%v %v\n", NAME, VERSION))

	updateCh := make(chan string, 1)
	go internal.CheckForUpdates(VERSION, &updateCh)

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
		AddFlag("warmup,w", "The number of warmup runs to perform.", commando.Int, 0).
		AddFlag("ignore-error,I", "Ignore if the process returns a non-zero return code", commando.Bool, false).
		AddFlag("export,e", "Comma separated list of benchmark export formats, including json, text, csv and markdown.", commando.String, "none").
		AddFlag("verbose,V", "Enable verbose output.", commando.Bool, false).
		AddFlag("no-color", "Disable colored output.", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {

			// * getting args and flag values
			if strings.TrimSpace(args["command"].Value) == "" {
				fmt.Println("Error: not enough arguments.")
				return
			}

			command, err := shlex.Split(args["command"].Value)
			if err != nil {
				internal.Log("red", "unable to parse the given command: "+args["command"].Value)
				internal.Log("white", err.Error())
				return
			}

			iterations, e := flags["iterations"].GetInt()
			if e != nil {
				internal.Log("red", "The number of iterations must be an integer!")
				internal.Log("white", e.Error())
				return
			}

			warmupRuns, e := flags["warmup"].GetInt()
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

			ignoreError, er := flags["ignore-error"].GetBool()
			if er != nil {
				internal.Log("red", "Application error: cannot parse flag values.")
			}

			// warmup runs
			for i := 0; i < warmupRuns; i++ {
				// todo replace running logs with a progress bar in non-verbose mode
				internal.Log("purple", fmt.Sprintf("***********\nRunning warmup %d\n***********", i+1))
				_, e := run(command, verbose, ignoreError)
				if e != nil {
					return
				}
			}

			// actual runs
			var runs []int64
			started := time.Now().Format("02-01-2006 15:04:05")
			// * looping for given iterations
			for i := 1; i <= iterations; i++ {
				internal.Log("purple", fmt.Sprintf("***********\nRunning iteration %d\n***********", i))

				dur, e := run(command, verbose, ignoreError)
				if e != nil {
					return
				}
				runs = append(runs, (dur.Microseconds()))
			}

			ended := time.Now().Format("02-01-2006 15:04:05")

			// * intialising the template struct
			avg, stddev := internal.ComputeAverageAndStandardDeviation(runs)
			avgDuration := internal.DurationFromNumber(avg, time.Microsecond)
			stddevDuration := internal.DurationFromNumber(stddev, time.Microsecond)
			result := internal.Result{
				Started:           started,
				Ended:             ended,
				Command:           strings.Join(command, " "),
				Iterations:        iterations,
				Average:           avgDuration.String(),
				StandardDeviation: stddevDuration.String(),
			}

			result.Consolify()

			// * getting export values
			exportFormat, ierr := flags["export"].GetString()
			if ierr != nil {
				internal.Log("red", "Application error: cannot parse flag values.")
				return
			}
			result.Export(exportFormat)

		})

	commando.Parse(nil)
	fmt.Println(<-updateCh)
}
