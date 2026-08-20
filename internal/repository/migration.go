package repository

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

type MigrationRepository struct {
	db *gorm.DB
}

func NewMigrationRepository(db *gorm.DB) *MigrationRepository {
	return &MigrationRepository{db: db}
}

func (r *MigrationRepository) WithTx(tx *gorm.DB) *MigrationRepository {
	return &MigrationRepository{db: tx}
}

type OrganizationalSummaryRow struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	NewJoining  int64  `json:"new_joining"`
	TotalLeft   int64  `json:"total_left"`
	TotalResign int64  `json:"total_resign"`
	ActiveTotal int64  `json:"active_total"`
}

type FullMigrationSummary struct {
	TotalJoining  int64                      `json:"total_joining"`
	TotalLeft     int64                      `json:"total_left"`
	TotalResign   int64                      `json:"total_resign"`
	TotalActive   int64                      `json:"total_active"`
	ByDepartment  []OrganizationalSummaryRow `json:"by_department"`
	BySection     []OrganizationalSummaryRow `json:"by_section"`
	ByDesignation []OrganizationalSummaryRow `json:"by_designation"`
	ByLine        []OrganizationalSummaryRow `json:"by_line"`
}

func (r *MigrationRepository) Create(m *models.EmployeeMigration) error {
	return r.db.Create(m).Error
}

func (r *MigrationRepository) FindByID(id string) (*models.EmployeeMigration, error) {
	var m models.EmployeeMigration
	err := r.db.Preload("OldDepartment").Preload("NewDepartment").
		Preload("OldSection").Preload("NewSection").
		Preload("OldDesignation").Preload("NewDesignation").
		Preload("OldLine").Preload("NewLine").
		Where("id = ? AND deleted_at IS NULL", id).First(&m).Error
	return &m, err
}

func (r *MigrationRepository) buildQuery(employee, employeeID, departmentID, sectionID, designationID, lineID, migrationType, status, companyID, dateFrom, dateTo string) *gorm.DB {
	base := r.db.Model(&models.EmployeeMigration{}).Where("employee_migrations.deleted_at IS NULL").Order("employee_migrations.effective_date DESC, employee_migrations.created_at DESC")

	if employee != "" {
		base = base.Where("employee_migrations.employee_name ILIKE ?", "%"+employee+"%")
	}
	if employeeID != "" {
		base = base.Where("employee_migrations.employee_id ILIKE ?", "%"+employeeID+"%")
	}
	if departmentID != "" {
		base = base.Where("employee_migrations.new_department_id = ? OR employee_migrations.old_department_id = ?", departmentID, departmentID)
	}
	if sectionID != "" {
		base = base.Where("employee_migrations.new_section_id = ? OR employee_migrations.old_section_id = ?", sectionID, sectionID)
	}
	if designationID != "" {
		base = base.Where("employee_migrations.new_designation_id = ? OR employee_migrations.old_designation_id = ?", designationID, designationID)
	}
	if lineID != "" {
		base = base.Where("employee_migrations.new_line_id = ? OR employee_migrations.old_line_id = ?", lineID, lineID)
	}
	if migrationType != "" {
		base = base.Where("employee_migrations.migration_type = ?", migrationType)
	}
	if status != "" {
		base = base.Where("employee_migrations.status = ?", status)
	}
	if companyID != "" {
		base = base.Where("employee_migrations.company_id = ?", companyID)
	}
	if dateFrom != "" {
		base = base.Where("employee_migrations.effective_date >= ?", dateFrom)
	}
	if dateTo != "" {
		base = base.Where("employee_migrations.effective_date <= ?", dateTo)
	}

	return base
}

func (r *MigrationRepository) ListFiltered(employee, employeeID, departmentID, sectionID, designationID, lineID, migrationType, status, companyID, dateFrom, dateTo string, page, limit int) ([]models.EmployeeMigration, int64, error) {
	base := r.buildQuery(employee, employeeID, departmentID, sectionID, designationID, lineID, migrationType, status, companyID, dateFrom, dateTo)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.EmployeeMigration
	err := base.Preload("OldDepartment").Preload("NewDepartment").
		Preload("OldSection").Preload("NewSection").
		Preload("OldDesignation").Preload("NewDesignation").
		Preload("OldLine").Preload("NewLine").
		Offset((page - 1) * limit).Limit(limit).Find(&items).Error

	return items, total, err
}

func (r *MigrationRepository) ListAllFiltered(employee, employeeID, departmentID, sectionID, designationID, lineID, migrationType, status, companyID, dateFrom, dateTo string) ([]models.EmployeeMigration, error) {
	base := r.buildQuery(employee, employeeID, departmentID, sectionID, designationID, lineID, migrationType, status, companyID, dateFrom, dateTo)

	var items []models.EmployeeMigration
	err := base.Preload("OldDepartment").Preload("NewDepartment").
		Preload("OldSection").Preload("NewSection").
		Preload("OldDesignation").Preload("NewDesignation").
		Preload("OldLine").Preload("NewLine").
		Find(&items).Error

	return items, err
}

func (r *MigrationRepository) GetFullSummary(dateFrom, dateTo, companyID, departmentID, sectionID, designationID, lineID string) (*FullMigrationSummary, error) {
	summary := &FullMigrationSummary{
		ByDepartment:  []OrganizationalSummaryRow{},
		BySection:     []OrganizationalSummaryRow{},
		ByDesignation: []OrganizationalSummaryRow{},
		ByLine:        []OrganizationalSummaryRow{},
	}

	// 1. Calculate Top Totals
	empJoinQ := r.db.Model(&models.Employee{}).Where("deleted_at IS NULL")
	if companyID != "" {
		empJoinQ = empJoinQ.Where("company_id = ?", companyID)
	}
	if dateFrom != "" {
		empJoinQ = empJoinQ.Where("joining_date >= ?", dateFrom)
	}
	if dateTo != "" {
		empJoinQ = empJoinQ.Where("joining_date <= ?", dateTo+" 23:59:59")
	}
	_ = empJoinQ.Count(&summary.TotalJoining)

	sepQ := r.db.Model(&models.Separation{}).Where("deleted_at IS NULL")
	if companyID != "" {
		sepQ = sepQ.Where("company_id = ?", companyID)
	}
	if dateFrom != "" {
		sepQ = sepQ.Where("date >= ?", dateFrom)
	}
	if dateTo != "" {
		sepQ = sepQ.Where("date <= ?", dateTo)
	}
	_ = sepQ.Count(&summary.TotalLeft)

	resignQ := r.db.Model(&models.Separation{}).Where("deleted_at IS NULL AND (type ILIKE '%Resign%' OR type ILIKE '%resignation%')")
	if companyID != "" {
		resignQ = resignQ.Where("company_id = ?", companyID)
	}
	if dateFrom != "" {
		resignQ = resignQ.Where("date >= ?", dateFrom)
	}
	if dateTo != "" {
		resignQ = resignQ.Where("date <= ?", dateTo)
	}
	_ = resignQ.Count(&summary.TotalResign)

	activeQ := r.db.Model(&models.Employee{}).Where("deleted_at IS NULL AND status = 'active'")
	if companyID != "" {
		activeQ = activeQ.Where("company_id = ?", companyID)
	}
	_ = activeQ.Count(&summary.TotalActive)

	// 2. Department Breakdown
	var depts []models.Department
	deptQuery := r.db.Where("deleted_at IS NULL")
	if companyID != "" {
		deptQuery = deptQuery.Where("company_id = ?", companyID)
	}
	if departmentID != "" {
		deptQuery = deptQuery.Where("id = ?", departmentID)
	}
	_ = deptQuery.Order("name ASC").Find(&depts)

	for _, d := range depts {
		var jCount, lCount, rCount, aCount int64

		jQ := r.db.Model(&models.Employee{}).Where("department_id = ? AND deleted_at IS NULL", d.ID)
		if dateFrom != "" {
			jQ = jQ.Where("joining_date >= ?", dateFrom)
		}
		if dateTo != "" {
			jQ = jQ.Where("joining_date <= ?", dateTo+" 23:59:59")
		}
		_ = jQ.Count(&jCount)

		sQ := r.db.Model(&models.Separation{}).Where("department_id = ? AND deleted_at IS NULL", d.ID)
		if dateFrom != "" {
			sQ = sQ.Where("date >= ?", dateFrom)
		}
		if dateTo != "" {
			sQ = sQ.Where("date <= ?", dateTo)
		}
		_ = sQ.Count(&lCount)

		rQ := r.db.Model(&models.Separation{}).Where("department_id = ? AND deleted_at IS NULL AND (type ILIKE '%Resign%' OR type ILIKE '%resignation%')", d.ID)
		if dateFrom != "" {
			rQ = rQ.Where("date >= ?", dateFrom)
		}
		if dateTo != "" {
			rQ = rQ.Where("date <= ?", dateTo)
		}
		_ = rQ.Count(&rCount)

		aQ := r.db.Model(&models.Employee{}).Where("department_id = ? AND status = 'active' AND deleted_at IS NULL", d.ID)
		_ = aQ.Count(&aCount)

		summary.ByDepartment = append(summary.ByDepartment, OrganizationalSummaryRow{
			ID:          d.ID,
			Name:        d.Name,
			NewJoining:  jCount,
			TotalLeft:   lCount,
			TotalResign: rCount,
			ActiveTotal: aCount,
		})
	}

	// 3. Section Breakdown
	var secs []models.Section
	secQuery := r.db.Where("deleted_at IS NULL")
	if departmentID != "" {
		secQuery = secQuery.Where("department_id = ?", departmentID)
	}
	if sectionID != "" {
		secQuery = secQuery.Where("id = ?", sectionID)
	}
	_ = secQuery.Order("name ASC").Find(&secs)

	for _, s := range secs {
		var jCount, lCount, rCount, aCount int64

		jQ := r.db.Model(&models.Employee{}).Where("section_id = ? AND deleted_at IS NULL", s.ID)
		if dateFrom != "" {
			jQ = jQ.Where("joining_date >= ?", dateFrom)
		}
		if dateTo != "" {
			jQ = jQ.Where("joining_date <= ?", dateTo+" 23:59:59")
		}
		_ = jQ.Count(&jCount)

		// Separation by employee section_id
		sQ := r.db.Model(&models.Separation{}).
			Joins("JOIN employees ON employees.employee_id = separations.employee_id").
			Where("employees.section_id = ? AND separations.deleted_at IS NULL", s.ID)
		if dateFrom != "" {
			sQ = sQ.Where("separations.date >= ?", dateFrom)
		}
		if dateTo != "" {
			sQ = sQ.Where("separations.date <= ?", dateTo)
		}
		_ = sQ.Count(&lCount)

		rQ := r.db.Model(&models.Separation{}).
			Joins("JOIN employees ON employees.employee_id = separations.employee_id").
			Where("employees.section_id = ? AND separations.deleted_at IS NULL AND (separations.type ILIKE '%Resign%' OR separations.type ILIKE '%resignation%')", s.ID)
		if dateFrom != "" {
			rQ = rQ.Where("separations.date >= ?", dateFrom)
		}
		if dateTo != "" {
			rQ = rQ.Where("separations.date <= ?", dateTo)
		}
		_ = rQ.Count(&rCount)

		aQ := r.db.Model(&models.Employee{}).Where("section_id = ? AND status = 'active' AND deleted_at IS NULL", s.ID)
		_ = aQ.Count(&aCount)

		summary.BySection = append(summary.BySection, OrganizationalSummaryRow{
			ID:          s.ID,
			Name:        s.Name,
			NewJoining:  jCount,
			TotalLeft:   lCount,
			TotalResign: rCount,
			ActiveTotal: aCount,
		})
	}

	// 4. Designation Breakdown
	var desigs []models.Designation
	desigQuery := r.db.Where("deleted_at IS NULL")
	if sectionID != "" {
		desigQuery = desigQuery.Where("section_id = ?", sectionID)
	}
	if designationID != "" {
		desigQuery = desigQuery.Where("id = ?", designationID)
	}
	_ = desigQuery.Order("name ASC").Find(&desigs)

	for _, des := range desigs {
		var jCount, lCount, rCount, aCount int64

		jQ := r.db.Model(&models.Employee{}).Where("designation_id = ? AND deleted_at IS NULL", des.ID)
		if dateFrom != "" {
			jQ = jQ.Where("joining_date >= ?", dateFrom)
		}
		if dateTo != "" {
			jQ = jQ.Where("joining_date <= ?", dateTo+" 23:59:59")
		}
		_ = jQ.Count(&jCount)

		sQ := r.db.Model(&models.Separation{}).
			Joins("JOIN employees ON employees.employee_id = separations.employee_id").
			Where("employees.designation_id = ? AND separations.deleted_at IS NULL", des.ID)
		if dateFrom != "" {
			sQ = sQ.Where("separations.date >= ?", dateFrom)
		}
		if dateTo != "" {
			sQ = sQ.Where("separations.date <= ?", dateTo)
		}
		_ = sQ.Count(&lCount)

		rQ := r.db.Model(&models.Separation{}).
			Joins("JOIN employees ON employees.employee_id = separations.employee_id").
			Where("employees.designation_id = ? AND separations.deleted_at IS NULL AND (separations.type ILIKE '%Resign%' OR separations.type ILIKE '%resignation%')", des.ID)
		if dateFrom != "" {
			rQ = rQ.Where("separations.date >= ?", dateFrom)
		}
		if dateTo != "" {
			rQ = rQ.Where("separations.date <= ?", dateTo)
		}
		_ = rQ.Count(&rCount)

		aQ := r.db.Model(&models.Employee{}).Where("designation_id = ? AND status = 'active' AND deleted_at IS NULL", des.ID)
		_ = aQ.Count(&aCount)

		summary.ByDesignation = append(summary.ByDesignation, OrganizationalSummaryRow{
			ID:          des.ID,
			Name:        des.Name,
			NewJoining:  jCount,
			TotalLeft:   lCount,
			TotalResign: rCount,
			ActiveTotal: aCount,
		})
	}

	// 5. Line Breakdown
	var lines []models.Line
	lineQuery := r.db.Where("deleted_at IS NULL")
	if sectionID != "" {
		lineQuery = lineQuery.Where("section_id = ?", sectionID)
	}
	if lineID != "" {
		lineQuery = lineQuery.Where("id = ?", lineID)
	}
	_ = lineQuery.Order("name ASC").Find(&lines)

	for _, l := range lines {
		var jCount, lCount, rCount, aCount int64

		jQ := r.db.Model(&models.Employee{}).Where("line_id = ? AND deleted_at IS NULL", l.ID)
		if dateFrom != "" {
			jQ = jQ.Where("joining_date >= ?", dateFrom)
		}
		if dateTo != "" {
			jQ = jQ.Where("joining_date <= ?", dateTo+" 23:59:59")
		}
		_ = jQ.Count(&jCount)

		sQ := r.db.Model(&models.Separation{}).
			Joins("JOIN employees ON employees.employee_id = separations.employee_id").
			Where("employees.line_id = ? AND separations.deleted_at IS NULL", l.ID)
		if dateFrom != "" {
			sQ = sQ.Where("separations.date >= ?", dateFrom)
		}
		if dateTo != "" {
			sQ = sQ.Where("separations.date <= ?", dateTo)
		}
		_ = sQ.Count(&lCount)

		rQ := r.db.Model(&models.Separation{}).
			Joins("JOIN employees ON employees.employee_id = separations.employee_id").
			Where("employees.line_id = ? AND separations.deleted_at IS NULL AND (separations.type ILIKE '%Resign%' OR separations.type ILIKE '%resignation%')", l.ID)
		if dateFrom != "" {
			rQ = rQ.Where("separations.date >= ?", dateFrom)
		}
		if dateTo != "" {
			rQ = rQ.Where("separations.date <= ?", dateTo)
		}
		_ = rQ.Count(&rCount)

		aQ := r.db.Model(&models.Employee{}).Where("line_id = ? AND status = 'active' AND deleted_at IS NULL", l.ID)
		_ = aQ.Count(&aCount)

		summary.ByLine = append(summary.ByLine, OrganizationalSummaryRow{
			ID:          l.ID,
			Name:        l.Name,
			NewJoining:  jCount,
			TotalLeft:   lCount,
			TotalResign: rCount,
			ActiveTotal: aCount,
		})
	}

	return summary, nil
}
