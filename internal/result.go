package internal

// Contains all the numerical quantities (in microseconds) for relative speed comparison.
type SpeedResult struct {
	Command           string
	Average           float64
	StandardDeviation float64
}

// PrintableResult struct which is shown at the end as benchmarking summary and is written to a file.
// Other numerical quantities except runs are represented as strings because they are
// durations, and time.Duration offers a .String() method.
type PrintableResult struct {
	Command           string // Command is different from OriginalCommand such that it doesn't include shell prefixes etc.
	Runs              int
	Average           string
	StandardDeviation string
	Min               string
	Max               string
}

// Implements [sort.Interface] for []Result based on the Average field.
type ByAverage []SpeedResult

func (a ByAverage) Len() int {
	return len(a)
}

func (a ByAverage) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByAverage) Less(i, j int) bool {
	return a[i].Average < a[j].Average
}
