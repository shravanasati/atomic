package internal

import "time"

// Contains all the numerical quantities (in microseconds) for relative speed comparison. Also used for export.
type SpeedResult struct {
	Command           string
	AverageElapsed    float64
	AverageUser       float64
	AverageSystem     float64
	StandardDeviation float64
	Max               float64
	Min               float64
	Times             []float64
	RelativeMean      float64
	RelativeStddev    float64
}

// PrintableResult struct which is shown at the end as benchmarking summary and is written to a file.
// Other numerical quantities except runs are represented as strings because they are
// durations, and time.Duration offers a .String() method.
type PrintableResult struct {
	Command           string // Command is different from OriginalCommand such that it doesn't include shell prefixes etc.
	Runs              int
	AverageElapsed    string
	AverageUser       string
	AverageSystem     string
	StandardDeviation string
	Min               string
	Max               string
}

func NewPrintableResult() *PrintableResult {
	var pr PrintableResult
	return &pr
}

func (pr *PrintableResult) FromSpeedResult(sr SpeedResult) *PrintableResult {
	pr.Command = sr.Command
	pr.Runs = len(sr.Times)
	pr.AverageElapsed = DurationFromNumber(sr.AverageElapsed, time.Microsecond).String()
	pr.AverageUser = DurationFromNumber(sr.AverageUser, time.Microsecond).String()
	pr.AverageSystem = DurationFromNumber(sr.AverageSystem, time.Microsecond).String()
	pr.StandardDeviation = DurationFromNumber(sr.StandardDeviation, time.Microsecond).String()
	pr.Max = DurationFromNumber(sr.Max, time.Microsecond).String()
	pr.Min = DurationFromNumber(sr.Min, time.Microsecond).String()
	return pr
}

// Implements [sort.Interface] for []Result based on the Average field.
type ByAverage []*SpeedResult

func (a ByAverage) Len() int {
	return len(a)
}

func (a ByAverage) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByAverage) Less(i, j int) bool {
	return a[i].AverageElapsed < a[j].AverageElapsed
}
