package service

import (
	"fmt"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

// OtEarlyExitService computes early-exit overtime deductions for a payroll
// month and persists them as an immutable ledger.
type OtEarlyExitService struct {
	otRepo      *repository.OtEarlyExitRepository
	holidayRepo *repository.HolidayRepository
}

func NewOtEarlyExitService(otRepo *repository.OtEarlyExitRepository, holidayRepo *repository.HolidayRepository) *OtEarlyExitService {
	return &OtEarlyExitService{otRepo: otRepo, holidayRepo: holidayRepo}
}

// ComputeResult summarizes a compute run.
type ComputeResult struct {
	CompanyID       string
	Month           int
	Year            int
	TotalRecords    int
	AffectedEmployees int
}

// ComputeEarlyExitDeductions recomputes the early-exit shortfall ledger for a
// month and replaces the stored records in one transaction.
func (s *OtEarlyExitService) ComputeEarlyExitDeductions(companyID string, month, year int, userID string) (*ComputeResult, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)
	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	rows, err := s.otRepo.ListShortfallRows(companyID, startStr, endStr)
	if err != nil {
		return nil, fmt.Errorf("compute shortfall: %w", err)
	}

	// Exclude days that are government/company holidays (no work baseline).
	holidaySet := make(map[string]bool)
	if s.holidayRepo != nil {
		if holidays, hErr := s.holidayRepo.ListActiveByDateRange(startStr, endStr, companyID); hErr == nil {
			for _, h := range holidays {
				if h.Type == "weekend_change" {
					continue
				}
				hDate := utils.NormalizeDate(h.Date)
				var hFrom, hTo string
				if h.FromDate != nil {
					hFrom = utils.NormalizeDate(*h.FromDate)
				}
				if h.ToDate != nil {
					hTo = utils.NormalizeDate(*h.ToDate)
				}
				if hDate != "" {
					holidaySet[hDate] = true
				}
				if hFrom != "" && hTo != "" {
					days, _ := utils.GenerateDateRange(hFrom, hTo)
					for _, d := range days {
						holidaySet[d] = true
					}
				}
			}
		}
	}

	// Load exemptions for this month so removed/exempted employees or dates are not recreated
	exemptions, _ := s.otRepo.ListExemptions(companyID, month, year)
	exemptMap := make(map[string]bool)
	for _, ex := range exemptions {
		if ex.Date != nil && *ex.Date != "" {
			d := utils.NormalizeDate(*ex.Date)
			exemptMap[ex.EmployeeID+"|"+d] = true
		} else {
			exemptMap[ex.EmployeeID] = true
		}
	}

	// Exclude weekend days (per the employee's resolved shift) and exempted entries.
	filtered := make([]repository.ShortfallRow, 0, len(rows))
	for _, r := range rows {
		if rr := utils.NormalizeDate(r.Date); rr == "" {
			continue
		}
		d := utils.NormalizeDate(r.Date)
		if holidaySet[d] {
			continue
		}
		if utils.IsWeekend(d, r.WeekendDays) {
			continue
		}
		// Skip exempted employee or exempted employee-date
		if exemptMap[r.EmployeeID] || exemptMap[r.EmployeeID+"|"+d] {
			continue
		}
		filtered = append(filtered, r)
	}

	if err := s.otRepo.UpsertMonth(companyID, month, year, filtered, userID); err != nil {
		return nil, fmt.Errorf("persist deductions: %w", err)
	}

	empSet := make(map[string]struct{}, len(filtered))
	for _, r := range filtered {
		empSet[r.EmployeeID] = struct{}{}
	}

	return &ComputeResult{
		CompanyID:         companyID,
		Month:             month,
		Year:              year,
		TotalRecords:      len(filtered),
		AffectedEmployees: len(empSet),
	}, nil
}

// RemoveDeduction deletes an individual deduction record and saves an exemption
// so future re-computations or salary processes will NOT recreate it.
func (s *OtEarlyExitService) RemoveDeduction(id string, reason, userID string) error {
	deduction, err := s.otRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("deduction not found: %w", err)
	}

	dateStr := utils.NormalizeDate(deduction.Date)
	if dateStr == "" {
		dateStr = deduction.Date
	}
	var datePtr *string
	if dateStr != "" {
		datePtr = &dateStr
	}

	ex := &models.OtEarlyExitExemption{
		CompanyID:  deduction.CompanyID,
		EmployeeID: deduction.EmployeeID,
		Month:      deduction.Month,
		Year:       deduction.Year,
		Date:       datePtr,
		Reason:     reason,
	}
	if userID != "" && len(userID) == 36 {
		ex.CreatedBy = &userID
	}

	if err := s.otRepo.CreateExemption(ex); err != nil {
		return fmt.Errorf("save exemption: %w", err)
	}

	return s.otRepo.DeleteDeduction(id)
}

// ExemptEmployee exempts an entire employee from early-exit deductions for a payroll month
// and deletes any existing deduction records for that employee in that month.
func (s *OtEarlyExitService) ExemptEmployee(companyID, employeeID string, month, year int, reason, userID string) error {
	ex := &models.OtEarlyExitExemption{
		CompanyID:  companyID,
		EmployeeID: employeeID,
		Month:      month,
		Year:       year,
		Reason:     reason,
	}
	if userID != "" && len(userID) == 36 {
		ex.CreatedBy = &userID
	}

	if err := s.otRepo.CreateExemption(ex); err != nil {
		return fmt.Errorf("save exemption: %w", err)
	}

	return s.otRepo.DeleteEmployeeDeductions(companyID, employeeID, month, year)
}

// ShortfallTotals exposes the per-employee monthly shortfall for salary.
func (s *OtEarlyExitService) ShortfallTotals(companyID string, month, year int) (map[string]float64, error) {
	return s.otRepo.MonthlyShortfallTotals(companyID, month, year)
}