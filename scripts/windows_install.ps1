Write-Host "Downloading bench..."

$url = "https://github.com/Shravan-1908/bench/releases/latest/download/bench-windows-amd64.exe"

$dir = $env:USERPROFILE + "\.bench"
$filepath = $env:USERPROFILE + "\.bench\bench.exe"

[System.IO.Directory]::CreateDirectory($dir)
(Invoke-WebRequest -Uri $url -OutFile $filepath)

Write-Host "Adding bench to PATH..."
[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::Machine) + ";"+$dir,
    [EnvironmentVariableTarget]::Machine)

Write-Host 'bench installation is successfull!'
Write-Host "You need to restart your shell to use bench."
