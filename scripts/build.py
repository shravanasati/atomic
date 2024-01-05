import subprocess
import os
from multiprocessing import Process
from typing import List

def build(appname:str, platform: str) -> None:
	try:
		goos = platform.split("/")[0]
		goarch = platform.split("/")[1]

		print(f"==> 🚧 Building executable for `{platform}`...")
		os.environ["GOOS"] = goos
		os.environ["GOARCH"] = goarch

		outpath = f"./bin/{appname}-{goos}-{goarch}"

		if goos == "windows":
			outpath += ".exe"

		subprocess.check_output(["go", "build", "-v", "-o", outpath])

		print(f"==> ✅ Built executable for `{platform}` at `{outpath}`.")

	except Exception as e:
		print(e)
		print("==> ❌ An error occured! Aborting script execution.")
		os._exit(1)

if __name__ == "__main__":
	# add all platforms to the tuple you want to build
	platforms = {"windows/amd64", "linux/amd64", "darwin/amd64"}
	appname = "atomic" # name of the executable
	multithreaded = True # set to True to enable multithreading

	if multithreaded:
		threads: List[Process] = []

		for p in platforms:
			threads.append(Process(target=build, args=(appname, p)))

		for t in threads:
			t.start()

		for t in threads:
			t.join()

	else:
		for p in platforms:
			build(appname, p)

	print(f"==> 👍 Executables for {len(platforms)} platforms built successfully!")
