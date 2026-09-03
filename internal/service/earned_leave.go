package service

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"gorm.io/gorm"
)

var elProcessLocks sync.Map // key = companyID|period → struct{}

// EarnedLeaveService centralizes EL logic (§4)
type EarnedLeaveService struct {
	policyRepo     *repository.EarnedLeavePolicyRepository
	ledgerRepo     *repository.EarnedLeaveLedgerRepository
	balanceRepo    *repository.EarnedLeaveBalanceRepository
	salaryRepo     *repository.EarnedLeaveSalaryRepository
	employeeRepo   *repository.EmployeeRepository
	separationRepo *repository.SeparationRepository
	leaveRepo      *repository.LeaveRepository
	db             *gorm.DB
}

func NewEarnedLeaveService(
	policyRepo *repository.EarnedLeavePolicyRepository,
	ledgerRepo *repository.EarnedLeaveLedgerRepository,
	balanceRepo *repository.EarnedLeaveBalanceRepository,
	salaryRepo *repository.EarnedLeaveSalaryRepository,
	employeeRepo *repository.EmployeeRepository,
	separationRepo *repository.SeparationRepository,
	leaveRepo *repository.LeaveRepository,
	db *gorm.DB,
) *EarnedLeaveService {
	return &EarnedLeaveService{
		policyRepo: policyRepo, ledgerRepo: ledgerRepo, balanceRepo: balanceRepo,
		salaryRepo: salaryRepo, employeeRepo: employeeRepo, separationRepo: separationRepo,
		leaveRepo: leaveRepo, db: db,
	}
}

// IsEligible — central eligibility (§4, §5)
func (s *EarnedLeaveService) IsEligible(emp *models.Employee, asOfDate string, policy *models.EarnedLeavePolicy, sepMap map[string]string) bool {
	if emp == nil || policy == nil {
		return false
	}
	if emp.Status != "active" {
		return false
	}
	asOf, err := time.Parse("2006-01-02", asOfDate)
	if err != nil {
		return false
	}
	if emp.JoiningDate.IsZero() {
		return false
	}
	// Minimum service
	if policy.MinServiceMonths > 0 {
		minDate := emp.JoiningDate.AddDate(0, policy.MinServiceMonths, 0)
		if asOf.Before(minDate) {
			return false
		}
	}
	if asOf.Before(emp.JoiningDate) {
		return false
	}
	// Separation
	if sepDateStr, ok := sepMap[emp.EmployeeID]; ok && sepDateStr != "" {
		if sep, err := time.Parse("2006-01-02", sepDateStr); err == nil {
			if asOf.After(sep) {
				return false
			}
		}
	}
	return true
}

func rounding(v float64, rule string) float64 {
	switch rule {
	case "FLOOR":
		return math.Floor(v*100) / 100
	case "CEIL":
		return math.Ceil(v*100) / 100
	default:
		return math.Round(v*100) / 100
	}
}

// ProcessResult summary (§34)
type ELProcessResult struct {
	CompanyID string `json:"company_id"`
	Period    string `json:"period"` // YYYY-MM
	Employees int    `json:"employees"`
	Eligible  int    `json:"eligible"`
	Processed int    `json:"processed"`
	Skipped   int    `json:"skipped"`
	Created   int    `json:"created"`
	Failed    int    `json:"failed"`
	DurationMs int64 `json:"duration_ms"`
}

// AccrualProcess — idempotent, bulk, transactional (§8, §19-21, §24-26)
func (s *EarnedLeaveService) AccrualProcess(companyID, period, userID string) (*ELProcessResult, error) {
	if period == "" {
		return nil, fmt.Errorf("period is required (YYYY-MM)")
	}
	if _, err := time.Parse("2006-01", period); err != nil {
		return nil, fmt.Errorf("invalid period, expected YYYY-MM")
	}
	lockKey := companyID + "|" + period
	if _, loaded := elProcessLocks.LoadOrStore(lockKey, struct{}{}); loaded {
		return nil, fmt.Errorf("EL accrual already processing for %s %s", companyID, period)
	}
	defer elProcessLocks.Delete(lockKey)

	start := time.Now()
	asOfDate := period + "-01"
	if t, err := time.Parse("2006-01-02", asOfDate); err == nil {
		// use last day of period for eligibility/balance
		last := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		asOfDate = last
	}

	// Bulk load
	policy, err := s.policyRepo.FindActive(companyID, asOfDate)
	if err != nil {
		return nil, fmt.Errorf("no active EL policy for %s as of %s", companyID, asOfDate)
	}
	employees, err := s.employeeRepo.ListForSalaryProcessing(companyID, 0, 0) // use 0,0 to get all? fallback to ListActive
	if err != nil || len(employees) == 0 {
		// fallback to active
		emps, _, _ := s.employeeRepo.ListActive(companyID, 1, 10000)
		employees = make([]models.Employee, len(emps))
		for i, e := range emps {
			employees[i] = e
		}
	}
	// For generic EL, if ListForSalaryProcessing returns 0 (month 0), use ListActive
	if len(employees) == 0 {
		allActive, _, _ := s.employeeRepo.ListActive(companyID, 1, 10000)
		employees = make([]models.Employee, len(allActive))
		for i, e := range allActive {
			employees[i] = e
		}
	}
	empIDs := make([]string, 0, len(employees))
	for _, e := range employees {
		empIDs = append(empIDs, e.EmployeeID)
	}
	// Separations bulk
	sepMap := make(map[string]string)
	if s.separationRepo != nil && len(empIDs) > 0 {
		if rows, err := s.separationRepo.ListByEmployeeIDs(empIDs); err == nil {
			for _, r := range rows {
				if _, ok := sepMap[r.EmployeeID]; !ok {
					sepMap[r.EmployeeID] = r.Date
				}
			}
		}
	}
	// Existing accrual check (idempotency)
	existingAccrual, _ := s.ledgerRepo.BatchExistingAccrual(companyID, period, empIDs)
	// Latest ledger per employee for opening
	latestMap, _ := s.ledgerRepo.BatchLatestByEmployee(companyID, empIDs, asOfDate)
	// Balances map via ledger sums (opening)
	balanceDeltas, _ := s.ledgerRepo.SumBalances(companyID, empIDs, asOfDate)

	var toCreate []models.EarnedLeaveLedger
	var toUpsertBalances []models.EarnedLeaveBalance
	eligible := 0
	skipped := 0
	year, _ := time.Parse("2006-01-02", asOfDate)
	yr := year.Year()

	for _, emp := range employees {
		if _, exists := existingAccrual[emp.EmployeeID]; exists {
			skipped++
			continue
		}
		if !s.IsEligible(&emp, asOfDate, policy, sepMap) {
			skipped++
			continue
		}
		eligible++
		// Opening = latest closing or 0
		opening := 0.0
		if latest, ok := latestMap[emp.EmployeeID]; ok && latest != nil {
			opening = latest.ClosingBalance
		} else if v, ok := balanceDeltas[emp.EmployeeID]; ok {
			// SumBalances returns delta, but opening from ledger's closing is more accurate
			_ = v
		}
		accrued := rounding(policy.AccrualRate, policy.RoundingRule)
		closing := opening + accrued
		// Max balance clamp
		if policy.MaxBalance > 0 && closing > policy.MaxBalance {
			accrued = policy.MaxBalance - opening
			if accrued < 0 {
				accrued = 0
			}
			accrued = rounding(accrued, policy.RoundingRule)
			closing = opening + accrued
			if closing > policy.MaxBalance {
				closing = policy.MaxBalance
			}
		}
		if accrued <= 0 {
			skipped++
			continue
		}
		toCreate = append(toCreate, models.EarnedLeaveLedger{
			CompanyID:       companyID,
			EmployeeID:      emp.EmployeeID,
			TransactionDate: asOfDate,
			Period:          period,
			TransactionType: "ACCRUAL",
			OpeningBalance:  opening,
			Accrued:         accrued,
			ClosingBalance:  closing,
			Remarks:         fmt.Sprintf("EL accrual %s", period),
			CreatedBy:       &userID,
		})
		// Balance upsert for year
		// closing for year = previous closing + accrued (yearly roll)
		toUpsertBalances = append(toUpsertBalances, models.EarnedLeaveBalance{
			CompanyID:  companyID,
			EmployeeID: emp.EmployeeID,
			Year:       yr,
			Period:     fmt.Sprintf("%d", yr),
			Opening:    opening,
			Accrued:    accrued,
			Closing:    closing,
		})
	}

	// Transaction
	created := 0
	if len(toCreate) > 0 {
		err = s.db.Transaction(func(tx *gorm.DB) error {
			lRepo := s.ledgerRepo.WithTx(tx)
			bRepo := s.balanceRepo.WithTx(tx)
			if err := lRepo.CreateBatch(toCreate); err != nil {
				return err
			}
			// Upsert balances: need to merge with existing (sum)
			// For simplicity, use OnConflict to increment accrued/closing
			// But our models has separate fields, so we do manual upsert via CreateInBatches with OnConflict
			// Here we just use UpsertBalances which does CreateInBatches OnConflict
			// To avoid overwriting, we fetch existing and sum — already handled via opening calc
			// So upsert is fine as replace
			if len(toUpsertBalances) > 0 {
				if err := bRepo.UpsertBalances(toUpsertBalances); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		created = len(toCreate)
	}

	duration := time.Since(start).Milliseconds()
	return &ELProcessResult{
		CompanyID: companyID, Period: period,
		Employees: len(employees), Eligible: eligible,
		Processed: created + skipped, Skipped: skipped,
		Created: created, Failed: 0, DurationMs: duration,
	}, nil
}

// GenerateSalarySheet — bulk calc §14, §18-21
func (s *EarnedLeaveService) GenerateSalarySheet(companyID string, month, year int, userID string) (*models.EarnedLeaveSalarySheet, error) {
	period := fmt.Sprintf("%04d-%02d", year, month)
	lockKey := companyID + "|salary|" + period
	if _, loaded := elProcessLocks.LoadOrStore(lockKey, struct{}{}); loaded {
		return nil, fmt.Errorf("EL salary sheet already processing for %s %s", companyID, period)
	}
	defer elProcessLocks.Delete(lockKey)

	// Idempotency: if sheet exists and finalized, block
	if existing, err := s.salaryRepo.FindSheetByPeriod(companyID, month, year); err == nil && existing != nil {
		if existing.Status == "FINALIZED" {
			return nil, fmt.Errorf("EL salary sheet %s already finalized, use reverse", period)
		}
		if existing.Status == "APPROVED" || existing.Status == "CALCULATED" {
			return nil, fmt.Errorf("EL salary sheet %s already exists with status %s, use reprocess", period, existing.Status)
		}
	}

	// Bulk load
	employees, _ := s.employeeRepo.ListForSalaryProcessing(companyID, month, year)
	if len(employees) == 0 {
		allActive, _, _ := s.employeeRepo.ListActive(companyID, 1, 10000)
		employees = make([]models.Employee, len(allActive))
		for i, e := range allActive {
			employees[i] = e
		}
	}
	empIDs := make([]string, len(employees))
	for i, e := range employees {
		empIDs[i] = e.EmployeeID
	}
	policy, _ := s.policyRepo.FindActive(companyID, fmt.Sprintf("%04d-%02d-01", year, month))
	if policy == nil {
		// fallback to any active
		policies, _ := s.policyRepo.ListByCompany(companyID)
		if len(policies) > 0 {
			policy = &policies[0]
		}
	}
	asOfDate := fmt.Sprintf("%04d-%02d-%02d", year, month, time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day())
	ledgerMap, _ := s.ledgerRepo.BatchLatestByEmployee(companyID, empIDs, asOfDate)

	var items []models.EarnedLeaveSalaryItem
	totalDays := 0.0
	totalAmount := 0.0
	for _, emp := range employees {
		latest := ledgerMap[emp.EmployeeID]
		opening, closing, accrued, used := 0.0, 0.0, 0.0, 0.0
		if latest != nil {
			opening = latest.OpeningBalance
			closing = latest.ClosingBalance
			accrued = latest.Accrued
			used = latest.Used
		}
		// Encashable = closing - used? For EL salary sheet, encashable = closing (available)
		encashable := closing
		if encashable < 0 {
			encashable = 0
		}
		// Salary basis per policy
		basis := emp.GrossSalary
		if policy != nil {
			switch policy.EncashmentBasis {
			case "BASIC":
				basis = emp.BasicSalary
			case "GROSS":
				basis = emp.GrossSalary
			default:
				basis = emp.GrossSalary
			}
		}
		divisor := 30.0
		if policy != nil && policy.EncashmentDivisor > 0 {
			divisor = float64(policy.EncashmentDivisor)
		}
		rate := 0.0
		if divisor > 0 {
			rate = basis / divisor
		}
		amount := rounding(encashable*rate, "ROUND")

		items = append(items, models.EarnedLeaveSalaryItem{
			CompanyID:      companyID,
			EmployeeID:     emp.EmployeeID,
			Opening:        opening,
			Accrued:        accrued,
			Used:           used,
			Closing:        closing,
			EncashableDays: encashable,
			SalaryBasis:    basis,
			Rate:           rate,
			Amount:         amount,
		})
		totalDays += encashable
		totalAmount += amount
	}

	sheet := &models.EarnedLeaveSalarySheet{
		CompanyID:      companyID,
		PeriodMonth:    month,
		PeriodYear:     year,
		Period:         period,
		Status:         "CALCULATED",
		TotalEmployees: len(items),
		TotalDays:      totalDays,
		TotalAmount:    totalAmount,
		CreatedBy:      &userID,
	}
	if err := s.salaryRepo.CreateSheetWithItems(sheet, items); err != nil {
		return nil, err
	}
	return sheet, nil
}
