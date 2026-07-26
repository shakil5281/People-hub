<#
.SYNOPSIS
  Build and deploy PeopleHub project on IIS + Windows Services
.DESCRIPTION
  This script builds the Go backend, Next.js frontend, and a Go reverse proxy gateway,
  then configures IIS, installs WinSW services, and starts everything.
  SAFETY: Does NOT stop/remove Default Web Site or existing IIS app pools on port 80.
.PARAMETER NoIIS
  Skip IIS configuration
.PARAMETER NoServices
  Skip WinSW service installation (build only)
.PARAMETER SkipBuild
  Skip build steps (deploy only)
#>

param(
  [switch]$NoIIS,
  [switch]$NoServices,
  [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"
$ROOT = (Resolve-Path "$PSScriptRoot\..").Path
$BINARY = "$ROOT\peoplehub.exe"
$GATEWAY = "$ROOT\peoplehub-gateway.exe"
$DEPLOY = "$ROOT\deploy"
$WWWROOT = "$ROOT\wwwroot"
$WINSW_URL = "https://github.com/winsw/winsw/releases/download/v2.12.0/WinSW-x64.exe"
$WINSW_EXE = "$DEPLOY\WinSW-x64.exe"

# --- 1. Prerequisites --------------------------------------------------------
Write-Host "=== Checking Prerequisites ===" -ForegroundColor Cyan

# Admin check
$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
  Write-Warning "Not running as Administrator. Some steps may fail (IIS, services, firewall)."
}

# Go
$goVer = go version 2>$null
if (-not $goVer) { Write-Error "Go not found. Install Go 1.26+"; return }
Write-Host "  Go: $goVer"

# Node
$nodeVer = node --version 2>$null
if (-not $nodeVer) { Write-Error "Node.js not found."; return }
Write-Host "  Node: $nodeVer"

# PostgreSQL
$psql = Get-Command psql -ErrorAction SilentlyContinue
if (-not $psql) { Write-Warning "  psql not in PATH. Ensure PostgreSQL is running." }

# IIS
$iis = Get-Service W3SVC -ErrorAction SilentlyContinue
if ($iis -and $iis.Status -eq "Running") {
  Write-Host "  IIS: Running (WARNING: Default Web Site on :80 will NOT be modified)"
} else {
  Write-Warning "  IIS W3SVC is not running."
}

# URL Rewrite module
$rewriteModule = Get-WebGlobalModule | Where-Object { $_.Name -eq "RewriteModule" } 2>$null
if ($rewriteModule) {
  Write-Host "  URL Rewrite Module: Installed"
} else {
  Write-Warning "  URL Rewrite Module not detected. Trying to install..."
  $rewriteUrl = "https://download.microsoft.com/download/1/2/8/128E2E22-C1B9-44A4-BE2A-5859ED1D4592/rewrite_amd64_en-US.msi"
  $rewriteOut = "$env:TEMP\rewrite_amd64_en-US.msi"
  try {
    Invoke-WebRequest -Uri $rewriteUrl -OutFile $rewriteOut -UseBasicParsing
    Start-Process msiexec.exe -ArgumentList "/i `"$rewriteOut`" /quiet /norestart" -Wait -NoNewWindow
    Write-Host "  URL Rewrite Module: Installed"
  } catch {
    Write-Warning "  Could not install URL Rewrite Module. Reverse proxy in web.config won't work."
  }
}

# --- 1.5 Clean Old Binaries --------------------------------------------------
Write-Host "`n=== Cleaning Old Binaries ===" -ForegroundColor Cyan
$oldBinaries = @(
    "$ROOT\hrhub.exe",
    "$ROOT\hub.exe",
    "$ROOT\hub-gateway.exe",
    "$ROOT\server.exe",
    "$ROOT\employee.exe",
    "$ROOT\reset.exe"
)
foreach ($f in $oldBinaries) {
    if (Test-Path $f) { Remove-Item $f -Force; Write-Host "  Removed old: $(Split-Path $f -Leaf)" }
}

# Clean old Hub-renamed services (HubAPI, HubWeb, HubGateway) if lingering
@("HubAPI", "HubWeb", "HubGateway") | ForEach-Object {
    $svc = Get-Service $_ -ErrorAction SilentlyContinue
    if ($svc) {
        Stop-Service $_ -Force -ErrorAction SilentlyContinue
        sc.exe delete $_ 2>$null
        Write-Host "  Removed old service: $_"
    }
}

# NOTE: We do NOT remove any IIS sites or app pools. Default Web Site on port 80
#       and all existing app pools (contactDB, ERPHubPool, HRMSAppPool, HubAppPool)
#       are left untouched for production coexistence.

# --- 2. Build ----------------------------------------------------------------
if (-not $SkipBuild) {
  Write-Host "`n=== Building Backend (peoplehub.exe) ===" -ForegroundColor Cyan
  Set-Location $ROOT
  go build -o $BINARY -ldflags="-s -w" ./cmd/server
  if ($LASTEXITCODE -ne 0) { Write-Error "Backend build failed"; return }
  Write-Host "  -> $BINARY"

  Write-Host "`n=== Building Gateway (peoplehub-gateway.exe) ===" -ForegroundColor Cyan
  go build -o $GATEWAY -ldflags="-s -w" ./cmd/gateway
  if ($LASTEXITCODE -ne 0) { Write-Error "Gateway build failed"; return }
  Write-Host "  -> $GATEWAY"

  Write-Host "`n=== Building Frontend (Next.js) ===" -ForegroundColor Cyan
  Set-Location "$ROOT\web"
  $env:NEXT_PUBLIC_BASE_PATH = "/people-hub"
  $env:NEXT_PUBLIC_API_URL = "http://localhost:8081/api/v1"
  npm install --silent 2>$null
  npm run build
  if ($LASTEXITCODE -ne 0) { Write-Error "Frontend build failed"; return }
  Write-Host "  -> web\build\standalone"

  # Copy standalone files correctly
  $standalone = "$ROOT\web\build\standalone"
  if (Test-Path "$standalone\web") {
    Write-Host "  Restructuring standalone output..."
    Copy-Item "$standalone\web\*" "$standalone\" -Recurse -Force
    Remove-Item "$standalone\web" -Recurse -Force -ErrorAction SilentlyContinue
  }
  # Copy public and static assets
  if (Test-Path "$ROOT\web\public") {
    $publicDir = "$standalone\public"
    New-Item -ItemType Directory -Force -Path $publicDir | Out-Null
    Copy-Item "$ROOT\web\public\*" "$publicDir\" -Recurse -Force -ErrorAction SilentlyContinue
  }
  if (Test-Path "$ROOT\web\build\static") {
    $staticDir = "$standalone\build\static"
    New-Item -ItemType Directory -Force -Path $staticDir | Out-Null
    Copy-Item "$ROOT\web\build\static\*" "$staticDir\" -Recurse -Force
  }

  Write-Host "`n=== Build Complete ===" -ForegroundColor Green
} else {
  Write-Host "`n=== Skipping Build ===" -ForegroundColor Yellow
}

# --- 3. WinSW Setup ---------------------------------------------------------
if (-not $NoServices) {
  Write-Host "`n=== Setting up Windows Services ===" -ForegroundColor Cyan

  # Download WinSW if not present
  if (-not (Test-Path $WINSW_EXE)) {
    Write-Host "  Downloading WinSW..."
    Invoke-WebRequest -Uri $WINSW_URL -OutFile $WINSW_EXE -UseBasicParsing
  }

  # Stop existing PeopleHub services if running (prep for reinstall)
  @("PeopleHubAPI", "PeopleHubWeb", "PeopleHubGateway") | ForEach-Object {
    $svc = Get-Service $_ -ErrorAction SilentlyContinue
    if ($svc -and $svc.Status -eq "Running") {
      Stop-Service $_ -Force
      Start-Sleep 1
    }
  }

  # Install services
  $services = @(
    @{ Name="PeopleHubAPI"; Xml="hub-service-backend.xml" },
    @{ Name="PeopleHubWeb"; Xml="hub-service-frontend.xml" },
    @{ Name="PeopleHubGateway"; Xml="hub-service-gateway.xml" }
  )

  foreach ($svc in $services) {
    $svcExe = "$DEPLOY\$($svc.Name)-service.exe"
    $svcXml = "$DEPLOY\$($svc.Xml)"

    # Remove old service if exists (clean install)
    $existing = Get-Service $svc.Name -ErrorAction SilentlyContinue
    if ($existing) {
      sc.exe delete $svc.Name 2>$null
      Start-Sleep 1
    }

    # Copy WinSW as the service exe
    Copy-Item $WINSW_EXE $svcExe -Force

    # Copy XML to match exe name
    $targetXml = "$DEPLOY\$($svc.Name)-service.xml"
    Copy-Item $svcXml $targetXml -Force

    # Replace %BASE% with actual path
    (Get-Content $targetXml) -replace '%BASE%', $ROOT.Replace('\','\\') | Set-Content $targetXml

    # Install
    Write-Host "  Installing $($svc.Name)..."
    & $svcExe install
    Start-Sleep 1
  }

  Write-Host "  Starting services..."
  Start-Service PeopleHubAPI
  Start-Sleep 3
  Start-Service PeopleHubWeb
  Start-Sleep 2
  Start-Service PeopleHubGateway

  Write-Host "  Services installed and started." -ForegroundColor Green
} else {
  Write-Host "`n=== Skipping Service Installation ===" -ForegroundColor Yellow
}

# --- 4. IIS Configuration (SAFE: Does NOT touch existing IIS sites/pools) ---
if (-not $NoIIS) {
  Write-Host "`n=== Configuring IIS ===" -ForegroundColor Cyan

  Import-Module WebAdministration -Force -ErrorAction SilentlyContinue

  # ===== CRITICAL: Default Web Site (port 80) is NOT modified =====
  # Other production apps run on it. PeopleHub uses its own port 8081 gateway.
  Write-Host "  Default Web Site on :80 preserved (other production apps)."
  Write-Host "  Existing App Pools preserved (contactDB, ERPHubPool, HRMSAppPool, HubAppPool)."

  # Create PeopleHub App Pool
  $appPoolName = "PeopleHubAppPool"
  $existingPool = Get-ChildItem "IIS:\AppPools\$appPoolName" -ErrorAction SilentlyContinue
  if (-not $existingPool) {
    New-Item "IIS:\AppPools\$appPoolName" -Force | Out-Null
    Write-Host "  App Pool '$appPoolName' created."
  } else {
    Write-Host "  App Pool '$appPoolName' already exists."
  }
  Set-ItemProperty "IIS:\AppPools\$appPoolName" -Name managedRuntimeVersion -Value ""
  Set-ItemProperty "IIS:\AppPools\$appPoolName" -Name startMode -Value "AlwaysRunning"

  # Create/Update PeopleHub IIS site on port 8083 (admin/status page only)
  # Main traffic routes through the Go gateway on :8081
  $siteName = "PeopleHubSite"
  $existingSite = Get-Website -Name $siteName -ErrorAction SilentlyContinue
  if (-not $existingSite) {
    New-Website -Name $siteName -PhysicalPath $WWWROOT -Port 8083 -ApplicationPool $appPoolName -Force
    Write-Host "  Site '$siteName' created on :8083 (admin/status)."
  } else {
    Set-ItemProperty "IIS:\Sites\$siteName" -Name physicalPath -Value $WWWROOT
    Set-ItemProperty "IIS:\Sites\$siteName" -Name applicationPool -Value $appPoolName
    Write-Host "  Site '$siteName' updated."
  }

  Start-Website -Name $siteName

  Write-Host "  IIS Configuration complete." -ForegroundColor Green
} else {
  Write-Host "`n=== Skipping IIS Configuration ===" -ForegroundColor Yellow
}

# --- 5. Firewall Rules ------------------------------------------------------
Write-Host "`n=== Configuring Firewall ===" -ForegroundColor Cyan
$fwRuleName = "PeopleHub-Gateway-8081"
$existingRule = netsh advfirewall firewall show rule name="$fwRuleName" 2>$null
if (-not $existingRule) {
  netsh advfirewall firewall add rule name="$fwRuleName" dir=in action=allow protocol=TCP localport=8081 profile=any
  Write-Host "  Firewall rule '$fwRuleName' created for port 8081."
} else {
  Write-Host "  Firewall rule '$fwRuleName' already exists."
}

# Remove old port-80 firewall rule if exists (no longer needed, port 80 is for Default Web Site)
netsh advfirewall firewall delete rule name="PeopleHub-Gateway-80" 2>$null

# --- 6. Summary ---------------------------------------------------------------
Write-Host "`n============================================" -ForegroundColor Green
Write-Host "  PeopleHub Deployment Complete!" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Green
Write-Host ""
Write-Host "  Frontend:  http://localhost:8081/people-hub (via gateway)"
Write-Host "  API:       http://localhost:8081/api/v1 (gateway → :5050)"
Write-Host "  Swagger:   http://localhost:8081/swagger/index.html"
Write-Host "  Health:    http://localhost:8081/health"
Write-Host ""
Write-Host "  IIS Admin Status: http://localhost:8083 (PeopleHub status page)"
Write-Host "  Default Web Site: http://localhost:80 (unchanged, other apps)"
Write-Host ""
Write-Host "  Services:"
Write-Host "    PeopleHubAPI     -> $ROOT\peoplehub.exe                (:5050)"
Write-Host "    PeopleHubWeb     -> Next.js on port 3050              (:3050)"
Write-Host "    PeopleHubGateway -> $ROOT\peoplehub-gateway.exe       (:8081)"
Write-Host ""
Write-Host "  Logs: $DEPLOY\PeopleHub*-service.*.log"
Write-Host "============================================" -ForegroundColor Green
