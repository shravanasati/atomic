package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"

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

type RunResult struct {
	elapsed time.Duration
	user    time.Duration
	system  time.Duration
	err     error
}

// runs the built command using os/exec and returns the duration the command lasted
func RunCommand(runOpts *RunOptions) *RunResult {
	var cmd *exec.Cmd
	runResult := &RunResult{
		elapsed: 0,
		user:    0,
		system:  0,
		err:     nil,
	}
	ctx, cancel := context.WithTimeout(context.Background(), runOpts.timeout)
	defer cancel()
	cmd = exec.CommandContext(ctx, runOpts.command[0], runOpts.command[1:]...)

	if runOpts.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	var e error
	if e = cmd.Start(); e != nil {
		runResult.err = &failedProcessError{command: runOpts.command, err: e, where: "starting"}
	}
	init := time.Now()
	e = cmd.Wait()
	duration := time.Since(init)

	if e != nil {
		if ctx.Err() == context.DeadlineExceeded {
			runResult.err = &failedProcessError{command: runOpts.command, err: context.DeadlineExceeded, where: "execution"}
		}
		if !runOpts.ignoreError {
			runResult.err = &failedProcessError{command: runOpts.command, err: e, where: "execution"}
		}
	}

	runResult.elapsed = duration
	runResult.user = cmd.ProcessState.UserTime()
	runResult.system = cmd.ProcessState.SystemTime()

	return runResult
}

// todo add graphing support
// todo --time-unit/-u flag for exports

var MinRuns = 10
var MaxRuns = math.MaxInt64
var MinDuration = (3 * time.Second).Microseconds()

// Determine the number of runs from a single run duration. This happens by meeting both
// of these criteria:
// 1. Minimum number of runs to be performed: 10
// 2. Minimum duration the benchmark should last: 3s
func determineRuns(singleRuntime time.Duration) int {
	// todo adjust runs based on running average
	if (singleRuntime.Microseconds() * int64(MinRuns)) > MinDuration {
		return MinRuns
	} else {
		runs := int(float64(MinDuration) / float64(singleRuntime.Microseconds()))
		return min(runs, MaxRuns)
	}
}

type benchmarkMode int

const (
	shellMode  benchmarkMode = 0
	warmupMode benchmarkMode = 1
	mainMode   benchmarkMode = 2
)

type BenchmarkOptions struct {
	command           []string
	runs              int
	verbose           bool
	ignoreError       bool
	executePrepareCmd bool
	prepareCmd        []string
	executeCleanupCmd bool
	cleanupCmd        []string
	shellCalibration  *RunResult
	mode              benchmarkMode
	timeout           time.Duration
}

// Benchmark runs the given command as per the given opts and returns a slice of durations in
// microseconds as well as the number of runs performed and whether the Benchmark was NOT successfull.
func Benchmark(opts BenchmarkOptions) ([]*RunResult, bool) {
	// actual runs, each entry stored in microseconds
	var runsData []*RunResult
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

	// * looping for given runs
	if opts.verbose {
		word, ok := wordMap[opts.mode]
		if !ok {
			// used internally, ok to panic
			panic(fmt.Sprintf("invalid mode passed to benchmark: %v", opts.mode))
		}
		startI := 1
		if opts.runs < 0 {
			var prepareResult, cleanupResult *RunResult
			if opts.executePrepareCmd {
				prepareResult = RunCommand(&prepareRunOpts)
				if errors.As(prepareResult.err, &processErr) {
					processErr.handle()
					return nil, true
				}
			}
			startI = 2
			singleRunResult := RunCommand(&runOpts)
			if errors.As(singleRunResult.err, &processErr) {
				processErr.handle()
				return nil, true
			}
			if opts.executeCleanupCmd {
				cleanupResult = RunCommand(&cleanupRunOpts)
				if errors.As(cleanupResult.err, &processErr) {
					processErr.handle()
					return nil, true
				}
			}
			opts.runs = determineRuns(singleRunResult.elapsed + prepareResult.elapsed + cleanupResult.elapsed)
			singleRunResult.elapsed -= opts.shellCalibration.elapsed
			singleRunResult.user -= opts.shellCalibration.user
			singleRunResult.system -= opts.shellCalibration.system
			runsData = append(runsData, singleRunResult)
		}
		for i := startI; i <= opts.runs; i++ {
			internal.Log("purple", fmt.Sprintf("***********\nRunning "+word+" %d\n***********", i))

			// dont output prepare command execution
			if opts.executePrepareCmd {
				prepareResult := RunCommand(&prepareRunOpts)
				if errors.As(prepareResult.err, &processErr) {
					processErr.handle()
					return nil, true
				}
			}
			runResult := RunCommand(&runOpts)
			if errors.As(runResult.err, &processErr) {
				processErr.handle()
				return nil, true
			}
			runResult.elapsed -= opts.shellCalibration.elapsed
			runResult.user -= opts.shellCalibration.user
			runResult.system -= opts.shellCalibration.system
			runsData = append(runsData, runResult)

			if opts.executeCleanupCmd {
				cleanupResult := RunCommand(&cleanupRunOpts)
				if errors.As(cleanupResult.err, &processErr) {
					processErr.handle()
					return nil, true
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
		barMax := opts.runs
		if barMax < 0 {
			barMax = 1
		}
		bar := progressbar.NewOptions(
			barMax, pbarOptions...,
		)
		startI := 1
		var prepareResult, cleanupResult *RunResult

		// automatically determine runs
		if opts.runs < 0 {
			startI = 2
			if opts.executePrepareCmd {
				prepareResult = RunCommand(&prepareRunOpts)
				if errors.As(prepareResult.err, &processErr) {
					processErr.handle()
					return nil, true
				}
			}
			singleRunResult := RunCommand(&runOpts)
			if errors.As(singleRunResult.err, &processErr) {
				processErr.handle()
				return nil, true
			}

			if opts.executeCleanupCmd {
				cleanupResult = RunCommand(&cleanupRunOpts)
				if errors.As(cleanupResult.err, &processErr) {
					processErr.handle()
					return nil, true
				}
			}
			opts.runs = determineRuns(singleRunResult.elapsed + prepareResult.elapsed + cleanupResult.elapsed)
			singleRunResult.elapsed -= opts.shellCalibration.elapsed
			singleRunResult.user -= opts.shellCalibration.user
			singleRunResult.system -= opts.shellCalibration.system
			bar.Reset()
			bar.ChangeMax(opts.runs)
			bar.Add(1)
			runsData = append(runsData, singleRunResult)
		}
		for i := startI; i <= opts.runs; i++ {
			// run the prepareCmd first
			// dont ignore errors in prepare command execution, dont output it either
			if opts.executePrepareCmd {
				prepareResult = RunCommand(&prepareRunOpts)
				if errors.As(prepareResult.err, &processErr) {
					processErr.handle()
					return nil, true
				}
			}
			runResult := RunCommand(&runOpts)
			if errors.As(runResult.err, &processErr) {
				bar.Clear()
				processErr.handle()
				return nil, true
			}
			runResult.elapsed -= opts.shellCalibration.elapsed
			runResult.user -= opts.shellCalibration.user
			runResult.system -= opts.shellCalibration.system
			runsData = append(runsData, runResult)

			if opts.mode == mainMode {
				bar.Describe(
					fmt.Sprintf("[magenta]Current estimate: [green]%s[reset]",
						internal.DurationFromNumber(
							internal.CalculateAverage(
								internal.MapFunc[[]*RunResult, []float64](func(r *RunResult) float64 { return float64(r.elapsed.Microseconds()) },
									runsData[:]),
							), time.Microsecond).String(),
					),
				)
			}
			bar.Add(1)
			if opts.executeCleanupCmd {
				cleanupResult := RunCommand(&cleanupRunOpts)
				if errors.As(cleanupResult.err, &processErr) {
					processErr.handle()
					return nil, true
				}
			}
		}
	}
	return runsData, false
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
		SetShortDescription("Benchmark a command for given number of runs.").
		SetDescription("Benchmark a command for given number of runs.").
		AddArgument("commands...", "The command to run for benchmarking.", "").
		AddFlag("min,m", "Minimum number of runs to perform.", commando.Int, MinRuns).
		AddFlag("max,M", "Maximum number of runs to perform.", commando.Int, MaxRuns).
		AddFlag("runs,r", "The number of runs to perform", commando.Int, -1).
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
		AddFlag("outlier-threshold", "Minimum number of runs to be outliers for the outlier warning to be displayed, in percentage.", commando.String, "0").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			// * getting args and flag values
			if strings.TrimSpace(args["commands"].Value) == "" {
				internal.Log("red", "error: not enough arguments.")
				return
			}
			runs, e := flags["runs"].GetInt()
			if e != nil {
				internal.Log("red", "The number of runs must be an integer!")
				internal.Log("white", e.Error())
				return
			}

			MinRuns, e = flags["min"].GetInt()
			if e != nil {
				internal.Log("red", "The number of minimum runs must be an integer!")
				internal.Log("white", e.Error())
				return
			}

			MaxRuns, e = flags["max"].GetInt()
			if e != nil {
				internal.Log("red", "The number of maximum runs must be an integer!")
				internal.Log("white", e.Error())
				return
			}

			warmupRuns, e := flags["warmup"].GetInt()
			if e != nil {
				internal.Log("red", "The number of runs must be an integer!")
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

			outlierThresholdString, e := flags["outlier-threshold"].GetString()
			if e != nil {
				internal.Log("red", "Application error: cannot parse flag values.")
				return
			}
			outlierThreshold, e := strconv.ParseFloat(outlierThresholdString, 64)
			if e != nil {
				internal.Log("red", "The outlier threshold percentage must be a decimal value.")
				return
			}
			if outlierThreshold < 0 && outlierThreshold > 100 {
				internal.Log("red", "The value outlier threshold can only be between 0 and 100, inclusive.")
				return
			}
			internal.OUTLIER_THRESHOLD = outlierThreshold

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

			var shellCalibration *RunResult
			if useShell {
				shellEmptyCommand, err := buildCommand("''", true, shellPath)
				if err != nil {
					internal.Log("red", "unable to calibrate shell: make sure you can run "+shellPath)
					internal.Log("red", "error: "+err.Error())
					return
				}
				calibrationOpts := BenchmarkOptions{
					command:           shellEmptyCommand,
					runs:              -1,
					verbose:           false,
					ignoreError:       true,
					executePrepareCmd: false,
					prepareCmd:        []string{},
					executeCleanupCmd: false,
					cleanupCmd:        []string{},
					mode:              shellMode,
					timeout:           LargestDuration,
					shellCalibration:  &RunResult{},
				}
				runs, failed := Benchmark(calibrationOpts)
				if failed {
					return
				}
				shellElapsedAvg := internal.CalculateAverage(internal.MapFunc[[]*RunResult, []float64](func(r *RunResult) float64 { return float64(r.elapsed.Microseconds()) }, runs))
				shellUserAvg := internal.CalculateAverage(internal.MapFunc[[]*RunResult, []float64](func(r *RunResult) float64 { return float64(r.user.Microseconds()) }, runs))
				shellSystemAvg := internal.CalculateAverage(internal.MapFunc[[]*RunResult, []float64](func(r *RunResult) float64 { return float64(r.system.Microseconds()) }, runs))
				shellElapsedAvgDuration := internal.DurationFromNumber(shellElapsedAvg, time.Microsecond)
				shellUserAvgDuration := internal.DurationFromNumber(shellUserAvg, time.Microsecond)
				shellSystemAvgDuration := internal.DurationFromNumber(shellSystemAvg, time.Microsecond)
				shellCalibration = &RunResult{
					elapsed: shellElapsedAvgDuration,
					user:    shellUserAvgDuration,
					system:  shellSystemAvgDuration,
				}
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
					runs:              warmupRuns,
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
				_, shouldSkip := Benchmark(warmupOpts)
				if shouldSkip {
					continue
				}

				benchmarkOpts := warmupOpts
				benchmarkOpts.runs = runs
				benchmarkOpts.mode = mainMode

				runsData, shouldSkip := Benchmark(benchmarkOpts)
				if shouldSkip {
					continue
				}
				elapsedTimes := internal.MapFunc[[]*RunResult, []float64](func(rr *RunResult) float64 { return float64(rr.elapsed.Microseconds()) }, runsData)
				userTimes := internal.MapFunc[[]*RunResult, []float64](func(rr *RunResult) float64 { return float64(rr.user.Microseconds()) }, runsData)
				systemTimes := internal.MapFunc[[]*RunResult, []float64](func(rr *RunResult) float64 { return float64(rr.system.Microseconds()) }, runsData)

				// * intialising the template struct
				avgElapsed := internal.CalculateAverage(elapsedTimes)
				avgUser := internal.CalculateAverage(userTimes)
				avgSystem := internal.CalculateAverage(systemTimes)
				if avgElapsed < 0 {
					internal.Log("red", "shell calibration is yielding inaccurate results")
					internal.Log("yellow", "Try executing the command without the -s/--shell flag.")
					continue
				}
				stddev := internal.CalculateStandardDeviation(elapsedTimes, avgElapsed)
				avgElapsedDuration := internal.DurationFromNumber(avgElapsed, time.Microsecond)
				avgUserDuration := internal.DurationFromNumber(avgUser, time.Microsecond)
				avgSystemDuration := internal.DurationFromNumber(avgSystem, time.Microsecond)
				stddevDuration := internal.DurationFromNumber(stddev, time.Microsecond)
				max_ := slices.Max(elapsedTimes)
				min_ := slices.Min(elapsedTimes)
				maxDuration := internal.DurationFromNumber(max_, time.Microsecond)
				minDuration := internal.DurationFromNumber(min_, time.Microsecond)
				speedResult := internal.SpeedResult{
					Command:           commandString,
					Average:           avgElapsed,
					StandardDeviation: stddev,
				}
				printableResult := internal.PrintableResult{
					Command:           strings.Join(command, " "),
					Runs:              len(runsData),
					AverageElapsed:    avgElapsedDuration.String(),
					AverageUser:       avgUserDuration.String(),
					AverageSystem:     avgSystemDuration.String(),
					StandardDeviation: stddevDuration.String(),
					Max:               maxDuration.String(),
					Min:               minDuration.String(),
				}
				speedResults = append(speedResults, speedResult)
				fmt.Print(printableResult.String())

				outliersDetected := internal.TestOutliers(elapsedTimes)
				if outliersDetected {
					internal.Log("yellow", "\nWarning: Statistical outliers were detected. Consider re-running this benchmark on a quiet system, devoid of any interferences from other programs.")
					if warmupRuns == 0 {
						internal.Log("yellow", "It might help to use the --warmup flag.")
					} else {
						internal.Log("yellow", "Since you're already using the --warmup flag, you can consider increasing the warmup count.")
					}
				}

				// min is in microseconds
				if min_ < float64((5 * time.Millisecond).Microseconds()) {
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

			// // * getting export values
			// exportFormats, ierr := flags["export"].GetString()
			// if ierr != nil {
			// 	internal.Log("red", "Application error: cannot parse flag values.")
			// 	return
			// }
			// // todo export all results
			// result.Export(exportFormats)

		})

	commando.Parse(nil)
}
