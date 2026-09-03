package repository

import (
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ── Policy ──────────────────────────────────────────────────────────

type EarnedLeavePolicyRepository struct{ db *gorm.DB }

func NewEarnedLeavePolicyRepository(db *gorm.DB) *EarnedLeavePolicyRepository {
	return &EarnedLeavePolicyRepository{db: db}
}
func (r *EarnedLeavePolicyRepository) WithTx(tx *gorm.DB) *EarnedLeavePolicyRepository {
	return &EarnedLeavePolicyRepository{db: tx}
}
func (r *EarnedLeavePolicyRepository) Create(p *models.EarnedLeavePolicy) error {
	return r.db.Create(p).Error
}
func (r *EarnedLeavePolicyRepository) Update(p *models.EarnedLeavePolicy) error {
	return r.db.Save(p).Error
}
func (r *EarnedLeavePolicyRepository) FindByID(id string) (*models.EarnedLeavePolicy, error) {
	var p models.EarnedLeavePolicy
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&p).Error
	return &p, err
}
func (r *EarnedLeavePolicyRepository) ListByCompany(companyID string) ([]models.EarnedLeavePolicy, error) {
	var list []models.EarnedLeavePolicy
	err := r.db.Where("company_id = ? AND deleted_at IS NULL", companyID).Order("effective_from DESC").Find(&list).Error
	return list, err
}
func (r *EarnedLeavePolicyRepository) FindActive(companyID string, asOfDate string) (*models.EarnedLeavePolicy, error) {
	var p models.EarnedLeavePolicy
	q := r.db.Where("company_id = ? AND status = 'active' AND effective_from <= ? AND deleted_at IS NULL", companyID, asOfDate)
	// effective_to IS NULL OR >= asOfDate
	q = q.Where("(effective_to IS NULL OR effective_to >= ?)", asOfDate)
	err := q.Order("effective_from DESC").First(&p).Error
	return &p, err
}

// ── Ledger ──────────────────────────────────────────────────────────

type EarnedLeaveLedgerRepository struct{ db *gorm.DB }

func NewEarnedLeaveLedgerRepository(db *gorm.DB) *EarnedLeaveLedgerRepository {
	return &EarnedLeaveLedgerRepository{db: db}
}
func (r *EarnedLeaveLedgerRepository) WithTx(tx *gorm.DB) *EarnedLeaveLedgerRepository {
	return &EarnedLeaveLedgerRepository{db: tx}
}
func (r *EarnedLeaveLedgerRepository) CreateBatch(rows []models.EarnedLeaveLedger) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.CreateInBatches(rows, 200).Error
}
func (r *EarnedLeaveLedgerRepository) BatchLatestByEmployee(companyID string, employeeIDs []string, asOfDate string) (map[string]*models.EarnedLeaveLedger, error) {
	if len(employeeIDs) == 0 {
		return map[string]*models.EarnedLeaveLedger{}, nil
	}
	var rows []models.EarnedLeaveLedger
	err := r.db.Where("company_id = ? AND employee_id IN ? AND transaction_date <= ? AND deleted_at IS NULL", companyID, employeeIDs, asOfDate).
		Order("employee_id ASC, transaction_date DESC, created_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]*models.EarnedLeaveLedger, len(employeeIDs))
	for i := range rows {
		eid := rows[i].EmployeeID
		if _, ok := m[eid]; !ok {
			cp := rows[i]
			m[eid] = &cp
		}
	}
	return m, nil
}
func (r *EarnedLeaveLedgerRepository) BatchExistingAccrual(companyID, period string, employeeIDs []string) (map[string]bool, error) {
	if len(employeeIDs) == 0 {
		return map[string]bool{}, nil
	}
	var ids []string
	err := r.db.Model(&models.EarnedLeaveLedger{}).
		Where("company_id = ? AND period = ? AND transaction_type = 'ACCRUAL' AND employee_id IN ? AND deleted_at IS NULL", companyID, period, employeeIDs).
		Pluck("employee_id", &ids).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	return m, nil
}
func (r *EarnedLeaveLedgerRepository) List(companyID, employeeID string, from, to string, txType string, page, limit int) ([]models.EarnedLeaveLedger, int64, error) {
	q := r.db.Model(&models.EarnedLeaveLedger{}).Where("company_id = ? AND deleted_at IS NULL", companyID)
	if employeeID != "" {
		q = q.Where("employee_id = ?", employeeID)
	}
	if from != "" {
		q = q.Where("transaction_date >= ?", from)
	}
	if to != "" {
		q = q.Where("transaction_date <= ?", to)
	}
	if txType != "" {
		q = q.Where("transaction_type = ?", txType)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.EarnedLeaveLedger
	err := q.Order("transaction_date DESC, created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&rows).Error
	return rows, total, err
}
func (r *EarnedLeaveLedgerRepository) SumBalances(companyID string, employeeIDs []string, asOfDate string) (map[string]float64, error) {
	if len(employeeIDs) == 0 {
		return map[string]float64{}, nil
	}
	type row struct {
		EmployeeID string  `gorm:"column:employee_id"`
		Balance    float64 `gorm:"column:balance"`
	}
	var rows []row
	// closing = SUM(accrued + adjusted - used - encashed - expired) + opening of first
	// For simplicity we sum deltas; opening is handled via first row's opening in service
	err := r.db.Model(&models.EarnedLeaveLedger{}).
		Select("employee_id, COALESCE(SUM(accrued + adjusted - used - encashed - expired),0) as balance").
		Where("company_id = ? AND employee_id IN ? AND transaction_date <= ? AND deleted_at IS NULL", companyID, employeeIDs, asOfDate).
		Group("employee_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]float64, len(rows))
	for _, r := range rows {
		m[r.EmployeeID] = r.Balance
	}
	return m, nil
}

// ── Balance (materialized) ───────────────────────────────────────────

type EarnedLeaveBalanceRepository struct{ db *gorm.DB }

func NewEarnedLeaveBalanceRepository(db *gorm.DB) *EarnedLeaveBalanceRepository {
	return &EarnedLeaveBalanceRepository{db: db}
}
func (r *EarnedLeaveBalanceRepository) WithTx(tx *gorm.DB) *EarnedLeaveBalanceRepository {
	return &EarnedLeaveBalanceRepository{db: tx}
}
func (r *EarnedLeaveBalanceRepository) UpsertBalances(balances []models.EarnedLeaveBalance) error {
	if len(balances) == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "company_id"}, {Name: "employee_id"}, {Name: "year"}},
		DoUpdates: clause.AssignmentColumns([]string{"opening", "accrued", "used", "adjusted", "encashed", "expired", "closing", "last_ledger_id", "updated_at"}),
	}).CreateInBatches(balances, 100).Error
}
func (r *EarnedLeaveBalanceRepository) GetByEmployeeYear(companyID, employeeID string, year int) (*models.EarnedLeaveBalance, error) {
	var b models.EarnedLeaveBalance
	err := r.db.Where("company_id = ? AND employee_id = ? AND year = ? AND deleted_at IS NULL", companyID, employeeID, year).First(&b).Error
	return &b, err
}
func (r *EarnedLeaveBalanceRepository) ListByCompanyYear(companyID string, year int, page, limit int) ([]models.EarnedLeaveBalance, int64, error) {
	q := r.db.Model(&models.EarnedLeaveBalance{}).Where("company_id = ? AND year = ? AND deleted_at IS NULL", companyID, year)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.EarnedLeaveBalance
	err := q.Order("employee_id ASC").Offset((page - 1) * limit).Limit(limit).Find(&rows).Error
	return rows, total, err
}

// ── Salary Sheet ─────────────────────────────────────────────────────

type EarnedLeaveSalaryRepository struct{ db *gorm.DB }

func NewEarnedLeaveSalaryRepository(db *gorm.DB) *EarnedLeaveSalaryRepository {
	return &EarnedLeaveSalaryRepository{db: db}
}
func (r *EarnedLeaveSalaryRepository) WithTx(tx *gorm.DB) *EarnedLeaveSalaryRepository {
	return &EarnedLeaveSalaryRepository{db: tx}
}
func (r *EarnedLeaveSalaryRepository) FindSheetByPeriod(companyID string, month, year int) (*models.EarnedLeaveSalarySheet, error) {
	var s models.EarnedLeaveSalarySheet
	err := r.db.Where("company_id = ? AND period_month = ? AND period_year = ? AND deleted_at IS NULL", companyID, month, year).First(&s).Error
	return &s, err
}
func (r *EarnedLeaveSalaryRepository) CreateSheetWithItems(sheet *models.EarnedLeaveSalarySheet, items []models.EarnedLeaveSalaryItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(sheet).Error; err != nil {
			return err
		}
		if len(items) > 0 {
			// assign sheet_id
			for i := range items {
				items[i].SheetID = sheet.ID
			}
			if err := tx.CreateInBatches(items, 100).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *EarnedLeaveSalaryRepository) UpdateSheetStatus(id, status, userID string) error {
	updates := map[string]interface{}{"status": status, "updated_at": time.Now()}
	if status == "APPROVED" {
		updates["approved_by"] = userID
		updates["approved_at"] = time.Now()
	}
	if status == "FINALIZED" {
		updates["finalized_by"] = userID
		updates["finalized_at"] = time.Now()
	}
	return r.db.Model(&models.EarnedLeaveSalarySheet{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates).Error
}
func (r *EarnedLeaveSalaryRepository) ListSheets(companyID string, page, limit int) ([]models.EarnedLeaveSalarySheet, int64, error) {
	q := r.db.Model(&models.EarnedLeaveSalarySheet{}).Where("company_id = ? AND deleted_at IS NULL", companyID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.EarnedLeaveSalarySheet
	err := q.Order("period_year DESC, period_month DESC").Offset((page - 1) * limit).Limit(limit).Find(&rows).Error
	return rows, total, err
}
func (r *EarnedLeaveSalaryRepository) GetSheet(id string) (*models.EarnedLeaveSalarySheet, error) {
	var s models.EarnedLeaveSalarySheet
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&s).Error
	return &s, err
}
func (r *EarnedLeaveSalaryRepository) ListItems(sheetID string) ([]models.EarnedLeaveSalaryItem, error) {
	var rows []models.EarnedLeaveSalaryItem
	err := r.db.Preload("Employee.Department").Preload("Employee.DesignationRef").
		Where("sheet_id = ? AND deleted_at IS NULL", sheetID).
		Order("LENGTH(employee_id) ASC, employee_id ASC").
		Find(&rows).Error
	return rows, err
}
