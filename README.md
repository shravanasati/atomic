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

**For Linux users:**

Execute the following command in bash:

```bash
curl https://raw.githubusercontent.com/shravanasati/atomic/master/scripts/linux_install.sh > bench_install.sh

chmod +x ./bench_install.sh

bash ./bench_install.sh
```


**For MacOS users:**

Execute the following command in bash:

```bash
curl https://raw.githubusercontent.com/shravanasati/atomic/master/scripts/macos_install.sh > bench_install.sh

chmod +x ./bench_install.sh

bash ./bench_install.sh
```

**For Windows users:**

Open Powershell **as Admin** and execute the following command:
```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force; (Invoke-WebRequest -Uri https://raw.githubusercontent.com/shravanasati/atomic/master/scripts/windows_install.ps1 -UseBasicParsing).Content | powershell -
```

To verify the installation of *atomic*, open a new shell and execute `atomic -v`. You should see output like this:
```
atomic 0.1.1

Version: 0.1.1
```
If the output isn't something like this, you need to repeat the above steps carefully.


<br>

## 💡 Usage
This section shows how you can use *atomic*.


You can benchmark anything with atomic, python programs, executables, shell commands or anything. To benchmark with atomic, simply execute:

```
atomic <command> [runs]
```

The `command` argument is the command to execute for benchmarking, like `python3 file` or `./executable`.

The `runs` argument defaults to 10, if not provided.

Example:
```
atomic "node speedtest.js" 20
```

You can export the benchmark summary in three different formats - markdown, text and json.

To export the results, use the `--export` flag. A file named `atomic-summary.format` will be created.

Example:
```
atomic "node speedtest.js" 20 --export json
```


### version
`$ atomic version`
>
The version command shows the version of *atomic* installed.

### help
`$ atomic help`

Renders assistance for *atomic* on a terminal, briefly showing its usage.

<br

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
