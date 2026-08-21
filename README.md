# PeopleHub — Enterprise HR Management & Payroll System

> **Target Industry:** Bangladeshi Garment & Factory Industries  
> **Architecture:** Monolithic (Go Backend + Next.js Frontend + PostgreSQL)  
> **Deployment:** Docker Compose (3 services)

---

## Table of Contents

1. [Features](#features)
2. [Tech Stack](#tech-stack)
3. [Prerequisites](#prerequisites)
4. [Installation & Setup](#installation--setup)
5. [Database & Migration](#database--migration)
6. [Running the Application](#running-the-application)
7. [API Documentation](#api-documentation)
8. [Docker Deployment](#docker-deployment)
9. [Environment Variables](#environment-variables)
10. [Seed Data](#seed-data)
11. [API Endpoints Overview](#api-endpoints-overview)
12. [Project Structure](#project-structure)
13. [Key Data Flows](#key-data-flows)
14. [Development Standards](#development-standards)
15. [Security Notes](#security-notes)

---

## Features

- **Employee Lifecycle Management:** Full employee profile, organizational hierarchy, ID cards, separations
- **Biometric Attendance Integration:** ZKTeco device integration via MDB file import
- **Shift Management:** Regular shifts + temporary shift overrides per employee per date
- **Leave Management:** Leave types, applications, approvals, balance tracking, monthly reports
- **Payroll Automation:** Monthly salary processing with absent deduction, overtime, attendance bonus
- **Organization Hierarchy:** Companies → Departments → Sections → Designations → Lines → Groups → Floors
- **Bangladesh Geo Data:** Divisions → Districts → Upazilas → Unions
- **Dashboard & Analytics:** Real-time stats, attendance trends, department/gender distribution
- **Database Admin:** Backup, restore, reset via API endpoints
- **RBAC Authentication:** JWT-based auth with roles, permissions, session management, password history
- **Additional HR Operations:** Night bill and tiffin bill management, punishment tracking, daily schedules, salary increment approvals

---

## Tech Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| **Backend Language** | Go | 1.26.2 |
| **Web Framework** | Gin | 1.12.0 |
| **ORM** | GORM | 1.31.2 |
| **Database** | PostgreSQL | 16 |
| **Authentication** | JWT (golang-jwt/v5) | 5.3.1 |
| **Password Hash** | bcrypt | — |
| **API Docs** | Swaggo | 1.16.6 |
| **Frontend Framework** | Next.js (App Router) | 16.2.10 |
| **UI Library** | React + Tailwind CSS + shadcn/ui | 19.2.4 / v4 |
| **Tables** | TanStack React Table | v8 |
| **Charts** | Recharts | 3.8.0 |
| **HTTP Client** | Axios | 1.18.1 |
| **Biometric Reader** | ZKTeco ADODB / ACE.OLEDB.12.0 | Windows only |
| **Containerization** | Docker & Docker Compose | — |

---

## Prerequisites

### For Local Development (without Docker)

- **Go** 1.26 or later
- **Node.js** 20 or later
- **PostgreSQL** 16 or later
- **ZKTeco Software** (for biometric MDB import — **Windows only**)
- **Swaggo CLI** (`go install github.com/swaggo/swag/cmd/swag@latest`)

### For Docker Development

- **Docker** 24.0+
- **Docker Compose** v2+

---

## Installation & Setup

### 1. Clone the Repository

```bash
git clone <repository-url>
cd PeopleHub
```

### 2. Configure Environment Variables

Copy the example environment file and adjust values:

```bash
cp .env.example .env
```

Edit `.env` with your local settings:

```env
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
NEXT_PUBLIC_API_URL=http://localhost:5000/api/v1
```

> **Security Warning:** Change `JWT_SECRET` and database credentials for production. Never commit `.env` to version control.

---

## Database & Migration

PeopleHub uses **GORM AutoMigrate** for schema management. Migrations run automatically when the backend starts.

### How Migration Works

1. **AutoMigrate:** `internal/database/postgres.go` calls `db.AutoMigrate(...)` with all 40 models. This creates tables if they don't exist.
2. **ALTER Fixes:** After AutoMigrate, `postgres.go` runs `ALTER TABLE ... TYPE varchar(50)` on `employee_id` columns because GORM 1.31.2 forces UUID type on `*_id` columns by default.
3. **Auto-Created Indexes:** On startup, `postgres.go` creates `IF NOT EXISTS` performance indexes for high-frequency queries (salaries, employees, leave_allocations, temporary_shifts, data_logs, attendances, leaves, sessions).
4. **No Foreign Keys:** `DisableForeignKeyConstraintWhenMigrating: true` is set. Referential integrity is managed at the application level.

### Manual Database Reset (DANGEROUS)

Via API (authenticated):
```bash
curl -X POST http://localhost:5000/api/v1/database/reset \
  -H "Authorization: Bearer <token>"
```

This drops **all** tables and recreates them via AutoMigrate.

### Manual Backup & Restore

**Create Backup:**
```bash
curl -X POST http://localhost:5000/api/v1/database/backup \
  -H "Authorization: Bearer <token>"
```

**Restore from SQL:**
```bash
curl -X POST http://localhost:5000/api/v1/database/import \
  -H "Authorization: Bearer <token>" \
  -F "file=@backup.sql"
```

---

## Running the Application

### Option A: Local Development (without Docker)

#### Step 1: Start PostgreSQL
Ensure PostgreSQL is running and the database `PeopleHub` exists:

```bash
createdb -U shakil PeopleHub   # or use pgAdmin / psql
```

#### Step 2: Run the Backend

```bash
# Install Go dependencies
go mod download

# Build the server
go build -o PeopleHub.exe ./cmd/server

# Run the server
.\PeopleHub.exe
```

The API will be available at: **http://localhost:5000**

#### Step 3: Run the Frontend

```bash
cd web

# Install dependencies
npm install

# Start development server
npm run dev
```

The frontend will be available at: **http://localhost:3000**

#### Step 4: Regenerate Swagger Docs (after handler changes)

```powershell
# Windows PowerShell
& "C:\Users\SHAKIL\go\bin\swag.exe" init -g cmd/server/main.go -o docs

# Or if swag is in PATH
swag init -g cmd/server/main.go -o docs
```

View Swagger UI at: **http://localhost:5000/swagger/index.html**

---

### Option B: Docker Deployment (Recommended)

```bash
docker-compose up --build
```

Services will be available at:

| Service | URL / Port |
|---------|-----------|
| PostgreSQL | `localhost:5432` |
| Backend API | `http://localhost:5000` |
| Frontend | `http://localhost:3000` |
| Swagger UI | `http://localhost:5000/swagger/index.html` |

To stop:
```bash
docker-compose down
```

To stop and remove volumes (deletes database data):
```bash
docker-compose down -v
```

---

## API Documentation

PeopleHub uses **Swagger/OpenAPI 2.0** for API documentation.

- **Swagger UI:** `http://localhost:5000/swagger/index.html`
- **OpenAPI JSON:** `http://localhost:5000/swagger/doc.json`
- **OpenAPI YAML:** `http://localhost:5000/swagger/doc.yaml`

All endpoints are prefixed with `/api/v1`.

### Authentication

Protected endpoints require a **Bearer Token** in the `Authorization` header:

```
Authorization: Bearer <access_token>
```

Obtain tokens via:
```bash
curl -X POST http://localhost:5000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

Response:
```json
{
  "access_token": "eyJ...",
  "refresh_token": "abc123...",
  "user": { "id": "...", "email": "...", "name": "..." }
}
```

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `shakil` | PostgreSQL username |
| `DB_PASS` | `123456` | PostgreSQL password |
| `DB_NAME` | `PeopleHub` | PostgreSQL database name |
| `DB_SSLMODE` | `disable` | SSL mode (disable/require/verify-full) |
| `PORT` | `5000` | Backend API port |
| `JWT_SECRET` | `PeopleHub-secret-key...` | JWT signing secret |
| `NEXT_PUBLIC_API_URL` | `http://localhost:5000/api/v1` | Frontend API base URL |

---

## Seed Data

After starting the application, run seeders to populate initial data:

### 1. Geo Data (Bangladesh Administrative Divisions)
```bash
go run cmd/seed/main.go
```
Fetches divisions, districts, upazilas, and unions from remote JSON. Safe to run multiple times (idempotent).

### 2. Superadmin User
```bash
go run cmd/superadmin/main.go
```
Creates the superadmin role, all permissions, and a superadmin user.

**Default credentials:**
- Email: `superadmin@PeopleHub.com`
- Password: `superadmin1234`

Or configure via environment:
```env
SUPERADMIN_EMAIL=admin@company.com
SUPERADMIN_PASSWORD=SecurePass123!
SUPERADMIN_NAME=System Admin
```

### 3. Organization Hierarchy
```bash
go run cmd/seed/organization/main.go
```
Creates:
- Company: PeopleHub Technologies Ltd.
- 4 Branches
- 12 Departments
- 20 Sections
- 27 Designations
- Production Lines
- Groups A–D
- Floors 1–10
- 5 Shifts

### 4. Leave Types
```bash
go run cmd/seed/leave/main.go
```
Creates 8 default leave types:
- AL (Annual Leave)
- SL (Sick Leave)
- CL (Casual Leave)
- ML (Maternity Leave)
- PL (Paternity Leave)
- EL (Earned Leave)
- STL (Special Leave)
- HL (Half-day Leave)

### 5. Sample Employees
```bash
go run cmd/seed/employee/main.go
```
Creates sample employee records for testing.

---

## API Endpoints Overview

### Public Endpoints (No Auth Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/auth/register` | User registration |
| POST | `/api/v1/auth/login` | Login (returns tokens) |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/logout` | Revoke refresh token |
| POST | `/api/v1/auth/forgot-password` | Request password reset email |
| POST | `/api/v1/auth/reset-password` | Reset password with token |
| POST | `/api/v1/auth/validate-token` | Check if access token is valid |

### Protected Endpoints (Bearer Token Required)

#### Auth
- `POST /auth/logout-all` — Revoke all sessions
- `PUT /auth/change-password` — Change password
- `GET /auth/me` — Current user profile
- `PUT /auth/profile` — Update profile
- `GET /auth/sessions` — Active sessions

#### Companies
- `GET /companies`, `GET /companies/:id`, `POST /companies`, `PUT /companies/:id`, `DELETE /companies/:id`

#### Organization Hierarchy
- `GET/POST/PUT/DELETE` for `/departments`, `/sections`, `/designations`, `/lines`, `/groups`, `/floors`

#### Geo Data
- `GET/POST/PUT/DELETE` for `/divisions`, `/districts`, `/upazilas`, `/unions`
- Filter with query params: `?division_id=`, `?district_id=`, `?upazila_id=`

#### Employees
- `GET /employees` — List with filters (company, department, section, designation, line, group, status, gender, blood_group, salary range)
- `GET /employees/:id` — Get by ID
- `POST /employees` — Create
- `PUT /employees/:id` — Update
- `DELETE /employees/:id` — Delete
- `GET /employees/import/template` — Download Excel template
- `POST /employees/import` — Bulk import from Excel
- `GET /employees/export/excel` — Export to Excel
- `GET /employees/export/pdf` — Export to PDF

#### Shifts
- `GET/POST/PUT/DELETE` for `/shifts`

#### Temporary Shifts
- `GET/POST/PUT/DELETE` for `/temporary-shifts`

#### Attendance (14 endpoints)
- `GET /attendance` — List by date (default today)
- `GET /attendance/:id` — Get record
- `POST /attendance` — Create manual attendance (upserts if duplicate)
- `PUT /attendance/:id` — Update
- `DELETE /attendance/:id` — Delete
- `DELETE /attendance/delete-all` — Permanently delete all
- `POST /attendance/clock-in` — Mark clock-in
- `POST /attendance/clock-out` — Mark clock-out
- `GET /attendance/summary` — Aggregated summary
- `GET /attendance/overtime` — Overtime sheet
- `GET /attendance/overtime-summary` — Department overtime summary
- `GET /attendance/job-card` — Job card view
- `GET /attendance/stats` — Today's count
- `GET /attendance/missing` — Missing attendance records
- `GET /attendance/absent` — Absent employees
- `GET /attendance/monthly-report` — Per-employee monthly summary

#### Data Logs (ZKTeco Integration)
- `GET /data-logs` — List raw punches by date range
- `GET /data-logs/stats` — Total/today counts
- `POST /data-logs/import` — Import from MDB file
- `POST /data-logs/process` — Process unprocessed logs → attendance
- `DELETE /data-logs/delete-all` — Permanently delete all logs

#### Leave Management
- `GET/POST/PUT/DELETE` for `/leave-types`
- `GET/POST/PUT/DELETE` for `/leaves`
- `PUT /leaves/:id/approve` — Approve leave
- `PUT /leaves/:id/reject` — Reject leave
- `GET /leave-balance` — Get leave balance
- `GET /leave-reports/monthly` — Monthly leave report

#### Payroll
- `POST /salary/process` — Process monthly salary
- `GET /salary/sheet` — Full salary sheet
- `GET /salary/payslip` — Single employee payslip
- `GET /salary/list` — Department summary
- `GET /salary/summary` — Salary summary with grand totals
- `GET /salary/daily-sheet`, `/salary/bank-sheet` — Specialized sheets
- `GET/PUT /salary/increments` — Salary increment CRUD + approve/reject

#### HR Operations
- `GET/POST/PUT/DELETE` for `/requirements`
- `GET/POST/PUT/DELETE` for `/separations`
- `GET/POST/PUT/DELETE` for `/id-cards` + `POST /id-cards/generate` (PDF)
- `GET/POST/PUT/DELETE` for `/punishments`
- `GET/POST/PUT/DELETE` for `/daily-schedules`
- `GET/POST/PUT/DELETE` for `/night-bills`, `/tiffin-bills`

#### User & Role Management
- `GET/POST/PUT/DELETE` for `/users`, `/roles`
- `POST /users/:id/roles` — Assign roles to user
- `POST /roles/:id/permissions` — Assign permissions to role
- `GET /permissions` — List all permissions

#### Settings
- `GET /settings` — List all settings
- `PUT /settings` — Bulk update settings

#### Dashboard & Admin
- `GET /dashboard/stats` — Dashboard statistics
- `POST /database/backup` — Create backup
- `GET /database/backups` — List backups
- `GET /database/export?filename=...` — Download backup
- `POST /database/import` — Restore from SQL
- `DELETE /database/backups` — Delete backups
- `POST /database/reset` — Drop all tables and remigrate

#### Upload
- `POST /upload` — Generic file upload (max 5MB, jpg/jpeg/png/gif)

**Pagination:** All `GET` list endpoints support `?page=` (default 1) and `?limit=` (default 20, max 100) query parameters. Responses are wrapped in `PaginatedResponse` with `data`, `total`, `page`, `limit`, `total_pages`.

**Total Endpoints:** ~170+ (7 Public, ~163 Protected)

---

## Project Structure

```
PeopleHub/
├── .env                          # Environment configuration
├── .env.example                  # Environment template
├── docker-compose.yml            # 3-service orchestration
├── Dockerfile                    # Backend Docker image
├── go.mod / go.sum               # Go module dependencies
├── cmd/
│   ├── server/main.go            # API entry point + Swagger annotations
│   ├── seed/main.go              # Geo data seeder (BD divisions/districts/...)
│   ├── seed/organization/main.go # Company & org hierarchy seeder
│   ├── seed/leave/main.go        # Leave types seeder
│   ├── seed/employee/main.go     # Sample employees seeder
│   └── superadmin/main.go        # Superadmin role + user seeder
├── docs/                         # Generated Swagger files
├── backups/                      # PostgreSQL dump files
├── uploads/                      # Static file uploads
├── internal/
│   ├── auth/                     # JWT & password hashing (bcrypt cost 12)
│   ├── config/                   # Environment config loader
│   ├── database/                 # GORM PostgreSQL connection + AutoMigrate(40 models) + ALTER fixes + 9 indexes
│   ├── handlers/                 # HTTP handlers (33 files)
│   ├── middleware/               # Auth, CORS, Logger, Audit, Permission
│   ├── models/                   # 40 GORM model files
│   ├── repository/               # 26 data access files
│   ├── routes/                   # Route registration (single 490-line file)
│   ├── server/                   # Dependency injection wiring (25+ repos, 7 services)
│   └── service/                  # Business logic (auth, salary, attendance, separation, MDB reader, user)
└── web/                          # Next.js 16 frontend
    ├── app/                      # App Router pages
    ├── components/               # UI components, tables, forms, layout
    ├── lib/                      # API client, axios config, utilities
    └── package.json
```

---

## Key Data Flows

### 1. ZKTeco Biometric → Attendance

```
ZKTeco Device
    └──► CHECKINOUT table in att2000.mdb
              │
              ▼
    POST /api/v1/data-logs/import
    PowerShell reads USERINFO + CHECKINOUT
              │
              ▼
    data_logs table (raw punches, processed=false)
              │
              ▼
    POST /api/v1/data-logs/process
    Groups by user, matches badge → employee,
    applies shift rules + leaves + weekends,
    creates attendance records
```

### 2. Attendance → Payroll

```
POST /api/v1/salary/process
    ├──► Load active employees
    ├──► Load monthly attendance report
    ├──► Load monthly overtime hours
     └──► For each employee:
         - core = gross - transport(450) - food(1250) - medical(750)
         - basic = core / 1.5
         - house_rent = core - basic
         - medical = 750 (fixed)
         - absent_deduction = (gross / total_days) × absent_days
         - ot_amount = ot_hours × (basic / total_days / 8)
         - attendance_bonus = 500 (if perfect attendance)
         - net_salary = gross - absent_deduction + ot_amount + attendance_bonus
         - Upsert salaries table
```

### 3. Leave Application Flow

All leave operations are wrapped in **database transactions** for data consistency.

```
POST /leaves (Apply)
    ├──► Validate dates
    ├──► Check allocation balance
    └──► Transaction:
         ├── Create pending leave
         └── Increment allocation.pending_days

PUT /leaves/:id/approve
    └──► Transaction:
         ├── Set approved status
         ├── Move allocation pending → used
         └── Mark attendance as on_leave

PUT /leaves/:id/reject
    └──► Transaction:
         ├── Set rejected status
         └── Decrement allocation.pending_days
```

---

## Frontend Architecture Details

| Layer | Technology | Details |
|-------|-----------|---------|
| **Framework** | Next.js 16 App Router | Client-side rendered pages with `"use client"` |
| **State** | React hooks (`useState`/`useEffect`) | No global state manager (Zustand/Redux) |
| **Data Fetching** | Axios via `lib/api.ts` | `lib/axios-instance.ts` handles 401 token refresh |
| **Forms** | React Hook Form + Zod | Domain-specific schemas in `components/data/*` |
| **Tables** | TanStack React Table v8 | Wrapped in `DataTable` component (528 lines) |
| **UI** | shadcn/ui + Radix | 129 components (Sidebar, Dialog, Select, DatePicker, etc.) |
| **Styling** | Tailwind CSS v4 + cn() | oklch color space, dark mode, custom Bengali font |
| **Charts** | Recharts | Dashboard and analytics pages |
| **Layout** | Sidebar-based | Collapsible groups with navigation (NavMain, NavGroup, NavSecondary) |

**Key pages (~40+ route files):** Dashboard, Employee CRUD (with 15+ filters), Attendance (daily, summary, job-card, OT, missing, absent), Leave (types, applications, reports), Payroll (process, sheet, payslip, increment), HR Ops (requirements, separations, ID cards, punishments, daily schedules, night bills, tiffin bills), Admin (users, roles, permissions, database, settings)

## Development Standards

For detailed coding standards, architecture decisions, business rules, and security model, refer to:

- **`AGENTS.md`** — Complete system reference (database schema, API inventory, data flows, known issues)
- **`.opencode/workflow.md`** — Agent/development workflow rules (DRY, SOLID, Clean Architecture, code quality gates)

### Quick Rules
- **Backend:** One handler file per domain. Use Swaggo godoc. Return proper HTTP codes.
- **Pagination:** Every list endpoint must use `utils.ParsePagination(c)` and return `utils.NewPaginatedResponse(data, total, p)`.
- **Transactions:** Multi-table updates (e.g., leave approval) must use `database.DB.Transaction(...)` with `repo.WithTx(tx)`.
- **Frontend:** Use `lib/api.ts` for all API calls. Reuse `DataTable` and `FilterBar`.
- **Database:** Add new models to `postgres.go` AutoMigrate. Run `ALTER` for `employee_id` varchar fixes.
- **No Duplicates:** Search existing code before creating new files. Extend existing utilities.

---

## Security Notes

| Layer | Status | Notes |
|-------|--------|-------|
| **Transport** | ⚠️ Dev uses HTTP | Use HTTPS in production |
| **Auth** | ✅ Strong | JWT + bcrypt + refresh token rotation |
| **RBAC** | ✅ Wired | `RequireRole("super_admin")` applied to destructive endpoints (`/database/*`, `/attendance/delete-all`, `/data-logs/delete-all`) |
| **Password Policy** | ✅ Enforced | Min 12 chars, history tracking, lockout support |
| **Audit Trail** | ✅ Active | `AuditMiddleware` captures all POST/PUT/PATCH/DELETE to `audit_logs` table |
| **Rate Limiting** | ❌ Missing | Add rate limiting for auth endpoints |
| **Input Sanitization** | ✅ Safe | All raw SQL uses parameterized `Where()` with `?` placeholders |
| **Test Coverage** | ❌ Missing | No `_test.go` files exist in the codebase |
| **Transaction Usage** | ⚠️ Partial | Leave operations use transactions; some multi-table updates do not |
| **Secrets** | ⚠️ Default values | Change `JWT_SECRET` and DB credentials for production |

---

## License

MIT

---

## Support

For issues or questions, contact: **support@PeopleHub.com**

---

*Document Version: 1.0*  
*Last Updated: 2026-07-19*
