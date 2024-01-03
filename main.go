package main

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/Shravan-1908/bench/internal"
	"github.com/google/shlex"
	"github.com/schollz/progressbar/v3"
	"github.com/thatisuday/commando"
)

const (
	// NAME is the executable name.
	NAME = "bench"
	// VERSION is the executable version.
	VERSION = "v0.4.0"
)

// NO_COLOR is a global variable that is used to determine whether or not to enable color output.
var NO_COLOR bool = false

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

func benchmark(iterations int, command []string, verbose bool, ignoreError bool, warmup bool) ([]int64, bool) {
	// actual runs, each entry stored in microseconds
	var runs []int64

	// * looping for given iterations
	if verbose {
		word := ""
		if warmup {
			word = "warmup"
		} else {
			word = "iteration"
		}
		for i := 1; i <= iterations; i++ {
			internal.Log("purple", fmt.Sprintf("***********\nRunning "+word+" %d\n***********", i))

			dur, e := run(command, verbose, ignoreError)
			if e != nil {
				return nil, true
			}
			runs = append(runs, (dur.Microseconds()))
		}
	} else {
		description := ""
		if warmup {
			description = "Performing warmup runs"
		} else {
			description = "Performing benchmark runs"
		}
		pbarOptions := []progressbar.Option{
			progressbar.OptionClearOnFinish(),
			progressbar.OptionSetDescription("[magenta]" + description + "[reset]"),
			progressbar.OptionSetPredictTime(true),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "[green]=[reset]",
				SaucerHead:    "[green]>[reset]",
				SaucerPadding: " ",
				BarStart:      "|",
				BarEnd:        "|",
			}),
		}
		if NO_COLOR {
			pbarOptions = append(pbarOptions, progressbar.OptionEnableColorCodes(true))
		}
		bar := progressbar.NewOptions(
			iterations, pbarOptions...,
		)
		for i := 1; i <= iterations; i++ {
			bar.Add(1)
			dur, e := run(command, verbose, ignoreError)
			if e != nil {
				bar.Finish()
				return nil, true
			}
			if !warmup {
				bar.Describe(fmt.Sprintf("Current estimate: %s", dur.String()))
			}
			runs = append(runs, (dur.Microseconds()))
		}
	}
	return runs, false
}

func main() {
	internal.Log("white", fmt.Sprintf("%v %v\n", NAME, VERSION))

	updateCh := make(chan string, 1)
	go internal.CheckForUpdates(VERSION, &updateCh)
	defer fmt.Println(<-updateCh)

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

			if iterations <= 0 {
				return
			}

			// no need for runs in warmups
			_, shouldReturn := benchmark(warmupRuns, command, verbose, ignoreError, true)
			if shouldReturn {
				return
			}

			started := time.Now().Format("02-01-2006 15:04:05")
			runs, shouldReturn := benchmark(iterations, command, verbose, ignoreError, false)
			if shouldReturn {
				return
			}
			ended := time.Now().Format("02-01-2006 15:04:05")

			// * intialising the template struct
			avg, stddev := internal.ComputeAverageAndStandardDeviation(runs)
			avgDuration := internal.DurationFromNumber(avg, time.Microsecond)
			stddevDuration := internal.DurationFromNumber(stddev, time.Microsecond)
			max_ := slices.Max(runs)
			min_ := slices.Min(runs)
			maxDuration := internal.DurationFromNumber(max_, time.Microsecond)
			minDuration := internal.DurationFromNumber(min_, time.Microsecond)
			result := internal.Result{
				Started:           started,
				Ended:             ended,
				Command:           strings.Join(command, " "),
				Iterations:        iterations,
				Average:           avgDuration.String(),
				StandardDeviation: stddevDuration.String(),
				Max:               maxDuration.String(),
				Min:               minDuration.String(),
			}

			result.Consolify()

			// * getting export values
			exportFormat, ierr := flags["export"].GetString()
			if ierr != nil {
				internal.Log("red", "Application error: cannot parse flag values.")
				return
			}
			result.Export(exportFormat)

			outliersDetected := internal.TestOutliers(runs)
			if outliersDetected {
				internal.Log("yellow", "\nWarning: Statistical outliers were detected. Consider re-running this benchmark on a quiet system, devoid of any interferences from other programs.")
				if warmupRuns == 0 {
					internal.Log("yellow", "It might help to use the --warmup flag.")
				}
			}

		})

	commando.Parse(nil)
}
