# python script to generate scoop schema
# must be ran after build.py

import json
from pathlib import Path
import hashlib

ATOMIC_BASE_URL = "https://github.com/shravanasati/atomic"



def hash_file(filename):
    h = hashlib.sha256()

    with open(filename, "rb") as file:
        chunk = 0
        while chunk != b"":
            chunk = file.read(1024)
            h.update(chunk)

    return h.hexdigest()


if __name__ == "__main__":
    schema = {
        "homepage": ATOMIC_BASE_URL,
        "version": "",
        "architecture": {"64bit": {}, "arm64": {}},
        "license": "MIT",
        "bin": "atomic.exe",
        "checkver": "github",
    }

    # read release config to obtain version and platforms
    project_base = Path(__file__).parent.parent
    release_config_file = project_base / "scripts" / "release.config.json"
    with open(str(release_config_file)) as f:
        release_config = json.load(f)

    # set version in schema
    schema["version"] = release_config["version"]

    # set architecture data in the manifest file
    for entry in schema["architecture"]:
        match entry:
            case "64bit":
                arch = "amd64"
            case "arm64":
                arch = "arm64"
            case _:
                raise Exception(f"Unkown architecture: {entry}")

        filename = f"atomic_windows_{arch}.zip"
        dist_file = project_base / "dist" / filename
        arch_data = {
            "url": f"{ATOMIC_BASE_URL}/releases/latest/download/{filename}",
            "hash": hash_file(str(dist_file))
        }

        schema["architecture"][entry] = arch_data

    # write the manifest file
    jsonfile_path = project_base / "scripts" / "atomic.json"
    with open(str(jsonfile_path), 'w') as f:
        f.write(json.dumps(schema, indent=2))

    print("Scoop app manifest file for atomic generated.")
