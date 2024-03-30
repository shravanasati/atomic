#!/bin/sh

# This script installs atomic.
#
# Quick install: `curl https://raw.githubusercontent.com/shravanasati/atomic/main/scripts/install.sh | bash`
#
# Acknowledgments:
#   - https://github.com/zyedidia/eget
#   - https://github.com/burntsushi/ripgrep

set -e -u

githubLatestTag() {
  finalUrl=$(curl "https://github.com/$1/releases/latest" -s -L -I -o /dev/null -w '%{url_effective}')
  printf "%s\n" "${finalUrl##*v}"
}

ensure() {
    if ! "$@"; then err "command failed: $*"; fi
}

platform=''
machine=$(uname -m)

if [ "${GETatomic_PLATFORM:-x}" != "x" ]; then
  platform="$GETatomic_PLATFORM"
else
  case "$(uname -s | tr '[:upper:]' '[:lower:]')" in
    "linux")
      case "$machine" in
        "arm64"* | "aarch64"* ) platform='linux_arm64' ;;
        *"86") platform='linux_386' ;;
        *"64") platform='linux_amd64' ;;
      esac
      ;;
    "darwin")
      case "$machine" in
        "arm64"* | "aarch64"* ) platform='darwin_arm64' ;;
        *"64") platform='darwin_amd64' ;;
      esac
      ;;
    "msys"*|"cygwin"*|"mingw"*|*"_nt"*|"win"*)
      case "$machine" in
        *"86") platform='windows_386' ;;
        *"64") platform='windows_amd64' ;;
        "arm64"* | "aarch64"* ) platform='windows_arm64' ;;
      esac
      ;;
  esac
fi

if [ "x$platform" = "x" ]; then
  cat << 'EOM'
/=====================================\\
|      COULD NOT DETECT PLATFORM      |
\\=====================================/
Uh oh! We couldn't automatically detect your operating system.
To continue with installation, please choose from one of the following values:
- linux_arm64
- linux_386
- linux_amd64
- darwin_amd64
- darwin_arm64
- windows_386
- windows_arm64
- windows_amd64
Export your selection as the GETatomic_PLATFORM environment variable, and then
re-run this script.
For example:
  $ export GETatomic_PLATFORM=linux_amd64
  $ curl https://raw.githubusercontent.com/shravanasati/atomic/main/scripts/install.sh | bash
EOM
  exit 1
else
  printf "Detected platform: %s\n" "$platform"
fi

TAG=$(githubLatestTag shravanasati/atomic)

if [ "x$platform" = "xwindows_amd64" ] || [ "x$platform" = "xwindows_386" ] || [ "x$platform" = "xwindows_arm64" ]; then
  extension='zip'
else
  extension='tar.gz'
fi

printf "Latest Version: %s\n" "$TAG"
printf "Downloading https://github.com/shravanasati/atomic/releases/download/v%s/atomic_%s.%s\n" "$TAG" "$platform" "$extension"

ensure curl -L "https://github.com/shravanasati/atomic/releases/download/v$TAG/atomic_$platform.$extension" > "atomic.$extension"

case "$extension" in
  "zip") ensure unzip -j "atomic.$extension" -d "./atomic" ;;
  "tar.gz") ensure tar -xvzf "atomic.$extension" "./atomic" ;;
esac

bin_dir="${HOME}/.local/bin"
ensure mkdir -p "${bin_dir}"

if [ -e "$bin_dir/atomic" ]; then
  echo "Existing atomic binary found at ${bin_dir}, removing it..."
  ensure rm "$bin_dir/atomic"
fi

ensure mv "./atomic" "${bin_dir}"
ensure chmod +x "${bin_dir}/atomic"

ensure rm "atomic.$extension"
ensure rm -rf "$platform"

echo 'atomic has been installed at' ${bin_dir}

if ! echo ":${PATH}:" | grep -Fq ":${bin_dir}:"; then
  echo "NOTE: ${bin_dir} is not on your \$PATH. atomic will not work unless it is added to \$PATH."
fi
