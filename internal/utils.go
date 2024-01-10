package internal

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

// formats the text in a javascript like syntax.
func format(text string, params map[string]string) string {
	for key, val := range params {
		text = strings.Replace(text, fmt.Sprintf("${%v}", key), val, -1)
	}
	return text
}

// MapFunc returns a slice of all elements in the given slice mapped by the given function.
func MapFunc[T any, S any](function func(T) S, slice []T) []S {
	mappedSlice := make([]S, len(slice))
	for i, v := range slice {
		mappedSlice[i] = function(v)
	}
	return mappedSlice
}

// FilterFunc takes a predicate function and returns all the elements of the slice which return true for the function.
func FilterFunc[T any](function func(T) bool, slice []T) []T {
	var filtered []T
	for _, v := range slice {
		if function(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
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

func roundFloat(num float64, digits int) float64 {
	tenMultiplier := math.Pow10(digits)
	return math.Round(num*tenMultiplier) / tenMultiplier
}

func checkPathExists(fp string) bool {
	_, e := os.Stat(fp)
	return !os.IsNotExist(e)
}

func getBenchDir() string {
	usr, e := user.Current()
	if e != nil {
		panic(e)
	}

	// * determining atomic's directory
	dir := filepath.Join(usr.HomeDir, ".atomic")

	if !checkPathExists(dir) {
		os.Mkdir(dir, os.ModePerm)
	}

	return dir
}

// readFile reads the given file and returns the string content of the same.
func readFile(file string) string {
	f, ferr := os.Open(file)
	if ferr != nil {
		panic(ferr)
	}
	defer f.Close()

	text := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text += scanner.Text()
	}

	return text
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
