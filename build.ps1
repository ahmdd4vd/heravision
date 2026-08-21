$ErrorActionPreference = "Stop"
$env:CGO_ENABLED = "0"
$version = "0.1.1"
go build -ldflags "-s -w -X heravision/internal/buildinfo.Version=$version" -o heravision.exe ./cmd/heravision
if ($LASTEXITCODE -eq 0) {
    Write-Host "[ok] built heravision.exe v$version"
    & .\heravision.exe version
}
