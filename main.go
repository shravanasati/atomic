package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/Shravan-1908/atomic/internal"
	"github.com/google/shlex"
	"github.com/schollz/progressbar/v3"
	"github.com/thatisuday/commando"
)

const (
	// NAME is the executable name.
	NAME = "atomic"
	// VERSION is the executable version.
	VERSION = "v0.4.0"
)

type benchmarkMode int

const (
	shellMode  benchmarkMode = 0
	warmupMode benchmarkMode = 1
	mainMode   benchmarkMode = 2
)

type BenchmarkOptions struct {
	command           []string
	iterations        int
	verbose           bool
	ignoreError       bool
	executePrepareCmd bool
	prepareCmd        []string
	mode              benchmarkMode
}

// NO_COLOR is a global variable that is used to determine whether or not to enable color output.
var NO_COLOR bool = false

// Tells if the current system is windows.
var WINDOWS = runtime.GOOS == "windows"

// returns true if powershell is available on the system
func _testPowershell() bool {
	cmd := exec.Command("powershell", "-h")
	err := cmd.Start()
	if err != nil {
		return false
	}
	err = cmd.Wait()
	return err == nil
}

// this value is used in cases of flags default values
// because empty default values in commando marks the flag as required
const dummyDefault = "~!_default_!~"

// returns the shell path
func getShell() (string, error) {
	if WINDOWS {
		// windows
		// first test if powershell is present and use it if present
		if _testPowershell() {
			return "powershell", nil
		} else {
			// fall back to cmd.exe if pwsh absent

			// lookup the comspec env variable -> it contains the path to cmd.exe
			comspec, ok := os.LookupEnv("ComSpec")
			if ok {
				return comspec, nil
			} else {
				// otherwise find cmd.exe in $SystemRoot/System32
				systemRoot, ok := os.LookupEnv("SystemRoot")
				if !ok {
					return "", fmt.Errorf("buildCommand with useShell=true on windows: neither ComSpec nor SystemRoot is set")
				}
				comspec = filepath.Join(systemRoot, "System32", "cmd.exe")
				return comspec, nil
			}
		}
	} else {
		// posix
		return "/bin/sh", nil
	}
}

// todo write tests

// builds the given command as per the given params.
// if useShell is true, adds a shell in front of the command.
// returns the built command and a boolean value indicating whether
// the application should quit in case it is unable to build a command.
func buildCommand(command string, useShell bool, shellPath string) ([]string, error) {
	var builtCommand []string
	var err error
	if useShell {
		// the flag that enables execution of command from a string
		// eg. -Command or -c on pwsh, /c on cmd.exe, -c on any other shell
		var commandSwitch string
		if strings.Contains(shellPath, "cmd.exe") || strings.Contains(shellPath, "cmd"){
			commandSwitch = "/c"
		} else {
			commandSwitch = "-c"
		}
		builtCommand, err = shlex.Split(fmt.Sprintf("%s %s \"%s\"", shellPath, commandSwitch, command))
	} else {
		builtCommand, err = shlex.Split(command)
	}
	return builtCommand, err
}

type failedProcessError struct {
	command []string
	err     error
	mode    string
}

func (fpe *failedProcessError) Error() string {
	return fmt.Sprintf("The command `%s` failed in the process of %s!\nerror: %s", strings.Join(fpe.command, " "), fpe.mode, fpe.err.Error())
}

// runs the built command using os/exec and returns the duration the command lasted
func run(command []string, verbose bool, ignoreError bool) (time.Duration, error) {
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

	if e := cmd.Start(); e != nil {
		return 0, &failedProcessError{command: command, err: e, mode: "starting"}
	}

	init := time.Now()
	e = cmd.Wait()
	duration := time.Since(init)

	if e != nil && !ignoreError {
		return 0, &failedProcessError{command: command, err: e, mode: "execution"}
	}

	return duration, nil
}

// todo automatically determine the number of runs
func benchmark(opts BenchmarkOptions) ([]int64, bool) {
	// actual runs, each entry stored in microseconds
	var runs []int64
	wordMap := map[benchmarkMode]string{
		shellMode:  "shell",
		warmupMode: "warmup",
		mainMode:   "iteration",
	}
	descriptionMap := map[benchmarkMode]string{
		shellMode:  "Measuring shell spawn time",
		warmupMode: "Performing warmup runs",
		mainMode:   "Performing benchmark runs",
	}
	var processErr *failedProcessError

	// * looping for given iterations
	if opts.verbose {
		word, ok := wordMap[opts.mode]
		if !ok {
			// used internally, ok to panic
			panic(fmt.Sprintf("invalid mode passed to benchmark: %v", opts.mode))
		}
		for i := 1; i <= opts.iterations; i++ {
			internal.Log("purple", fmt.Sprintf("***********\nRunning "+word+" %d\n***********", i))

			// dont ignore errors in prepare command execution, dont output it either
			if opts.executePrepareCmd {
				_, e := run(opts.prepareCmd, false, false)
				if errors.As(e, &processErr) {
					internal.Log("red", processErr.Error())
					return nil, true
				}
			}
			dur, e := run(opts.command, opts.verbose, opts.ignoreError)
			if errors.As(e, &processErr) {
				internal.Log("red", processErr.Error())
				return nil, true
			}
			runs = append(runs, (dur.Microseconds()))
		}
	} else {
		description, ok := descriptionMap[opts.mode]
		if !ok {
			panic(fmt.Sprintf("invalid mode passed to benchmark: %v", opts.mode))
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
			opts.iterations, pbarOptions...,
		)
		for i := 1; i <= opts.iterations; i++ {
			// run the prepareCmd first
			// dont ignore errors in prepare command execution, dont output it either
			if opts.executePrepareCmd {
				_, e := run(opts.prepareCmd, false, false)
				if errors.As(e, &processErr) {
					internal.Log("red", processErr.Error())
					return nil, true
				}
			}
			bar.Add(1)
			dur, e := run(opts.command, opts.verbose, opts.ignoreError)
			if errors.As(e, &processErr) {
				bar.Clear()
				internal.Log("red", processErr.Error())
				return nil, true
			}
			runs = append(runs, (dur.Microseconds()))
			if opts.mode == mainMode {
				bar.Describe(
					fmt.Sprintf("[magenta]Current estimate:[reset] [green]%s[reset]",
						internal.DurationFromNumber(
							internal.CalculateAverage(runs), time.Microsecond).String(),
					),
				)
			}
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
		SetDescription("atomic is a simple CLI tool to make benchmarking easy. \nFor more info visit https://github.com/Shravan-1908/atomic.")

	defaultShellValue, err := getShell()
	if err != nil {
		defaultShellValue = dummyDefault
	}
		
	// * root command
	// todo add a timeout flag
	commando.
		Register(nil).
		SetShortDescription("Benchmark a command for given number of iterations.").
		SetDescription("Benchmark a command for given number of iterations.").
		AddArgument("command...", "The command to run for benchmarking.", "").
		AddFlag("iterations,i", "The number of iterations to perform", commando.Int, 10).
		AddFlag("warmup,w", "The number of warmup runs to perform.", commando.Int, 0).
		AddFlag("prepare,p", "The command to execute once before every run.", commando.String, dummyDefault).
		AddFlag("ignore-error,I", "Ignore if the process returns a non-zero return code", commando.Bool, false).
		AddFlag("shell,s", "Whether to use shell to execute the given command.", commando.Bool, false).
		AddFlag("shell-path", "Path to the shell to use.", commando.String, defaultShellValue).
		AddFlag("export,e", "Comma separated list of benchmark export formats, including json, text, csv and markdown.", commando.String, "none").
		AddFlag("verbose,V", "Enable verbose output.", commando.Bool, false).
		AddFlag("no-color", "Disable colored output.", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			// * getting args and flag values
			if strings.TrimSpace(args["command"].Value) == "" {
				internal.Log("red", "Error: not enough arguments.")
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
				return
			}
			NO_COLOR, e = (flags["color"].GetBool())
			if e != nil {
				internal.Log("red", "Application error: cannot parse flag values.")
				return
			}
			internal.NO_COLOR = !NO_COLOR

			ignoreError, er := flags["ignore-error"].GetBool()
			if er != nil {
				internal.Log("red", "Application error: cannot parse flag values.")
				return
			}

			useShell, er := flags["shell"].GetBool()
			if er != nil {
				internal.Log("red", "Application error: cannot parse flag values.")
				return
			}
			shellPath, er := flags["shell-path"].GetString()
			if er != nil {
				internal.Log("red", "Application error: cannot parse flag values.")
				return
			}

			if iterations <= 0 {
				return
			}

			commandString := (args["command"].Value)
			command, err := buildCommand(commandString, useShell, shellPath)
			if err != nil {
				internal.Log("error", "unable to parse the given command: "+commandString)
				internal.Log("error", "error: "+err.Error())
				return
			}

			prepareCmdString, err := flags["prepare"].GetString()
			if err != nil {
				internal.Log("error", "unable to parse the given command: "+commandString)
				internal.Log("error", "error: "+err.Error())
				return
			}
			executePrepareCmd := true
			if prepareCmdString == dummyDefault {
				executePrepareCmd = false
			}
			var prepareCmd []string
			prepareCmd, err = buildCommand(prepareCmdString, useShell, shellPath)
			if err != nil {
				internal.Log("error", "unable to parse the given command: "+commandString)
				internal.Log("error", "error: "+err.Error())
				return
			}

			warmupOpts := BenchmarkOptions{
				command:           command,
				iterations:        warmupRuns,
				verbose:           verbose,
				ignoreError:       ignoreError,
				prepareCmd:        prepareCmd,
				executePrepareCmd: executePrepareCmd,
				mode:              warmupMode,
			}

			// no need for runs in warmups
			_, shouldReturn := benchmark(warmupOpts)
			if shouldReturn {
				return
			}

			benchmarkOpts := warmupOpts
			benchmarkOpts.iterations = iterations
			benchmarkOpts.mode = mainMode

			started := time.Now().Format("02-01-2006 15:04:05")
			runs, shouldReturn := benchmark(benchmarkOpts)
			if shouldReturn {
				return
			}
			ended := time.Now().Format("02-01-2006 15:04:05")

			// * intialising the template struct
			avg := internal.CalculateAverage(runs)
			stddev := internal.CalculateStandardDeviation(runs, avg)
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
				} else {
					internal.Log("yellow", "Since you're already using the --warmup flag, you can consider increasing the warmup count.")
				}
			}

		})

	commando.Parse(nil)
}
