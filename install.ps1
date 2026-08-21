# HeraVision one-line installer (Windows)
# usage: irm https://raw.githubusercontent.com/ahmdd4vd/heravision/main/install.ps1 | iex
$ErrorActionPreference = "Stop"

$repo = "ahmdd4vd/heravision"
$rel = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$asset = $rel.assets | Where-Object {
  ($_.name -like "heravision_*_Windows_*x86_64.zip") -or ($_.name -like "heravision_*_Windows_*amd64.zip")
} | Select-Object -First 1
if (-not $asset) { throw "no matching Windows asset in the latest release" }

$tmp = Join-Path $env:TEMP ("heravision_" + [guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmp | Out-Null
$zip = Join-Path $tmp $asset.name
Write-Host "downloading $($asset.name)"
Invoke-WebRequest $asset.browser_download_url -OutFile $zip

$dir = Join-Path $env:LOCALAPPDATA "Programs\heravision"
New-Item -ItemType Directory -Force -Path $dir | Out-Null
Expand-Archive -Force $zip $dir
Remove-Item $tmp -Recurse -Force

$exe = Join-Path $dir "heravision.exe"
if (($env:Path -split ";") -notcontains $dir) {
  [Environment]::SetEnvironmentVariable("Path", "$dir;" + [Environment]::GetEnvironmentVariable("Path", "User"), "User")
  $env:Path = "$dir;$env:Path"
  Write-Host "added to user PATH: $dir"
}

Write-Host "installed: $exe"
& $exe setup
