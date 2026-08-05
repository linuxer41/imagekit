param(
    [string]$Filter = ""
)

# Runs the image transform tests in a Docker container with libvips and
# writes the generated output PNGs to internal/transform/testdata/out/ so
# they can be inspected visually on the host.

# Note: do NOT use $ErrorActionPreference = "Stop" here. Under PowerShell 5.1
# a "Stop" preference turns docker's stderr progress output into a terminating
# error, so native commands are guarded with $LASTEXITCODE + throw instead.
$ErrorActionPreference = "Continue"

$root = Split-Path -Parent $PSScriptRoot
$outDir = Join-Path $root "internal\transform\testdata\out"

New-Item -ItemType Directory -Force -Path $outDir | Out-Null

Write-Host "Building test stage (golang:alpine + vips-dev)..." -ForegroundColor Cyan
docker build --target test -f "$root\Dockerfile.imager" -t imagekit-transform-test "$root"
if ($LASTEXITCODE -ne 0) {
    throw "docker build failed with exit code $LASTEXITCODE"
}

Write-Host "Running transform tests..." -ForegroundColor Cyan
$filterArg = if ($Filter) { "-test.run $Filter" } else { "" }
docker run --rm -v "${outDir}:/src/internal/transform/testdata/out" imagekit-transform-test sh -c "CGO_ENABLED=1 go test ./internal/transform/ -count=1 -v $filterArg"
if ($LASTEXITCODE -ne 0) {
    throw "tests failed with exit code $LASTEXITCODE"
}

Write-Host ""
Write-Host "Output images written to: $outDir" -ForegroundColor Green
