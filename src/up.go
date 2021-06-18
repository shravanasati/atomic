package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"os/exec"
)

// execute executes a command in the shell.
func execute(base string, command ...string) error {
	cmd := exec.Command(base, command...)
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}

func update() {
	fmt.Println("Updating bench...")

	fmt.Println("Downloading the bench executable...")
	// * determining the os-specific url
	url := ""
	switch runtime.GOOS {
	case "windows":
		url = "https://github.com/Shravan-1908/bench/releases/latest/download/bench-windows-amd64.exe"
	case "linux":
		url = "https://github.com/Shravan-1908/bench/releases/latest/download/bench-linux-amd64"
	case "darwin":
		url = "https://github.com/Shravan-1908/bench/releases/latest/download/bench-darwin-amd64"
	default:
		fmt.Println("Your OS isnt supported by bench.")
		return
	}

	// * sending a request
	res, err := http.Get(url)
	
	if err != nil {
		fmt.Println("Error: Unable to download the executable. Check your internet connection.")
		fmt.Println(err)
		return
	}

	defer res.Body.Close()

	// * determining the executable path
	downloadPath, e := os.UserHomeDir()
	if e != nil {
		fmt.Println("Error: Unable to retrieve bench path.")
		fmt.Println(e)
		return
	}
	downloadPath += "/.bench/bench"
	if runtime.GOOS == "windows" {downloadPath += ".exe"}

	os.Rename(downloadPath, downloadPath + "-old")

	exe, er := os.Create(downloadPath)
	if er != nil {
		fmt.Println("Error: Unable to access file permissions.")
		fmt.Println(er)
		return
	}
	defer exe.Close()

	// * writing the recieved content to the bench executable
	_, errr := io.Copy(exe, res.Body)
	if errr != nil {
		fmt.Println("Error: Unable to write the executable.")
		fmt.Println(errr)
		return
	}

	// * performing an additional `chmod` utility for linux and mac 
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		execute("chmod", "u+x", downloadPath)
	}

	fmt.Println("Update completed!")
}

func deletePreviousInstallation() {
	benchDir, _ := os.UserHomeDir()
	
	benchDir += "/.bench"

	files, _ := ioutil.ReadDir(benchDir)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), "-old") {
			// fmt.Println("found existsing installation")
			os.Remove(benchDir + "/" + f.Name())
		}
		// fmt.Println(f.Name())
	}
}