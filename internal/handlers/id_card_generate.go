package handlers

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/service"
)

type GenerateIdCardRequest struct {
	EmployeeIDs []string `json:"employee_ids" binding:"required"`
}

// GenerateIdCards godoc
//
//	@Summary      Generate ID cards PDF
//	@Description  Generate printable ID cards (front + back on same page, 3 employees per A4 portrait page, 35x45mm photo). English design without QR or barcode.
//	@Tags         ID Cards
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        request body GenerateIdCardRequest true "Employee IDs (business keys)"
//	@Success      200  {object}  map[string]string
//	@Failure      400  {object}  map[string]string
//	@Failure      500  {object}  map[string]string
//	@Router       /id-cards/generate [post]
func (h *IdCardHandler) Generate(c *gin.Context) {
	var req GenerateIdCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if len(req.EmployeeIDs) == 0 {
		c.JSON(400, gin.H{"error": "at least one employee ID is required"})
		return
	}

	var employees []models.Employee
	if err := database.DB.
		Preload("Company").
		Preload("Department").
		Preload("DesignationRef").
		Preload("SectionRef").
		Preload("LineRef").
		Where("employee_id IN ? AND deleted_at IS NULL", req.EmployeeIDs).
		Order("LENGTH(employee_id) ASC, employee_id ASC").
		Find(&employees).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// --- Font setup (SutonnyMJ or embedded Noto Sans Bengali) ---
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	font := loadBanglaFont(pdf)

	// --- Layout constants (Portrait A4: 210 x 297) ---
	// 3 columns x 1 row per half -> 3 front cards on top half, 3 back cards on bottom half.
	const (
		pageW = 210.0
		pageH = 297.0
		margin = 8.0
		gapX   = 3.0
		cols   = 3
		cardH  = 95.0
	)

	cardW := (pageW - 2*margin - float64(cols-1)*gapX) / float64(cols)
	halfH := pageH / 2
	offsetY := (halfH - cardH) / 2
	cardTopY := margin + offsetY
	cardBottomY := pageH/2 + offsetY

	// Generate pages (cols employees per page: fronts top, backs bottom)
	for i := 0; i < len(employees); i += cols {
		pdf.AddPage()
		if pdf.Error() != nil {
			service.WriteErrorLog("idcard", "AddPage error: "+pdf.Error().Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF page error: " + pdf.Error().Error()})
			return
		}
		end := i + cols
		if end > len(employees) {
			end = len(employees)
		}
		for j, emp := range employees[i:end] {
			x := margin + float64(j)*(cardW+gapX)
			drawCardFront(pdf, x, cardTopY, cardW, cardH, emp, font)
			if pdf.Error() != nil {
				service.WriteErrorLog("idcard", "drawCardFront error at emp="+emp.EmployeeID+": "+pdf.Error().Error())
			}
			drawCardBack(pdf, x, cardBottomY, cardW, cardH, emp, font)
			if pdf.Error() != nil {
				service.WriteErrorLog("idcard", "drawCardBack error at emp="+emp.EmployeeID+": "+pdf.Error().Error())
			}
		}
	}

	if pdf.Error() != nil {
		service.WriteErrorLog("idcard", "final PDF error: "+pdf.Error().Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF generation failed: " + pdf.Error().Error()})
		return
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		service.WriteErrorLog("idcard", "PDF output error: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF: " + err.Error()})
		return
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	filename := "id_cards_" + strings.ReplaceAll(time.Now().Format("2006-01-02"), "-", "_") + ".pdf"
	c.JSON(http.StatusOK, gin.H{"data": encoded, "filename": filename})
}

func companyDisplayName(company models.Company) string {
	if company.CompanyNameEn != "" {
		return company.CompanyNameEn
	}
	if company.CompanyNameBn != "" {
		return company.CompanyNameBn
	}
	return "Company"
}

func drawCardFront(pdf *gofpdf.Fpdf, x, y, w, h float64, emp models.Employee, font string) {
	// Card border
	pdf.SetDrawColor(29, 78, 137)
	pdf.SetLineWidth(0.6)
	pdf.Rect(x, y, w, h, "D")

	companyName := companyDisplayName(emp.Company)

	// Header: logo placeholder + company name + subtitle
	logoSize := 9.0
	pdf.SetFillColor(29, 78, 137)
	pdf.SetDrawColor(29, 78, 137)
	pdf.SetLineWidth(0.3)
	pdf.Rect(x+2, y+2, logoSize, logoSize, "FD")
	pdf.SetFont(font, "B", 4)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(x+2, y+2+logoSize/2-1.5)
	pdf.CellFormat(logoSize, 3, "LOGO", "", 0, "C", false, 0, "")

	pdf.SetFont(font, "B", 7)
	pdf.SetTextColor(29, 78, 137)
	pdf.SetXY(x+12.5, y+1.8)
	pdf.MultiCell(w-14.5, 3.6, companyName, "", "L", false)

	pdf.SetFont(font, "B", 5.5)
	pdf.SetTextColor(120, 120, 120)
	pdf.SetXY(x, y+9)
	pdf.CellFormat(w, 3.5, "EMPLOYEE ID CARD", "", 0, "C", false, 0, "")

	pdf.SetDrawColor(29, 78, 137)
	pdf.SetLineWidth(0.5)
	pdf.Line(x+2, y+12.2, x+w-2, y+12.2)

	// Photo area 35 x 45 mm
	photoW, photoH := 35.0, 45.0
	photoX := x + (w-photoW)/2
	photoY := y + 13.5
	pdf.SetDrawColor(180, 180, 180)
	pdf.SetLineWidth(0.4)
	pdf.Rect(photoX, photoY, photoW, photoH, "D")

	hasImage := false
	if emp.ImageURL != "" {
		imgPath := resolveImagePath(emp.ImageURL)
		if _, err := os.Stat(imgPath); err == nil {
			pdf.ImageOptions(imgPath, photoX+1, photoY+1, photoW-2, 0, false, gofpdf.ImageOptions{ImageType: filepath.Ext(imgPath)}, 0, "")
			hasImage = true
		}
	}
	if !hasImage {
		pdf.SetFont(font, "", 4.5)
		pdf.SetTextColor(150, 150, 150)
		pdf.SetXY(photoX, photoY+14)
		pdf.CellFormat(photoW, 3.5, "EMPLOYEE PHOTO", "", 0, "C", false, 0, "")
		pdf.SetFont(font, "", 4)
		pdf.SetXY(photoX, photoY+18.5)
		pdf.CellFormat(photoW, 3, "35 x 45 mm", "", 0, "C", false, 0, "")
	}

	// Employee info below photo
	labelW := 18.0
	infoX := x + 3
	infoY := photoY + photoH + 1.5
	lineH := 4.6

	name := emp.NameEn
	if name == "" {
		name = emp.NameBn
	}
	name = strings.ToUpper(name)

	desig := ""
	if emp.DesignationRef != nil {
		desig = emp.DesignationRef.Name
		if desig == "" {
			desig = emp.DesignationRef.NameBn
		}
	}
	dept := ""
	if emp.Department != nil {
		dept = emp.Department.Name
		if dept == "" {
			dept = emp.Department.NameBn
		}
	}
	lineName := ""
	if emp.LineRef != nil {
		lineName = emp.LineRef.Name
		if lineName == "" {
			lineName = emp.LineRef.NameBn
		}
	}
	joinDate := ""
	if !emp.JoiningDate.IsZero() {
		joinDate = emp.JoiningDate.Format("02 January 2006")
	}

	rows := [][2]string{
		{"Employee ID", emp.EmployeeID},
		{"Employee Name", name},
		{"Joining Date", joinDate},
		{"Designation", desig},
		{"Department", dept},
		{"Line", lineName},
	}

	for i, r := range rows {
		ry := infoY + float64(i)*lineH
		pdf.SetFont(font, "B", 4.8)
		pdf.SetTextColor(110, 110, 110)
		pdf.SetXY(infoX, ry)
		pdf.CellFormat(labelW, 3.2, r[0], "", 0, "L", false, 0, "")
		pdf.SetFont(font, "B", 5.5)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetXY(infoX+labelW, ry)
		pdf.CellFormat(w-labelW-6, 3.4, truncateString(r[1], 24), "", 0, "L", false, 0, "")
	}

	// Signatures
	sigY := y + h - 7
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.3)
	pdf.Line(x+3, sigY, x+26, sigY)
	pdf.Line(x+36.5, sigY, x+w-3, sigY)
	pdf.SetFont(font, "", 4)
	pdf.SetTextColor(130, 130, 130)
	pdf.SetXY(x+3, sigY+1)
	pdf.CellFormat(23, 2.5, "Employee Signature", "", 0, "C", false, 0, "")
	pdf.SetXY(x+36.5, sigY+1)
	pdf.CellFormat(w-39.5, 2.5, "Authorized Signature", "", 0, "C", false, 0, "")
}

func drawCardBack(pdf *gofpdf.Fpdf, x, y, w, h float64, emp models.Employee, font string) {
	pdf.SetDrawColor(29, 78, 137)
	pdf.SetLineWidth(0.6)
	pdf.Rect(x, y, w, h, "D")

	companyName := companyDisplayName(emp.Company)

	// Header
	pdf.SetFont(font, "B", 7)
	pdf.SetTextColor(29, 78, 137)
	pdf.SetXY(x, y+1.8)
	pdf.CellFormat(w, 3.6, companyName, "", 0, "C", false, 0, "")
	pdf.SetFont(font, "B", 5.5)
	pdf.SetTextColor(120, 120, 120)
	pdf.SetXY(x, y+6)
	pdf.CellFormat(w, 3.4, "PEOPLEHUB HR & PAYROLL", "", 0, "C", false, 0, "")

	pdf.SetDrawColor(29, 78, 137)
	pdf.SetLineWidth(0.5)
	pdf.Line(x+2, y+9.5, x+w-2, y+9.5)

	m := 3.0
	cy := y + 11.5

	// Company address
	pdf.SetFont(font, "B", 5)
	pdf.SetTextColor(110, 110, 110)
	pdf.SetXY(x+m, cy)
	pdf.CellFormat(w-2*m, 3.2, "Company Address", "", 0, "L", false, 0, "")
	address := emp.Company.AddressEn
	if address == "" {
		address = emp.Company.AddressBn
	}
	if address == "" {
		address = "-"
	}
	pdf.SetFont(font, "", 5)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetXY(x+m, cy+3.4)
	pdf.MultiCell(w-2*m, 3.4, address, "", "L", false)
	cy = pdf.GetY() + 1.5

	// Contact lines
	if emp.Company.Phone != "" {
		pdf.SetFont(font, "", 4.8)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(w-2*m, 3.2, "Phone : "+emp.Company.Phone, "", 0, "L", false, 0, "")
		cy += 3.6
	}
	if emp.Company.Email != "" {
		pdf.SetFont(font, "", 4.8)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(w-2*m, 3.2, "Email : "+emp.Company.Email, "", 0, "L", false, 0, "")
		cy += 3.6
	}

	// Emergency contact
	cy += 1.0
	pdf.SetFont(font, "B", 5)
	pdf.SetTextColor(110, 110, 110)
	pdf.SetXY(x+m, cy)
	pdf.CellFormat(w-2*m, 3.2, "Emergency Contact", "", 0, "L", false, 0, "")
	emergency := emp.EmergencyPhone
	if emergency == "" {
		emergency = emp.EmergencyContact
	}
	if emergency == "" {
		emergency = "-"
	}
	cy += 3.4
	pdf.SetFont(font, "B", 5)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetXY(x+m, cy)
	pdf.CellFormat(w-2*m, 3.4, emergency, "", 0, "L", false, 0, "")
	cy += 4.4

	// Return note
	pdf.SetFont(font, "", 4.5)
	pdf.SetTextColor(90, 90, 90)
	pdf.SetXY(x+m, cy)
	pdf.MultiCell(w-2*m, 3.2, "If found, please return this card to the Human Resource Department.", "", "C", false)
	cy = pdf.GetY() + 1.0

	// Property note
	pdf.SetFont(font, "", 4.5)
	pdf.SetTextColor(90, 90, 90)
	pdf.SetXY(x+m, cy)
	pdf.MultiCell(w-2*m, 3.2, "This card remains the property of "+companyName+".", "", "C", false)
}

func resolveImagePath(url string) string {
	path := strings.TrimPrefix(url, "/")
	path = strings.TrimPrefix(path, "uploads/")
	return filepath.Join("uploads", path)
}

func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-2]) + ".."
}
