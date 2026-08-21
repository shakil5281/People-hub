package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
)

type ZKTecoSyncHandler struct {
	employeeRepo *repository.EmployeeRepository
}

func NewZKTecoSyncHandler(employeeRepo *repository.EmployeeRepository) *ZKTecoSyncHandler {
	return &ZKTecoSyncHandler{employeeRepo: employeeRepo}
}

type ZKTecoSyncRequest struct {
	MDBPath     string   `json:"mdb_path"`
	EmployeeIDs []string `json:"employee_ids"`
	SyncAll     bool     `json:"sync_all"`
}

type ZKTecoSyncItemResult struct {
	EmployeeID  string `json:"employee_id"`
	PunchNumber string `json:"punch_number"`
	Name        string `json:"name"`
	Status      string `json:"status"` // "synced", "failed", "skipped"
	Message     string `json:"message"`
}

type ZKTecoEmployeeStatus struct {
	ID          string `json:"id"`
	EmployeeID  string `json:"employee_id"`
	PunchNumber string `json:"punch_number"`
	Name        string `json:"name"`
	Department  string `json:"department"`
	Designation string `json:"designation"`
	IsSynched   bool   `json:"is_synched"`
}

type syncItemInput struct {
	EmployeeID  string `json:"emp_id"`
	PunchNumber string `json:"punch"`
	Name        string `json:"name"`
}

// GetStatus godoc
// @Summary      Get ZKTeco Sync Status
// @Tags         Admin
// @Produce      json
// @Param        mdb_path query string false "Path to att2000.mdb"
// @Success      200  {object}  map[string]interface{}
// @Router       /admin/zkteco-sync/status [get]
func (h *ZKTecoSyncHandler) GetStatus(c *gin.Context) {
	requestedPath := c.DefaultQuery("mdb_path", `C:\Program Files (x86)\ZKTeco\att2000.mdb`)
	if requestedPath == "" {
		requestedPath = `C:\Program Files (x86)\ZKTeco\att2000.mdb`
	}

	targetPath, pathErr := resolveMDBPath(requestedPath)
	mdbExists := pathErr == nil

	employees, err := h.employeeRepo.GetAllActive()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	syncedBadges := make(map[string]bool)
	if mdbExists {
		badges, err := getMDBUserBadgesFast(targetPath)
		if err == nil {
			for _, b := range badges {
				syncedBadges[strings.TrimSpace(b)] = true
			}
		}
	}

	var list []ZKTecoEmployeeStatus
	for _, emp := range employees {
		punch := strings.TrimSpace(emp.PunchNumber)
		if punch == "" {
			punch = strings.TrimSpace(emp.EmployeeID)
		}
		dept := ""
		if emp.Department != nil {
			dept = emp.Department.Name
		}
		desig := ""
		if emp.DesignationRef != nil {
			desig = emp.DesignationRef.Name
		}

		isSynced := syncedBadges[punch] || syncedBadges[emp.EmployeeID]

		list = append(list, ZKTecoEmployeeStatus{
			ID:          emp.ID,
			EmployeeID:  emp.EmployeeID,
			PunchNumber: punch,
			Name:        emp.NameEn,
			Department:  dept,
			Designation: desig,
			IsSynched:   isSynced,
		})
	}

	resPath := targetPath
	if !mdbExists {
		resPath = requestedPath
	}

	c.JSON(http.StatusOK, gin.H{
		"mdb_path":     resPath,
		"mdb_exists":   mdbExists,
		"total":        len(list),
		"synced_count": countSynced(list),
		"data":         list,
	})
}

// TestConnection godoc
// @Summary      Test ZKTeco MDB Connection
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /admin/zkteco-sync/test-connection [post]
func (h *ZKTecoSyncHandler) TestConnection(c *gin.Context) {
	var body struct {
		MDBPath string `json:"mdb_path"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.MDBPath == "" {
		body.MDBPath = `C:\Program Files (x86)\ZKTeco\att2000.mdb`
	}

	targetPath, pathErr := resolveMDBPath(body.MDBPath)
	if pathErr != nil {
		c.JSON(http.StatusOK, gin.H{
			"connected": false,
			"mdb_path":  body.MDBPath,
			"message":   pathErr.Error(),
		})
		return
	}

	badges, connErr := getMDBUserBadgesFast(targetPath)
	if connErr != nil {
		c.JSON(http.StatusOK, gin.H{
			"connected": false,
			"mdb_path":  targetPath,
			"message":   fmt.Sprintf("Database file found at '%s', but connection failed: %v", targetPath, connErr),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"connected":  true,
		"mdb_path":   targetPath,
		"user_count": len(badges),
		"message":    fmt.Sprintf("Successfully connected to att2000.mdb at '%s'! Found %d existing users in USERINFO table.", targetPath, len(badges)),
	})
}

// Sync godoc
// @Summary      Sync Employees to ZKTeco att2000.mdb
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /admin/zkteco-sync [post]
func (h *ZKTecoSyncHandler) Sync(c *gin.Context) {
	var req ZKTecoSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	targetPath, pathErr := resolveMDBPath(req.MDBPath)
	if pathErr != nil {
		c.JSON(http.StatusOK, gin.H{
			"success":      false,
			"synced_count": 0,
			"failed_count": len(req.EmployeeIDs),
			"message":      pathErr.Error(),
			"details":      []ZKTecoSyncItemResult{},
		})
		return
	}

	var employees []models.Employee
	var err error
	if req.SyncAll || len(req.EmployeeIDs) == 0 {
		employees, err = h.employeeRepo.GetAllActive()
	} else {
		employees, err = h.employeeRepo.GetByIDs(req.EmployeeIDs)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(employees) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"synced_count": 0,
			"message":      "No employees found to sync",
			"details":      []ZKTecoSyncItemResult{},
		})
		return
	}

	var batchItems []syncItemInput
	for _, emp := range employees {
		punch := strings.TrimSpace(emp.PunchNumber)
		if punch == "" {
			punch = strings.TrimSpace(emp.EmployeeID)
		}
		name := strings.TrimSpace(emp.NameEn)
		batchItems = append(batchItems, syncItemInput{
			EmployeeID:  emp.EmployeeID,
			PunchNumber: punch,
			Name:        name,
		})
	}

	details, syncErr := syncBatchUsersToMDB(targetPath, batchItems)
	if syncErr != nil {
		c.JSON(http.StatusOK, gin.H{
			"success":      false,
			"synced_count": 0,
			"failed_count": len(employees),
			"total":        len(employees),
			"mdb_path":     targetPath,
			"message":      syncErr.Error(),
			"details":      []ZKTecoSyncItemResult{},
		})
		return
	}

	syncedCount := 0
	failedCount := 0
	for _, item := range details {
		if item.Status == "synced" {
			syncedCount++
		} else if item.Status == "failed" {
			failedCount++
		}
	}

	// Re-read the MDB badges and build updated status so frontend gets real-time data
	allEmployees, _ := h.employeeRepo.GetAllActive()
	syncedBadges := make(map[string]bool)
	badges, badgeErr := getMDBUserBadgesFast(targetPath)
	if badgeErr == nil {
		for _, b := range badges {
			syncedBadges[strings.TrimSpace(b)] = true
		}
	}

	var statusData []ZKTecoEmployeeStatus
	statusSyncedCount := 0
	for _, emp := range allEmployees {
		punch := strings.TrimSpace(emp.PunchNumber)
		if punch == "" {
			punch = strings.TrimSpace(emp.EmployeeID)
		}
		dept := ""
		if emp.Department != nil {
			dept = emp.Department.Name
		}
		desig := ""
		if emp.DesignationRef != nil {
			desig = emp.DesignationRef.Name
		}
		isSynced := syncedBadges[punch] || syncedBadges[emp.EmployeeID]
		if isSynced {
			statusSyncedCount++
		}
		statusData = append(statusData, ZKTecoEmployeeStatus{
			ID:          emp.ID,
			EmployeeID:  emp.EmployeeID,
			PunchNumber: punch,
			Name:        emp.NameEn,
			Department:  dept,
			Designation: desig,
			IsSynched:   isSynced,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":            failedCount == 0,
		"synced_count":       syncedCount,
		"failed_count":       failedCount,
		"total":              len(employees),
		"mdb_path":           targetPath,
		"message":            fmt.Sprintf("Synced %d of %d employees to ZKTeco MDB USERINFO table at '%s'.", syncedCount, len(employees), targetPath),
		"details":            details,
		"status_data":        statusData,
		"status_synced_count": statusSyncedCount,
		"status_total":       len(allEmployees),
		"mdb_exists":         true,
	})
}

func countSynced(list []ZKTecoEmployeeStatus) int {
	cnt := 0
	for _, item := range list {
		if item.IsSynched {
			cnt++
		}
	}
	return cnt
}

func resolveMDBPath(userPath string) (string, error) {
	if userPath != "" {
		cleaned := filepath.Clean(strings.TrimSpace(userPath))
		if _, err := os.Stat(cleaned); err == nil {
			return cleaned, nil
		}
	}

	homeDir, _ := os.UserHomeDir()
	desktopMDB := ""
	if homeDir != "" {
		desktopMDB = filepath.Join(homeDir, "Desktop", "att2000.mdb")
	}

	candidates := []string{
		`C:\Program Files (x86)\ZKTeco\att2000.mdb`,
		`C:\Program Files\ZKTeco\att2000.mdb`,
		`C:\att2000.mdb`,
		`C:\Users\shaki\Desktop\att2000.mdb`,
		desktopMDB,
		`C:\att2000\att2000.mdb`,
		`D:\att2000.mdb`,
	}

	for _, cand := range candidates {
		if cand != "" {
			if _, err := os.Stat(cand); err == nil {
				return cand, nil
			}
		}
	}

	reqStr := userPath
	if reqStr == "" {
		reqStr = `C:\Program Files (x86)\ZKTeco\att2000.mdb`
	}
	return "", fmt.Errorf("att2000.mdb file not found at '%s' or default installation directories. Please verify the file path.", reqStr)
}

func execPowerShellScript(script string) (string, error) {
	// Always prefer SysWOW64 (32-bit) PowerShell first for Jet OLE DB compatibility
	sysWOW64PS := `C:\Windows\SysWOW64\WindowsPowerShell\v1.0\powershell.exe`
	if _, statErr := os.Stat(sysWOW64PS); statErr == nil {
		cmd32 := exec.Command(sysWOW64PS, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
		out32, err32 := cmd32.CombinedOutput()
		if err32 == nil {
			return string(out32), nil
		}
		// Fall through to default PowerShell if 32-bit fails
	}

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%v (output: %s)", err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func getMDBUserBadgesFast(mdbPath string) ([]string, error) {
	winPath := filepath.FromSlash(mdbPath)
	escapedPath := strings.ReplaceAll(winPath, `\`, `\\`)

	psScript := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
$path = "%s"

if (-not (Test-Path $path)) { Write-Error "MDB not found"; exit 1 }

$tempMDB = Join-Path $env:TEMP ("zkteco_badges_" + [guid]::NewGuid().ToString() + ".mdb")
Copy-Item -Path $path -Destination $tempMDB -Force -ErrorAction Stop

# Remove read-only attribute on temp copy if it exists
$fi = Get-Item $tempMDB
if ($fi.Attributes -band [System.IO.FileAttributes]::ReadOnly) {
	$fi.Attributes = $fi.Attributes -bxor [System.IO.FileAttributes]::ReadOnly
}

function Get-MDBConnection($mdbPath) {
	$providers = @(
		"Provider=Microsoft.ACE.OLEDB.12.0;Data Source=$mdbPath;User ID=Admin;Password=;",
		"Provider=Microsoft.Jet.OLEDB.4.0;Data Source=$mdbPath;User ID=Admin;Password=;"
	)
	$errs = @()
	foreach ($cs in $providers) {
		try {
			$c = New-Object System.Data.OleDb.OleDbConnection($cs)
			$c.Open()
			return $c
		} catch {
			$errs += $_.Exception.Message
		}
	}
	throw "Cannot open MDB: $($errs -join ' | ')"
}

try {
	$conn = Get-MDBConnection -mdbPath $tempMDB
	$cmd = $conn.CreateCommand()
	$cmd.CommandText = "SELECT Badgenumber FROM USERINFO"
	$reader = $cmd.ExecuteReader()
	$badges = [System.Collections.Generic.List[string]]::new()
	while ($reader.Read()) {
		if (-not $reader.IsDBNull(0)) {
			$badges.Add($reader.GetValue(0).ToString().Trim())
		}
	}
	$reader.Close()
	$conn.Close()
	$conn.Dispose()
	$badges -join ","
} catch {
	Write-Error $_.Exception.Message
	exit 1
} finally {
	if (Test-Path $tempMDB) { Remove-Item -Path $tempMDB -Force -ErrorAction SilentlyContinue }
	$tempLdb = [System.IO.Path]::ChangeExtension($tempMDB, ".ldb")
	if (Test-Path $tempLdb) { Remove-Item -Path $tempLdb -Force -ErrorAction SilentlyContinue }
}
`, escapedPath)

	resStr, err := execPowerShellScript(psScript)
	if err != nil {
		return nil, err
	}
	resStr = strings.TrimSpace(resStr)
	if resStr == "" {
		return []string{}, nil
	}
	parts := strings.Split(resStr, ",")
	var badges []string
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			badges = append(badges, trimmed)
		}
	}
	return badges, nil
}

func syncBatchUsersToMDB(mdbPath string, items []syncItemInput) ([]ZKTecoSyncItemResult, error) {
	if len(items) == 0 {
		return []ZKTecoSyncItemResult{}, nil
	}

	winPath := filepath.FromSlash(mdbPath)
	escapedPath := strings.ReplaceAll(winPath, `\`, `\\`)

	tmpJSON := filepath.Join(os.TempDir(), "zkteco_sync_payload.json")
	payloadBytes, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("failed to encode sync payload: %v", err)
	}
	if err := os.WriteFile(tmpJSON, payloadBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write sync payload file: %v", err)
	}
	defer os.Remove(tmpJSON)

	escapedJSONPath := strings.ReplaceAll(filepath.FromSlash(tmpJSON), `\`, `\\`)

	// Strategy:
	// 1. Copy MDB to writable temp dir (bypasses Att.exe lock + Program Files permissions)
	// 2. Write all employee data to the temp copy
	// 3. Stop Att.exe, delete .ldb lock, grant NTFS permissions
	// 4. Copy updated MDB back to original location
	// 5. Report clear error if copy-back fails
	psScript := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
$mdbPath = "%s"
$jsonPath = "%s"

if (-not (Test-Path $mdbPath)) {
	Write-Error "MDB file not found at $mdbPath"
	exit 1
}
if (-not (Test-Path $jsonPath)) {
	Write-Error "Sync payload file not found at $jsonPath"
	exit 1
}

$rawJson = Get-Content -Path $jsonPath -Raw -Encoding UTF8 | ConvertFrom-Json

# --- Step 1: Copy MDB to a writable temp location ---
$tempDir = Join-Path $env:TEMP "zkteco_sync_work"
if (-not (Test-Path $tempDir)) { New-Item -ItemType Directory -Path $tempDir -Force | Out-Null }
$tempMDB = Join-Path $tempDir "att2000_work.mdb"

Copy-Item -Path $mdbPath -Destination $tempMDB -Force

# Remove read-only attribute from temp copy
$fi = Get-Item $tempMDB
if ($fi.Attributes -band [System.IO.FileAttributes]::ReadOnly) {
	$fi.Attributes = $fi.Attributes -bxor [System.IO.FileAttributes]::ReadOnly
}

# Delete any stale lock file on the temp copy
$tempLdb = [System.IO.Path]::ChangeExtension($tempMDB, ".ldb")
if (Test-Path $tempLdb) { Remove-Item -Path $tempLdb -Force -ErrorAction SilentlyContinue }

# --- Step 2: Connect to the writable temp copy ---
function Get-MDBConnection($path) {
	$providers = @(
		"Provider=Microsoft.ACE.OLEDB.12.0;Data Source=$path;User ID=Admin;Password=;",
		"Provider=Microsoft.Jet.OLEDB.4.0;Data Source=$path;User ID=Admin;Password=;"
	)
	$errs = @()
	foreach ($cs in $providers) {
		try {
			$c = New-Object System.Data.OleDb.OleDbConnection($cs)
			$c.Open()
			return $c
		} catch {
			$errs += $_.Exception.Message
		}
	}
	throw "Cannot open MDB database connection ($($errs -join ' | '))."
}

try {
	$conn = Get-MDBConnection -path $tempMDB

	# --- Step 3: Read existing badge numbers for upsert logic ---
	$checkCmd = $conn.CreateCommand()
	$checkCmd.CommandText = "SELECT Badgenumber FROM USERINFO"
	$reader = $checkCmd.ExecuteReader()
	$existingBadges = @{}
	while ($reader.Read()) {
		if (-not $reader.IsDBNull(0)) {
			$val = $reader.GetValue(0).ToString().Trim()
			if ($val -ne "") { $existingBadges[$val] = $true }
		}
	}
	$reader.Close()
	$reader.Dispose()
	$checkCmd.Dispose()

	# --- Step 4: Get next USERID ---
	$idCmd = $conn.CreateCommand()
	$idCmd.CommandText = "SELECT MAX(USERID) FROM USERINFO"
	$maxVal = $idCmd.ExecuteScalar()
	$idCmd.Dispose()
	$currentId = 1
	if ($maxVal -ne $null -and $maxVal -ne [System.DBNull]::Value) {
		$currentId = [int]$maxVal + 1
	}

	# --- Step 5: Upsert each employee using parameterized queries ---
	$results = [System.Collections.Generic.List[hashtable]]::new()

	# Pre-create the update command template
	$updCmd = $conn.CreateCommand()
	$updCmd.CommandText = "UPDATE USERINFO SET [Name] = ? WHERE Badgenumber = ?"
	$updParamName = $updCmd.Parameters.Add("pName", [System.Data.OleDb.OleDbType]::VarChar, 100)
	$updParamBadge = $updCmd.Parameters.Add("pBadge", [System.Data.OleDb.OleDbType]::VarChar, 50)

	# Pre-create the insert command template
	$insCmd = $conn.CreateCommand()
	$insCmd.CommandText = "INSERT INTO USERINFO (USERID, Badgenumber, [Name], DEFAULTDEPTID) VALUES (?, ?, ?, ?)"
	$insParamId = $insCmd.Parameters.Add("pId", [System.Data.OleDb.OleDbType]::Integer)
	$insParamBadge = $insCmd.Parameters.Add("pBadge", [System.Data.OleDb.OleDbType]::VarChar, 50)
	$insParamName = $insCmd.Parameters.Add("pName", [System.Data.OleDb.OleDbType]::VarChar, 100)
	$insParamDept = $insCmd.Parameters.Add("pDept", [System.Data.OleDb.OleDbType]::Integer)

	foreach ($item in $rawJson) {
		$empId = $item.emp_id
		$badge = $item.punch.Trim()
		$name = $item.name.Trim()

		try {
			if ($existingBadges.ContainsKey($badge)) {
				# Update existing
				$updParamName.Value = $name
				$updParamBadge.Value = $badge
				$updCmd.ExecuteNonQuery() | Out-Null
				$results.Add(@{ employee_id = $empId; punch_number = $badge; name = $name; status = "synced"; message = "Updated existing user in USERINFO" })
			} else {
				# Insert new
				$insParamId.Value = $currentId
				$insParamBadge.Value = $badge
				$insParamName.Value = $name
				$insParamDept.Value = 1
				$insCmd.ExecuteNonQuery() | Out-Null
				$existingBadges[$badge] = $true
				$currentId++
				$results.Add(@{ employee_id = $empId; punch_number = $badge; name = $name; status = "synced"; message = "Inserted new user into USERINFO" })
			}
		} catch {
			$results.Add(@{ employee_id = $empId; punch_number = $badge; name = $name; status = "failed"; message = $_.Exception.Message })
		}
	}

	$updCmd.Dispose()
	$insCmd.Dispose()
	$conn.Close()
	$conn.Dispose()

	# --- Step 6: Stop Att.exe and release .ldb lock, then copy back ---
	# Stop ZKTeco processes that may hold the database lock
	Get-Process -Name "Att" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
	Get-Process -Name "MSACCESS" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
	Start-Sleep -Milliseconds 800

	# Delete .ldb lock file on original
	$origLdb = [System.IO.Path]::ChangeExtension($mdbPath, ".ldb")
	if (Test-Path $origLdb) { Remove-Item -Path $origLdb -Force -ErrorAction SilentlyContinue }

	# Grant NTFS permissions on original file and its directory
	$mdbDir = Split-Path $mdbPath -Parent
	$currentUser = [System.Security.Principal.WindowsIdentity]::GetCurrent().Name
	try { & icacls $mdbDir /grant "${currentUser}:(OI)(CI)F" /T /Q 2>$null | Out-Null } catch {}
	try { & icacls $mdbPath /grant "${currentUser}:(F)" /Q 2>$null | Out-Null } catch {}

	# Remove read-only on original
	$origFi = Get-Item $mdbPath
	if ($origFi.Attributes -band [System.IO.FileAttributes]::ReadOnly) {
		$origFi.Attributes = $origFi.Attributes -bxor [System.IO.FileAttributes]::ReadOnly
	}

	# Copy updated MDB back to original location
	$copyBackError = ""
	try {
		Copy-Item -Path $tempMDB -Destination $mdbPath -Force -ErrorAction Stop
	} catch {
		$copyBackError = $_.Exception.Message
		# Try one more time using .NET File.Copy which can sometimes bypass PS issues
		try {
			[System.IO.File]::Copy($tempMDB, $mdbPath, $true)
			$copyBackError = ""
		} catch {
			$copyBackError = "COPY_BACK_FAILED: Could not copy updated database back to '$mdbPath'. Error: $($_.Exception.Message). The updated file is at '$tempMDB'."
		}
	}

	# Cleanup temp lock file
	if (Test-Path $tempLdb) { Remove-Item -Path $tempLdb -Force -ErrorAction SilentlyContinue }

	# Build output: append copy-back status
	$output = @{
		results = $results
		copy_back_error = $copyBackError
	}
	$output | ConvertTo-Json -Compress -Depth 4
} catch {
	# Cleanup on error
	if (Test-Path $tempMDB) { Remove-Item -Path $tempMDB -Force -ErrorAction SilentlyContinue }
	Write-Error $_.Exception.Message
	exit 1
}
`, escapedPath, escapedJSONPath)

	outStr, err := execPowerShellScript(psScript)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "file already in use") || strings.Contains(msg, "already opened exclusively") || strings.Contains(msg, "locked") {
			return nil, fmt.Errorf("att2000.mdb at '%s' is currently locked by ZKTeco software (Att.exe) or MS Access. Please close ZKTeco software / MS Access on your computer and try again", mdbPath)
		}
		if strings.Contains(msg, "updateable query") || strings.Contains(msg, "Read-only") || strings.Contains(msg, "PermissionDenied") {
			return nil, fmt.Errorf("att2000.mdb at '%s' is protected by Windows permissions. Please run the server as Administrator, or move the .mdb file to a writable location", mdbPath)
		}
		return nil, fmt.Errorf("%s", strings.TrimSpace(msg))
	}

	// Parse the wrapper output that contains results + copy_back_error
	var results []ZKTecoSyncItemResult
	outStr = strings.TrimSpace(outStr)

	if strings.HasPrefix(outStr, "{") {
		var wrapper struct {
			Results       []ZKTecoSyncItemResult `json:"results"`
			CopyBackError string                 `json:"copy_back_error"`
		}
		if err := json.Unmarshal([]byte(outStr), &wrapper); err == nil {
			results = wrapper.Results
			if wrapper.CopyBackError != "" {
				return nil, fmt.Errorf("sync completed but failed to save back to '%s': %s", mdbPath, wrapper.CopyBackError)
			}
		} else {
			// Fallback: try parsing as single result
			var single ZKTecoSyncItemResult
			if err := json.Unmarshal([]byte(outStr), &single); err == nil {
				results = append(results, single)
			}
		}
	} else if strings.HasPrefix(outStr, "[") {
		_ = json.Unmarshal([]byte(outStr), &results)
	}

	// Cleanup temp file after successful copy-back
	tempMDB := filepath.Join(os.TempDir(), "zkteco_sync_work", "att2000_work.mdb")
	os.Remove(tempMDB)

	return results, nil
}
