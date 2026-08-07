package service

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/utils"
	"gorm.io/gorm"
)

type NightBillService struct {
	db                *gorm.DB
	nightBillRepo     *repository.NightBillRepository
	nightBillListRepo *repository.NightBillEmployeeListRepository
}

func NewNightBillService(db *gorm.DB, nightBillRepo *repository.NightBillRepository, nightBillListRepo *repository.NightBillEmployeeListRepository) *NightBillService {
	return &NightBillService{
		db:                db,
		nightBillRepo:     nightBillRepo,
		nightBillListRepo: nightBillListRepo,
	}
}

type ProcessNightBillParams struct {
	StartDate        string  `json:"start_date"`
	EndDate          string  `json:"end_date"`
	CompanyID        string  `json:"company_id"`
	Mode             string  `json:"mode"`          // fallback mode when not using the employee list
	FixedAmount      float64 `json:"fixed_amount"`  // fallback amount
	HourlyRate       float64 `json:"hourly_rate"`   // fallback rate
	UseNightBillList bool    `json:"use_night_bill_list"`
	UserID           string  `json:"-"`
}

type ProcessNightBillResult struct {
	ProcessedCount int     `json:"processed_count"`
	TotalAmount    float64 `json:"total_amount"`
}

// ProcessConfigParams is the payload of the top-level POST /api/v1/night-bill/process endpoint.
type ProcessConfigParams struct {
	CompanyID string `json:"companyId" binding:"required"`
	FromDate  string `json:"fromDate" binding:"required"`
	ToDate    string `json:"toDate" binding:"required"`
	UserID    string `json:"-"`
}

// ProcessConfigResult mirrors the spec API response.
type ProcessConfigResult struct {
	Success    bool     `json:"success"`
	Processed  int      `json:"processed"`
	Generated  int      `json:"generated"`
	Skipped    int      `json:"skipped"`
	Duplicates int      `json:"duplicates"`
	Errors     []string `json:"errors"`
}

// skipStatuses are attendance statuses that must not generate a night bill:
// leave, holiday, weekend, missing/absent attendance.
var skipStatuses = map[string]bool{
	"absent":   true,
	"on_leave": true,
	"leave":    true,
	"weekend":  true,
	"holiday":  true,
}

func parseTimeOnDate(dateStr, timeStr string) (time.Time, error) {
	timeStr = strings.TrimSpace(timeStr)
	if len(timeStr) == 5 { // HH:MM
		timeStr += ":00"
	}
	fullStr := dateStr + " " + timeStr
	return time.Parse("2006-01-02 15:04:05", fullStr)
}

// shiftEndOnDate returns the shift end datetime on the attendance date, pushing
// overnight shift ends to the next day when they fall before the check-in time.
func shiftEndOnDate(att models.Attendance, dateStr string) (time.Time, string) {
	var shiftEndStr string
	if att.Shift != nil && att.Shift.EndTime != "" {
		shiftEndStr = att.Shift.EndTime
	} else {
		shiftEndStr = "17:00"
	}
	if att.CheckIn == nil {
		return time.Time{}, shiftEndStr
	}
	parsedEnd, err := parseTimeOnDate(dateStr, shiftEndStr)
	if err != nil {
		return time.Time{}, shiftEndStr
	}
	if parsedEnd.Before(*att.CheckIn) {
		parsedEnd = parsedEnd.Add(24 * time.Hour)
	}
	return parsedEnd, shiftEndStr
}

// computeNightBill applies the fixed/hourly rule for one attendance row.
//
//	Hourly: eligible hours = floor(OutTime - 20:00), minutes ignored.
//	Fixed:  eligible when the employee worked until Shift End + 7 hours (one bill only).
func computeNightBill(att models.Attendance, dateStr, billType string, fixedAmount, hourlyRate float64) (eligibleHours, rate, amount float64, qualifies bool) {
	if att.CheckIn == nil || att.CheckOut == nil {
		return 0, 0, 0, false
	}
	checkOut := *att.CheckOut

	if billType == "fixed" {
		shiftEndTime, _ := shiftEndOnDate(att, dateStr)
		if shiftEndTime.IsZero() {
			return 0, 0, 0, false
		}
		shiftEndPlus7 := shiftEndTime.Add(7 * time.Hour)
		if checkOut.Equal(shiftEndPlus7) || checkOut.After(shiftEndPlus7) {
			rate = fixedAmount
			if rate == 0 {
				rate = 100.0 // default 100 BDT
			}
			return 1, rate, rate, true
		}
		return 0, 0, 0, false
	}

	// hourly
	eightPm, err := parseTimeOnDate(dateStr, "20:00:00")
	if err != nil {
		return 0, 0, 0, false
	}
	if !checkOut.After(eightPm) {
		return 0, 0, 0, false
	}
	hours := math.Floor(checkOut.Sub(eightPm).Hours())
	if hours <= 0 {
		return 0, 0, 0, false
	}
	rate = hourlyRate
	if rate == 0 {
		rate = 20.0 // default 20 BDT/hr
	}
	return hours, rate, hours * rate, true
}

// Process is the legacy per-attendance-range processing (used by /attendance/night-bill/process).
func (s *NightBillService) Process(params ProcessNightBillParams) (*ProcessNightBillResult, error) {
	if params.StartDate == "" || params.EndDate == "" {
		return nil, fmt.Errorf("start_date and end_date are required")
	}

	fallbackMode := strings.ToLower(params.Mode)
	if fallbackMode == "" {
		fallbackMode = "fixed"
	}

	// Build eligible employee set from the Employee Night Bill List.
	eligible := map[string]*models.NightBillEmployeeList{}
	if params.UseNightBillList {
		list, err := s.nightBillListRepo.ListEligible(params.CompanyID)
		if err != nil {
			return nil, err
		}
		for i := range list {
			eligible[list[i].EmployeeID] = &list[i]
		}
	}

	// Fetch attendances in date range
	var attendances []models.Attendance
	query := s.db.Preload("Shift").
		Where("date BETWEEN ? AND ? AND check_in IS NOT NULL AND check_out IS NOT NULL AND deleted_at IS NULL", params.StartDate, params.EndDate)

	if params.CompanyID != "" {
		query = query.Where("company_id = ?", params.CompanyID)
	}
	if params.UseNightBillList && len(eligible) > 0 {
		ids := make([]string, 0, len(eligible))
		for id := range eligible {
			ids = append(ids, id)
		}
		query = query.Where("employee_id IN ?", ids)
	}

	if err := query.Find(&attendances).Error; err != nil {
		return nil, err
	}

	processedCount := 0
	totalAmount := 0.0

	for _, att := range attendances {
		if att.CheckIn == nil || att.CheckOut == nil {
			continue
		}

		// Determine the mode and rate for this employee.
		mode := fallbackMode
		fixedAmount := params.FixedAmount
		hourlyRate := params.HourlyRate
		if params.UseNightBillList {
			entry, ok := eligible[att.EmployeeID]
			if !ok {
				continue // only listed employees get night bills
			}
			mode = entry.BillType
			fixedAmount = entry.FixedAmount
			hourlyRate = entry.HourlyRate
		}

		dateStr := utils.NormalizeDate(att.Date)
		eligibleHours, rate, amount, qualifies := computeNightBill(att, dateStr, mode, fixedAmount, hourlyRate)
		if !qualifies || amount <= 0 {
			continue
		}

		shiftEndTime, shiftEndStr := shiftEndOnDate(att, dateStr)
		_ = shiftEndTime

		nb := &models.NightBill{
			CompanyID:      att.CompanyID,
			AttendanceID:   att.ID,
			EmployeeID:     att.EmployeeID,
			AttendanceDate: dateStr,
			ShiftID:        att.ShiftID,
			InTime:         att.CheckIn,
			OutTime:        att.CheckOut,
			BillType:       mode,
			ShiftEndTime:   &shiftEndStr,
			EligibleHours:  eligibleHours,
			Rate:           rate,
			Amount:         amount,
			Remarks:        fmt.Sprintf("Auto-calculated %s night bill", mode),
			CreatedBy:      &params.UserID,
			UpdatedBy:      &params.UserID,
		}

		if err := s.nightBillRepo.Upsert(nb); err == nil {
			processedCount++
			totalAmount += amount
		}
	}

	return &ProcessNightBillResult{
		ProcessedCount: processedCount,
		TotalAmount:    totalAmount,
	}, nil
}

// ProcessFromConfig implements the spec POST /api/v1/night-bill/process flow:
//  1. Load Employee Night Bill List.
//  2. Load Attendance + Shift.
//  3. Skip employees not configured for Night Bill.
//  4. Skip leave / holiday / weekend / missing attendance / rows without out time.
//  5. Calculate Hourly or Fixed Night Bill.
//  6. Prevent duplicates (employee + attendance date + bill type).
//  7. Save inside a single transaction.
func (s *NightBillService) ProcessFromConfig(params ProcessConfigParams) (*ProcessConfigResult, error) {
	from := utils.NormalizeDate(params.FromDate)
	to := utils.NormalizeDate(params.ToDate)
	if from == "" || to == "" {
		return nil, fmt.Errorf("fromDate and toDate must be valid dates")
	}

	eligible, err := s.nightBillListRepo.ListEligible(params.CompanyID)
	if err != nil {
		return nil, err
	}
	eligibleMap := make(map[string]*models.NightBillEmployeeList, len(eligible))
	for i := range eligible {
		eligibleMap[eligible[i].EmployeeID] = &eligible[i]
	}

	// Load holidays for the company in range.
	var holidays []models.Holiday
	if err := s.db.Where("company_id = ? AND deleted_at IS NULL", params.CompanyID).Find(&holidays).Error; err != nil {
		return nil, err
	}
	holidaySet := buildHolidaySet(holidays)

	var attendances []models.Attendance
	if err := s.db.Preload("Shift").
		Where("company_id = ? AND date BETWEEN ? AND ? AND deleted_at IS NULL", params.CompanyID, from, to).
		Find(&attendances).Error; err != nil {
		return nil, err
	}

	res := &ProcessConfigResult{Success: true}
	now := time.Now()

	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, att := range attendances {
			entry, ok := eligibleMap[att.EmployeeID]
			if !ok {
				res.Skipped++
				continue
			}
			if att.CheckIn == nil || att.CheckOut == nil {
				res.Skipped++
				continue
			}
			if skipStatuses[att.Status] {
				res.Skipped++
				continue
			}
			dateStr := utils.NormalizeDate(att.Date)
			if holidaySet[dateStr] {
				res.Skipped++
				continue
			}
			if att.Shift != nil && att.Shift.WeekendDays != "" && utils.IsWeekend(dateStr, att.Shift.WeekendDays) {
				res.Skipped++
				continue
			}

			billType := entry.BillType
			eligibleHours, rate, amount, qualifies := computeNightBill(att, dateStr, billType, entry.FixedAmount, entry.HourlyRate)
			if !qualifies || amount <= 0 {
				res.Skipped++
				continue
			}

			var dupCount int64
			if err := tx.Model(&models.NightBill{}).
				Where("employee_id = ? AND attendance_date = ? AND bill_type = ? AND deleted_at IS NULL", att.EmployeeID, dateStr, billType).
				Count(&dupCount).Error; err != nil {
				res.Errors = append(res.Errors, err.Error())
				continue
			}
			if dupCount > 0 {
				res.Duplicates++
				res.Processed++
				continue
			}

			_, shiftEndStr := shiftEndOnDate(att, dateStr)
			nb := &models.NightBill{
				CompanyID:      att.CompanyID,
				AttendanceID:   att.ID,
				EmployeeID:     att.EmployeeID,
				AttendanceDate: dateStr,
				ShiftID:        att.ShiftID,
				InTime:         att.CheckIn,
				OutTime:        att.CheckOut,
				BillType:       billType,
				ShiftEndTime:   &shiftEndStr,
				EligibleHours:  eligibleHours,
				Rate:           rate,
				Amount:         amount,
				Remarks:        fmt.Sprintf("Auto-calculated %s night bill", billType),
				ProcessedAt:    &now,
				CreatedBy:      &params.UserID,
				UpdatedBy:      &params.UserID,
			}
			if err := tx.Create(nb).Error; err != nil {
				res.Errors = append(res.Errors, err.Error())
				continue
			}
			res.Generated++
			res.Processed++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	res.Success = len(res.Errors) == 0
	return res, nil
}

func buildHolidaySet(holidays []models.Holiday) map[string]bool {
	set := map[string]bool{}
	for _, h := range holidays {
		if d := utils.NormalizeDate(h.Date); d != "" {
			set[d] = true
		}
		if h.FromDate == nil || h.ToDate == nil {
			continue
		}
		fd := utils.NormalizeDate(*h.FromDate)
		td := utils.NormalizeDate(*h.ToDate)
		if fd == "" || td == "" {
			continue
		}
		cur, err := time.Parse("2006-01-02", fd)
		if err != nil {
			continue
		}
		end, err := time.Parse("2006-01-02", td)
		if err != nil {
			continue
		}
		for !cur.After(end) {
			set[cur.Format("2006-01-02")] = true
			cur = cur.AddDate(0, 0, 1)
		}
	}
	return set
}
