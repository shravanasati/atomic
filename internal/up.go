package internal

// todo replace self updater with check for updates

import (
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
)

// Update updates bench by downloading the latest executable from github, and renaming the
// old executable to `bench-old` so that it can be deleted by `DeletePreviousInstallation`.
func Update() {
	Log("yellow", "Updating bench...")
	Log("yellow", "Downloading the bench executable...")

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
		Log("red", "Your OS isn't supported by bench.")
		return
	}

	// * sending a request
	res, err := http.Get(url)

	if err != nil {
		Log("red", "Error: Unable to download the executable. Check your internet connection.")
		Log("white", err.Error())
		return
	}

	defer res.Body.Close()

	// * determining the executable path
	downloadPath, e := os.UserHomeDir()
	if e != nil {
		Log("red", "Error: Unable to determine bench's location.")
		Log("white", e.Error())
		return
	}
	downloadPath += "/.bench/bench"
	if runtime.GOOS == "windows" {
		downloadPath += ".exe"
	}

	os.Rename(downloadPath, downloadPath+"-old")

	exe, er := os.Create(downloadPath)
	if er != nil {
		Log("red", "Error: Unable to access file permissions.")
		Log("white", er.Error())
		return
	}
	defer exe.Close()

	// * writing the received content to the bench executable
	_, errr := io.Copy(exe, res.Body)
	if errr != nil {
		Log("red", "Error: Unable to write the executable.")
		Log("white", errr.Error())
		return
	}

	// * performing an additional `chmod` utility for linux and mac
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		os.Chmod(downloadPath, 0700)
	}

	Log("green", "Bench was updated successfully.")
}

// DeletePreviousInstallation deletes previous installation if it exists.
func DeletePreviousInstallation(wg *sync.WaitGroup) {
	benchDir, _ := os.UserHomeDir()
	benchDir += "/.bench"

	files, _ := os.ReadDir(benchDir)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), "-old") {
			// fmt.Println("found existsing installation")
			os.Remove(benchDir + "/" + f.Name())
		}
		// fmt.Println(f.Name())
	}
	wg.Done()
}
