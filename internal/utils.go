package internal

import (
	"fmt"
	"math"
	"os"
	"strings"
)

// formats the text in a javascript like syntax.
func format(text string, params map[string]string) string {
	for key, val := range params {
		text = strings.Replace(text, fmt.Sprintf("${%v}", key), val, -1)
	}
	return text
}

// writeToFile writes text string to the given filename.
func writeToFile(text, filename string) (err error) {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(text)
	return err
}

type numberLike interface {
	~int | ~float64 | ~int32 | ~int64 | ~float32
}

func ComputeAverageAndStandardDeviation[T numberLike](population []T) (float64, float64) {
	var deviationSum float64 = 0
	var avg float64 = 0
	n := float64(len(population))
	for _, v := range population {
		avg += float64(v)
	}
	avg /= n
	for _, v := range population {
		deviationSum += math.Pow((float64(v) - avg), 2)
	}
	deviationSum /= n
	deviationSum = math.Sqrt(deviationSum)

	return avg, deviationSum
}