Write-Host "Downloading atomic..."

$url = "https://github.com/shravanasati/atomic/releases/latest/download/atomic-windows-amd64.exe"

$dir = $env:USERPROFILE + "\.atomic"
$filepath = $env:USERPROFILE + "\.atomic\atomic.exe"

[System.IO.Directory]::CreateDirectory($dir)
(Invoke-WebRequest -Uri $url -OutFile $filepath)

Write-Host "Adding atomic to PATH..."
[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::Machine) + ";"+$dir,
    [EnvironmentVariableTarget]::Machine)

Write-Host 'atomic installation is successfull!'
Write-Host "You need to restart your shell to use atomic."
