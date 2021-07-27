package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
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

// updates bench.
func update() {
	log("yellow", "Updating bench...")
	log("yellow", "Downloading the bench executable...")

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
		log("red", "Your OS isnt supported by bench.")
		return
	}

	// * sending a request
	res, err := http.Get(url)

	if err != nil {
		log("red", "Error: Unable to download the executable. Check your internet connection.")
		log("white", err.Error())
		return
	}

	defer res.Body.Close()

	// * determining the executable path
	downloadPath, e := os.UserHomeDir()
	if e != nil {
		log("red", "Error: Unable to determine bench's location.")
		log("white", e.Error())
		return
	}
	downloadPath += "/.bench/bench"
	if runtime.GOOS == "windows" {
		downloadPath += ".exe"
	}

	os.Rename(downloadPath, downloadPath+"-old")

	exe, er := os.Create(downloadPath)
	if er != nil {
		log("red", "Error: Unable to access file permissions.")
		log("white", er.Error())
		return
	}
	defer exe.Close()

	// * writing the recieved content to the bench executable
	_, errr := io.Copy(exe, res.Body)
	if errr != nil {
		log("red", "Error: Unable to write the executable.")
		log("white", errr.Error())
		return
	}

	// * performing an additional `chmod` utility for linux and mac
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		execute("chmod", "u+x", downloadPath)
	}

	log("green", "Bench was updated successfully.")
}

// deletes previous installation if it exists.
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
