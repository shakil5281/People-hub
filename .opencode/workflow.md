# OpenCode Workflow & Agent Rules — PeopleHub

> **Purpose:** This file governs how OpenCode interacts with the PeopleHub project. Every change must pass the rules below. Read this file before every task.

---

## 1. Universal Agent Rules (Non-Negotiable)

### 1.1 DRY — Do Not Repeat Yourself
- **Never** write duplicate code.
- **Never** create a new function if an existing one in the project does the same thing.
- **Never** create a new component, service, utility, or repository if one already exists.
- If you find yourself copy-pasting code, **refactor it into a shared helper** instead.

### 1.2 Reuse First, Create Second
Before creating **any** new file, you **MUST**:
1. Search the project for existing implementations using `grep`, `glob`, or file reads.
2. Check if a similar model, repository, handler, service, or utility already exists.
3. If similar functionality exists, **extend it** (add a method, add a parameter, use generics).
4. Only create a new file if the existing one cannot reasonably accommodate the new behavior.

### 1.3 Search Before You Write
- Use `grep` to find existing patterns.
- Use `glob` to discover file naming conventions.
- Read the most similar existing file before writing a new one.
- Example: Before creating a new repository for "Advance Salary", read `internal/repository/salary.go` and `internal/repository/leave.go` to follow the exact same pattern.

### 1.4 SOLID Principles
- **Single Responsibility:** One file/class/function should do one thing.
- **Open/Closed:** Extend behavior without modifying existing code where possible.
- **Liskov Substitution:** Substitutable implementations (use interfaces where appropriate).
- **Interface Segregation:** Do not force handlers to depend on things they don't use.
- **Dependency Inversion:** Depend on abstractions (repositories, services) not concrete details.

### 1.5 Clean Architecture
- **Models** know nothing about HTTP or business logic.
- **Repositories** know nothing about HTTP or business rules (only data access).
- **Services** contain business logic (auth, external readers, complex calculations).
- **Handlers** deal with HTTP only: binding, validation, calling services/repos, returning status codes.
- **Middleware** deals with cross-cutting concerns (auth, logging, CORS, permissions).
- Business logic must **NOT** leak into handlers beyond simple orchestration. Complex calculations belong in `service/`.

### 1.6 Clean Code Checklist
- Use **meaningful variable and function names** (`FindByEmployeeID`, not `GetData`).
- Keep **functions small** (under 50 lines ideally, under 100 lines absolutely).
- **Separate business logic** from infrastructure concerns.
- **Avoid magic numbers/strings** — use constants or configuration.
- **Comment only "why", not "what"** — code should be self-documenting.
- **No dead code** — remove commented-out blocks before finishing.

---

## 2. PeopleHub-Specific Conventions

### 2.1 Go Backend Conventions

#### Models (`internal/models/`)
- GORM struct tags for both JSON and DB.
- Primary keys: `uuid` with `default:gen_random_uuid()`.
- Soft delete: `gorm.DeletedAt` with `json:"-"`.
- Audit fields: `created_by`, `updated_by`, `deleted_by` as `*string` with `gorm:"type:uuid"`.
- Foreign key relations: use `gorm:"foreignKey:..."` but remember **no actual DB FK constraints exist**.

#### Repositories (`internal/repository/`)
- One file per aggregate root.
- Constructor: `NewXxxRepository(db *gorm.DB) *XxxRepository`.
- Standard methods: `Create`, `FindByID`, `List`, `Update`, `Delete`.
- Use `Preload()` for relationships.
- For list methods, accept filter structs or variadic parameters.
- Raw SQL is acceptable for aggregations, but **MUST use parameter placeholders** (`?`) — never string concatenation for user input.

#### Handlers (`internal/handlers/`)
- One file per domain.
- Request DTOs named `CreateXxxRequest`, `UpdateXxxRequest`.
- Use `c.ShouldBindJSON(&req)` for binding.
- HTTP status codes:
  - `201 Created` for new resources
  - `200 OK` for updates/reads
  - `400 Bad Request` for validation errors
  - `401 Unauthorized` for auth failures
  - `403 Forbidden` for permission failures
  - `404 Not Found` for missing resources
  - `409 Conflict` for business rule violations (e.g., insufficient leave balance)
  - `500 Internal Server Error` for unexpected errors
- Return errors as `gin.H{"error": "message"}`.
- Every handler method **must** have Swaggo godoc comments.

#### Services (`internal/service/`)
- Business logic lives here, not in handlers.
- `auth.go` handles login, register, refresh, password changes.
- `mdb_reader.go` handles ZKTeco integration.
- If a new domain has complex rules (e.g., loan eligibility), create a service.

#### Routes (`internal/routes/routes.go`)
- All routes in one file.
- Group by domain with `api.Group("/xxx")`.
- Apply `middleware.AuthMiddleware(jwtSecret)` to protected groups.
- Apply `middleware.RequirePermission(...)` or `middleware.RequireRole(...)` where RBAC is needed.

#### Server Wiring (`internal/server/server.go`)
- Manual dependency injection.
- Order: DB → JWT config → Repositories → Services → Handlers → Routes.
- When adding a new feature, add the repo and handler here, then pass to `routes.Setup(...)`.

#### Middleware (`internal/middleware/`)
- `auth.go`: Bearer token validation, sets context values.
- `permission.go`: RBAC checks. Currently **not applied to routes** — must be explicitly added when needed.
- `cors.go`: CORS headers.
- `logger.go`: Request logging.

### 2.2 Next.js Frontend Conventions

#### API Layer (`web/lib/`)
- **Always** use `lib/api.ts` functions. **Never** call `axios` directly from pages.
- Add new domain API functions to `lib/api.ts` following existing patterns.
- `axios-instance.ts` handles auth headers and token refresh. Do not modify unless fixing bugs.

#### Pages (`web/app/(root)/` and `web/app/(auth)/`)
- Use `"use client"` for pages with data fetching.
- Use `next/navigation` `useRouter()` for navigation.
- Use `sonner` `toast` for notifications.

#### Components
- Reuse `DataTable` from `components/table/` for all list views.
- Reuse `FilterBar` from `components/` for filter UIs.
- Reuse shadcn/ui components from `components/ui/`.
- Use `cn()` from `lib/utils.ts` for conditional Tailwind classes.
- Use Lucide React for icons.

#### Forms
- Use React Hook Form + Zod where possible.
- Use existing form abstractions in `components/form/`.

---

## 3. Business Logic & Condition Rules

### 3.1 Employee & Attendance Rules
- `employees.employee_id` (VARCHAR business key) is used for attendance, salary, and leave foreign references. `employees.id` (UUID) is the PK but rarely used in child tables.
- ZKTeco badge matching: **primary** match is `punch_number`, **fallback** is `employee_id`.
- Attendance status hierarchy (highest to lowest precedence):
  1. `on_leave` (if approved leave covers the date)
  2. `weekend` (if shift weekend_days matches)
  3. `present` / `late` / `absent` / `half_day` (derived from punches)
- Temporary shifts override regular shifts for attendance processing.

### 3.2 Leave Rules
- Leave balance formula: `remaining = total_days - used_days - pending_days`.
- Only `pending` leaves can be updated, cancelled, approved, or rejected.
- Approving a leave: `pending_days -= total_days`, `used_days += total_days`, mark attendance as `on_leave`.
- Rejecting a leave: `pending_days -= total_days` (min 0).

### 3.3 Payroll Rules
- Salary calculations are derived from `gross_salary`:
  - `basic_salary = gross * 0.5`
  - `house_rent = gross * 0.25`
  - `medical_allowance = gross * 0.1`
  - `transport_allowance = employee.transport_allowance` (default 450)
  - `food_allowance = employee.food_allowance` (default 1250)
- Absent deduction: `per_day = gross / total_days_in_month`, then `per_day * absent_days`.
- OT rate: `basic / total_days / 8`.
- Attendance bonus: 500 BDT if `absent_days == 0` AND `present_days > 0`.
- PF and Tax are always 0 (placeholders).
- Net salary: `gross - absent_deduction + ot_amount + attendance_bonus` (min 0).

### 3.4 Auth Rules
- Access token TTL: 15 minutes.
- Refresh token TTL: 7 days.
- Password minimum: 12 characters.
- Password history prevents reuse.
- Account lockout tracked via `failed_attempts` and `locked_at`.
- Force password change flag: `force_password_change`.

---

## 4. Feature Development Workflow

When asked to add a new feature (e.g., "Loan", "Advance Salary", "Training", "Complaints"):

### Step 1: Research (Mandatory)
1. Read `AGENTS.md` Section 15 (Quick Reference).
2. Search existing code for similar features using `grep`.
3. Identify the closest existing domain (e.g., `leave` is closest to `loan`).
4. Read the model, repository, handler, and route files for the closest domain.

### Step 2: Backend Implementation
1. **Model:** Create `internal/models/<feature>.go` following the pattern of the closest existing model.
2. **Repository:** Create `internal/repository/<feature>.go` with `New`, `Create`, `FindByID`, `List`, `Update`, `Delete`. Copy the pattern from `repository/leave.go` or `repository/salary.go`.
3. **Service:** Only if the feature has complex business rules. Otherwise, handler can call repository directly.
4. **Handler:** Create `internal/handlers/<feature>.go`. Follow `leave.go` or `salary.go` patterns.
   - Define request DTOs.
   - Add Swaggo godoc comments.
   - Return correct HTTP status codes.
5. **Routes:** Register in `internal/routes/routes.go` inside an `api.Group("/<feature>s")` with auth middleware.
6. **Server Wiring:** Add repository and handler in `internal/server/server.go`, pass to `routes.Setup(...)`.
7. **Migration:** Add model to `postgres.go` AutoMigrate list.
   - If the model has a column named `*_id` that should be `varchar(50)` instead of UUID, add an `ALTER TABLE` in `postgres.go`.
8. **Swagger:** Regenerate with `swag init -g cmd/server/main.go -o docs`.

### Step 3: Frontend Implementation
1. **API:** Add functions to `web/lib/api.ts`.
2. **Page:** Create `web/app/(root)/<feature>s/page.tsx`.
   - Use `DataTable` for lists.
   - Use `FilterBar` for filters if needed.
   - Use `useEffect` + `api.ts` for data fetching.
3. **Navigation:** Add sidebar link in the layout component (if applicable).

### Step 4: Verification (Mandatory)
1. Build the backend: `go build ./cmd/server`.
2. Build the frontend: `cd web && npm run build`.
3. Verify no duplicate files were created.
4. Verify existing tests (if any) still pass.
5. Review the diff to ensure minimal changes.

---

## 5. File Creation Rules

### 5.1 When to Create a New File
- A new domain aggregate root needs a model (e.g., "Loan" model does not exist).
- A new repository is needed because no existing repository manages the new entity.
- A new service is needed because business logic is complex and not covered by existing services.
- A new handler is needed because a new API domain is required.
- A new frontend page is needed because a new UI screen is required.

### 5.2 When NOT to Create a New File
- **Do not** create a new utility for string manipulation if `lib/utils.ts` or existing helpers cover it.
- **Do not** create a new repository method in a new file if the existing repository file can hold it.
- **Do not** create a new middleware if the existing `auth`, `cors`, `logger`, or `permission` middleware can be extended.
- **Do not** create a new generic CRUD handler if the existing patterns cover it.
- **Do not** create a new database connection wrapper — use `database.DB` everywhere.

### 5.3 Naming Conventions for New Files
- Models: `internal/models/<domain>.go` (e.g., `loan.go`)
- Repositories: `internal/repository/<domain>.go` (e.g., `loan.go`)
- Handlers: `internal/handlers/<domain>.go` (e.g., `loan.go`)
- Services: `internal/service/<domain>.go` (e.g., `loan.go`)
- Frontend pages: `web/app/(root)/<domain>s/page.tsx` (e.g., `loans/page.tsx`)
- Frontend components: `web/components/<domain>/` (only if reusable)

---

## 6. Code Quality Gates (Check Before Finishing)

Before completing any task, verify:

- [ ] **No duplicate code:** I searched the project and did not create redundant implementations.
- [ ] **Reused existing code:** I extended existing services, repositories, or utilities where possible.
- [ ] **Clean Architecture:** Business logic is in service or repository, not mixed with HTTP concerns in handlers.
- [ ] **Small functions:** No function exceeds 100 lines.
- [ ] **Proper naming:** Variables and functions have meaningful names.
- [ ] **HTTP status codes:** Correct status codes returned (201, 200, 400, 401, 403, 404, 409, 500).
- [ ] **Swagger docs:** Every new handler method has Swaggo godoc comments.
- [ ] **GORM models:** Proper JSON and DB tags, soft delete, audit columns where appropriate.
- [ ] **Error handling:** Errors are returned to the client, not swallowed.
- [ ] **Input validation:** Request DTOs use `binding:"required"` and similar tags.
- [ ] **No dead code:** No commented-out blocks remain.
- [ ] **Build passes:** `go build ./cmd/server` succeeds.
- [ ] **Frontend builds:** `cd web && npm run build` succeeds.
- [ ] **Consistent with existing style:** Code looks like it was written by the same author as the rest of the project.

---

## 7. Forbidden Actions

- **Never** run `git commit`, `git push`, `git reset`, `git rebase`, or any git mutations unless explicitly asked.
- **Never** delete or modify `.env` without explicit instruction.
- **Never** run `docker-compose down -v` or destructive Docker commands without confirmation.
- **Never** modify `go.mod` or `package.json` unless adding a required dependency.
- **Never** change the database schema (ALTER, DROP, TRUNCATE) in production-like environments without backup confirmation.
- **Never** expose secrets (JWT_SECRET, DB_PASS) in code comments or API responses.
- **Never** create a file outside the project root (`F:\softwer\PeopleHub`).

---

## 8. Agent Self-Check Prompt

Before generating any code, ask yourself:

1. "Has this been done before in the project?"
2. "Can I extend an existing file instead of creating a new one?"
3. "Am I duplicating logic that already exists in a repository, service, or utility?"
4. "Is my business logic separated from HTTP handling?"
5. "Am I following the exact pattern used by `leave`, `salary`, or `attendance`?"
6. "Will this build and run without errors?"

If the answer to question 1 or 3 is "yes", **stop and refactor**.

---

*Version: 1.0*  
*Effective: 2026-07-19*  
*Applies to: All OpenCode sessions on PeopleHub*
