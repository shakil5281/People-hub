package service

import (
	"errors"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"gorm.io/gorm"
)

type MigrationService struct {
	db            *gorm.DB
	migrationRepo *repository.MigrationRepository
	employeeRepo  *repository.EmployeeRepository
}

func NewMigrationService(db *gorm.DB, migrationRepo *repository.MigrationRepository, employeeRepo *repository.EmployeeRepository) *MigrationService {
	return &MigrationService{
		db:            db,
		migrationRepo: migrationRepo,
		employeeRepo:  employeeRepo,
	}
}

type CreateMigrationInput struct {
	EmployeeID       string  `json:"employee_id" binding:"required"`
	MigrationType    string  `json:"migration_type" binding:"required"`
	NewDepartmentID  *string `json:"new_department_id"`
	NewSectionID     *string `json:"new_section_id"`
	NewDesignationID *string `json:"new_designation_id"`
	NewLineID        *string `json:"new_line_id"`
	EffectiveDate    string  `json:"effective_date"`
	Reason           string  `json:"reason"`
	Remarks          string  `json:"remarks"`
	CreatedBy        *string `json:"created_by"`
}

func (s *MigrationService) Create(input CreateMigrationInput) (*models.EmployeeMigration, error) {
	emp, err := s.employeeRepo.FindByEmployeeID(input.EmployeeID)
	if err != nil || emp == nil {
		return nil, errors.New("employee not found")
	}

	effDate := input.EffectiveDate
	if effDate == "" {
		effDate = time.Now().Format("2006-01-02")
	}

	mig := &models.EmployeeMigration{
		EmployeeID:       emp.EmployeeID,
		EmployeeName:     emp.NameEn,
		CompanyID:        emp.CompanyID,
		MigrationType:    input.MigrationType,
		OldDepartmentID:  emp.DepartmentID,
		NewDepartmentID:  input.NewDepartmentID,
		OldSectionID:     emp.SectionID,
		NewSectionID:     input.NewSectionID,
		OldDesignationID: emp.DesignationID,
		NewDesignationID: input.NewDesignationID,
		OldLineID:        emp.LineID,
		NewLineID:        input.NewLineID,
		EffectiveDate:    effDate,
		Status:           "Completed",
		Reason:           input.Reason,
		Remarks:          input.Remarks,
		CreatedBy:        input.CreatedBy,
	}

	// Transaction to save migration and update employee record
	err = s.db.Transaction(func(tx *gorm.DB) error {
		txMigRepo := s.migrationRepo.WithTx(tx)
		if err := txMigRepo.Create(mig); err != nil {
			return err
		}

		// Apply updates to employee
		updates := map[string]interface{}{}
		if input.NewDepartmentID != nil {
			updates["department_id"] = *input.NewDepartmentID
		}
		if input.NewSectionID != nil {
			updates["section_id"] = *input.NewSectionID
		}
		if input.NewDesignationID != nil {
			updates["designation_id"] = *input.NewDesignationID
		}
		if input.NewLineID != nil {
			updates["line_id"] = *input.NewLineID
		}

		if len(updates) > 0 {
			if err := tx.Model(&models.Employee{}).Where("employee_id = ?", emp.EmployeeID).Updates(updates).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.migrationRepo.FindByID(mig.ID)
}
