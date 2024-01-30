package internal

import (
	"bufio"
	"fmt"
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

// MapFunc returns a slice of all elements in the given slice mapped by the given function.
func MapFunc[Ts ~[]T, Ss ~[]S, T, S any](function func(T) S, slice Ts) Ss {
	mappedSlice := make(Ss, len(slice))
	for i, v := range slice {
		mappedSlice[i] = function(v)
	}
	return mappedSlice
}

// FilterFunc takes a predicate function and returns all the elements of the slice which return true for the function.
func FilterFunc[T any, Ts ~[]T](function func(T) bool, slice Ts) Ts {
	var filtered Ts
	for _, v := range slice {
		if function(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// ReduceFunc reduces the given slice to a single value by repeatedly applying the given function over the slice.
// func ReduceFunc[S ~[]T, T any, O any](function func(T, T) O, slice S, initial O) T {
// 	var accumulated O = initial
// 	for _, v := range slice {
// 		accumulated = function(accumulated, v)
// 	}
// 	return accumulated
// }

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

func addExtension(filename, ext string) string {
	if !strings.HasSuffix(filename, "."+ext) {
		return filename + "." + ext
	}
	return filename
}
