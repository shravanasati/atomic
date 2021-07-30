package internal

// Result struct which is shown at the end as benchmarking summary and is written to a file.
type Result struct {
	Started    string
	Ended      string
	Command    string
	Iterations int
	Average    string
}