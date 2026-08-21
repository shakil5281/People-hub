<#
.SYNOPSIS
  Quick rebuild and publish PeopleHub (backend + frontend + gateway)
.DESCRIPTION
  Builds all components and restarts Windows services.
  Assumes services are already installed (WinSW + XML configured).
  Skips IIS config, firewall, WinSW download — pure build + restart.
#>

param(
  [switch]$SkipBackend,
  [switch]$SkipGateway,
  [switch]$SkipFrontend
)

$ErrorActionPreference = "Stop"
$ROOT = (Resolve-Path "$PSScriptRoot\..").Path
$DEPLOY = "$ROOT\deploy"
$BINARY = "$ROOT\peoplehub.exe"
$GATEWAY = "$ROOT\peoplehub-gateway.exe"

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  PeopleHub Build & Publish" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# --- 1. Build Backend ----------------------------------------------------
if (-not $SkipBackend) {
  Write-Host ">>> Building Backend (peoplehub.exe)..." -ForegroundColor Yellow
  Set-Location $ROOT
  go build -o $BINARY -ldflags="-s -w" ./cmd/server
  if ($LASTEXITCODE -ne 0) { Write-Error "Backend build failed"; exit 1 }
  $size = "{0:N0} KB" -f ((Get-Item $BINARY).Length / 1KB)
  Write-Host "  -> $BINARY ($size)" -ForegroundColor Green
} else {
  Write-Host ">>> Skipping Backend" -ForegroundColor DarkGray
}

# --- 2. Build Gateway ----------------------------------------------------
if (-not $SkipGateway) {
  Write-Host ">>> Building Gateway (peoplehub-gateway.exe)..." -ForegroundColor Yellow
  Set-Location $ROOT
  go build -o $GATEWAY -ldflags="-s -w" ./cmd/gateway
  if ($LASTEXITCODE -ne 0) { Write-Error "Gateway build failed"; exit 1 }
  $size = "{0:N0} KB" -f ((Get-Item $GATEWAY).Length / 1KB)
  Write-Host "  -> $GATEWAY ($size)" -ForegroundColor Green
} else {
  Write-Host ">>> Skipping Gateway" -ForegroundColor DarkGray
}

# --- 3. Build Frontend ---------------------------------------------------
if (-not $SkipFrontend) {
  Write-Host ">>> Stopping PeopleHubWeb service (free build lock)..." -ForegroundColor Yellow
  $webSvcExe = "$DEPLOY\PeopleHubWeb-service.exe"
  if (Test-Path $webSvcExe) {
    & $webSvcExe stop 2>$null
    Start-Sleep 2
  }

  Write-Host ">>> Building Frontend (Next.js)..." -ForegroundColor Yellow
  Set-Location "$ROOT\web"

  # Remove old standalone to avoid EBUSY
  if (Test-Path "$ROOT\web\build\standalone") {
    Remove-Item -LiteralPath "$ROOT\web\build\standalone" -Recurse -Force -ErrorAction SilentlyContinue
  }

  $env:NEXT_PUBLIC_BASE_PATH = "/people-hub"
  $env:NEXT_PUBLIC_API_URL = "http://localhost:8081/api/v1"

  yarn run build
  if ($LASTEXITCODE -ne 0) { Write-Error "Frontend build failed"; exit 1 }

  Write-Host ">>> Copying static files to standalone..." -ForegroundColor Yellow
  $standalone = "$ROOT\web\build\standalone"
  $staticDir = "$standalone\build\static"
  New-Item -ItemType Directory -Force -Path $staticDir | Out-Null
  Copy-Item "$ROOT\web\build\static\*" "$staticDir\" -Recurse -Force
  $chunks = (Get-ChildItem "$staticDir\chunks" -ErrorAction SilentlyContinue).Count
  Write-Host "  -> $chunks static chunks copied" -ForegroundColor Green
} else {
  Write-Host ">>> Skipping Frontend" -ForegroundColor DarkGray
}

# --- 4. Restart Services -------------------------------------------------
Write-Host ">>> Restarting services..." -ForegroundColor Yellow

$services = @(
  @{ Exe = "$DEPLOY\PeopleHubAPI-service.exe"; Name = "PeopleHubAPI" },
  @{ Exe = "$DEPLOY\PeopleHubGateway-service.exe"; Name = "PeopleHubGateway" }
)
if (-not $SkipFrontend) {
  $services += @{ Exe = "$DEPLOY\PeopleHubWeb-service.exe"; Name = "PeopleHubWeb" }
}

foreach ($svc in $services) {
  $exe = $svc.Exe
  $name = $svc.Name
  if (Test-Path $exe) {
    Write-Host "  Restarting $name..."
    & $exe restart 2>&1 | Out-Null
    Start-Sleep 3
  } else {
    Write-Warning "  Service exe not found: $exe"
  }
}

# --- 5. Verify -----------------------------------------------------------
Write-Host ">>> Verifying health check..." -ForegroundColor Yellow
Start-Sleep 2
try {
  $response = Invoke-WebRequest -Uri "http://localhost:8081/health" -UseBasicParsing -TimeoutSec 10
  if ($response.StatusCode -eq 200) {
    Write-Host "  Health check: OK ($($response.Content))" -ForegroundColor Green
  }
} catch {
  Write-Warning "  Health check failed. Services may still be starting."
}

# --- 6. Summary ----------------------------------------------------------
Write-Host ""
Write-Host "============================================" -ForegroundColor Green
Write-Host "  Build & Publish Complete!" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Green
Write-Host ""
Write-Host "  Services:"
Get-Service PeopleHubAPI, PeopleHubWeb, PeopleHubGateway -ErrorAction SilentlyContinue |
  Select-Object Name, Status | ForEach-Object {
    $icon = if ($_.Status -eq "Running") { "✅" } else { "❌" }
    Write-Host "    $icon $($_.Name)`: $($_.Status)"
  }
Write-Host ""
Write-Host "  Access: http://localhost:8081/people-hub/" -ForegroundColor Cyan
Write-Host "  API:    http://localhost:8081/api/v1" -ForegroundColor Cyan
Write-Host "  Health: http://localhost:8081/health" -ForegroundColor Cyan
Write-Host ""
