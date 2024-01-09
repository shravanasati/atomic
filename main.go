package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	// "path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/Shravan-1908/atomic/internal"
	"github.com/Shravan-1908/commando"
	"github.com/google/shlex"
	"github.com/mitchellh/colorstring"
	"github.com/schollz/progressbar/v3"
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
	executeCleanupCmd bool
	cleanupCmd        []string
	shellCalibration  time.Duration
	mode              benchmarkMode
	timeout           time.Duration
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

// used as default value for the timeout flag,
// borrowed from [time.Duration.String]
const LARGEST_DURATION_STRING = "2540400h10m10.000000000s"

var LargestDuration, _ = time.ParseDuration(LARGEST_DURATION_STRING)

// returns the default shell path (pwsh/cmd on windows, /bin/sh on unix based systems) and an error.
func getDefaultShell() (string, error) {
	if WINDOWS {
		// windows
		// yield cmd.exe first because its shell calibration is more accurate than powershell
		// lookup the comspec env variable -> it contains the path to cmd.exe
		return "cmd.exe", nil
		// comspec, ok := os.LookupEnv("ComSpec")
		// if ok {
		// 	return filepath.ToSlash(comspec), nil
		// } else {
		// 	// otherwise find cmd.exe in $SystemRoot/System32
		// 	systemRoot, ok := os.LookupEnv("SystemRoot")
		// 	if !ok {
		// 		// fall back to powershell
		// 		if _testPowershell() {
		// 			return "powershell", nil
		// 		}
		// 		return "", fmt.Errorf("buildCommand with useShell=true on windows: neither ComSpec nor SystemRoot is set. powershell not found either")
		// 	}
		// 	comspec = filepath.Join(systemRoot, "System32", "cmd.exe")
		// 	return filepath.ToSlash(comspec), nil
		// }

	} else {
		// posix
		return "/bin/sh", nil
	}
}

// todo write tests

// builds the given command as per the given params.
// if useShell is true, adds a shell in front of the command.
// the shell is determined by the shellPath.
// returns the built command and an error.
func buildCommand(command string, useShell bool, shellPath string) ([]string, error) {
	var builtCommand []string
	var err error
	if useShell {
		// the flag that enables execution of command from a string
		// eg. -Command or -c on pwsh, /c on cmd.exe, -c on any other shell
		var commandSwitch string
		if strings.Contains(shellPath, "cmd.exe") || strings.Contains(shellPath, "cmd") {
			commandSwitch = "/c"
		} else {
			commandSwitch = "-c"
		}
		builtCommand, err = shlex.Split(fmt.Sprintf("\"%s\" %s \"%s\"", shellPath, commandSwitch, command))
	} else {
		builtCommand, err = shlex.Split(command)
	}
	return builtCommand, err
}

type failedProcessError struct {
	command []string
	err     error
	where   string
}

func (fpe *failedProcessError) Error() string {
	return fmt.Sprintf("The command `%s` failed in the process of %s!\nerror: %s", strings.Join(fpe.command, " "), fpe.where, fpe.err.Error())
}

func (fpe *failedProcessError) handle() {
	internal.Log("red", fpe.Error())
	if fpe.err == context.DeadlineExceeded {
		internal.Log("yellow", "This happened due to the -t/--timeout flag. Consider increasing the timeout duration for successfull execution of the command.")
		return
	}
	internal.Log("yellow", "You should consider using -I/--ignore-error flag to ignore failures in the command execution. Alternatively, you can also try the -V/--verbose flag to show the output of the command. If the command is actually a shell function, use -s/--shell flag to execute it via a shell.")
}

type RunOptions struct {
	command     []string
	verbose     bool
	ignoreError bool
	timeout     time.Duration
}

// runs the built command using os/exec and returns the duration the command lasted
func RunCommand(runOpts *RunOptions) (time.Duration, error) {
	// todo refactor to use runresponse and benchmark-response
	var cmd *exec.Cmd
	ctx, cancel := context.WithTimeout(context.Background(), runOpts.timeout)
	defer cancel()
	cmd = exec.CommandContext(ctx, runOpts.command[0], runOpts.command[1:]...)
	_, e := cmd.StdoutPipe()
	if e != nil {
		panic(e)
	}

	if runOpts.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if e := cmd.Start(); e != nil {
		return 0, &failedProcessError{command: runOpts.command, err: e, where: "starting"}
	}

	init := time.Now()
	e = cmd.Wait()
	duration := time.Since(init)

	if e != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return 0, &failedProcessError{command: runOpts.command, err: context.DeadlineExceeded, where: "execution"}
		}
		if !runOpts.ignoreError {
			return 0, &failedProcessError{command: runOpts.command, err: e, where: "execution"}
		}
	}

	return duration, nil
}

// Determine the number of iterations from a single run duration. This happens by meeting both
// of these criteria:
// 1. Minimum number of iterations to be performed: 10
// 2. Minimum duration the benchmark should last: 3s
func determineIterations(singleRuntime int64) int {
	minIterations := 10
	minDuration := (3 * time.Second).Microseconds()
	if (singleRuntime * int64(minIterations)) > minDuration {
		return minIterations
	} else {
		return int(minDuration / singleRuntime)
	}
}

// Benchmark runs the given command as per the given opts and returns a slice of durations in
// microseconds as well as the number of runs performed and whether the Benchmark was successfull.
func Benchmark(opts BenchmarkOptions) ([]int64, int, bool) {
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
	// dont ignore errors in prepare and cleanup command
	prepareRunOpts := RunOptions{
		command:     opts.prepareCmd,
		verbose:     opts.verbose,
		ignoreError: false,
		timeout:     opts.timeout,
	}
	runOpts := RunOptions{
		command:     opts.command,
		verbose:     opts.verbose,
		ignoreError: opts.ignoreError,
		timeout:     opts.timeout,
	}
	cleanupRunOpts := RunOptions{
		command:     opts.cleanupCmd,
		verbose:     opts.verbose,
		ignoreError: false,
		timeout:     opts.timeout,
	}
	// todo refactor this code to eliminate code repetition

	// * looping for given iterations
	if opts.verbose {
		word, ok := wordMap[opts.mode]
		if !ok {
			// used internally, ok to panic
			panic(fmt.Sprintf("invalid mode passed to benchmark: %v", opts.mode))
		}
		var e error
		startI := 1
		if opts.iterations < 0 {
			if opts.executePrepareCmd {
				_, e = RunCommand(&prepareRunOpts)
				if errors.As(e, &processErr) {
					processErr.handle()
					return nil, 0, true
				}
			}
			startI = 2
			singleRuntime, e := RunCommand(&runOpts)
			if errors.As(e, &processErr) {
				processErr.handle()
				return nil, 0, true
			}
			opts.iterations = determineIterations(singleRuntime.Microseconds())
			singleRuntime -= opts.shellCalibration
			runs = append(runs, singleRuntime.Microseconds())
			if opts.executeCleanupCmd {
				_, e := RunCommand(&cleanupRunOpts)
				if errors.As(e, &processErr) {
					processErr.handle()
					return nil, 0, true
				}
			}
		}
		for i := startI; i <= opts.iterations; i++ {
			internal.Log("purple", fmt.Sprintf("***********\nRunning "+word+" %d\n***********", i))

			// dont output prepare command execution
			if opts.executePrepareCmd {
				_, e := RunCommand(&prepareRunOpts)
				if errors.As(e, &processErr) {
					processErr.handle()
					return nil, 0, true
				}
			}
			dur, e := RunCommand(&runOpts)
			if errors.As(e, &processErr) {
				processErr.handle()
				return nil, 0, true
			}
			runs = append(runs, (dur - opts.shellCalibration).Microseconds())

			if opts.executeCleanupCmd {
				_, e := RunCommand(&cleanupRunOpts)
				if errors.As(e, &processErr) {
					processErr.handle()
					return nil, 0, true
				}
			}
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
		barMax := opts.iterations
		if barMax < 0 {
			barMax = 1
		}
		bar := progressbar.NewOptions(
			barMax, pbarOptions...,
		)
		var e error
		startI := 1

		// automatically determine iterations
		if opts.iterations < 0 {
			startI = 2
			if opts.executePrepareCmd {
				_, e = RunCommand(&prepareRunOpts)
				if errors.As(e, &processErr) {
					processErr.handle()
					return nil, 0, true
				}
			}
			singleRuntime, e := RunCommand(&runOpts)
			if errors.As(e, &processErr) {
				processErr.handle()
				return nil, 0, true
			}
			opts.iterations = determineIterations(singleRuntime.Microseconds())
			singleRuntime -= opts.shellCalibration
			bar.Reset()
			bar.ChangeMax(opts.iterations)
			bar.Add(1)
			runs = append(runs, singleRuntime.Microseconds())

			if opts.executeCleanupCmd {
				_, e := RunCommand(&cleanupRunOpts)
				if errors.As(e, &processErr) {
					processErr.handle()
					return nil, 0, true
				}
			}
		}
		for i := startI; i <= opts.iterations; i++ {
			// run the prepareCmd first
			// dont ignore errors in prepare command execution, dont output it either
			if opts.executePrepareCmd {
				_, e = RunCommand(&prepareRunOpts)
				if errors.As(e, &processErr) {
					processErr.handle()
					return nil, 0, true
				}
			}
			dur, e := RunCommand(&runOpts)
			if errors.As(e, &processErr) {
				bar.Clear()
				processErr.handle()
				return nil, 0, true
			}
			runs = append(runs, (dur - opts.shellCalibration).Microseconds())

			if opts.mode == mainMode {
				bar.Describe(
					fmt.Sprintf("[magenta]Current estimate: [green]%s[reset]",
						internal.DurationFromNumber(
							internal.CalculateAverage(runs), time.Microsecond).String(),
					),
				)
			}
			bar.Add(1)
			if opts.executeCleanupCmd {
				_, e := RunCommand(&cleanupRunOpts)
				if errors.As(e, &processErr) {
					processErr.handle()
					return nil, 0, true
				}
			}
		}
	}
	return runs, opts.iterations, false
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
		SetDescription("atomic is a simple CLI tool to benchmark commands. \nFor more info visit https://github.com/Shravan-1908/atomic.")

	defaultShellValue, err := getDefaultShell()
	if err != nil {
		defaultShellValue = dummyDefault
	}

	// todo track memory usage
	// * root command
	commando.
		Register(nil).
		SetShortDescription("Benchmark a command for given number of iterations.").
		SetDescription("Benchmark a command for given number of iterations.").
		AddArgument("commands...", "The command to run for benchmarking.", "").
		AddFlag("iterations,i", "The number of iterations to perform", commando.Int, -1).
		AddFlag("warmup,w", "The number of warmup runs to perform.", commando.Int, 0).
		AddFlag("prepare,p", "The command to execute once before every run.", commando.String, dummyDefault).
		AddFlag("cleanup,c", "The command to execute once after every run.", commando.String, dummyDefault).
		AddFlag("ignore-error,I", "Ignore if the process returns a non-zero return code", commando.Bool, false).
		AddFlag("shell,s", "Whether to use shell to execute the given command.", commando.Bool, false).
		AddFlag("shell-path", "Path to the shell to use.", commando.String, defaultShellValue).
		AddFlag("timeout,t", "The timeout for a single command.", commando.String, LARGEST_DURATION_STRING).
		AddFlag("export,e", "Comma separated list of benchmark export formats, including json, text, csv and markdown.", commando.String, "none").
		AddFlag("verbose,V", "Enable verbose output.", commando.Bool, false).
		AddFlag("no-color", "Disable colored output.", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			// * getting args and flag values
			if strings.TrimSpace(args["commands"].Value) == "" {
				internal.Log("red", "error: not enough arguments.")
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

			// todo NO_COLOR functionality is broken due to colorstring
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

			if (shellPath == dummyDefault) && useShell {
				internal.Log("red", "unable to determine the shell to use! supply the name of the shell (if present in $PATH) or the path to the shell using the --shell-path flag.")
				return
			}
			prepareCmdString, err := flags["prepare"].GetString()
			if err != nil {
				internal.Log("red", "unable to parse the given command: "+prepareCmdString)
				internal.Log("red", "error: "+err.Error())
				return
			}
			executePrepareCmd := prepareCmdString != dummyDefault
			var prepareCmd []string
			prepareCmd, err = buildCommand(prepareCmdString, useShell, shellPath)
			if err != nil {
				internal.Log("red", "unable to parse the given command: "+prepareCmdString)
				internal.Log("red", "error: "+err.Error())
				return
			}

			cleanupCmdString, err := flags["cleanup"].GetString()
			if err != nil {
				internal.Log("red", "unable to parse the given command: "+cleanupCmdString)
				internal.Log("red", "error: "+err.Error())
				return
			}
			executeCleanupCmd := cleanupCmdString != dummyDefault
			var cleanupCmd []string
			cleanupCmd, err = buildCommand(cleanupCmdString, useShell, shellPath)
			if err != nil {
				internal.Log("red", "unable to parse the given command: "+cleanupCmdString)
				internal.Log("red", "error: "+err.Error())
				return
			}

			timeoutString, err := flags["timeout"].GetString()
			if err != nil {
				internal.Log("red", "Application error: cannot parse flag values.")
				return
			}
			timeout, err := time.ParseDuration(timeoutString)
			if err != nil {
				internal.Log("red", "unable to parse timeout: "+timeoutString)
				internal.Log("red", "error: ")
				return
			}

			var shellCalibration time.Duration
			if useShell {
				shellEmptyCommand, err := buildCommand("''", true, shellPath)
				if err != nil {
					internal.Log("red", "unable to calibrate shell: make sure you can run "+shellPath)
					internal.Log("red", "error: "+err.Error())
					return
				}
				calibrationOpts := BenchmarkOptions{
					command:           shellEmptyCommand,
					iterations:        -1,
					verbose:           false,
					ignoreError:       false,
					executePrepareCmd: false,
					prepareCmd:        []string{},
					executeCleanupCmd: false,
					cleanupCmd:        []string{},
					mode:              shellMode,
					timeout:           LargestDuration,
					shellCalibration:  shellCalibration,
				}
				if strings.Contains(shellPath, "cmd.exe") {
					calibrationOpts.ignoreError = true
				}
				runs, _, failed := Benchmark(calibrationOpts)
				if failed {
					return
				}
				shellAvg := internal.CalculateAverage(runs)
				shellCalibration = internal.DurationFromNumber(shellAvg, time.Microsecond)
			}
			// fmt.Println(shellCalibration)

			var speedResults []internal.SpeedResult
			// * benchmark each command given
			givenCommands := strings.Split(args["commands"].Value, commando.VariadicSeparator)
			nCommands := len(givenCommands)
			for index, commandString := range givenCommands {
				if _, err := colorstring.Printf("[bold][magenta]Benchmark %d: [cyan]%s", index+1, commandString); err != nil {
					panic(err)
				}
				// ! don't remove this println: for some weird reason the above colorstring.Printf
				// doesnt' work without this
				fmt.Println()

				command, err := buildCommand(commandString, useShell, shellPath)
				if err != nil {
					internal.Log("red", "unable to parse the given command: "+commandString)
					internal.Log("red", "error: "+err.Error())
					continue
				}

				warmupOpts := BenchmarkOptions{
					command:           command,
					iterations:        warmupRuns,
					verbose:           verbose,
					ignoreError:       ignoreError,
					prepareCmd:        prepareCmd,
					executePrepareCmd: executePrepareCmd,
					executeCleanupCmd: executeCleanupCmd,
					cleanupCmd:        cleanupCmd,
					shellCalibration:  shellCalibration,
					mode:              warmupMode,
					timeout:           timeout,
				}

				// no need for runs in warmups
				_, _, shouldSkip := Benchmark(warmupOpts)
				if shouldSkip {
					continue
				}

				benchmarkOpts := warmupOpts
				benchmarkOpts.iterations = iterations
				benchmarkOpts.mode = mainMode

				runs, iterations, shouldSkip := Benchmark(benchmarkOpts)
				if shouldSkip {
					continue
				}
				if len(runs) != iterations {
					panic(fmt.Sprintf("mismatch between len(runs)=%d and iterations=%d", len(runs), iterations))
				}

				// * intialising the template struct
				avg := internal.CalculateAverage(runs)
				if avg < 0 {
					internal.Log("red", "shell calibration is yielding inaccurate results")
					internal.Log("yellow", "Try executing the command without the -s/--shell flag.")
					continue
				}
				stddev := internal.CalculateStandardDeviation(runs, avg)
				avgDuration := internal.DurationFromNumber(avg, time.Microsecond)
				stddevDuration := internal.DurationFromNumber(stddev, time.Microsecond)
				max_ := slices.Max(runs)
				min_ := slices.Min(runs)
				maxDuration := internal.DurationFromNumber(max_, time.Microsecond)
				minDuration := internal.DurationFromNumber(min_, time.Microsecond)
				speedResult := internal.SpeedResult{
					Command:           commandString,
					Average:           avg,
					StandardDeviation: stddev,
				}
				printableResult := internal.PrintableResult{
					Command:           strings.Join(command, " "),
					Iterations:        iterations,
					Average:           avgDuration.String(),
					StandardDeviation: stddevDuration.String(),
					Max:               maxDuration.String(),
					Min:               minDuration.String(),
				}
				speedResults = append(speedResults, speedResult)
				fmt.Print(printableResult.String())

				outliersDetected := internal.TestOutliers(runs)
				if outliersDetected {
					internal.Log("yellow", "\nWarning: Statistical outliers were detected. Consider re-running this benchmark on a quiet system, devoid of any interferences from other programs.")
					if warmupRuns == 0 {
						internal.Log("yellow", "It might help to use the --warmup flag.")
					} else {
						internal.Log("yellow", "Since you're already using the --warmup flag, you can consider increasing the warmup count.")
					}
				}

				// 5000us = 5ms, avg is in microseconds
				if min_ < (5 * time.Millisecond).Microseconds() {
					internal.Log("yellow", "\nWarning: The command took less than 5ms to execute, the results might be inaccurate.")
					if useShell {
						internal.Log("yellow", "Try running the command without the -s/--shell flag.")
					}
				}

				if index != (nCommands-1) || nCommands > 1 {
					// print new line b/w each benchmark
					// and at the end one too if relative summary
					// has to be printed
					fmt.Println()
				}

			}

			internal.RelativeSummary(speedResults)

			// * getting export values
			// exportFormats, ierr := flags["export"].GetString()
			// if ierr != nil {
			// 	internal.Log("red", "Application error: cannot parse flag values.")
			// 	return
			// }
			// todo export all results
			// result.Export(exportFormats)

		})

	commando.Parse(nil)
}
