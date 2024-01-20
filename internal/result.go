package internal

import "time"

// Contains all the numerical quantities (in microseconds) for relative speed comparison. Also used for export.
type SpeedResult struct {
	Command           string    `json:"command,omitempty"`
	AverageElapsed    float64   `json:"average_elapsed,omitempty"`
	AverageUser       float64   `json:"average_user,omitempty"`
	AverageSystem     float64   `json:"average_system,omitempty"`
	StandardDeviation float64   `json:"standard_deviation,omitempty"`
	Max               float64   `json:"max,omitempty"`
	Min               float64   `json:"min,omitempty"`
	Times             []float64 `json:"times,omitempty"`
	RelativeMean      float64   `json:"relative_mean,omitempty"`
	RelativeStddev    float64   `json:"relative_stddev,omitempty"`
}

// PrintableResult struct which is shown at the end as benchmarking summary and is written to a file.
// Other numerical quantities except runs are represented as strings because they are
// durations, and time.Duration offers a .String() method.
type PrintableResult struct {
	Command           string
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
