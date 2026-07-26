# Hub HR System — Server Build & Deployment Guide

> Complete guide to building, deploying, and understanding the Hub HR system on IIS + Windows.

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [System Requirements](#2-system-requirements)
3. [Project Structure](#3-project-structure)
4. [Quick Deploy (One-Command)](#4-quick-deploy-one-command)
5. [Manual Step-by-Step Build](#5-manual-step-by-step-build)
6. [How It All Works](#6-how-it-all-works)
7. [Services Management](#7-services-management)
8. [IIS Configuration](#8-iis-configuration)
9. [Firewall & Security](#9-firewall--security)
10. [Troubleshooting](#10-troubleshooting)
11. [Re-Deploy & Updates](#11-re-deploy--updates)

---

## 1. Architecture Overview

```
┌─ Internet/Network ──────────────────────────────────┐
│                                                      │
│   http://hub-server:80  (main entry)                 │
│   http://hub-server:8081 (IIS admin)                 │
│                                                      │
└────────────────────────┬─────────────────────────────┘
                         │
                         ▼
┌── HubGateway (port 80) ────────────────────────────┐
│  Go reverse proxy  │  hub-gateway.exe              │
│                                                     │
│  ┌─────────────┐   ┌──────────────────────────┐   │
│  │ /api/*      │   │  /* (all other routes)    │   │
│  │ /swagger/*  │   │                          │   │
│  │ /uploads/*  │   │                          │   │
│  │ /health     │   │                          │   │
│  └──────┬──────┘   └──────────┬───────────────┘   │
└─────────┼─────────────────────┼────────────────────┘
          │                     │
          ▼                     ▼
┌── HubAPI ─────────┐  ┌── HubWeb ───────────────────┐
│  Port 5000        │  │  Port 3000                  │
│  hub.exe          │  │  Next.js standalone server  │
│  Go + Gin + GORM  │  │  Runs from:                 │
│  PostgreSQL       │  │  web\.next\standalone\      │
│  JWT Auth         │  │  Client-side rendered       │
│  Swagger Docs     │  │  React + Tailwind + shadcn  │
└───────────────────┘  └─────────────────────────────┘

┌── IIS HubSite (port 8081) ─────────────────────────┐
│  Admin/Status page → wwwroot\index.html            │
│  Serves: static content only                        │
│  App Pool: HubAppPool (no .NET runtime)            │
└─────────────────────────────────────────────────────┘

┌── PostgreSQL (port 5432) ──────────────────────────┐
│  40 tables, UUID PKs, GORM-managed migrations      │
│  Database: PeopleHub                                    │
│  User: shakil                                       │
└─────────────────────────────────────────────────────┘
```

### How requests flow

```
Browser → http://localhost/api/v1/employees
           ↓ (DNS resolves to localhost)
         HubGateway (port 80)
           ↓ (matches /api/* prefix)
         HubAPI (port 5000)
           ↓ (JWT auth middleware)
         Handler → Repository → PostgreSQL
           ↓
         JSON response → HubGateway → Browser

Browser → http://localhost/dashboard
           ↓
         HubGateway (port 80)
           ↓ (no /api/ prefix → sends to web)
         HubWeb (port 3000)
           ↓ (Next.js serves the React app)
         HTML page with client-side JS
           ↓ (browser then calls /api/* for data)
         HubGateway → HubAPI → PostgreSQL
```

---

## 2. System Requirements

### Software

| Component | Version | Notes |
|-----------|---------|-------|
| Windows | 10/11 or Server 2016+ | IIS + admin rights needed |
| PostgreSQL | 16+ | Database server |
| Go | 1.26+ | For building backend |
| Node.js | 20+ | For building frontend |
| IIS | 7+ | With Management Console |
| URL Rewrite Module | 2.1 | Installed automatically |

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
F:\softwer\hub\
├── hub.exe                  # Go backend binary (built)
├── hub-gateway.exe          # Go reverse proxy binary (built)
├── deploy/                  # Deployment files
│   ├── build-hub.ps1        # ★ ONE-CLICK DEPLOY SCRIPT
│   ├── WinSW-x64.exe        # Windows Service Wrapper
│   ├── hub-service-backend.xml   # HubAPI service config template
│   ├── hub-service-frontend.xml  # HubWeb service config template
│   ├── hub-service-gateway.xml   # HubGateway service config template
│   ├── HubAPI-service.xml        # Processed (paths resolved)
│   ├── HubWeb-service.xml        # Processed
│   └── HubGateway-service.xml    # Processed
├── cmd/                     # Go source code
│   ├── server/main.go       # API server entry point
│   └── gateway/main.go      # Reverse proxy entry point
├── internal/                # Go internal packages
│   ├── handlers/            # HTTP handlers (33 files)
│   ├── models/              # GORM models (40 files)
│   ├── repository/          # Data access layer (26 files)
│   ├── service/             # Business logic (7 files)
│   ├── routes/              # All route registrations
│   ├── middleware/          # Auth, CORS, logger, audit
│   ├── auth/                # JWT + password utilities
│   ├── config/              # Env config loader
│   ├── database/            # GORM connection + migrations
│   ├── server/              # Dependency injection wiring
│   └── utils/               # Pagination, date helpers
├── web/                     # Next.js frontend
│   ├── app/                 # App Router pages (65 routes)
│   ├── components/          # React components
│   │   ├── ui/              # shadcn/ui components
│   │   ├── table/           # DataTable, SimpleTable
│   │   ├── form/            # Form components
│   │   └── layout/          # Sidebar, header, nav
│   ├── lib/                 # API client, axios, utilities
│   ├── .next/standalone/    # ★ Built output (frontend)
│   └── next.config.ts       # ★ Modified: output: "standalone"
├── wwwroot/                 # IIS site root
│   ├── index.html           # Admin/status page
│   └── web.config           # IIS configuration
├── uploads/                 # File uploads directory
├── backups/                 # Database backups
├── README.md                # Original project README
├── AGENTS.md                # Project knowledge base
└── .env                     # Environment variables
```

---

## 4. Quick Deploy (One-Command)

### Prerequisites Check

Ensure these are installed on your machine:

```powershell
# Check Go
go version

# Check Node.js
node --version

# Check PostgreSQL (should be running)
psql --version

# Check IIS (should be running)
Get-Service W3SVC
```

### Deploy

```powershell
# Open PowerShell as Administrator, then:
cd F:\softwer\hub
.\deploy\build-hub.ps1
```

The script will:

1. ✅ Clean old `PeopleHub.exe` / `server.exe` binaries (rename safety)
2. ✅ Check prerequisites (Go, Node.js, IIS)
3. ✅ Build `hub.exe` (Go backend)
4. ✅ Build `hub-gateway.exe` (Go reverse proxy)
5. ✅ Build Next.js frontend (standalone output)
6. ✅ Download WinSW (if needed)
7. ✅ Install 3 Windows services (HubAPI, HubWeb, HubGateway)
8. ✅ Create IIS site on port 8081
9. ✅ Open firewall port 80
10. ✅ Start all services
11. ✅ Display access URLs

### Access After Deploy

| URL | What | Port |
|-----|------|------|
| `http://localhost/` | **Frontend** (Hub HR login) | 80 |
| `http://localhost/api/v1` | **API** endpoints | 80 → 5000 |
| `http://localhost/swagger/index.html` | **Swagger** API docs | 80 → 5000 |
| `http://localhost/health` | **Health** check | 80 → 5000 |
| `http://localhost:8081/` | **IIS** admin status page | 8081 |

---

## 5. Manual Step-by-Step Build

If you prefer to build each component manually (or the script fails), follow these steps:

### Step 1: Build Go Backend

```powershell
cd F:\softwer\hub
go build -o hub.exe -ldflags="-s -w" ./cmd/server
```

This produces `hub.exe` (~42 MB) — the API server.

### Step 2: Build Go Reverse Proxy

```powershell
go build -o hub-gateway.exe -ldflags="-s -w" ./cmd/gateway
```

This produces `hub-gateway.exe` (~6.4 MB) — the routing gateway.

### Step 3: Build Next.js Frontend

```powershell
cd F:\softwer\hub\web

# Ensure standalone output is enabled in next.config.ts:
#   output: "standalone",

npm install
npm run build
```

The built output is at `web\.next\standalone\`.

### Step 4: Prepare Standalone Folder

```powershell
$standalone = "F:\softwer\hub\web\.next\standalone"

# Copy public assets
Copy-Item "F:\softwer\hub\web\public\*" "$standalone\public\" -Recurse -Force

# Copy static JS/CSS
New-Item -ItemType Directory -Force -Path "$standalone\.next\static" | Out-Null
Copy-Item "F:\softwer\hub\web\.next\static\*" "$standalone\.next\static\" -Recurse -Force
```

### Step 5: Install Windows Services

```powershell
# Download WinSW
Invoke-WebRequest -Uri "https://github.com/winsw/winsw/releases/download/v2.12.0/WinSW-x64.exe" -OutFile "deploy\WinSW-x64.exe"

# Install each service
$ROOT = "F:\softwer\hub"

# HubAPI (Go backend on port 5000)
Copy-Item deploy\WinSW-x64.exe deploy\HubAPI-service.exe
$xml = (Get-Content deploy\hub-service-backend.xml -Raw) -replace '%BASE%', $ROOT
Set-Content deploy\HubAPI-service.xml $xml
deploy\HubAPI-service.exe install

# HubWeb (Next.js on port 3000)
Copy-Item deploy\WinSW-x64.exe deploy\HubWeb-service.exe
$xml = (Get-Content deploy\hub-service-frontend.xml -Raw) -replace '%BASE%', $ROOT
Set-Content deploy\HubWeb-service.xml $xml
deploy\HubWeb-service.exe install

# HubGateway (reverse proxy on port 80)
Copy-Item deploy\WinSW-x64.exe deploy\HubGateway-service.exe
$xml = (Get-Content deploy\hub-service-gateway.xml -Raw) -replace '%BASE%', $ROOT
Set-Content deploy\HubGateway-service.xml $xml
deploy\HubGateway-service.exe install
```

### Step 6: Configure IIS

```powershell
Import-Module WebAdministration

# Create App Pool
New-Item "IIS:\AppPools\HubAppPool" -Force
Set-ItemProperty "IIS:\AppPools\HubAppPool" -Name managedRuntimeVersion -Value ""
Set-ItemProperty "IIS:\AppPools\HubAppPool" -Name startMode -Value "AlwaysRunning"

# Create Site on port 8081
New-Website -Name "HubSite" -PhysicalPath "F:\softwer\hub\wwwroot" -Port 8081 -ApplicationPool "HubAppPool" -Force
Start-Website -Name "HubSite"
```

### Step 7: Configure Firewall

```powershell
netsh advfirewall firewall add rule name="Hub-Gateway-80" dir=in action=allow protocol=TCP localport=80 profile=any
```

### Step 8: Start Services

```powershell
Start-Service HubAPI    # Waits for :5000 to be ready
Start-Sleep 3
Start-Service HubWeb    # Waits for :3000 to be ready
Start-Sleep 2
Start-Service HubGateway  # Routes :80 → :5000 and :3000
```

### Step 9: Verify

```powershell
Invoke-WebRequest -Uri "http://localhost/health" -UseBasicParsing
Invoke-WebRequest -Uri "http://localhost/" -UseBasicParsing
Invoke-WebRequest -Uri "http://localhost:8081/" -UseBasicParsing
```

---

## 6. How It All Works

### 6.1 The Go Reverse Proxy (hub-gateway.exe)

This is the **heart of the deployment**. It's a small Go program that:

- **Listens on port 80** (no IIS needed for public traffic)
- **Inspects every incoming URL path**
- **Routes based on prefix matching**:

| Path Prefix | Routes To | Target Port |
|-------------|-----------|-------------|
| `/api/` | Backend API | :5000 |
| `/swagger/` | Swagger docs | :5000 |
| `/uploads/` | Uploaded files | :5000 |
| `/health` | Health check | :5000 |
| Everything else | Frontend app | :3000 |

**Source:** `cmd/gateway/main.go` — uses Go's `httputil.NewSingleHostReverseProxy`.

**Why not IIS ARR?** The IIS Application Request Routing (ARR) module requires Windows Server and is not available on Windows 10. The Go reverse proxy is a lightweight (~6 MB binary) alternative that works on any Windows version.

### 6.2 Go Backend API (hub.exe)

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

### 6.3 Next.js Frontend (HubWeb)

**Port:** 3000  
**Stack:** Next.js 16.2 + React 19.2 + Tailwind CSS 4 + shadcn/ui  
**Rendering:** Client-Side (all pages use `"use client"`)  
**Auth:** JWT stored in localStorage + Axios interceptor for auto-refresh

**Built as standalone output** (`output: "standalone"` in `next.config.ts`) so it runs via a plain Node.js server without requiring Next.js CLI.

### 6.4 IIS (HubSite on port 8081)

**Purpose:** Admin/management interface, not the main entry point.

The IIS site serves `wwwroot\index.html` — a simple status page with links to the gateway URLs. The `web.config` configures custom headers and request filtering.

**Why keep IIS?** The user requested IIS hosting. IIS provides:
- Centralized management via IIS Manager
- SSL certificate management (future)
- Windows authentication integration (future)
- Monitoring and logging via IIS logs

### 6.5 Windows Services (WinSW)

Each component runs as a Windows service managed by **WinSW** (Windows Service Wrapper):

| Service Name | Executable | XML Config |
|-------------|------------|------------|
| `HubAPI` | `hub.exe` | `deploy\HubAPI-service.xml` |
| `HubWeb` | `node server.js` | `deploy\HubWeb-service.xml` |
| `HubGateway` | `hub-gateway.exe` | `deploy\HubGateway-service.xml` |

**Auto-restart on failure:** All services restart after 10 seconds if they crash.  
**Auto-start on boot:** All services set to `Automatic` start type.

### 6.6 Service Configuration Details

**HubAPI** (`deploy\hub-service-backend.xml`):
```xml
<service>
  <id>HubAPI</id>
  <name>Hub API Server</name>
  <executable>F:\softwer\hub\hub.exe</executable>
  <workingdirectory>F:\softwer\hub</workingdirectory>
  <env name="PORT" value="5000"/>
  <env name="DB_HOST" value="localhost"/>
  <env name="DB_PORT" value="5432"/>
  <env name="DB_USER" value="shakil"/>
  <env name="DB_PASS" value="123456"/>
  <env name="DB_NAME" value="PeopleHub"/>
  <env name="JWT_SECRET" value="PeopleHub-secret-key-change-in-production-2024"/>
  <onfailure action="restart" delay="10 sec"/>
  <startmode>Automatic</startmode>
</service>
```

**HubWeb** (`deploy\hub-service-frontend.xml`):
```xml
<service>
  <id>HubWeb</id>
  <name>Hub Web Frontend</name>
  <executable>node</executable>
  <arguments>F:\softwer\hub\web\.next\standalone\server.js</arguments>
  <workingdirectory>F:\softwer\hub\web</workingdirectory>
  <env name="PORT" value="3000"/>
  <env name="NODE_ENV" value="production"/>
  <env name="NEXT_PUBLIC_API_URL" value="http://localhost:5000/api/v1"/>
  <onfailure action="restart" delay="10 sec"/>
</service>
```

**HubGateway** (`deploy\hub-service-gateway.xml`):
```xml
<service>
  <id>HubGateway</id>
  <name>Hub Gateway (Reverse Proxy)</name>
  <executable>F:\softwer\hub\hub-gateway.exe</executable>
  <workingdirectory>F:\softwer\hub</workingdirectory>
  <env name="GATEWAY_PORT" value="80"/>
  <env name="API_TARGET" value="http://localhost:5000"/>
  <env name="WEB_TARGET" value="http://localhost:3000"/>
  <onfailure action="restart" delay="10 sec"/>
</service>
```

---

## 7. Services Management

### Start/Stop/Restart

```powershell
# All services
Start-Service HubAPI; Start-Service HubWeb; Start-Service HubGateway
Stop-Service HubAPI, HubWeb, HubGateway
Restart-Service HubAPI, HubWeb, HubGateway

# Individual
Start-Service HubAPI
Stop-Service HubWeb
Restart-Service HubGateway
```

### Check Status

```powershell
Get-Service HubAPI, HubWeb, HubGateway | Select-Object Name, Status, StartType
```

### View Logs

```powershell
# Service logs are in deploy/ folder
Get-Content "F:\softwer\hub\deploy\HubAPI-service.log" -Tail 20
Get-Content "F:\softwer\hub\deploy\HubWeb-service.log" -Tail 20
Get-Content "F:\softwer\hub\deploy\HubGateway-service.log" -Tail 20
```

### Uninstall Services

```powershell
Stop-Service HubAPI, HubWeb, HubGateway
F:\softwer\hub\deploy\HubAPI-service.exe uninstall
F:\softwer\hub\deploy\HubWeb-service.exe uninstall
F:\softwer\hub\deploy\HubGateway-service.exe uninstall
```

---

## 8. IIS Configuration

### 8.1 IIS Site Details

| Setting | Value |
|---------|-------|
| Site Name | `HubSite` |
| App Pool | `HubAppPool` (No Managed Code) |
| Port | 8081 |
| Physical Path | `F:\softwer\hub\wwwroot` |
| Protocol | HTTP |

### 8.2 web.config

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

**Note:** URL Rewrite rules are NOT in this web.config because ARR (Application Request Routing) is not available on Windows 10. If you deploy to Windows Server, you can install ARR 3.0 and add reverse proxy rewrite rules to the web.config instead of using the Go gateway.

### 8.3 Adding SSL/HTTPS (Future)

In IIS Manager:
1. Select HubSite → Bindings → Add
2. Type: `https`, Port: `443`
3. Select an SSL certificate
4. Update `NEXT_PUBLIC_API_URL` in `.env` and HubWeb config to `https://`

### 8.4 Changing the IIS Admin Port

```powershell
Remove-WebBinding -Name "HubSite" -Protocol "http" -Port 8081
New-WebBinding -Name "HubSite" -Protocol "http" -Port 9090 -IPAddress "*"
```

---

## 9. Firewall & Security

### Ports Used

| Port | Service | Direction | Purpose |
|------|---------|-----------|---------|
| 80 | HubGateway | Inbound | Main web traffic |
| 8081 | IIS HubSite | Inbound (optional) | Admin page |
| 5000 | HubAPI | Local only | Backend API |
| 3000 | HubWeb | Local only | Next.js frontend |
| 5432 | PostgreSQL | Local only | Database |

### Firewall Rules

```powershell
# List existing rule
netsh advfirewall firewall show rule name="Hub-Gateway-80"

# Remove if needed
netsh advfirewall firewall delete rule name="Hub-Gateway-80"

# Re-create
netsh advfirewall firewall add rule name="Hub-Gateway-80" dir=in action=allow protocol=TCP localport=80 profile=any
```

### Environment Variables (`.env`)

```
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=shakil
DB_PASS=123456
DB_NAME=PeopleHub
DB_SSLMODE=disable

# API Server
PORT=5000
JWT_SECRET=PeopleHub-secret-key-change-in-production-2024

# Frontend
FRONTEND_PORT=3000
NEXT_PUBLIC_API_URL=http://localhost:5000/api/v1
```

**⚠️ Security Note:** Change the `JWT_SECRET` to a strong, random value before production use. Keep `.env` out of version control (it's in `.gitignore`).

### Project Rename: PeopleHub → hub

The project was renamed from `PeopleHub` → `hub` to avoid conflicts. During deployment:

- Old binaries (`PeopleHub.exe`, `server.exe`, `employee.exe`, `reset.exe`) are **automatically deleted**
- Old services (`PeopleHub*`) are stopped and removed
- Old IIS sites (`ERPHub`) are removed
- Only `hub.exe` (backend) and `hub-gateway.exe` (proxy) remain
- All new services use the **Hub** prefix (`HubAPI`, `HubWeb`, `HubGateway`)

If any old `PeopleHub` process or service is still running, the build script cleans it up automatically. The gateway on port 80 routes all traffic to the new Hub services only.

---

## 10. Troubleshooting

### 10.1 Service Won't Start

```powershell
# Check Windows Event Log
Get-WinEvent -LogName "Application" -MaxEvents 10 | Where-Object { $_.LevelDisplayName -eq "Error" } | Format-Table TimeCreated, Message -Wrap

# Check service logs
Get-Content "F:\softwer\hub\deploy\HubAPI-service.log" -Tail 30
```

### 10.2 Port Already in Use

```powershell
# Find what's using port 80
netstat -ano | Select-String ":80 "

# Stop the process
Stop-Process -Id <PID> -Force

# Or change gateway port
$env:GATEWAY_PORT = "8080"
```

### 10.3 Database Connection Error

```powershell
# Test PostgreSQL
psql -U shakil -d PeopleHub -c "SELECT 1"

# Check if PostgreSQL service is running
Get-Service postgresql*

# Verify .env credentials match
```

### 10.4 Frontend Not Loading

```powershell
# Check if the standalone server exists
Test-Path "F:\softwer\hub\web\.next\standalone\server.js"

# Check if the standalone has static assets
Test-Path "F:\softwer\hub\web\.next\standalone\.next\static"

# Rebuild if missing
cd F:\softwer\hub\web
npm run build
```

### 10.5 IIS Site 500 Error

```powershell
# Check applicationHost.config syntax
& "$env:SystemRoot\system32\inetsrv\appcmd.exe" list config "HubSite"

# Restart IIS
iisreset /restart

# Check if web.config has locked sections
# Most common: staticContent with duplicate MIME types
# Fix: Remove the web.config and start minimal
```

### 10.6 Gateway Not Routing

```powershell
# Check if backend is reachable
Invoke-WebRequest -Uri "http://localhost:5000/health" -UseBasicParsing

# Check if frontend is reachable
Invoke-WebRequest -Uri "http://localhost:3000/" -UseBasicParsing

# Test gateway routing
Invoke-WebRequest -Uri "http://localhost/health" -UseBasicParsing
Invoke-WebRequest -Uri "http://localhost/" -UseBasicParsing
```

---

## 11. Re-Deploy & Updates

### Full Re-Deploy (Build + Services + IIS)

```powershell
.\deploy\build-hub.ps1
```

This runs the complete pipeline: build all binaries → install services → configure IIS → start everything.

### Update Backend Only

```powershell
cd F:\softwer\hub
git pull (or update source code)
go build -o hub.exe -ldflags="-s -w" ./cmd/server
Stop-Service HubAPI
Start-Sleep 2
Copy-Item hub.exe F:\softwer\hub\hub.exe -Force
Start-Service HubAPI
```

### Update Frontend Only

```powershell
cd F:\softwer\hub\web
git pull (or update source code)
npm install
npm run build

# Copy assets to standalone
Copy-Item "web\public\*" "web\.next\standalone\public\" -Recurse -Force
Copy-Item "web\.next\static\*" "web\.next\standalone\.next\static\" -Recurse -Force

Stop-Service HubWeb
Start-Sleep 2
Start-Service HubWeb
```

### Update Gateway Only

```powershell
cd F:\softwer\hub
go build -o hub-gateway.exe -ldflags="-s -w" ./cmd/gateway
Stop-Service HubGateway
Start-Sleep 2
Copy-Item hub-gateway.exe F:\softwer\hub\hub-gateway.exe -Force
Start-Service HubGateway
```

### Database Migrations

The Go backend automatically runs GORM AutoMigrate on startup. To run migrations:

```powershell
# Simply restart HubAPI
Restart-Service HubAPI

# To seed data
go run cmd/superadmin/main.go
go run cmd/seed/main.go
go run cmd/seed/organization/main.go
go run cmd/seed/leave/main.go
go run cmd/seed/employee/main.go
```

---

## Appendix: Key Files Reference

| File | Absolute Path |
|------|---------------|
| Backend binary | `F:\softwer\hub\hub.exe` |
| Gateway binary | `F:\softwer\hub\hub-gateway.exe` |
| Frontend standalone | `F:\softwer\hub\web\.next\standalone\server.js` |
| Frontend config | `F:\softwer\hub\web\next.config.ts` |
| Environment | `F:\softwer\hub\.env` |
| IIS web.config | `F:\softwer\hub\wwwroot\web.config` |
| IIS status page | `F:\softwer\hub\wwwroot\index.html` |
| Gateway source | `F:\softwer\hub\cmd\gateway\main.go` |
| Backend entry | `F:\softwer\hub\cmd\server\main.go` |
| Service configs | `F:\softwer\hub\deploy\Hub*-service.xml` |
| Deploy script | `F:\softwer\hub\deploy\build-hub.ps1` |
| WinSW executable | `F:\softwer\hub\deploy\WinSW-x64.exe` |
| Service logs | `F:\softwer\hub\deploy\Hub*-service.log` |
| Uploaded files | `F:\softwer\hub\uploads\` |
| Database backups | `F:\softwer\hub\backups\` |

---

## Architecture Decision Record

| Decision | Rationale |
|----------|-----------|
| **Go reverse proxy instead of IIS ARR** | ARR 3.0 is incompatible with Windows 10. HttpPlatformHandler download is discontinued. Go proxy is lightweight, cross-platform, and reliable. |
| **WinSW for services** | Native Windows services with auto-restart and logging. No extra dependencies. |
| **Next.js standalone output** | Allows frontend to run as a service without Next.js CLI. Enables production `node server.js` mode. |
| **IIS on port 8081** | Keeps IIS available for management while gateway handles main traffic. Port 8081 avoids conflicts. |
| **Two separate Go binaries** | Clear separation of concerns. Gateway handles routing; backend handles business logic. Each can be updated independently. |

---

*Document Version: 1.0*  
*Last Updated: 2026-07-23*  
*Project: Hub HR System (formerly PeopleHub)*
