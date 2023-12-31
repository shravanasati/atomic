package internal

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"os/user"
	"path/filepath"
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

func checkPathExists(fp string) bool {
	_, e := os.Stat(fp)
	return !os.IsNotExist(e)
}

func getBenchDir() string {
	usr, e := user.Current()
	if e != nil {
		panic(e)
	}

	// * determining iris's directory
	dir := filepath.Join(usr.HomeDir, ".iris")

	if !checkPathExists(dir) {
		os.Mkdir(dir, os.ModePerm)
	}

	subDirs := []string{"wallpapers", "temp", "cache"}
	for _, subDir := range subDirs {
		dirPath := filepath.Join(dir, subDir)
		if !checkPathExists(dirPath) {
			os.Mkdir(dirPath, os.ModePerm)
		}
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
