package service

import (
	"fmt"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
)

type EidBonusService struct {
	employeeRepo  *repository.EmployeeRepository
	eidBonusRepo  *repository.EidBonusRepository
}

func NewEidBonusService(
	employeeRepo *repository.EmployeeRepository,
	eidBonusRepo *repository.EidBonusRepository,
) *EidBonusService {
	return &EidBonusService{
		employeeRepo:  employeeRepo,
		eidBonusRepo:  eidBonusRepo,
	}
}

type EidBonusResult struct {
	Processed int
	Total     int
	Year      int
}

func (s *EidBonusService) ProcessYear(companyID string, year int, userID string) (*EidBonusResult, error) {
	employees, err := s.employeeRepo.ListActiveAll(companyID)
	if err != nil {
		return nil, fmt.Errorf("fetch employees: %w", err)
	}

	if len(employees) == 0 {
		return &EidBonusResult{Processed: 0, Total: 0, Year: year}, nil
	}

	processed := 0

	for _, emp := range employees {
		bonusAmount := emp.GrossSalary

		bonus := &models.EidBonus{
			CompanyID:   emp.CompanyID,
			EmployeeID:  emp.EmployeeID,
			Year:        year,
			BonusType:   "eid",
			GrossSalary: emp.GrossSalary,
			BonusAmount: bonusAmount,
			Status:      "processed",
			CreatedBy:   &userID,
		}

		if err := s.eidBonusRepo.Upsert(bonus); err != nil {
			continue
		}
		processed++
	}

	return &EidBonusResult{
		Processed: processed,
		Total:     len(employees),
		Year:      year,
	}, nil
}
