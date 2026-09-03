# PeopleHub — Earned Leave (EL) Implementation Plan

> **Spec:** `Earned Leave (EL) Process, EL Balance & EL Salary Sheet` — 40 sections, date-based, ledger-driven, batch-optimized.

## 1. Inspection Summary (OLD)

| Area | File:Line | Old State |
|------|-----------|-----------|
| **Leave models** | `models/leave.go:9-31` `leave_allocation:33` `leave_type:9` | Generic `Leave` (varchar employee_id, status pending/approved), balance via `LeaveAllocation (employee_id, leave_type, year, total/used/pending)`. No EL-specific rate/max/carry/encashment. Single `LeaveType` code not per-policy. |
| **Employee** | `models/employee.go:50 joining_date`, `51 resign_date`, `83 status`, `models/separation.go:9` | `ListForSalaryProcessing:160` already handles 4-way separation OR; `JoiningDate` used for `startDay` in `salary.go:288`. No EL eligibility centralization. |
| **Payroll** | `service/salary.go:66 ProcessMonth`, `salary.go:264 calculateEmployeeSalary` | `BatchEffectiveSalaries` for increments (`198`) exists, but `ProcessMonth:73` never reads `salary_increments` temporal. `employees.gross_salary` is single mutable column, no version. Daily process `SyncLeaveLockedStatus` locks Lv. |
| **DB** | `database/postgres.go:72 alterCol 120, 171 indexes` | `DisableForeignKeyConstraintWhenMigrating:true`, partial `WHERE deleted_at IS NULL` indexes, `alterCol("salaries","employee_id")` pattern. No EL tables. |
| **Routes** | `routes/routes.go:569 salary`, `507 leave` | `api.Group("/salary").Use(AuthMiddleware)` + `api.Group("/leaves").Use(AuthMiddleware)` — no policy-specific group. |
| **Frontend** | `web/lib/api.ts:506`, `DataTable` | `salaryApi.process` pattern, no EL UI. |

---

## 2. Problem → New → Why → Perf → Business

### 2.1 No EL Policy (Hard-coded)

**OLD:** `leave_type.total_days = 12` assumed. Spec §2 forbids.  
**Problem:** Different companies need `1/1.5/2 per month`, `max 40/60`, `carry 0/20`, `encash Basic vs Gross`.  
**NEW:** `EarnedLeavePolicy` table (company-scoped, see §3 model). All accrual/encashment reads policy via `employee.company_id`.  
**Why:** Configurable without code deploy. Follows `SalaryIncrement` temporal policy pattern.  
**Performance:** Single `SELECT * FROM earned_leave_policies WHERE company_id=? AND status='active' AND effective_from <= ? AND (effective_to IS NULL OR >= ?)` per process (one row, indexed). Map `companyID → policy`.  
**Business:** HR can clone policy per branch/year.

### 2.2 No Ledger (Single Mutable Balance)

**OLD:** `LeaveAllocation.used/pending` mutable, no history, `employee.el_balance` would be similar anti-pattern Spec §6.  
**Problem:** Silent overwrites, no audit, cannot reprocess historical period.  
**NEW:** `EarnedLeaveLedger` append-only (see §3). Each row stores `opening, accrued, used, adjusted, encashed, expired, closing`, `transaction_type` enum, `reference_id` (leave/adjustment), `created_by`. Balance is `SUM(accrued+adjusted - used - encashed - expired) + opening`.  
**Why:** Audit, §7 reversible adjustments, §11 reversal. `created_at` + `transaction_date` dual for reprocessing.  
**Perf:** Batch `CreateInBatches(200)`; indexes `(employee_id, transaction_date)`, `(company_id, period)`; balance computed via `SUM` per employee in memory, not N+1.

### 2.3 Single Eligibility Duplicated

**OLD:** `ListForSalaryProcessing` OR repeated in accrual/balance/sheet.  
**NEW:** One `EarnedLeaveService.IsEligible(emp, asOfDate, policy) bool` central (§4). Checks `joining_date <= asOfDate - minServiceMonths*30`, `status=active`, `separation.date > asOfDate OR NULL`, `policy != nil`. Used by accrual, balance, encashment, salary sheet.  
**Why:** §4 DRY, Spec forbids duplication.  
**Perf:** Eligibility evaluated in-memory after bulk load of `employees + separations + policies`, zero extra queries.

### 2.4 Separation Accrual Beyond Employment

**OLD:** `salary.go:308 effectiveDays` prorates salary, but EL had no guard — accrual could run after separation.  
**NEW:** Eligibility returns `false` if `asOfDate > separationDate` or `asOfDate < joiningDate`. Accrual loop skips.  
**Why:** §5 separation handling.  
**Perf:** `separationMap := BatchSeparationMap(employeeIDs)` one query.

### 2.5 Accrual Not Idempotent

**OLD:** No EL accrual yet; `salary increments CreateBatch` had no unique guard.  
**NEW:** Unique partial `ux_earned_ledger_company_employee_period_type WHERE transaction_type='ACCRUAL'` + `ux_earned_salary_sheet_company_period WHERE status != 'reversed'`. Service does `BatchExistingLedger(company, period)` map check before calc; `CreateInBatches` with `OnConflict DoNothing` as fallback. Reprocess blocked per §25/30 unless `status != finalized` and explicit `reprocess=true`.  
**Perf:** One `SELECT employee_id FROM earned_leave_ledger WHERE company_id=? AND period=? AND transaction_type='ACCRUAL'` to build set.

### 2.6 Leave Usage Not Connected

**OLD:** `Leave` approved deducts `LeaveAllocation` but no EL ledger entry.  
**NEW:** Hook in `LeaveService.Approve` (or `EarnedLeaveService.RecordUsage`): when `leave.leave_type.code='EL'` and approved, create `LEAVE_USED` ledger row (`used = total_days`, `closing = opening - used`). Cancel → `REVERSAL`.  
**Why:** §10/11.

### 2.7 Salary Basis Hard-coded

**OLD:** No EL salary.  
**NEW:** Policy `encashment_basis` enum `BASIC, GROSS, BASIC_PLUS_ALLOWANCES, AVERAGE_3M` + `encashment_rate` (e.g. `perDay = basis / 30`). Service `CalculateEncashment(emp, days, policy)` returns `amount = days * (basis / divisor)`. Stores `salary_basis, el_rate, el_amount` in `EarnedLeaveSalaryItem` so history immutable per §14.

### 2.8 N+1 Salary Sheet

**OLD:** No EL sheet.  
**NEW:** §18-21 bulk: `Bulk Load (employees, policies, ledger latest per emp, salary structures, separations, existing sheet)` → `Build Maps (empID→balance, empID→salaryBasis)` → `Calc in memory (accrued = policy.accrual_rate clamped by max_balance, used from ledger, closing = opening+accrued-adjusted-encashed)` → `Batch Create/Upsert` (`CreateInBatches 100` + `OnConflict(company,employee,period) DoUpdates`).

---

## 3. Models (New Tables, GORM Conventions)

```go
// EarnedLeavePolicy — configurable per company, date-based
type EarnedLeavePolicy struct {
  ID uuid PK, CompanyID uuid not null index, Name varchar100,
  AccrualRate float64 `decimal(5,2)` // e.g. 1.0, 1.5 per period
  AccrualFrequency varchar20 // MONTHLY, YEARLY
  MaxBalance float64 decimal5,2
  MinServiceMonths int
  CarryForwardAllowed bool, CarryForwardLimit float64,
  EncashmentAllowed bool, EncashmentBasis varchar20 // BASIC, GROSS, BASIC_PLUS_ALLOWANCES
  EncashmentDivisor int // e.g. 30 or 26
  RoundingRule varchar20 // FLOOR, ROUND, CEIL
  EffectiveFrom type:date not null, EffectiveTo *type:date,
  Status varchar20 default:active,
  CreatedAt/UpdatedAt/DeletedAt soft, CreatedBy/UpdatedBy *uuid
}

// EarnedLeaveLedger — append-only, audit trail
type EarnedLeaveLedger struct {
  ID uuid PK, CompanyID uuid not null index, EmployeeID varchar50 not null index:idx_ledger_emp_date
  TransactionDate type:date not null, // effective/accrual date
  Period string // YYYY-MM for accrual, indexed
  TransactionType varchar30 not null // ACCRUAL, LEAVE_USED, ADJUSTMENT_ADD, ADJUSTMENT_DEDUCT, ENCASHMENT, EXPIRY, OPENING_BALANCE, REVERSAL, CARRY_FORWARD
  OpeningBalance float64 decimal6,2, Accrued float64, Used float64, Adjusted float64, Encashed float64, Expired float64, ClosingBalance float64,
  ReferenceID *string varchar50, // leave_id / adjustment_id / sheet_id
  ReferenceType varchar30, Remarks text, CreatedBy *uuid
  // indexes: (employee_id, transaction_date), (company_id, period, transaction_type) partial uniq for ACCRUAL
}

// EarnedLeaveBalance — materialized per employee per year (fast balance read), derived from ledger but cached
type EarnedLeaveBalance struct {
  ID uuid PK, CompanyID uuid not null, EmployeeID varchar50 not null uniqueIndex:ux_balance_emp_year, Year int not null same index, Period string // YYYY
  Opening float64, Accrued float64, Used float64, Adjusted float64, Encashed float64, Expired float64, Closing float64,
  LastLedgerID *uuid, UpdatedAt, DeletedAt
}

// EarnedLeaveSalarySheet — header
type EarnedLeaveSalarySheet struct {
  ID uuid PK, CompanyID uuid not null, PeriodMonth int, PeriodYear int, Period string // YYYY-MM uniqueIndex:ux_sheet_company_period
  Status varchar20 default:DRAFT // DRAFT, CALCULATED, VERIFIED, APPROVED, FINALIZED, REVERSED
  TotalEmployees int, TotalDays float64, TotalAmount float64,
  CreatedBy/ApprovedBy/FinalizedBy *uuid, CreatedAt/ApprovedAt/FinalizedAt *time, DeletedAt
}

// EarnedLeaveSalaryItem — per employee
type EarnedLeaveSalaryItem struct {
  ID uuid PK, SheetID uuid not null index, CompanyID uuid not null, EmployeeID varchar50 not null uniqueIndex:ux_item_sheet_employee (sheet_id, employee_id)
  Opening float64, Accrued float64, Used float64, Closing float64,
  EncashableDays float64, SalaryBasis float64, Rate float64, Amount float64,
  Remarks text,
}
```

**Migration:** Add to `postgres.go:72 AutoMigrate(...)` + `alterCol("earned_leave_ledger","employee_id")` etc. + `CREATE INDEX IF NOT EXISTS` per §23.

---

## 4. Repositories

```
NewEarnedLeavePolicyRepository(db)
NewEarnedLeaveLedgerRepository(db)
NewEarnedLeaveBalanceRepository(db)
NewEarnedLeaveSalaryRepository(db) // sheet + item
```

Methods (each `WithTx`):

- `Policy: FindActive(companyID, asOfDate) (*Policy, error)` // one row, `effective_from <= asOf AND (effective_to IS NULL OR >= asOf) AND status=active ORDER BY effective_from DESC LIMIT 1`
- `Ledger: BatchLatestByEmployee(companyID, employeeIDs, asOfDate) map[empID]*Ledger` // latest per emp `DISTINCT ON` ordered `employee_id, transaction_date DESC`
- `Ledger: BatchExistingAccrual(companyID, period, employeeIDs) map[empID]bool` // for idempotency `transaction_type=ACCRUAL AND period=?`
- `Ledger: CreateInBatches(rows, 200)` + `BulkUpsert` via `OnConflict(company,employee,period,transaction_type) DoNothing`
- `Balance: UpsertBalances(balances)` // `OnConflict(company,employee,year) DoUpdates`
- `SalarySheet: FindByPeriod(company, month,year)`, `CreateSheetWithItems(sheet, items)` // Tx `Create sheet + CreateInBatches items`, `UpdateStatus(id, status, user)`
- Indexes: `idx_ledger_emp_date (employee_id, transaction_date)`, `idx_ledger_company_period_type`, `ux_ledger_company_employee_period_type WHERE transaction_type='ACCRUAL'`, `ux_sheet_company_period WHERE deleted_at IS NULL`, `ux_item_sheet_employee`

---

## 5. Services (Centralized)

```
EarnedLeaveService struct {
  policyRepo, ledgerRepo, balanceRepo, salaryRepo,
  employeeRepo, separationRepo, leaveRepo, salaryIncrementRepo
  db *gorm.DB
}

- IsEligible(emp, asOfDate, policy) bool // §4 central
  -> joining: asOf >= joining + minServiceMonths*30
  -> separation: asOf <= separationDate (via BatchSeparationMap)
  -> status active
- CalculateAccrual(emp, policy, asOfDate) float64 // policy.accrual_rate clamped by max_balance - closing, rounding per policy.rounding
- GetBalance(empID, asOfDate) float64 // sum ledger: opening + accrued + adjusted - used - encashed - expired, or from Balance cache
- AccrualProcess(companyID, period YYYY-MM, userID) (*ProcessResult, error)
  -> resolve period → start/end dates
  -> acquire lock (see §6)
  -> bulk load: employees (ListForSalaryProcessing), policies, ledger latest, separations, existing accrual map
  -> build maps
  -> calc in memory: for each eligible, accrued = policy.rate, closing = opening + accrued (clamp max)
  -> batch create ledger ACCRUAL + upsert balances
  -> tx commit, release lock, return summary {eligible, processed, skipped, created, duration_ms}
- GenerateSalarySheet(companyID, month,year, userID) // §14/18 bulk: employees, policies, ledger balances, salary structures (employees.gross/basic), existing sheet check, calc amount = encashableDays * (basis/divisor), batch create sheet+items
```

---

## 6. Handlers & Routes

```
earnedLeaveHandler := NewEarnedLeaveHandler(service)

protected := api.Group("/earned-leaves")
protected.Use(AuthMiddleware)
{
  POST   "/policies"                  -> CreatePolicy
  GET    "/policies"                  -> ListPolicies
  POST   "/accrual/process"           -> AccrualProcess  // company_id, period YYYY-MM
  GET    "/balance"                   -> GetBalance       // employee_id, asOfDate
  GET    "/ledger"                    -> ListLedger       // employee_id, date range, type
  POST   "/adjustment"                -> CreateAdjustment // ADD/DEDUCT
  POST   "/encashment"                -> CreateEncashment
  POST   "/salary-sheet/generate"     -> GenerateSalarySheet
  GET    "/salary-sheet"              -> ListSalarySheets
  GET    "/salary-sheet/:id"          -> GetSalarySheet
  POST   "/salary-sheet/:id/approve"  -> ApproveSheet
  POST   "/salary-sheet/:id/finalize" -> FinalizeSheet
  POST   "/salary-sheet/:id/reverse"  -> ReverseSheet
}
```

Each handler: `ShouldBindJSON` + `binding:"required"` + `c.GetString("user_id")` company isolation check (`policy.CompanyID == req.CompanyID && user.company_id`), `201` on create, `409` on insufficient balance, `400 validation`, `500 internal` not leaking SQL. Swaggo godoc per `salary.go:41`.

**Permissions:** `EL_VIEW, EL_PROCESS, EL_ADJUST, EL_ENCASH, EL_SALARY_GENERATE, EL_APPROVE, EL_FINALIZE` via `RequirePermission` on each route group (reuse `middleware/permission.go`).

**Concurrency:** `processingLocks sync.Map` key `companyID|period` + DB advisory `SELECT pg_try_advisory_xact_lock(hashtext(key))` inside Tx; second request → `409 Already processing`.

---

## 7. Wire

`database/postgres.go:72 AutoMigrate(&models.EarnedLeavePolicy{}, &models.EarnedLeaveLedger{}, &models.EarnedLeaveBalance{}, &models.EarnedLeaveSalarySheet{}, &models.EarnedLeaveSalaryItem{})`, `alterCol` for `earned_* employee_id`, `CREATE INDEX` per §23.

`server/server.go:44` create 4 repos, `service.NewEarnedLeaveService(...)`, `handlers.NewEarnedLeaveHandler(service)`, `routes.Setup(..., earnedLeaveHandler, ...)`.

---

## 8. Frontend (Plan Only, Backend First)

Module `web/app/(root)/earned-leave/` with `policy/page.tsx`, `accrual/page.tsx` (Company + Period picker + Process button + summary), `balance/page.tsx` (DataTable), `ledger/page.tsx`, `salary-sheet/page.tsx` (generate, DataTable, approve/finalize, Excel export via `excelize` pattern).

---

## 9. Verification

`go vet ./...`, `go test ./internal/service -run TestIsEligible -run TestAccrual -run TestBalance`, `go build ./cmd/server`, `swag init`, `web npm run build`, `EXPLAIN ANALYZE` on ledger queries, load test 4k employees, 1M ledger.

