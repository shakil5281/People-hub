package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrEmployeeNotFound        = errors.New("employee not found or not active regular")
	ErrDuplicateSeparation     = errors.New("employee already has a pending or approved separation")
	ErrAlreadyProcessed        = errors.New("separation already processed or cancelled")
	ErrInvalidSeparationType   = errors.New("separation type must be Resign, Lefty, or Close")
	ErrSeparationBeforeJoining = errors.New("separation date cannot be before joining date")
	ErrDepartmentRequired      = errors.New("employee has no department assigned — department is required for separation")
)

type SeparationService struct {
	db             *gorm.DB
	sepRepo        *repository.SeparationRepository
	employeeRepo   *repository.EmployeeRepository
	attendanceRepo *repository.AttendanceRepository
}

func NewSeparationService(
	db *gorm.DB,
	sepRepo *repository.SeparationRepository,
	employeeRepo *repository.EmployeeRepository,
	attendanceRepo *repository.AttendanceRepository,
) *SeparationService {
	return &SeparationService{
		db:             db,
		sepRepo:        sepRepo,
		employeeRepo:   employeeRepo,
		attendanceRepo: attendanceRepo,
	}
}

type CreateSeparationInput struct {
	EmployeeID   string
	DepartmentID string
	SepType      string
	Date         string
	Reason       string
}

type ProcessResult struct {
	EmployeeID        string
	EmployeeName      string
	OldStatus         string
	NewType           string
	NewStatus         string
	AttendanceDeleted int64
}

// Create creates a separation record. If the separation date is today or in the past,
// it auto-processes and applies employee changes in the same transaction.
func (s *SeparationService) Create(input CreateSeparationInput) (*models.Separation, *ProcessResult, error) {
	sepType := strings.TrimSpace(input.SepType)
	empCode := strings.TrimSpace(input.EmployeeID)

	if empCode == "" {
		return nil, nil, ErrEmployeeNotFound
	}

	emp, err := s.employeeRepo.FindByEmployeeID(empCode)
	if err != nil {
		emp, err = s.employeeRepo.FindByPunchNumber(empCode)
		if err != nil {
			return nil, nil, ErrEmployeeNotFound
		}
	}

	if !strings.EqualFold(emp.Status, "active") || !strings.EqualFold(strings.TrimSpace(emp.EmployeeType), "regular") {
		return nil, nil, ErrEmployeeNotFound
	}

	if input.Date != "" {
		joinDay := emp.JoiningDate
		sepDate, parseErr := time.Parse("2006-01-02", input.Date)
		if parseErr == nil && !joinDay.IsZero() {
			joinOnlyDate := time.Date(joinDay.Year(), joinDay.Month(), joinDay.Day(), 0, 0, 0, 0, time.UTC)
			if sepDate.Before(joinOnlyDate) {
				return nil, nil, ErrSeparationBeforeJoining
			}
		}
	}

	exists, dupErr := s.sepRepo.ExistsPendingOrApproved(emp.EmployeeID)
	if dupErr != nil {
		return nil, nil, dupErr
	}
	if exists {
		return nil, nil, ErrDuplicateSeparation
	}

	newType, ok := mapSeparationTypeToEmployeeType(sepType)
	if !ok {
		return nil, nil, ErrInvalidSeparationType
	}

	shouldAutoProcess := false
	if input.Date != "" {
		sepDate, parseErr := time.Parse("2006-01-02", input.Date)
		if parseErr == nil {
			today := time.Now().Truncate(24 * time.Hour)
			if !sepDate.After(today) {
				shouldAutoProcess = true
			}
		}
	}

	employeeName := emp.NameEn
	if employeeName == "" {
		employeeName = emp.NameBn
	}

	var result *ProcessResult

	deptID := ""
	if emp.DepartmentID != nil {
		deptID = *emp.DepartmentID
	}
	if input.DepartmentID != "" {
		deptID = input.DepartmentID
	}
	if deptID == "" {
		return nil, nil, ErrDepartmentRequired
	}

	var created *models.Separation
	err = s.db.Transaction(func(tx *gorm.DB) error {
		sep := &models.Separation{
			Employee:     employeeName,
			EmployeeID:   emp.EmployeeID,
			CompanyID:    emp.CompanyID,
			DepartmentID: deptID,
			Type:         sepType,
			Date:         input.Date,
			Status:       "Pending",
			Reason:       input.Reason,
		}
		if createErr := s.sepRepo.WithTx(tx).Create(sep); createErr != nil {
			return createErr
		}
		created = sep

		if shouldAutoProcess {
			r, processErr := s.processOne(tx, sep, emp, newType)
			if processErr != nil {
				return processErr
			}
			result = r
		} else {
			result = &ProcessResult{
				EmployeeID:   emp.EmployeeID,
				EmployeeName: employeeName,
			}
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return created, result, nil
}

// ProcessBatch processes all eligible pending/approved separations with date <= processDate.
func (s *SeparationService) ProcessBatch(processDate string) ([]ProcessResult, error) {
	eligible, err := s.sepRepo.FindPendingDue(processDate)
	if err != nil {
		return nil, err
	}

	var results []ProcessResult
	var errs []string

	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, sep := range eligible {
			emp, empErr := s.employeeRepo.WithTx(tx).FindByEmployeeID(sep.EmployeeID)
			if empErr != nil {
				errs = append(errs, fmt.Sprintf("employee %s: not found", sep.EmployeeID))
				continue
			}
			if !strings.EqualFold(emp.Status, "active") {
				errs = append(errs, fmt.Sprintf("employee %s: not active", sep.EmployeeID))
				continue
			}

			newType, ok := mapSeparationTypeToEmployeeType(strings.TrimSpace(sep.Type))
			if !ok {
				errs = append(errs, fmt.Sprintf("employee %s: invalid separation type %s", sep.EmployeeID, sep.Type))
				continue
			}

			r, processErr := s.processOne(tx, &sep, emp, newType)
			if processErr != nil {
				errs = append(errs, fmt.Sprintf("employee %s: %s", sep.EmployeeID, processErr.Error()))
				continue
			}
			results = append(results, *r)
		}
		return nil
	})

	if len(errs) > 0 {
		return results, fmt.Errorf("processed %d separations with %d errors: %s", len(results), len(errs), strings.Join(errs, "; "))
	}

	return results, nil
}

// ProcessOne processes a single separation by ID.
func (s *SeparationService) ProcessOne(id string) (*ProcessResult, error) {
	sep, err := s.sepRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("separation not found")
	}
	if sep.Status == "Processed" || sep.Status == "Cancelled" {
		return nil, ErrAlreadyProcessed
	}

	emp, err := s.employeeRepo.FindByEmployeeID(sep.EmployeeID)
	if err != nil {
		return nil, ErrEmployeeNotFound
	}

	newType, ok := mapSeparationTypeToEmployeeType(strings.TrimSpace(sep.Type))
	if !ok {
		return nil, ErrInvalidSeparationType
	}

	var result *ProcessResult
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		r, processErr := s.processOne(tx, sep, emp, newType)
		if processErr != nil {
			return processErr
		}
		result = r
		return nil
	})

	return result, txErr
}

// Cancel marks a pending/approved separation as cancelled.
func (s *SeparationService) Cancel(id string) error {
	sep, err := s.sepRepo.FindByID(id)
	if err != nil {
		return errors.New("separation not found")
	}
	if sep.Status == "Processed" {
		return errors.New("cannot cancel a processed separation; use reactivate instead")
	}
	if sep.Status == "Cancelled" {
		return nil
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		return tx.Model(sep).Update("status", "Cancelled").Error
	})
}

// Reactivate reverts a processed separation, restoring the employee to active/Regular status.
func (s *SeparationService) Reactivate(id string) error {
	sep, err := s.sepRepo.FindByID(id)
	if err != nil {
		return errors.New("separation not found")
	}
	if sep.Status != "Processed" {
		return errors.New("only processed separations can be reactivated")
	}

	emp, err := s.employeeRepo.FindByEmployeeID(sep.EmployeeID)
	if err != nil {
		return ErrEmployeeNotFound
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"employee_type": "Regular",
			"status":        "active",
		}
		if err := tx.Model(emp).Updates(updates).Error; err != nil {
			return err
		}
		return tx.Model(sep).Update("status", "Cancelled").Error
	})
}

// processOne applies employee changes and marks separation as Processed.
// Must be called within a transaction.
func (s *SeparationService) processOne(tx *gorm.DB, sep *models.Separation, emp *models.Employee, newType string) (*ProcessResult, error) {
	updates := map[string]interface{}{
		"employee_type": newType,
		"status":        "inactive",
	}
	if err := tx.Model(emp).Updates(updates).Error; err != nil {
		return nil, err
	}

	var deleted int64
	if sep.Date != "" {
		d, delErr := s.attendanceRepo.WithTx(tx).DeleteAfterDate(sep.EmployeeID, sep.Date)
		if delErr != nil {
			return nil, delErr
		}
		deleted = d
	}

	if err := tx.Model(sep).Update("status", "Processed").Error; err != nil {
		return nil, err
	}

	return &ProcessResult{
		EmployeeID:        emp.EmployeeID,
		EmployeeName:      emp.NameEn,
		OldStatus:         "active",
		NewType:           newType,
		NewStatus:         "inactive",
		AttendanceDeleted: deleted,
	}, nil
}

func mapSeparationTypeToEmployeeType(t string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "resign":
		return "Resign", true
	case "lefty":
		return "Lefty", true
	case "close":
		return "Close", true
	default:
		return "", false
	}
}
