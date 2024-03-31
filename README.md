# atomic

[![Continuous integration](https://github.com/shravanasati/atomic/actions/workflows/integrate.yml/badge.svg)](https://github.com/shravanasati/atomic/actions/workflows/integrate.yml)

![atomic demo](assets/demo.gif)


*atomic* is a simple CLI tool for making benchmarking easy.


<br>

## ✨ Features

- Detailed benchmark summary at the end
- Export the results in markdown, json, csv format
- Statistical Outlier Detection
- Plot the benchmarking data, comparing the different commands
- Arbitrary command support 
- Constant feedback about the benchmark progress and current estimates.
- Warmup runs can be executed before the actual benchmark.
- Cache-clearing commands can be set up before each timing run.

<br>

## ⚡️ Installation


### Installation Scripts

#### Linux and macOS

```bash
curl https://raw.githubusercontent.com/shravanasati/atomic/master/scripts/install.sh | bash
```

### Package Managers

#### Windows
```powershell
scoop install https://github.com/shravanasati/atomic/raw/master/scripts/atomic.json
```

### GitHub Releases

atomic binaries for all operating systems are available on the [GitHub Releases](https://github.com/shravanasati/atomic/releases/latest) tab. You can download them manually and place them on `PATH` in order to use them.

To simplify this process, you can use [eget](https://github.com/zyedidia/eget):
```
eget shravanasati/atomic
```

### Using Go compiler

If you've Go compiler (v1.21 or above) installed on your system, you can install atomic via the following command. 

```
go install github.com/shravanasati/atomic@latest
```


### Build from source

You can alternatively build atomic from source via the following commands (again, requires go1.21 or above):

```
git clone https://github.com/shravanasati/atomic.git
cd ./atomic
go build
```

If you want to build atomic in release mode (stripped binaries, compressed distribution and cross compilation), execute the following command. You can also control the release builds behavior using the [`release.config.json`](./scripts/release.config.json) file.

```
python ./scripts/build.py
```

<br>

To verify the installation of *atomic*, open a new shell and execute `atomic -v`. You should see output like this:
```
atomic 0.4.0

Version: 0.4.0
```
If the output isn't something like this, you need to repeat the above steps carefully.


<br>

## 💡 Usage


### Simple benchmarks

Let's benchmark our first CLI command using atomic.

```
atomic "grep -iFr 'type'"
```

This grep command searches the codebase for the plain text 'type', recursively and case sensitively.

Notice how atomic automatically determines the number of runs to perform in the benchmark. atomic will run any given command atleast 10 times and for atleast 3 seconds, when determining the number of runs on its own.

You can alter this behavior using the `--min/-m X` and `--max/-M Y` flags. When the `--min/-m X` flag is passed, atomic will run the command atleast X number of times. If the `--max/-M Y` flag is passed, atomic will perform maximum Y number of runs.

```
atomic "grep -iFr 'type'" -m 50 -M 200
```

Both these flags are independent of each other, you can use any one of those at a time.

You can pass the exact number of runs to perform using the `--runs/-r N` flag.

```
atomic "grep -iFr 'type'" --runs 50
```

### Warmup runs and preparation & cleanup commands

You might get a warning that says atomic found statistical outliers in the benchmark, which generally happens due to missing filesystem caches (especially for IO heavy programs like grep) and/or interferences from other running programs (OS context switches).

You can use the `--warmup/-w N` flag to ask atomic to run the command N number of times before beginning the actual benchmark.

```
atomic "grep -iFr 'type'" --runs 50 --warmup 10
```

atomic raises the statistical outlier warning even if one of the data point (execution time in this case) is an outlier. You can raise this threshold using the `--outlier-threshold P` flag where P is the minimum percentage of outliers that should be present in the benchmark data for atomic to raise the warning. 

A command you are running may require some additional setup before every time it is executed, or need to remove some assets it has generated after the execution. You can use the `--prepare/-p command` and `--cleanup/-c command` flags respectively to achieve above tasks.

```
atomic "go build" --prepare "go generate ./..." --cleanup "rm *.exe"
```

### Intermediate shells

> This feature is under development.

Let's look at another command to benchmark.

```
atomic ls
```

This command might give the error on Windows, because `ls` is not a executable but a shell function, provided by powershell, unlike Unix systems.

In such cases, use the `--shell/-s` flag to ask atomic to run the command within a shell context.
atomic will use `cmd.exe` on Windows and `/bin/sh` on Unix-based systems by default. 
You can choose a custom shell too using the `--shell-path path` flag. Since `ls` is not provided by `cmd.exe` either, we'll use powershell.

```
atomic ls -s --shell-path powershell
```

atomic will perform shell calibration (substracting shell spawn time from total process execution time) but it's far from perfect (even working state) and may yield negative runtimes. 

### Timeouts & Debugging failed benchmarks

atomic also offers a `--timeout/-t D` flag, which tells atomic to cancel the benchmark if any of the run of given commands takes longer than D, where D is the time duration which can expressed as following: `number{ns|us|ms|s|m|h}`.

```
atomic "grep -iFr 'type'" --timeout 100ms
```

Sometimes the command you've given might finish with non-zero exit code, which generally indicates that it failed to execute successfully.

In such cases, atomic will stop the benchmark immediately. 

You can alter this behaviour with the `--ignore-error/-I` flag, which will make atomic continue the benchmark even if commands return non-zero exit codes.

Use this flag cautiously, advisably only when you know why the command is behaving that way and whether it is desired.

A good example is Go compiler when called with zero arguments:

```
atomic go
```

This will fail - to investigate what's wrong you can use the `--verbose/-V` flag to show the output of the command. 

Turns out, the Go compiler just renders the help text with an exit status of 2, which causes the benchmark to fail. It's safe to use the `--ignore-error/-I` flag now.

On another note, use the `--verbose/-V` flag sparingly (writing to stdout is expensive).

### Comparing several commands

atomic accepts multiple commands to benchmark, and then also displays relative summary at the end, comparing those commands.

Let's use atomic to compare [scc](https://github.com/boyter/scc) and [tokei](https://github.com/XAMPPRocky/tokei), two popular code counting tools.

```
atomic scc tokei -w 20
```

All the flags you provide to atomic will be applied for all the commands.
In this example, 20 warmup runs will be executed for both scc and tokei.



### Exports and plots

atomic can export the benchmarking data in JSON, markdown, CSV and text formats.

Use the `--export/-e formats` flag where `formats` is the comma-separated list of supported export formats.

```
atomic 'grep -iFr "type"' --export json,csv,md
```

atomic will create three files named `atomic-summary.{ext}` with the benchmark data.

The default name for exports is `atomic-summary` but can be modified using the `--filename/-f name` flag. 

The default time unit for exports is `ms` (milliseconds) but can be modified using the `--time-unit/-u unit` flag. Valid values for time units are {ns, us, ms, s, m, h}.

```
atomic 'grep -iFr "type"' -e json,csv,md -f grep_data -u s
```

JSON export has the most amount of data, since it also contains a key named `times` which is an array of all run-times. This JSON output also matches with that of [hyperfine](https://github.com/sharkdp/hyperfine)'s, so you can utilize the Python scripts in the hyperfine repository that visualize the benchmark data (and vice versa with the upcoming `plot` command). 

You can also plot this data using the `--plot T` flag, where `T` is the comma-separated list of chart formats. Valid values for T include {hist, histogram, bar, **all**}. If all is used as `T`, atomic will plot the data with all chart types.

```
atomic 'ag pattern' 'rg pattern' --plot all
```

> The plot feature is also under development.

<br>

## Acknowledgement

This tool is heavily inspired by [*hyperfine*](https://github.com/sharkdp/hyperfine). I learnt a lot of stuff looking at the code of this project and tried matching the feature-set as close as possible.

## Known Issues and Missing Features

- [ ] Shell calibration yields negative process run times
- [ ] No Color functionality is broken
- [ ] Implementation of errorbar, boxplot and bubble chart is pending
- [ ] Plot command is missing

## 🔖 Versioning
*atomic* releases follow semantic versioning, every release is in the *x.y.z* form, where:
- *x* is the MAJOR version and is incremented when a backwards incompatible change to atomic is made.
- *y* is the MINOR version and is incremented when a backwards compatible change to atomic is made, like changing dependencies or adding a new function, method, struct field, or type.
- *z* is the PATCH version and is incremented after making minor changes that don't affect atomic's public API or dependencies, like fixing a bug.

<br>

## 📄 License
License
© 2021-Present Shravan Asati

This repository is licensed under the MIT license. See [LICENSE](LICENSE) for details.

<br>
