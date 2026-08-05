package service

import (
	"fmt"
	"time"

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

	// Exclude weekend days (per the employee's resolved shift).
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

// ShortfallTotals exposes the per-employee monthly shortfall for salary.
func (s *OtEarlyExitService) ShortfallTotals(companyID string, month, year int) (map[string]float64, error) {
	return s.otRepo.MonthlyShortfallTotals(companyID, month, year)
}