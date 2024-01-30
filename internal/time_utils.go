package internal

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

var ErrInvalidTimeUnit = errors.New("invalid time unit")

func roundFloat(num float64, digits int) float64 {
	tenMultiplier := math.Pow10(digits)
	return math.Round(num*tenMultiplier) / tenMultiplier
}

type numberLike interface {
	~int | ~float64 | ~int32 | ~int64 | ~float32
}

func DurationFromNumber[T numberLike](number T, unit time.Duration) time.Duration {
	unitToSuffixMap := map[time.Duration]string{
		time.Nanosecond:  "ns",
		time.Microsecond: "us",
		time.Millisecond: "ms",
		time.Second:      "s",
		time.Minute:      "m",
		time.Hour:        "h",
	}
	suffix, ok := unitToSuffixMap[unit]
	if !ok {
		// this function is only used internally, panic if unknown time unit is passed
		panic("unknown time unit in DurationFromNumber: " + unit.String())
	}
	numberFloat := roundFloat(float64(number), 2)
	timeString := fmt.Sprintf("%.2f%v", numberFloat, suffix)
	duration, err := time.ParseDuration(timeString)
	if err != nil {
		// again, function only used internally, invalid duration must not be present
		panic("unable to parse duration: " + timeString + " in DurationFromNumber \n" + err.Error())
	}
	return duration.Round(time.Microsecond)
}

func ParseTimeUnit(unitString string) (time.Duration, error) {
	switch strings.TrimSpace(strings.ToLower(unitString)) {
	case "ns":
		return time.Nanosecond, nil
	case "us", "µs":
		return time.Microsecond, nil
	case "ms":
		return time.Millisecond, nil
	case "s":
		return time.Second, nil
	case "m":
		return time.Minute, nil
	case "h":
		return time.Hour, nil
	default:
		return 0, ErrInvalidTimeUnit
	}
}

func convertToTimeUnit(given float64, unit time.Duration) float64 {
	// first get duration from microseconds
	duration := DurationFromNumber(given, time.Microsecond)
	switch unit {
	case time.Nanosecond:
		return float64(duration.Nanoseconds())
	case time.Microsecond:
		return float64(duration.Microseconds())
	case time.Millisecond:
		return float64(duration.Nanoseconds()) / float64(1e6)
	case time.Second:
		return duration.Seconds()
	case time.Minute:
		return duration.Minutes()
	case time.Hour:
		return duration.Hours()
	default:
		panic("convertToTimeUnit: unknown time unit: " + unit.String())
	}
}

// ModifyTimeUnit takes a slice of [SpeedResult] and modifies its every attribute to suit accordingly
// to the given timeUnit.
func ModifyTimeUnit(results []*SpeedResult, timeUnit time.Duration) {
	// except for microseconds, because that's what used internally
	if timeUnit != time.Microsecond {
		var wg sync.WaitGroup
		for _, sr := range results {
			wg.Add(1)
			go func(sr *SpeedResult) {
				sr.AverageElapsed = convertToTimeUnit(sr.AverageElapsed, timeUnit)
				sr.AverageUser = convertToTimeUnit(sr.AverageUser, timeUnit)
				sr.AverageSystem = convertToTimeUnit(sr.AverageSystem, timeUnit)
				sr.StandardDeviation = convertToTimeUnit(sr.StandardDeviation, timeUnit)
				sr.Max = convertToTimeUnit(sr.Max, timeUnit)
				sr.Min = convertToTimeUnit(sr.Min, timeUnit)
				for i, t := range sr.Times {
					sr.Times[i] = convertToTimeUnit(t, timeUnit)
				}
				wg.Done()
			}(sr)
		}
		wg.Wait()
	}
}
