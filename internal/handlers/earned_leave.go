package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/service"
)

type EarnedLeaveHandler struct {
	service    *service.EarnedLeaveService
	policyRepo *repository.EarnedLeavePolicyRepository
	ledgerRepo *repository.EarnedLeaveLedgerRepository
	salaryRepo *repository.EarnedLeaveSalaryRepository
}

func NewEarnedLeaveHandler(
	svc *service.EarnedLeaveService,
	policyRepo *repository.EarnedLeavePolicyRepository,
	ledgerRepo *repository.EarnedLeaveLedgerRepository,
	salaryRepo *repository.EarnedLeaveSalaryRepository,
) *EarnedLeaveHandler {
	return &EarnedLeaveHandler{service: svc, policyRepo: policyRepo, ledgerRepo: ledgerRepo, salaryRepo: salaryRepo}
}

// ── Policy ──────────────────────────────────────────────────────────

type CreatePolicyRequest struct {
	CompanyID           string  `json:"company_id" binding:"required"`
	Name                string  `json:"name" binding:"required"`
	AccrualRate         float64 `json:"accrual_rate" binding:"required"`
	AccrualFrequency    string  `json:"accrual_frequency"`
	MaxBalance          float64 `json:"max_balance"`
	MinServiceMonths    int     `json:"min_service_months"`
	CarryForwardAllowed bool    `json:"carry_forward_allowed"`
	CarryForwardLimit   float64 `json:"carry_forward_limit"`
	EncashmentAllowed   bool    `json:"encashment_allowed"`
	EncashmentBasis     string  `json:"encashment_basis"`
	EncashmentDivisor   int     `json:"encashment_divisor"`
	RoundingRule        string  `json:"rounding_rule"`
	EffectiveFrom       string  `json:"effective_from" binding:"required"`
	EffectiveTo         *string `json:"effective_to"`
}

// CreatePolicy godoc
// @Summary Create EL policy
// @Tags Earned Leave
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreatePolicyRequest true "policy"
// @Success 201 {object} models.EarnedLeavePolicy
// @Failure 400 {object} map[string]string
// @Router /earned-leaves/policies [post]
func (h *EarnedLeaveHandler) CreatePolicy(c *gin.Context) {
	var req CreatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	companyID := req.CompanyID
	if companyID == "" {
		companyID = c.GetString("company_id")
	}
	if companyID == "" {
		companyID = c.Query("company_id")
	}
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id is required"})
		return
	}
	userID := c.GetString("user_id")
	p := &models.EarnedLeavePolicy{
		CompanyID:           companyID,
		Name:                req.Name,
		AccrualRate:         req.AccrualRate,
		AccrualFrequency:    req.AccrualFrequency,
		MaxBalance:          req.MaxBalance,
		MinServiceMonths:    req.MinServiceMonths,
		CarryForwardAllowed: req.CarryForwardAllowed,
		CarryForwardLimit:   req.CarryForwardLimit,
		EncashmentAllowed:   req.EncashmentAllowed,
		EncashmentBasis:     req.EncashmentBasis,
		EncashmentDivisor:   req.EncashmentDivisor,
		RoundingRule:        req.RoundingRule,
		EffectiveFrom:       req.EffectiveFrom,
		EffectiveTo:         req.EffectiveTo,
		Status:              "active",
		CreatedBy:           &userID,
	}
	if p.AccrualFrequency == "" {
		p.AccrualFrequency = "MONTHLY"
	}
	if p.EncashmentBasis == "" {
		p.EncashmentBasis = "BASIC"
	}
	if p.EncashmentDivisor == 0 {
		p.EncashmentDivisor = 30
	}
	if p.RoundingRule == "" {
		p.RoundingRule = "ROUND"
	}
	if err := h.policyRepo.Create(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// ListPolicies godoc
// @Summary List EL policies
// @Tags Earned Leave
// @Security BearerAuth
// @Produce json
// @Param company_id query string true "Company ID"
// @Success 200 {array} models.EarnedLeavePolicy
// @Router /earned-leaves/policies [get]
func (h *EarnedLeaveHandler) ListPolicies(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		companyID = c.GetString("company_id")
	}
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id required"})
		return
	}
	list, err := h.policyRepo.ListByCompany(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": len(list)})
}

// ── Accrual ──────────────────────────────────────────────────────────

type AccrualProcessRequest struct {
	CompanyID string `json:"company_id" binding:"required"`
	Period    string `json:"period" binding:"required"` // YYYY-MM
}

// AccrualProcess godoc
// @Summary Process EL accrual
// @Tags Earned Leave
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body AccrualProcessRequest true "accrual"
// @Success 200 {object} service.ELProcessResult
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /earned-leaves/accrual/process [post]
func (h *EarnedLeaveHandler) AccrualProcess(c *gin.Context) {
	var req AccrualProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	companyID := req.CompanyID
	if companyID == "" {
		companyID = c.GetString("company_id")
	}
	userID := c.GetString("user_id")
	res, err := h.service.AccrualProcess(companyID, req.Period, userID)
	if err != nil {
		if strings.Contains(err.Error(), "already processing") || strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// ── Balance / Ledger ─────────────────────────────────────────────────

 // GetBalance godoc
 // @Summary Get EL balance
 // @Tags Earned Leave
 // @Security BearerAuth
 // @Produce json
 // @Param company_id query string true "Company ID"
 // @Param employee_id query string true "Employee ID"
 // @Param as_of query string false "As of date YYYY-MM-DD"
 // @Success 200 {object} map[string]interface{}
 // @Router /earned-leaves/balance [get]
func (h *EarnedLeaveHandler) GetBalance(c *gin.Context) {
	companyID := c.Query("company_id")
	employeeID := c.Query("employee_id")
	asOf := c.Query("as_of")
	if asOf == "" {
		asOf = c.Query("as_of_date")
	}
	if companyID == "" || employeeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id and employee_id required"})
		return
	}
	// Use ledger sum + latest opening
	// For MVP, compute via ledgerRepo.SumBalances for single employee
	m, err := h.ledgerRepo.SumBalances(companyID, []string{employeeID}, asOf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	closing := m[employeeID]
	// Add opening from latest ledger if needed
	latestMap, _ := h.ledgerRepo.BatchLatestByEmployee(companyID, []string{employeeID}, asOf)
	opening := 0.0
	if latest, ok := latestMap[employeeID]; ok && latest != nil {
		opening = latest.OpeningBalance
		closing = latest.ClosingBalance
	}
	c.JSON(http.StatusOK, gin.H{"employee_id": employeeID, "as_of": asOf, "opening": opening, "balance": closing, "closing": closing})
}

// ListLedger godoc
// @Summary List EL ledger
// @Tags Earned Leave
// @Security BearerAuth
// @Produce json
// @Param company_id query string true "Company ID"
// @Param employee_id query string false "Employee ID"
// @Param from query string false "From YYYY-MM-DD"
// @Param to query string false "To YYYY-MM-DD"
// @Param transaction_type query string false "Type"
// @Param page query int false "Page"
// @Param limit query int false "Limit"
// @Success 200 {object} map[string]interface{}
// @Router /earned-leaves/ledger [get]
func (h *EarnedLeaveHandler) ListLedger(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		companyID = c.GetString("company_id")
	}
	employeeID := c.Query("employee_id")
	from := c.Query("from")
	to := c.Query("to")
	txType := c.Query("transaction_type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	rows, total, err := h.ledgerRepo.List(companyID, employeeID, from, to, txType, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows, "total": total, "page": page, "limit": limit})
}

// ── Adjustment ──────────────────────────────────────────────────────

type AdjustmentRequest struct {
	CompanyID     string  `json:"company_id" binding:"required"`
	EmployeeID    string  `json:"employee_id" binding:"required"`
	AdjustmentDate string `json:"adjustment_date" binding:"required"`
	Type          string  `json:"type" binding:"required"` // ADD, DEDUCT
	Days          float64 `json:"days" binding:"required"`
	Reason        string  `json:"reason" binding:"required"`
	Remarks       string  `json:"remarks"`
}

// CreateAdjustment godoc
// @Summary EL adjustment
// @Tags Earned Leave
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body AdjustmentRequest true "adjustment"
// @Success 201 {object} models.EarnedLeaveLedger
// @Router /earned-leaves/adjustment [post]
func (h *EarnedLeaveHandler) CreateAdjustment(c *gin.Context) {
	var req AdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Type != "ADD" && req.Type != "DEDUCT" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be ADD or DEDUCT"})
		return
	}
	if req.Days <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "days must be > 0"})
		return
	}
	userID := c.GetString("user_id")
	// Get current closing for opening
	latestMap, _ := h.ledgerRepo.BatchLatestByEmployee(req.CompanyID, []string{req.EmployeeID}, req.AdjustmentDate)
	opening := 0.0
	if latest, ok := latestMap[req.EmployeeID]; ok && latest != nil {
		opening = latest.ClosingBalance
	}
	adjusted := req.Days
	if req.Type == "DEDUCT" {
		adjusted = -req.Days
	}
	closing := opening + adjusted
	if closing < 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "insufficient EL balance"})
		return
	}
	txType := "ADJUSTMENT_ADD"
	if req.Type == "DEDUCT" {
		txType = "ADJUSTMENT_DEDUCT"
	}
	row := models.EarnedLeaveLedger{
		CompanyID:       req.CompanyID,
		EmployeeID:      req.EmployeeID,
		TransactionDate: req.AdjustmentDate,
		Period:          req.AdjustmentDate[:7],
		TransactionType: txType,
		OpeningBalance:  opening,
		Adjusted:        adjusted,
		ClosingBalance:  closing,
		Remarks:         req.Remarks + " | " + req.Reason,
		CreatedBy:       &userID,
	}
	if err := h.ledgerRepo.CreateBatch([]models.EarnedLeaveLedger{row}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, row)
}

// ── Encashment ──────────────────────────────────────────────────────

type EncashmentRequest struct {
	CompanyID     string  `json:"company_id" binding:"required"`
	EmployeeID    string  `json:"employee_id" binding:"required"`
	EncashmentDate string `json:"encashment_date" binding:"required"`
	Days          float64 `json:"days" binding:"required"`
	Remarks       string  `json:"remarks"`
}

// CreateEncashment godoc
// @Summary EL encashment
// @Tags Earned Leave
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body EncashmentRequest true "encashment"
// @Success 201 {object} models.EarnedLeaveLedger
// @Router /earned-leaves/encashment [post]
func (h *EarnedLeaveHandler) CreateEncashment(c *gin.Context) {
	var req EncashmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Days <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "days must be > 0"})
		return
	}
	userID := c.GetString("user_id")
	latestMap, _ := h.ledgerRepo.BatchLatestByEmployee(req.CompanyID, []string{req.EmployeeID}, req.EncashmentDate)
	opening := 0.0
	if latest, ok := latestMap[req.EmployeeID]; ok && latest != nil {
		opening = latest.ClosingBalance
	}
	if opening < req.Days {
		c.JSON(http.StatusConflict, gin.H{"error": "insufficient EL balance for encashment"})
		return
	}
	closing := opening - req.Days
	row := models.EarnedLeaveLedger{
		CompanyID:       req.CompanyID,
		EmployeeID:      req.EmployeeID,
		TransactionDate: req.EncashmentDate,
		Period:          req.EncashmentDate[:7],
		TransactionType: "ENCASHMENT",
		OpeningBalance:  opening,
		Encashed:        req.Days,
		ClosingBalance:  closing,
		Remarks:         req.Remarks,
		CreatedBy:       &userID,
	}
	if err := h.ledgerRepo.CreateBatch([]models.EarnedLeaveLedger{row}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, row)
}

// ── Salary Sheet ─────────────────────────────────────────────────────

type GenerateSheetRequest struct {
	CompanyID string `json:"company_id" binding:"required"`
	Month     int    `json:"month" binding:"required"`
	Year      int    `json:"year" binding:"required"`
}

// GenerateSalarySheet godoc
// @Summary Generate EL salary sheet
// @Tags Earned Leave
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body GenerateSheetRequest true "generate"
// @Success 201 {object} models.EarnedLeaveSalarySheet
// @Router /earned-leaves/salary-sheet/generate [post]
func (h *EarnedLeaveHandler) GenerateSalarySheet(c *gin.Context) {
	var req GenerateSheetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Month < 1 || req.Month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "month must be 1-12"})
		return
	}
	userID := c.GetString("user_id")
	sheet, err := h.service.GenerateSalarySheet(req.CompanyID, req.Month, req.Year, userID)
	if err != nil {
		if contains(err.Error(), "already exists") || contains(err.Error(), "already processing") || contains(err.Error(), "finalized") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, sheet)
}

// ListSheets godoc
// @Summary List EL salary sheets
// @Tags Earned Leave
// @Security BearerAuth
// @Produce json
// @Param company_id query string true "Company ID"
// @Param page query int false "Page"
// @Param limit query int false "Limit"
// @Success 200 {object} map[string]interface{}
// @Router /earned-leaves/salary-sheet [get]
func (h *EarnedLeaveHandler) ListSheets(c *gin.Context) {
	companyID := c.Query("company_id")
	if companyID == "" {
		companyID = c.GetString("company_id")
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	rows, total, err := h.salaryRepo.ListSheets(companyID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows, "total": total, "page": page, "limit": limit})
}

// GetSheet godoc
// @Summary Get EL salary sheet
// @Tags Earned Leave
// @Security BearerAuth
// @Produce json
// @Param id path string true "Sheet ID"
// @Success 200 {object} map[string]interface{}
// @Router /earned-leaves/salary-sheet/:id [get]
func (h *EarnedLeaveHandler) GetSheet(c *gin.Context) {
	id := c.Param("id")
	sheet, err := h.salaryRepo.GetSheet(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sheet not found"})
		return
	}
	items, _ := h.salaryRepo.ListItems(id)
	c.JSON(http.StatusOK, gin.H{"sheet": sheet, "items": items, "total_items": len(items)})
}

// ApproveSheet godoc
// @Summary Approve EL salary sheet
// @Tags Earned Leave
// @Security BearerAuth
// @Produce json
// @Param id path string true "Sheet ID"
// @Success 200 {object} map[string]string
// @Router /earned-leaves/salary-sheet/:id/approve [post]
func (h *EarnedLeaveHandler) ApproveSheet(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")
	if err := h.salaryRepo.UpdateSheetStatus(id, "APPROVED", userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sheet approved"})
}

// FinalizeSheet godoc
// @Summary Finalize EL salary sheet
// @Tags Earned Leave
// @Security BearerAuth
// @Produce json
// @Param id path string true "Sheet ID"
// @Success 200 {object} map[string]string
// @Router /earned-leaves/salary-sheet/:id/finalize [post]
func (h *EarnedLeaveHandler) FinalizeSheet(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")
	if err := h.salaryRepo.UpdateSheetStatus(id, "FINALIZED", userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sheet finalized"})
}

// ReverseSheet godoc
// @Summary Reverse EL salary sheet
// @Tags Earned Leave
// @Security BearerAuth
// @Produce json
// @Param id path string true "Sheet ID"
// @Success 200 {object} map[string]string
// @Router /earned-leaves/salary-sheet/:id/reverse [post]
func (h *EarnedLeaveHandler) ReverseSheet(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")
	if err := h.salaryRepo.UpdateSheetStatus(id, "REVERSED", userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sheet reversed"})
}

func contains(s, substr string) bool { return strings.Contains(s, substr) }
