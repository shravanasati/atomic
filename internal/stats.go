package internal

import (
	"math"
	"sort"

	"github.com/mitchellh/colorstring"
)

// borrowed from hyperfine
// https://github.com/sharkdp/hyperfine/blob/master/src/outlier_detection.rs
const zSCORE_THRESHOLD = 14.826

// can be modified by the outlier-threshold flag
var OUTLIER_THRESHOLD = 0.0

type numberLike interface {
	~int | ~float64 | ~int32 | ~int64 | ~float32
}

func CalculateAverage[T numberLike](data []T) float64 {
	var sum float64
	for _, v := range data {
		sum += float64(v)
	}
	return sum / float64(len(data))
}

// Computes the standard deviation of the given data.
func CalculateStandardDeviation[T numberLike](data []T, avg float64) float64 {
	var deviationSum float64 = 0
	n := float64(len(data))
	for _, v := range data {
		deviationSum += math.Pow((float64(v) - avg), 2)
	}
	deviationSum /= n
	deviationSum = math.Sqrt(deviationSum)

	return roundFloat(deviationSum, 2)
}

// returns a slice of absolute z-scores of each data point
func calculateModifiedZScore(data []float64) []float64 {
	median := calculateMedian(data)
	mad := calculateMAD(data, median)

	modifiedZScores := make([]float64, len(data))
	for i, value := range data {
		modifiedZScores[i] = math.Abs(0.6745 * (value - median) / mad)
	}

	return modifiedZScores
}

// calculates the median of data
func calculateMedian(data []float64) float64 {
	sort.Float64s(data)
	n := len(data)
	if n%2 == 0 {
		return (data[n/2-1] + data[n/2]) / 2
	}
	return data[n/2]
}

// calculates the median absolute deviation of data
func calculateMAD(data []float64, median float64) float64 {
	absoluteDeviations := make([]float64, len(data))
	for i, value := range data {
		absoluteDeviations[i] = math.Abs(value - median)
	}
	return calculateMedian(absoluteDeviations)
}

// Returns true if there are any statistical outliers in the data.
func TestOutliers[T numberLike](data []T) bool {
	zScores := calculateModifiedZScore(
		MapFunc[T, float64, []float64, []T](
			func(x T) float64 { return float64(x) },
			data,
		),
	)
	nOutliers := float64(len(
		FilterFunc(
			func(z float64) bool { return z > zSCORE_THRESHOLD },
			zScores,
		),
	))
	totalDataPoints := float64(len(data))
	
	return (nOutliers / totalDataPoints * 100) > OUTLIER_THRESHOLD
}

// todo represent this in a struct for export
func RelativeSummary(results []SpeedResult) {
	if len(results) <= 1 {
		return
	}
	sort.Sort(ByAverage(results))
	fastest := results[0]
	colorstring.Println("[bold][white]Summary")
	colorstring.Printf("  [cyan]%s[reset] ran \n", fastest.Command)
	for _, r := range results[1:] {
		ratio := r.Average / fastest.Average
		ratioStddev := ratio * math.Sqrt(
			math.Pow(r.StandardDeviation/r.Average, 2)+
				math.Pow(fastest.StandardDeviation/fastest.Average, 2),
		)
		colorstring.Printf("    [green]%.2f[reset] ± [light_green]%.2f[reset] times faster than [magenta]%s \n", ratio, ratioStddev, r.Command)
	}
}
