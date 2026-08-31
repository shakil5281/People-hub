# Roles & Permissions — System Audit, Architecture & Application UI Design

> PeopleHub RBAC | Audit Date: 2026-08-30 | Branch: `HEAD`

---

## 1) API Audit (As-Is)

### 1.1 Endpoints — `internal/routes/routes.go:651-667`

| Method | Path | Handler `internal/handlers/role.go:line` | Auth | Current Gate | Notes |
|---|---|---|---|---|---|
| `GET` | `/api/v1/roles` | `List:19` | `AuthMiddleware` | none | `?company_id=` else JWT `company_id`. Returns `{data: Role[]}`. No pagination, no search. |
| `GET` | `/api/v1/roles/:id` | `GetByID:32` | `AuthMiddleware` | none | Returns `{role, permissions}`. `GetRolePermissions` `repository/role.go:40` via join. |
| `POST` | `/api/v1/roles` | `Create:51` | `AuthMiddleware` | none | `CreateRoleRequest` `name*, description`. Sets `company_id=JWT`, `created_by=JWT`. No uniqueness check, no `is_system` guard (always `false`). |
| `PUT` | `/api/v1/roles/:id` | `Update:78` | `AuthMiddleware` | `is_system` → `403` | Partial update `name*, description`. |
| `DELETE` | `/api/v1/roles/:id` | `Delete:111` | `AuthMiddleware` | `is_system` → `403` | Soft-delete (`gorm.DeletedAt`). No cascade check (users still referencing `user_roles`). |
| `PUT` | `/api/v1/roles/:id/permissions` | `AssignPermissions:142` | `AuthMiddleware` | none | `permission_ids*`. Replaces via `ReplaceRolePermissions` `repo/role.go:73` (DELETE + INSERT in Tx). No existence validation, no audit log. |
| `GET` | `/api/v1/permissions` | `ListPermissions:129` | `AuthMiddleware` | none | Returns `{data: Permission[]}` ordered `resource, action`. No pagination/filter. |

**User↔Role:**
`PUT /users/:id/roles` `handlers/user.go:224` → `service.UserService.AssignRoles` — checked `currentRoles` superset but never `RequirePermission("users.assignRoles")`.

### 1.2 Data Model — `internal/models/*.go`

```
User --< user_roles >-- Role --< role_permissions >-- Permission
  id (uuid PK)            id (uuid PK)                    id (uuid PK)
  email, name, status     company_id* (nullable FK)       resource varchar(50)  e.g. "employees"
  ...                     name varchar(100)                action varchar(50)    e.g. "create"
                          description                       description
                          is_system bool                    created_at
                          deleted_at
```
* `roles.company_id` nullable → system-global roles (`is_system=true`, `company_id=NULL`) shared across tenants. `ListByCompany(companyID)` `repo/role.go:61` filters `company_id=?` so global roles invisible to non-null queries — they appear only if explicitly queried with empty string. Seed `cmd/superadmin/main.go:56-108`, `cmd/reset/main.go:165-218` creates `superadmin` (system) + 30-40 `resource.action` permissions.
* No `UNIQUE(company_id, name)` constraint — duplicate names possible.
* `permissions` unique is app-level (`resource+action`).

### 1.3 Auth Flow — `internal/auth/jwt.go:20-46`, `internal/service/auth.go:55-164`

```
Login → GetUserRoles + GetUserPermissions → claims.roles=[name], claims.permissions=[resource.action]
        → GenerateAccessToken(15m) {sub, email, company_id, roles, permissions}
        → GenerateRefreshToken(7d) hash stored in refresh_tokens
AuthMiddleware → ValidateAccessToken → c.Set(permissions, roles)
RequirePermission/RequireRole → slices.Contains(perm, "*") bypass
Refresh → re-read roles/permissions → new access token
```

JWT bloat: every request carries `*` expansion if superadmin (`permissions=["*"]`). No short-circuit for large permission sets.

### 1.4 Middleware Gaps

* `internal/middleware/permission.go:10,36` exists but **never mounted** on admin/role routes (comment in `.opencode/workflow.md:107` confirms). Any authenticated user can CRUD roles.
* `RequirePermission` checks `permissions` slice exact match `resource.action`; seed creates `users`, `roles`, `permissions` but UI never declares `resource.action` catalog. No `GET /permissions` grouping in handler.
* `WriteLog`/audit `audit.go:18` fires only for 2xx mutating; role permission changes audited implicitly but not explicitly diffed.

### 1.5 Repository Gaps

* `ReplaceRolePermissions` does not validate `permissionIDs` exist → orphan FK if bad UUID.
* `ListByCompany` does not preload `Permissions` nor user count; `GetRolePermissions` N+1 if listing many roles.
* No `CountUsersByRole`, no `Search`, no pagination.

### 1.6 Frontend — `web/app/(root)/admin/roles/page.tsx:1-248`, `web/lib/api.ts:67-78`

* `fetchData` loads `roleApi.list()` + `permissionApi.list()` in parallel — ok.
* `DataTable` row is not clickable; must add explicit "Manage" button (currently table has only delete icon, no edit/manage). `onEdit` undefined, delete only for non-system.
* `Dialog` groups permissions by `resource` `page.tsx:118-126` — good, but no **Select All per resource**, no search/filter, no count badges, no unsaved-changes diff.
* `handleCreate` posts `{name, description}` — no sanitized slug (`hr_manager` vs `HR Manager`).
* `users/page.tsx:173-187` shows read-only badges for assigned roles, no bulk assignment.

---

## 2) Target Architecture (Principles: DRY, SOLID, Clean)

### 2.1 RBAC Model

```
Company (tenant)
  └─ Role {id, company_id?, name, description, is_system, created_by}
       └─ Permission {resource, action, description}
User ──< user_roles (user_id, role_id)  [soft-delete aware]
```

* `is_system` → cannot be deleted/renamed, visible to all companies (`company_id IS NULL`).
* Custom roles are company-scoped (`company_id = JWT company_id`). Enforce `UNIQUE INDEX WHERE deleted_at IS NULL (COALESCE(company_id,'00000000-0000-0000-0000-000000000000'), name)`.
* Permission canonical form `resource.action` (e.g., `employees.create`, `attendance.process`, `salary.export`). Seed from `cmd/superadmin` enum.
* User effective permissions = `UNION(DISTINCT role.permissions)` resolved at login/refresh and cached in JWT; refresh forced on role change (revoke refresh tokens).

### 2.2 API Contract (Proposed)

| Method | Path | Gate | Body/Query | Response |
|---|---|---|---|---|
| `GET` | `/roles` | `roles.list` | `?company_id=&q=&include_system=true&page=&limit=` | `Paginated{data: Role+permCount+userCount}` |
| `GET` | `/roles/:id` | `roles.read` | — | `{role, permissions, users:[]}` |
| `POST` | `/roles` | `roles.create` | `{name*, description}` | `201 Role` |
| `PUT` | `/roles/:id` | `roles.update` | `{name, description}` | `200 Role` |
| `DELETE` | `/roles/:id` | `roles.delete` | — | `200 {message}` |
| `PUT` | `/roles/:id/permissions` | `roles.assignPermissions` | `{permission_ids*}` | `200 {message}` |
| `GET` | `/permissions` | `permissions.list` | `?resource=&q=` | `{data: Permission[], grouped: {resource: Permission[]}}` |
| `GET` | `/roles/:id/users` | `roles.read` | — | `{data: User[]}` |
| `PUT` | `/users/:id/roles` | `users.assignRoles` | `{role_ids*}` | `200 {message}` |

Validation: `name` regex `^[a-z0-9_]{3,50}$` lower-snake, uniqueness; `permission_ids` must exist and be active; `is_system` guard; audit log `user_id, company_id, action, resource, before/after`.

### 2.3 Route Protection

Mount `RequirePermission` per group (extend `routes.go`):

```go
roleR := api.Group("/roles", AuthMiddleware(jwt))
roleR.GET("", RequirePermission("roles.list"), handler.List)
roleR.POST("", RequirePermission("roles.create"), handler.Create)
// …
permR := api.Group("/permissions", AuthMiddleware(jwt), RequirePermission("permissions.list"))
```

Super-admin `["*"]` bypass remains in `RequirePermission` `permission.go:26`.

### 2.4 JWT & Refresh

On `AssignPermissions`/`AssignRoles`, call `authRepo.RevokeAllUserTokens(affectedUserIDs)` or at least bump `users.permission_version` and short-circuit `ValidateAccessToken` to re-issue on next refresh. Current behavior leaves old JWT valid for up to 15m.

### 2.5 Performance & Indexes

```sql
CREATE UNIQUE INDEX ux_roles_company_name ON roles(COALESCE(company_id::text,''), name) WHERE deleted_at IS NULL;
CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);
CREATE INDEX idx_user_roles_user ON user_roles(user_id) WHERE deleted_at IS NULL;
```

Add repo methods: `ListWithCounts(companyID, q, page, limit)`, `CountUsersByRole`, `SearchPermissions`.

### 2.6 Observability

Audit each mutating role call via `AuditMiddleware` plus explicit diff in `ReplaceRolePermissions` (before/after ids). SystemLog `source="rbac"`.

---

## 3) Application UI Design

### 3.1 Information Architecture

```
Administrator → Roles & Permissions (/admin/roles)
                ├─ Roles (tabs: All | System | Custom)
                ├─ Permissions Matrix
                └─ User Assignment (/admin/users → Manage Roles)
```

### 3.2 Page: Roles & Permissions — `web/app/(root)/admin/roles/page.tsx`

**Layout (3-tier):**

```
Header: [ShieldCheck] Roles & Permissions — subtitle + Primary [Add Role]
StatsRow: 4 Cards [Total Roles, System Roles, Custom Roles, Permissions]
FilterBar: [Search q] [Segmented: All/System/Custom] [GroupBy toggle] [Export]
Table: DataTable< Role >
  Sl | Name (capitalize + slug hint) | Description | Type (System/Custom Badge)
     | Permissions (count badge + overflow) | Users (count badge)
     | Created (dd-MM-yyyy) | Actions: [Manage] [Edit] [Delete]
Row click → open Manage Permissions dialog
Empty: “No roles — create your first custom role”
Skeleton: 6 rows shimmer
```

**Dialog A — Manage Permissions (max-w-3xl, 80vh scroll):**

```
Title: <RoleName> — Description
Search: [Filter permissions...] + [Select All] [Clear]
Sections: Accordion per resource (employees, attendance, leave, payroll, ... )
  Header: RESOURCE + [n/m selected] + [✓ select-all per resource]
  Grid 2 cols: [☐ action — description]
Footer: [Cancel] [Save Changes] (disable if no diff, show spinner)
Toast on success + revoke notice: “Users with this role will get new permissions on next login/refresh”
```

**Dialog B — Create/Edit Role (max-w-md):**

```
Name*: Input slugify live (hr_manager), hint “lowercase, underscores”
Description: Textarea 120 chars
Preview: Example JWT claim chip: hr_manager
Validation: required, min 3, uniqueness async (409 → “Role name already exists in this company”)
Buttons: Cancel / Create / Save (loading)
```

**Dialog C — Delete Confirm:**

```
Are you sure? Custom role deletion removes it from {n} users. System roles cannot be deleted.
[Cancel] [Delete] destructive
```

**States:**
* `loading`, `error` (retry), `saving` (disable), `is_system` (lock name/delete).
* Permission grouping memoized: `useMemo(() => groupBy(resource), [allPermissions])` — already present `page.tsx:118`.
* Diff detection: `initialPermsRef` vs `rolePermissions` → enable Save only if changed.

### 3.3 Page: User Management — `web/app/(root)/admin/users/page.tsx`

Add **Manage Roles** drawer:

```
Trigger: Shield icon per row (already exists  page.tsx:242) → opens Dialog
Content:
  Available Roles: Multi-select checkboxes grouped System/Custom
  Selected badges: removable chips
  Search roles filter
  Info: “Permissions are union of all roles. Changes take effect after user re-login”
  Save: PUT /users/:id/roles {role_ids}
```

Enhancement: column `Roles` (badge list) in users table, similar to current but make `allRoles` actually fetched from `/roles` (currently bug: `openRoles` `page.tsx:179-182` reuses `userData.roles` as `allRoles`).

### 3.4 Component Library Reuse

* `DataTable` `components/table/data-table.tsx` — serverSide for users, client for roles.
* `FilterBar` `components/filter-bar.tsx` for search + segmented.
* `Badge`, `Dialog`, `Checkbox`, `Accordion`, `Skeleton`, `Input`, `Textarea`, `Button` from `components/ui/*`.
* `cn()` `lib/utils.ts` for conditional classes.
* `sonner` toast for feedback.

### 3.5 Permission Catalog (Seed)

Resources: `users, roles, permissions, companies, employees, groups, floors, departments, sections, designations, lines, divisions, districts, upazilas, unions, post-offices, shifts, temporary-shifts, rosters, attendance, data-logs, leaves, leave-types, holidays, salaries, tiffin-bills, night-bills, separations, requirements, migrations, id-cards, punishments, daily-schedules, dashboard, database, settings, notifications, system-logs`.

Actions: `create, read, update, delete, list, export, approve, reject, assign, process, sync` — seed generates `resource.action` (e.g., `employees.create`). Total ~70 permissions — rendered grouped.

### 3.6 Accessibility & Polish

* Keyboard: `Enter` creates role, `Esc` closes dialog, focus trap.
* Aria: `role`, `aria-label` on icon buttons.
* Contrast: System badge `secondary`, Custom `outline`.
* Mobile: StatsRow wraps `grid grid-cols-2 lg:grid-cols-4`, table horizontal scroll.

### 3.7 Future Iterations (not over-optimized)

* Redis cache `permissions:user:<id>` TTL 15m if DB perm lookup becomes hot.
* `PATCH /roles/:id` partial, `POST /roles/clone`.
* Audit timeline view in settings.

---

## 4) Definition of Done

- [ ] `UNIQUE(company_id,name)` index + handler 409 on duplicate
- [ ] `RequirePermission` mounted on every `/roles`, `/permissions`, `/users/:id/roles`
- [ ] `ReplaceRolePermissions` validates IDs exist, transactional, audited, revokes refresh tokens
- [ ] `List` paginated + `q` search + counts
- [ ] UI: Search, filter, stats, permission matrix with select-all per resource, diff-aware Save
- [ ] Frontend `roleApi.list` uses search param, `users` column shows roles
- [ ] `is_system` locked in UI and API (`403`)

---

*Generated for PeopleHub — apply via `internal/routes/routes.go`, `internal/handlers/role.go`, `internal/repository/role.go`, `web/app/(root)/admin/roles/page.tsx`, `web/app/(root)/admin/users/page.tsx`.*
