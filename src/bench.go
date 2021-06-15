package main

import (
	"fmt"
	"github.com/thatisuday/commando"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const (
	NAME    = "bench"
	VERSION = "0.1.0"
)

type Result struct {
	Command    string
	Iterations int
	Average    time.Duration
}

func main() {
	fmt.Println(NAME, VERSION)

	// * basic configuration
	commando.
		SetExecutableName(NAME).
		SetVersion(VERSION).
		SetDescription("bench is a simple CLI tool to make benchmarking easy.")

	commando.
		Register(nil).
		AddArgument("command", "The command to run for benchmarking.", "").
		AddArgument("iterations", "The number of times to iterate for benchmarking", "10").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			executable := args["command"].Value
			command := strings.Fields(executable)
			x, e := strconv.Atoi(args["iterations"].Value)
			if e != nil {
				fmt.Println("Wrong input for iterations.")
				return
			}
			var sum time.Duration

			for i := 1; i <= x; i++ {
				cmd := exec.Command(command[0], command[1:]...)
				_, e := cmd.StdoutPipe()
				if e != nil {
					panic(e)
				}

				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				init := time.Now()
				cmd.Start()
				cmd.Wait()
				sum += time.Since(init)
			}
			text := `
Benchmarking Summary
--------------------

Executed Command: {{ .Command }}
Total iterations: {{ .Iterations }}
Average time taken: {{ .Average }}
`
			result := Result{
				executable,
				x,
				((sum) / time.Duration(x)),
			}

			tmpl, err := template.New("result").Parse(text)
			if err != nil {
				panic(err)
			}
			err = tmpl.Execute(os.Stdout, result)
			if err != nil {
				panic(err)
			}
		})

	commando.Parse(nil)
}
