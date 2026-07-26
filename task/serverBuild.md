# PeopleHub — Server Build & Deployment Guide

> Complete guide to building, deploying, and understanding the PeopleHub system on IIS + Windows with Go reverse proxy.

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [System Requirements](#2-system-requirements)
3. [Project Structure](#3-project-structure)
4. [Coexisting with Existing IIS Servers](#4-coexisting-with-existing-iis-servers)
5. [Quick Deploy (One-Command)](#5-quick-deploy-one-command)
6. [Manual Step-by-Step Build](#6-manual-step-by-step-build)
7. [How It All Works](#7-how-it-all-works)
8. [Services Management](#8-services-management)
9. [IIS Configuration](#9-iis-configuration)
10. [Firewall & Security](#10-firewall--security)
11. [Troubleshooting](#11-troubleshooting)
12. [Re-Deploy & Updates](#12-re-deploy--updates)
13. [Key Files Reference](#13-key-files-reference)
14. [Rebuild After Code Changes](#14-rebuild-after-code-changes)
15. [Architecture Decision Record](#architecture-decision-record)

---

## 1. Architecture Overview

```
┌─ Internet/Network ──────────────────────────────────┐
│                                                      │
│   http://hub-server/              (IIS Default Site) │
│   http://hub-server:8081/         (PeopleHub entry)  │
│                                                      │
└────────────────────────┬─────────────────────────────┘
                         │
              ┌──────────┴──────────┐
              ▼                     ▼
   ┌── Port 80 ──────────┐  ┌── Port 8081 ──────────────────┐
   │ Default Web Site     │  │ PeopleHubGateway              │
   │ (EXISTING — DO NOT   │  │ peoplehub-gateway.exe         │
   │  MODIFY OR STOP!)    │  │ Go reverse proxy              │
   │ IIS serves other     │  │                               │
   │ production apps      │  │ ┌───────────┐ ┌────────────┐ │
   └──────────────────────┘  │ │ /people-hub │ │ /api/v1   | │
                              │ │ /swagger  │ │ /uploads   | │
                              │ │ /health   │ │            | │
                              │ └─────┬─────┘ └──────┬─────┘ │
                              └───────┼───────────────┼───────┘
                                       │               │
                                       ▼               ▼
                            ┌── Port 3050 ──┐  ┌── Port 5050 ──────────┐
                            │ PeopleHubWeb  │  │ PeopleHubAPI           │
                            │ Next.js       │  │ peoplehub.exe          │
                            │ standalone    │  │ Go + Gin + GORM        │
                            │ React +       │  │ JWT Auth               │
                            │ Tailwind      │  │ Swagger Docs           │
                            └───────────────┘  └───────────┬───────────┘
                                                          │
                                                          ▼
                                               ┌── Port 5432 ──────────┐
                                               │ PostgreSQL 16          │
                                               │ Database: peoplehub    │
                                               │ 40 tables, UUID PKs    │
                                               └────────────────────────┘
```

### Dual Port Scheme

PeopleHub supports two port configurations:

| Context | API Port | Frontend Port | Gateway Port |
|---------|----------|--------------|--------------|
| **Local development** (`npm run dev` / `go run`) | `:5000` | `:3000` | — |
| **IIS/Gateway deployment** (Windows Services) | `:5050` | `:3050` | `:8081` |

### Port Allocation Map (Deployment)

| Port | Service | Direction | Purpose | Status |
|------|---------|-----------|---------|--------|
| 80 | Default Web Site | Inbound | **EXISTING production apps** | **DO NOT TOUCH** |
| 8081 | PeopleHubGateway | Inbound (new) | PeopleHub main entry | Gateway reverse proxy |
| 3050 | PeopleHubWeb | 127.0.0.1 only | Next.js frontend (bound to 127.0.0.1) | Internal only — not accessible externally |
| 5050 | PeopleHubAPI | Local only | Go backend API | Internal service |
| 5432 | PostgreSQL | Local only | Database | Already running |
| 8082 | contact IIS site | Inbound | Existing contact app | **DO NOT TOUCH** |

### How requests flow

```
Browser → http://localhost:8081/people-hub
            ↓
          PeopleHubGateway (port 8081)
            ↓ (matches /people-hub → web)
           PeopleHubWeb (port 3050)
            ↓ (Next.js serves React app)
          HTML page + client-side JS

Browser → http://localhost:8081/api/v1/employees
            ↓
          PeopleHubGateway (port 8081)
            ↓ (matches /api/* → API)
           PeopleHubAPI (port 5050)
            ↓ (JWT auth middleware)
          Handler → Repository → PostgreSQL
            ↓
          JSON response → Gateway → Browser

Browser → http://localhost/           ← EXISTING Default Web Site (port 80)
          Handled by IIS, PeopleHub gateway does NOT touch this.
```

### How the Gateway Works (Why 3 Services Must Run)

The gateway is **only a reverse proxy** — it has no business logic and no static files of its own. Think of it as a receptionist:

```
Browser → :8081 (Gateway)      → :3050 (Frontend)  → serves HTML/CSS/JS pages
  "receptionist"                → :5050 (API)       → database, auth, salary logic
```

#### What Each Service Does

| Service | Port | Analogy | If stopped |
|---------|------|---------|------------|
| **Gateway** | 8081 | Receptionist | Nobody can enter the building (connection refused) |
| **Frontend** | 3050 | Showroom | Building is open but empty — users see blank page or loading errors |
| **API** | 5050 | Back-office workers | Building has a showroom but nobody to process orders — login fails, data won't load |

#### Can we reduce to 2 processes?

Yes — embed the Next.js standalone output into the Go gateway binary. The gateway would serve frontend files directly from disk instead of proxying to :3050. This eliminates the Node.js process and port 3050 entirely. Only 2 processes remain:

- Gateway (serves frontend static files + proxies `/api/*` to API)
- API on :5050 (unchanged)

The API **cannot be eliminated** — it is the brain of the system containing all business logic (database, JWT auth, salary calculation, attendance processing, leave management). The gateway is just a router to it.

---

## 2. System Requirements

### Software

| Component | Version | Notes |
|-----------|---------|-------|
| Windows | 10/11 or Server 2016+ | IIS + admin rights needed |
| PostgreSQL | 16+ | Database server |
| Go | 1.26+ | For building backend |
| Node.js | 20+ | For building frontend |
| IIS | 7+ | Already running (DO NOT reinstall) |
| URL Rewrite Module | 2.1 | Optional (not required by gateway) |

### Hardware (Minimum)

| Resource | Requirement |
|----------|-------------|
| CPU | 2 cores |
| RAM | 4 GB |
| Disk | 10 GB free |
| Network | 1 GbE |

---

## 3. Project Structure

```
G:\softwer\People-hub\
├── peoplehub.exe              # Go backend binary (built)
├── peoplehub-gateway.exe      # Go reverse proxy binary (built)
├── deploy/                    # Deployment files
│   ├── build-hub.ps1          # ★ ONE-CLICK DEPLOY SCRIPT
│   ├── WinSW-x64.exe          # Windows Service Wrapper
│   ├── hub-service-backend.xml     # PeopleHubAPI service config template
│   ├── hub-service-frontend.xml    # PeopleHubWeb service config template
│   ├── hub-service-gateway.xml     # PeopleHubGateway service config template
│   ├── PeopleHubAPI-service.xml    # Processed (paths resolved)
│   ├── PeopleHubWeb-service.xml    # Processed
│   └── PeopleHubGateway-service.xml # Processed
├── cmd/                       # Go source code
│   ├── server/main.go         # API server entry point
│   └── gateway/main.go        # Reverse proxy entry point
├── internal/                  # Go internal packages
│   ├── handlers/              # HTTP handlers (33 files)
│   ├── models/                # GORM models (40 files)
│   ├── repository/            # Data access layer (26 files)
│   ├── service/               # Business logic (7 files)
│   ├── routes/                # All route registrations
│   ├── middleware/             # Auth, CORS, logger, audit
│   ├── auth/                  # JWT + password utilities
│   ├── config/                # Env config loader
│   ├── database/              # GORM connection + migrations
│   ├── server/                # Dependency injection wiring
│   └── utils/                 # Pagination, date helpers
├── web/                       # Next.js frontend
│   ├── app/                   # App Router pages (65 routes)
│   ├── components/            # React components
│   │   ├── ui/                # shadcn/ui components
│   │   ├── table/             # DataTable, SimpleTable
│   │   ├── form/              # Form components
│   │   └── layout/            # Sidebar, header, nav
│   ├── lib/                   # API client, axios, utilities
│   ├── build/standalone/      # ★ Built output (frontend)
│   └── next.config.ts         # ★ output: "standalone", distDir: "build"
├── wwwroot/                   # IIS site root (admin/status page)
│   ├── index.html             # Status & links page
│   └── web.config             # IIS configuration
├── uploads/                   # File uploads directory
├── backups/                   # Database backups
├── docs/                      # Swagger generated files
├── task/                      # Documentation & reference
├── .env                       # Environment variables
└── .env.example               # Template for env vars
```

---

## 4. Coexisting with Existing IIS Servers

> **CRITICAL WARNING — READ BEFORE DEPLOYING**

This machine already has production IIS applications running. **You must NOT stop, modify, or remove them.**

### Existing IIS Resources (DO NOT TOUCH)

| Resource | Type | Status | Purpose |
|----------|------|--------|---------|
| `Default Web Site` | IIS Site (port 80) | **Running** | Main production site |
| `contactDB` | App Pool | Running | Contact application |
| `ERPHubPool` | App Pool | Running | ERP Hub application |
| `HRMSAppPool` | App Pool | Running | HRMS application |
| `HubAppPool` | App Pool | Running | Hub application pool |
| `.NET v2.0 / v4.5` | App Pool | Running | Classic ASP.NET apps |

### PeopleHub's Safe Deployment Rules

1. **Gateway runs on port 8081** — NOT port 80 (that's used by Default Web Site)
2. **IIS PeopleHubSite runs on its own port** — separate app pool, separate bindings
3. **Never call `Stop-Website "Default Web Site"`** in any script
4. **Never remove any existing App Pool** or IIS site
5. **Only create/update the PeopleHub-specific resources**: `PeopleHubAppPool`, `PeopleHubSite`

### Ports NOT to Use

| Port | Used By | Conflict Risk |
|------|---------|--------------|
| 80 | Default Web Site (IIS) | **BLOCKED** |
| 443 | Potential HTTPS | Avoid |
| 8080 | Common alt HTTP | Avoid |
| 8082 | contact app (IIS) | **BLOCKED** |

### PeopleHub Dedicated Ports

| Port | Used By | Safe? |
|------|---------|-------|
| **8081** | PeopleHubGateway (reverse proxy) | **YES** |
| 5000 | PeopleHubAPI (backend) | **YES** (localhost) |
| 3000 | PeopleHubWeb (frontend) | **YES** (localhost) |

---

## 5. Quick Deploy (One-Command)

### Prerequisites Check

Ensure these are installed and running:

```powershell
# Check Go
go version

# Check Node.js
node --version

# Check PostgreSQL (should be running)
psql --version

# Check IIS (should be running — DO NOT RESTART)
Get-Service W3SVC

# Verify IIS Default Web Site is running (leave it alone!)
Get-Website -Name "Default Web Site"
```

### Deploy

```powershell
# Open PowerShell as Administrator, then:
Set-Location G:\softwer\People-hub
.\deploy\build-hub.ps1
```

The script will:

1. Check prerequisites (Go, Node.js, IIS running)
2. Clean old binaries (`hrhub.exe`, `hub.exe`, `server.exe`, `employee.exe`, `reset.exe`)
3. **NOT touch existing IIS sites or app pools**
4. Build `peoplehub.exe` (Go backend)
5. Build `peoplehub-gateway.exe` (Go reverse proxy)
6. Build Next.js frontend (standalone output to `web\build\standalone`)
7. Download WinSW (if needed)
8. Install/update 3 Windows services (PeopleHubAPI, PeopleHubWeb, PeopleHubGateway)
9. Create/update IIS site `PeopleHubSite` on a dedicated port
10. Open firewall port 8081
11. Start all PeopleHub services
12. Display access URLs

### Access After Deploy

| URL | What | Port |
|-----|------|------|
| `http://localhost:8081/people-hub/` | **Frontend** (PeopleHub login) | 8081 |
| `http://localhost:8081/api/v1` | **API** endpoints | 8081 → 5000 |
| `http://localhost:8081/swagger/index.html` | **Swagger** API docs | 8081 → 5000 |
| `http://localhost:8081/health` | **Health** check | 8081 → 5000 |
| `http://localhost/` | **Existing Default Site** (unchanged) | 80 |

---

## 6. Manual Step-by-Step Build

If you prefer to build each component manually (or the script fails), follow these steps:

### Step 1: Build Go Backend

```powershell
Set-Location G:\softwer\People-hub
go build -o peoplehub.exe -ldflags="-s -w" ./cmd/server
```

This produces `peoplehub.exe` (~42 MB) — the API server.

### Step 2: Build Go Reverse Proxy

```powershell
go build -o peoplehub-gateway.exe -ldflags="-s -w" ./cmd/gateway
```

This produces `peoplehub-gateway.exe` (~6.4 MB) — the routing gateway.

### Step 3: Build Next.js Frontend

```powershell
Set-Location G:\softwer\People-hub\web

# Ensure standalone output is enabled in next.config.ts:
#   output: "standalone",
#   distDir: "build",

npm install
npm run build
```

The built output is at `web\build\standalone\`.

### Step 4: Prepare Standalone Folder

```powershell
$standalone = "G:\softwer\People-hub\web\build\standalone"

# Copy public assets
Copy-Item "G:\softwer\People-hub\web\public\*" "$standalone\public\" -Recurse -Force

# Copy static JS/CSS
New-Item -ItemType Directory -Force -Path "$standalone\build\static" | Out-Null
Copy-Item "G:\softwer\People-hub\web\build\static\*" "$standalone\build\static\" -Recurse -Force
```

### Step 5: Install Windows Services

```powershell
# Download WinSW
Invoke-WebRequest -Uri "https://github.com/winsw/winsw/releases/download/v3.1.0/WinSW-x64.exe" -OutFile "deploy\WinSW-x64.exe"

Set-Location G:\softwer\People-hub
$ROOT = "G:\softwer\People-hub"
$DEPLOY = "$ROOT\deploy"

# PeopleHubAPI (Go backend on port 5050)
Copy-Item "$DEPLOY\WinSW-x64.exe" "$DEPLOY\PeopleHubAPI-service.exe" -Force
$xml = (Get-Content "$DEPLOY\hub-service-backend.xml" -Raw) -replace '%BASE%', $ROOT
Set-Content "$DEPLOY\hub-service-backend.xml" $xml
Copy-Item "$DEPLOY\hub-service-backend.xml" "$DEPLOY\PeopleHubAPI-service.xml" -Force
& "$DEPLOY\PeopleHubAPI-service.exe" install

# PeopleHubWeb (Next.js on port 3050)
Copy-Item "$DEPLOY\WinSW-x64.exe" "$DEPLOY\PeopleHubWeb-service.exe" -Force
$xml = (Get-Content "$DEPLOY\hub-service-frontend.xml" -Raw) -replace '%BASE%', $ROOT
Set-Content "$DEPLOY\hub-service-frontend.xml" $xml
Copy-Item "$DEPLOY\hub-service-frontend.xml" "$DEPLOY\PeopleHubWeb-service.xml" -Force
& "$DEPLOY\PeopleHubWeb-service.exe" install

# PeopleHubGateway (reverse proxy on port 8081)
Copy-Item "$DEPLOY\WinSW-x64.exe" "$DEPLOY\PeopleHubGateway-service.exe" -Force
$xml = (Get-Content "$DEPLOY\hub-service-gateway.xml" -Raw) -replace '%BASE%', $ROOT
Set-Content "$DEPLOY\hub-service-gateway.xml" $xml
Copy-Item "$DEPLOY\hub-service-gateway.xml" "$DEPLOY\PeopleHubGateway-service.xml" -Force
& "$DEPLOY\PeopleHubGateway-service.exe" install
```

### Step 6: Configure IIS (Safe — Does Not Touch Existing Sites)

```powershell
Import-Module WebAdministration -Force

$appPoolName = "PeopleHubAppPool"
$existingPool = Get-ChildItem "IIS:\AppPools\$appPoolName" -ErrorAction SilentlyContinue
if (-not $existingPool) {
    New-Item "IIS:\AppPools\$appPoolName" -Force | Out-Null
}
Set-ItemProperty "IIS:\AppPools\$appPoolName" -Name managedRuntimeVersion -Value ""
Set-ItemProperty "IIS:\AppPools\$appPoolName" -Name startMode -Value "AlwaysRunning"

# Create PeopleHub IIS site on a dedicated port
$siteName = "PeopleHubSite"
$wwwrootPath = "G:\softwer\People-hub\wwwroot"
$existingSite = Get-Website -Name $siteName -ErrorAction SilentlyContinue
if (-not $existingSite) {
    New-Website -Name $siteName -PhysicalPath $wwwrootPath -Port 8083 -ApplicationPool $appPoolName -Force
    Write-Host "Created IIS site '$siteName' on port 8083."
}
Start-Website -Name $siteName

# NOTE: Default Web Site (port 80) is NOT touched. Other IIS app pools are NOT modified.
```

### Step 7: Configure Firewall

```powershell
netsh advfirewall firewall add rule name="PeopleHub-Gateway-8081" dir=in action=allow protocol=TCP localport=8081 profile=any
```

### Step 8: Start Services

```powershell
Start-Service PeopleHubAPI    # Waits for :5050 to be ready
Start-Sleep 3
Start-Service PeopleHubWeb    # Waits for :3050 to be ready
Start-Sleep 2
Start-Service PeopleHubGateway # Routes :8081 → :5050 and :3050
```

### Step 9: Verify

```powershell
# Verify PeopleHub (on dedicated port 8081)
Invoke-WebRequest -Uri "http://localhost:8081/health" -UseBasicParsing

# Verify IIS Default Web Site still works (not affected)
Invoke-WebRequest -Uri "http://localhost/" -UseBasicParsing
```

---

## 7. How It All Works

### 7.1 The Go Reverse Proxy (peoplehub-gateway.exe)

This is the **heart of the deployment**. It's a small Go program that:

- **Listens on port 8081** (NOT port 80 — that belongs to existing IIS Default Web Site)
- **Inspects every incoming URL path**
- **Routes based on prefix matching**:

| Path Prefix | Routes To | Target Port |
|-------------|-----------|-------------|
| `/api/` | Backend API | :5050 |
| `/swagger/` | Swagger docs | :5050 |
| `/uploads/` | Uploaded files | :5050 |
| `/health` | Health check | :5050 |
| `/people-hub/` | Frontend app | :3050 |
| Everything else | Frontend app | :3050 |

**Source:** `cmd/gateway/main.go` — uses Go's `httputil.NewSingleHostReverseProxy`.

**Why not IIS ARR?** The IIS Application Request Routing (ARR) module requires Windows Server and is not available on Windows 10. The Go reverse proxy is a lightweight (~6 MB binary) alternative that works on any Windows version. It also avoids touching the existing IIS configuration at all.

### 7.2 Go Backend API (peoplehub.exe)

**Port:** 5000  
**Stack:** Go 1.26 + Gin + GORM + PostgreSQL  
**Auth:** JWT (HS256, 15-min access + 7-day refresh)  
**Docs:** Swagger 2.0 at `/swagger/index.html`

**Key features:**
- 170+ API endpoints (employees, attendance, leave, salary, etc.)
- ZKTeco biometric integration (MDB reader via PowerShell, Windows-only)
- Monthly payroll calculation engine
- Leave management with balance tracking
- RBAC (Role-Based Access Control) with permissions

### 7.3 Next.js Frontend (PeopleHubWeb)

**Port:** 3000  
**Stack:** Next.js 16.2 + React 19.2 + Tailwind CSS 4 + shadcn/ui  
**Rendering:** Client-Side (all pages use `"use client"`)  
**Auth:** JWT stored in localStorage + Axios interceptor for auto-refresh

**Built as standalone output** (`output: "standalone"`, `distDir: "build"` in `next.config.ts`) so it runs via a plain Node.js server without requiring Next.js CLI.

### 7.4 IIS (PeopleHubSite)

**Purpose:** Admin/management interface only, not the main entry point. Main traffic goes through the Go gateway on port 8081.

The IIS site serves `wwwroot\index.html` — a status page with links to the PeopleHub URLs. The `web.config` configures custom headers and request filtering.

**Why keep IIS?** 
- Centralized management via IIS Manager
- SSL certificate management (future)
- Windows authentication integration (future)
- Monitoring and logging via IIS logs
- Existing IIS sites continue running undisturbed

### 7.5 Windows Services (WinSW)

Each component runs as a Windows service managed by **WinSW** (Windows Service Wrapper):

| Service Name | Executable | XML Config |
|-------------|------------|------------|
| `PeopleHubAPI` | `peoplehub.exe` | `deploy\hub-service-backend.xml` |
| `PeopleHubWeb` | `node server.js` | `deploy\hub-service-frontend.xml` |
| `PeopleHubGateway` | `peoplehub-gateway.exe` | `deploy\hub-service-gateway.xml` |

**Auto-restart on failure:** All services restart after 10 seconds if they crash.  
**Auto-start on boot:** All services set to `Automatic` start type.

### 7.6 Service Configuration Details

**PeopleHubAPI** (`deploy\hub-service-backend.xml`):
```xml
<service>
  <id>PeopleHubAPI</id>
  <name>PeopleHub API Server</name>
  <description>PeopleHub Go Backend API Server</description>
  <executable>G:\softwer\People-hub\peoplehub.exe</executable>
  <workingdirectory>G:\softwer\People-hub</workingdirectory>
  <env name="PORT" value="5050"/>
  <env name="DB_HOST" value="localhost"/>
  <env name="DB_PORT" value="5432"/>
  <env name="DB_USER" value="shakil"/>
  <env name="DB_PASS" value="123456"/>
  <env name="DB_NAME" value="peoplehub"/>
  <env name="DB_SSLMODE" value="disable"/>
  <env name="JWT_SECRET" value="peoplehub-secret-key-change-in-production-2025"/>
  <onfailure action="restart" delay="10 sec"/>
  <startmode>Automatic</startmode>
</service>
```

**PeopleHubWeb** (`deploy\hub-service-frontend.xml`):
```xml
<service>
  <id>PeopleHubWeb</id>
  <name>PeopleHub Web Frontend</name>
  <description>PeopleHub Next.js Frontend Server</description>
  <executable>node</executable>
  <arguments>G:\softwer\People-hub\web\build\standalone\server.js</arguments>
  <workingdirectory>G:\softwer\People-hub\web</workingdirectory>
  <env name="HOSTNAME" value="127.0.0.1"/>
  <env name="PORT" value="3050"/>
  <env name="NODE_ENV" value="production"/>
  <env name="NEXT_PUBLIC_API_URL" value="http://localhost:8081/api/v1"/>
  <env name="NEXT_PUBLIC_BASE_PATH" value="/people-hub"/>
  <onfailure action="restart" delay="10 sec"/>
  <startmode>Automatic</startmode>
</service>
```

**PeopleHubGateway** (`deploy\hub-service-gateway.xml`):
```xml
<service>
  <id>PeopleHubGateway</id>
  <name>PeopleHub Gateway (Reverse Proxy)</name>
  <description>Reverse proxy routing PeopleHub traffic on port 8081</description>
  <executable>G:\softwer\People-hub\peoplehub-gateway.exe</executable>
  <workingdirectory>G:\softwer\People-hub</workingdirectory>
  <env name="GATEWAY_PORT" value="8081"/>
  <env name="API_TARGET" value="http://localhost:5050"/>
  <env name="WEB_TARGET" value="http://localhost:3050"/>
  <onfailure action="restart" delay="10 sec"/>
  <startmode>Automatic</startmode>
</service>
```

---

## 8. Services Management

### Start/Stop/Restart

```powershell
# All PeopleHub services (does NOT affect IIS or other services)
Start-Service PeopleHubAPI; Start-Service PeopleHubWeb; Start-Service PeopleHubGateway
Stop-Service PeopleHubAPI, PeopleHubWeb, PeopleHubGateway
Restart-Service PeopleHubAPI, PeopleHubWeb, PeopleHubGateway

# Individual
Start-Service PeopleHubAPI
Stop-Service PeopleHubWeb
Restart-Service PeopleHubGateway
```

### Check Status

```powershell
Get-Service PeopleHubAPI, PeopleHubWeb, PeopleHubGateway | Select-Object Name, Status, StartType

# Also check IIS is still running (should always be)
Get-Service W3SVC | Select-Object Name, Status
```

### View Logs

```powershell
# Service logs are in deploy/ folder
Get-Content "G:\softwer\People-hub\deploy\PeopleHubAPI-service.out.log" -Tail 20
Get-Content "G:\softwer\People-hub\deploy\PeopleHubWeb-service.out.log" -Tail 20
Get-Content "G:\softwer\People-hub\deploy\PeopleHubGateway-service.out.log" -Tail 20

# Wrapper logs
Get-Content "G:\softwer\People-hub\deploy\PeopleHubAPI-service.wrapper.log" -Tail 20
```

### Uninstall Services

```powershell
Stop-Service PeopleHubAPI, PeopleHubWeb, PeopleHubGateway
& "G:\softwer\People-hub\deploy\PeopleHubAPI-service.exe" uninstall
& "G:\softwer\People-hub\deploy\PeopleHubWeb-service.exe" uninstall
& "G:\softwer\People-hub\deploy\PeopleHubGateway-service.exe" uninstall
```

---

## 9. IIS Configuration

### 9.1 IIS Site Details (PeopleHub)

| Setting | Value |
|---------|-------|
| Site Name | `PeopleHubSite` |
| App Pool | `PeopleHubAppPool` (No Managed Code) |
| Port | 8083 (dedicated admin/status page) |
| Physical Path | `G:\softwer\People-hub\wwwroot` |
| Protocol | HTTP |

**Note:** This IIS site serves only the admin status page. All PeopleHub application traffic goes through the Go gateway on port 8081.

### 9.2 Existing IIS Sites (DO NOT TOUCH)

| Site Name | Port | Status | Owner |
|-----------|------|--------|-------|
| `Default Web Site` | 80 | **Running** | Main production apps |

### 9.3 web.config

```xml
<?xml version="1.0" encoding="UTF-8"?>
<configuration>
  <system.webServer>
    <httpProtocol>
      <customHeaders>
        <add name="X-Frame-Options" value="SAMEORIGIN" />
        <add name="X-Content-Type-Options" value="nosniff" />
      </customHeaders>
    </httpProtocol>
    <security>
      <requestFiltering>
        <requestLimits maxAllowedContentLength="52428800" />
      </requestFiltering>
    </security>
  </system.webServer>
</configuration>
```

### 9.4 Adding SSL/HTTPS (Future)

In IIS Manager:
1. Select PeopleHubSite → Bindings → Add
2. Type: `https`, Port: `8443`
3. Select an SSL certificate
4. Update `NEXT_PUBLIC_API_URL` in `.env` and PeopleHubWeb config to `https://`

### 9.5 Important: Never Disable or Stop Default Web Site

The Default Web Site on port 80 serves existing production applications. The build script and all deployment procedures are designed to coexist with it:

- Gateway uses port 8081 (not 80)
- No script calls `Stop-Website "Default Web Site"`
- No script removes existing App Pools

---

## 10. Firewall & Security

### Ports Used

| Port | Service | Direction | Purpose | Can Change? |
|------|---------|-----------|---------|-------------|
| 80 | Default IIS | Inbound | Existing production | **NO** |
| 8081 | PeopleHubGateway | Inbound | PeopleHub main entry | Yes (env: GATEWAY_PORT) |
| 8082 | contact (IIS) | Inbound | Existing contact app | **NO** |
| 8083 | PeopleHub IIS | Inbound (optional) | Admin status page | Yes |
| 5050 | PeopleHubAPI (deploy) | Local only | Backend API (service mode) | Yes (env: PORT) |
| 3050 | PeopleHubWeb (deploy) | 127.0.0.1 only | Next.js frontend (service mode, bound to 127.0.0.1) | Yes (env: PORT, HOSTNAME) |
| 5432 | PostgreSQL | Local only | Database | No |

### Local Development Ports

For local development (running directly, not through gateway):

| Port | Service | Purpose |
|------|---------|---------|
| 5000 | PeopleHubAPI | Backend API (`go run` or `.\peoplehub.exe`) |
| 3000 | PeopleHubWeb | Next.js frontend (`npm run dev`) |

### Firewall Rules

```powershell
# List existing PeopleHub rule
netsh advfirewall firewall show rule name="PeopleHub-Gateway-8081"

# Remove if needed
netsh advfirewall firewall delete rule name="PeopleHub-Gateway-8081"

# Re-create
netsh advfirewall firewall add rule name="PeopleHub-Gateway-8081" dir=in action=allow protocol=TCP localport=8081 profile=any
```

### Environment Variables (`.env`)

```
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=shakil
DB_PASS=123456
DB_NAME=peoplehub
DB_SSLMODE=disable

# API Server
PORT=5000
JWT_SECRET=peoplehub-secret-key-change-in-production-2025

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:5000/api/v1
NEXT_PUBLIC_BASE_PATH=/people-hub
```

**Security Note:** Change the `JWT_SECRET` to a strong, random value before production use. Keep `.env` out of version control (it's in `.gitignore`).

---

## 11. Troubleshooting

### 11.1 Service Won't Start

```powershell
# Check Windows Event Log
Get-WinEvent -LogName "Application" -MaxEvents 10 | Where-Object { $_.LevelDisplayName -eq "Error" } | Format-Table TimeCreated, Message -Wrap

# Check service logs
Get-Content "G:\softwer\People-hub\deploy\PeopleHubAPI-service.out.log" -Tail 30
Get-Content "G:\softwer\People-hub\deploy\PeopleHubAPI-service.err.log" -Tail 30
```

### 11.2 Port Already in Use

```powershell
# Find what's using port 8081
netstat -ano | Select-String ":8081 "

# Stop the process (if it's a PeopleHub process)
Stop-Process -Id <PID> -Force

# Or change gateway port to another unused port
# Edit G:\softwer\People-hub\deploy\hub-service-gateway.xml
# Change GATEWAY_PORT from 8081 to another port (e.g., 8084)
```

### 11.3 Database Connection Error

```powershell
# Test PostgreSQL
psql -U shakil -d peoplehub -c "SELECT 1"

# Check if PostgreSQL service is running
Get-Service postgresql*

# Verify .env credentials match
Get-Content "G:\softwer\People-hub\.env"
```

### 11.4 Frontend Not Loading

```powershell
# Check if the standalone server exists
Test-Path "G:\softwer\People-hub\web\build\standalone\server.js"

# Check if the standalone has static assets
Test-Path "G:\softwer\People-hub\web\build\standalone\build\static"

# Rebuild if missing
Set-Location G:\softwer\People-hub\web
npm run build
```

### 11.5 Gateway Not Routing

```powershell
# Check if backend is reachable (service mode: :5050, local dev: :5000)
Invoke-WebRequest -Uri "http://localhost:5050/health" -UseBasicParsing

# Check if frontend is reachable (service mode: :3050, local dev: :3000)
Invoke-WebRequest -Uri "http://localhost:3050/" -UseBasicParsing

# Test gateway routing
Invoke-WebRequest -Uri "http://localhost:8081/health" -UseBasicParsing
Invoke-WebRequest -Uri "http://localhost:8081/people-hub/" -UseBasicParsing
```

### 11.6 IIS Default Web Site Broken (EMERGENCY)

If the PeopleHub deployment accidentally affected the Default Web Site:

```powershell
# Restart IIS
iisreset /restart

# Verify Default Web Site is running
Get-Website -Name "Default Web Site" | Select-Object Name, State
```

---

## 12. Re-Deploy & Updates

### Full Re-Deploy (Build + Services + IIS)

```powershell
Set-Location G:\softwer\People-hub
.\deploy\build-hub.ps1
```

This runs the complete pipeline: build all binaries → install services → configure IIS → start everything. **It does NOT stop or modify the Default Web Site or any existing IIS app pools.**

### Update Backend Only

```powershell
Set-Location G:\softwer\People-hub
go build -o peoplehub.exe -ldflags="-s -w" ./cmd/server
Stop-Service PeopleHubAPI
Start-Sleep 2
Start-Service PeopleHubAPI
```

### Update Frontend Only

```powershell
Set-Location G:\softwer\People-hub\web
npm install
npm run build

# Copy assets to standalone
$standalone = "G:\softwer\People-hub\web\build\standalone"
Copy-Item "public\*" "$standalone\public\" -Recurse -Force
Copy-Item "build\static\*" "$standalone\build\static\" -Recurse -Force

Stop-Service PeopleHubWeb
Start-Sleep 2
Start-Service PeopleHubWeb
```

### Update Gateway Only

```powershell
Set-Location G:\softwer\People-hub
go build -o peoplehub-gateway.exe -ldflags="-s -w" ./cmd/gateway
Stop-Service PeopleHubGateway
Start-Sleep 2
Start-Service PeopleHubGateway
```

### Database Migrations

The Go backend automatically runs GORM AutoMigrate on startup. To run migrations:

```powershell
# Simply restart PeopleHubAPI
Restart-Service PeopleHubAPI

# To seed data
go run cmd/superadmin/main.go
go run cmd/seed/main.go
go run cmd/seed/organization/main.go
go run cmd/seed/leave/main.go
go run cmd/seed/employee/main.go
```

---

## 13. Key Files Reference

| File | Absolute Path |
|------|---------------|
| Backend binary | `G:\softwer\People-hub\peoplehub.exe` |
| Gateway binary | `G:\softwer\People-hub\peoplehub-gateway.exe` |
| Frontend standalone | `G:\softwer\People-hub\web\build\standalone\server.js` |
| Frontend config | `G:\softwer\People-hub\web\next.config.ts` |
| Environment | `G:\softwer\People-hub\.env` |
| IIS web.config | `G:\softwer\People-hub\wwwroot\web.config` |
| IIS status page | `G:\softwer\People-hub\wwwroot\index.html` |
| Gateway source | `G:\softwer\People-hub\cmd\gateway\main.go` |
| Backend entry | `G:\softwer\People-hub\cmd\server\main.go` |
| Service templates | `G:\softwer\People-hub\deploy\hub-service-*.xml` |
| Deploy script | `G:\softwer\People-hub\deploy\build-hub.ps1` |
| WinSW executable | `G:\softwer\People-hub\deploy\WinSW-x64.exe` |
| Service logs | `G:\softwer\People-hub\deploy\PeopleHub*-service.*.log` |
| Uploaded files | `G:\softwer\People-hub\uploads\` |
| Database backups | `G:\softwer\People-hub\backups\` |

---

## Architecture Decision Record

| Decision | Rationale |
|----------|-----------|
| **Gateway on port 8081, not 80** | Port 80 is already in use by Default Web Site for existing production apps. Using 8081 avoids conflicts entirely. |
| **Go reverse proxy instead of IIS ARR** | ARR 3.0 is incompatible with Windows 10. Go proxy is lightweight, cross-platform, and reliable. |
| **WinSW for services** | Native Windows services with auto-restart and logging. No extra dependencies. |
| **Next.js standalone output** | Allows frontend to run as a service without Next.js CLI. Enables production `node server.js` mode. |
| **IIS on dedicated port for admin** | Keeps IIS available for management while gateway handles main traffic on 8081. Does not interfere with existing IIS sites. |
| **Two separate Go binaries** | Clear separation of concerns. Gateway handles routing; backend handles business logic. Each can be updated independently. |
| **Never stop Default Web Site** | Critical safety rule. Existing production applications must continue running undisturbed. All scripts are designed for coexistence. |

---

## 14. Rebuild After Code Changes

When you modify the source code and need to rebuild the running services, follow these steps based on what changed.

### Quick Reference

| What changed | Build command | Restart service |
|-------------|---------------|-----------------|
| **Backend Go code** (`internal/handlers/`, `internal/repository/`, `internal/models/`, `internal/service/`) | `go build -o peoplehub.exe ./cmd/server` | Kill old `peoplehub.exe`, start new one |
| **Gateway code** (`cmd/gateway/main.go`) | `go build -o peoplehub-gateway.exe ./cmd/gateway` | Kill old `peoplehub-gateway.exe`, start new one |
| **Frontend code** (`web/` pages, components, api.ts) | See [Full Frontend Rebuild](#full-frontend-rebuild) below | Kill old `node server.js`, start new one |
| **Everything** | Run `.\deploy\build-hub.ps1 -SkipIIS -NoServices` (build only) | Then restart manually |

### 14.1 Rebuild Backend (Go API)

```powershell
Set-Location G:\softwer\People-hub
# Build new binary
go build -o peoplehub.exe -ldflags="-s -w" ./cmd/server

# Kill old API process
Get-Process peoplehub -ErrorAction SilentlyContinue | ForEach-Object { $_.Kill() }
Start-Sleep 2

# Start new API on port 5050
$pinfo = New-Object System.Diagnostics.ProcessStartInfo
$pinfo.FileName = "G:\softwer\People-hub\peoplehub.exe"
$pinfo.WorkingDirectory = "G:\softwer\People-hub"
$pinfo.UseShellExecute = $false
$pinfo.EnvironmentVariables["PORT"] = "5050"
$pinfo.EnvironmentVariables["DB_HOST"] = "localhost"
$pinfo.EnvironmentVariables["DB_PORT"] = "5432"
$pinfo.EnvironmentVariables["DB_USER"] = "shakil"
$pinfo.EnvironmentVariables["DB_PASS"] = "123456"
$pinfo.EnvironmentVariables["DB_NAME"] = "hrhub"
$pinfo.EnvironmentVariables["DB_SSLMODE"] = "disable"
$pinfo.EnvironmentVariables["JWT_SECRET"] = "peoplehub-secret-key-change-in-production-2025"
[System.Diagnostics.Process]::Start($pinfo)
Start-Sleep 3

# Verify
Invoke-WebRequest -Uri "http://localhost:5050/health" -UseBasicParsing
```

### 14.2 Rebuild Gateway

```powershell
Set-Location G:\softwer\People-hub
# Build new gateway
go build -o peoplehub-gateway.exe -ldflags="-s -w" ./cmd/gateway

# Kill old gateway
Get-Process peoplehub-gateway -ErrorAction SilentlyContinue | ForEach-Object { $_.Kill() }
Start-Sleep 2

# Start new gateway on port 8081
$pinfo = New-Object System.Diagnostics.ProcessStartInfo
$pinfo.FileName = "G:\softwer\People-hub\peoplehub-gateway.exe"
$pinfo.WorkingDirectory = "G:\softwer\People-hub"
$pinfo.UseShellExecute = $false
$pinfo.EnvironmentVariables["GATEWAY_PORT"] = "8081"
$pinfo.EnvironmentVariables["API_TARGET"] = "http://localhost:5050"
$pinfo.EnvironmentVariables["WEB_TARGET"] = "http://localhost:3050"
[System.Diagnostics.Process]::Start($pinfo)
Start-Sleep 2

# Verify
Invoke-WebRequest -Uri "http://localhost:8081/health" -UseBasicParsing
```

### 14.3 Full Frontend Rebuild

The frontend requires extra steps because `NEXT_PUBLIC_*` env vars are baked into the JavaScript at **build time**. The `.env` file is only for local development — for deployment builds you must set the env vars manually.

```powershell
# 1. Set required env vars (MUST be set before npm run build)
$env:NEXT_PUBLIC_BASE_PATH = "/people-hub"
$env:NEXT_PUBLIC_API_URL = "http://localhost:8081/api/v1"

# 2. Build
Set-Location G:\softwer\People-hub\web
npm run build

# 3. Copy static assets to standalone folder (CRITICAL — Next.js 16 doesn't auto-copy these)
$standalone = "G:\softwer\People-hub\web\build\standalone"
$staticDir = "$standalone\build\static"
New-Item -ItemType Directory -Force -Path $staticDir | Out-Null
Copy-Item "G:\softwer\People-hub\web\build\static\*" "$staticDir\" -Recurse -Force

# 4. Kill old frontend process
Get-Process node -ErrorAction SilentlyContinue | ForEach-Object { $_.Kill() }
Start-Sleep 2

# 5. Start new frontend on port 3050 (bound to 127.0.0.1 only — not externally accessible)
$pinfo = New-Object System.Diagnostics.ProcessStartInfo
$pinfo.FileName = "node"
$pinfo.Arguments = "$standalone\server.js"
$pinfo.WorkingDirectory = "G:\softwer\People-hub\web"
$pinfo.UseShellExecute = $false
$pinfo.EnvironmentVariables["PORT"] = "3050"
$pinfo.EnvironmentVariables["HOSTNAME"] = "127.0.0.1"
$pinfo.EnvironmentVariables["NODE_ENV"] = "production"
[System.Diagnostics.Process]::Start($pinfo)
Start-Sleep 4

# 6. Verify
Invoke-WebRequest -Uri "http://127.0.0.1:3050/people-hub" -UseBasicParsing
Invoke-WebRequest -Uri "http://localhost:8081/people-hub" -UseBasicParsing
```

### 14.4 Common Rebuild Issues

| Issue | Cause | Fix |
|-------|-------|-----|
| `EBUSY: resource busy` | Old process holds lock on `build/` directory | Kill node process first with `Get-Process node \| % { $_.Kill() }` |
| `Failed to load chunk` | Static files not in `standalone/build/static/` | Rerun the static copy step (14.3 step 3) |
| `listen tcp :5050: bind: address already in use` | Old API process still running | Kill it: `Get-Process peoplehub \| % { $_.Kill() }` |
| Chunk 404 in browser | Frontend rebuilt but browser cached old chunks | Hard refresh (Ctrl+F5) or clear browser cache |
| `cannot find module` | Missing `npm install` | Run `npm install` before `npm run build` |

---

*Document Version: 2.0*  
*Last Updated: 2026-07-26*  
*Project: PeopleHub HR System*  
*Workspace: G:\softwer\People-hub*
