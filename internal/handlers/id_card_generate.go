package handlers

import (
	"bytes"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
)

type GenerateIdCardRequest struct {
	EmployeeIDs []string `json:"employee_ids" binding:"required"`
}

// findWorkingFont tries candidate fonts in a throwaway PDF to avoid polluting the real PDF on failure.
func findWorkingFont() (fontName string, fontPath string, boldPath string) {
	candidates := []struct{ name, path, bold string }{
		{"Bangla", "C:\\Windows\\Fonts\\SutonnyMJ.ttf", "C:\\Windows\\Fonts\\SutonnyMJ-Bold.ttf"},
		{"Bangla", "C:\\Windows\\Fonts\\Nirmala.ttf", "C:\\Windows\\Fonts\\Nirmala.ttf"},
	}
	for _, c := range candidates {
		if _, err := os.Stat(c.path); err != nil {
			log.Printf("[IDCard] font not found: %s", c.path)
			continue
		}
		log.Printf("[IDCard] testing font: %s", c.path)
		testPdf := gofpdf.New("P", "mm", "A4", "")
		testPdf.AddUTF8Font(c.name, "", c.path)
		if c.bold != "" {
			testPdf.AddUTF8Font(c.name, "B", c.bold)
		}
		if testPdf.Error() == nil {
			log.Printf("[IDCard] font OK: %s", c.path)
			return c.name, c.path, c.bold
		}
		log.Printf("[IDCard] font failed: %s error=%v", c.path, testPdf.Error())
	}
	log.Printf("[IDCard] falling back to Helvetica")
	return "Helvetica", "", ""
}

// GenerateIdCards godoc
//
//	@Summary      Generate ID cards PDF
//	@Description  Generate printable ID cards (front + back on same page, 6 per A4 portrait page). Uses SutonnyMJ or Nirmala for Bangla.
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
		Where("employee_id IN ? AND deleted_at IS NULL", req.EmployeeIDs).
		Order("employee_id ASC").
		Find(&employees).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// --- Font setup ---
	fontName, fontPath, boldPath := findWorkingFont()
	log.Printf("[IDCard] selected font=%s path=%s bold=%s", fontName, fontPath, boldPath)
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)

	if fontPath != "" {
		pdf.AddUTF8Font(fontName, "", fontPath)
		if pdf.Error() != nil {
			log.Printf("[IDCard] AddUTF8Font normal failed: %v — falling back to Helvetica", pdf.Error())
			fontName = "Helvetica"
			pdf = gofpdf.New("P", "mm", "A4", "")
			pdf.SetMargins(0, 0, 0)
			pdf.SetAutoPageBreak(false, 0)
		} else if boldPath != "" {
			pdf.AddUTF8Font(fontName, "B", boldPath)
			if pdf.Error() != nil {
				log.Printf("[IDCard] AddUTF8Font bold failed: %v", pdf.Error())
			}
		}
	}

	// --- Layout constants (Portrait A4: 210 x 297) ---
	// 2 columns × 3 rows per half → 6 cards per page (front top, back bottom)
	const (
		pageW        = 210.0
		pageH        = 297.0
		margin       = 8.0
		gapX         = 3.0
		gapY         = 4.0
		halfGap      = 10.0
		cols         = 2
		rowsPerHalf  = 3
		cardsPerPage = 6
	)

	cardW := (pageW - 2*margin - float64(cols-1)*gapX) / float64(cols)                           // ~62mm
	cardH := (pageH/2 - margin - halfGap/2 - float64(rowsPerHalf-1)*gapY) / float64(rowsPerHalf) // ~64mm

	// Build positions for top half (front) and bottom half (back)
	var frontPositions, backPositions [][2]float64
	for row := 0; row < rowsPerHalf; row++ {
		for col := 0; col < cols; col++ {
			x := margin + float64(col)*(cardW+gapX)
			yFront := margin + float64(row)*(cardH+gapY)
			yBack := pageH/2 + halfGap/2 + float64(row)*(cardH+gapY)
			frontPositions = append(frontPositions, [2]float64{x, yFront})
			backPositions = append(backPositions, [2]float64{x, yBack})
		}
	}

	// Generate pages
	for i := 0; i < len(employees); i += cardsPerPage {
		pdf.AddPage()
		if pdf.Error() != nil {
			log.Printf("[IDCard] AddPage error: %v", pdf.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF page error: " + pdf.Error().Error()})
			return
		}
		end := i + cardsPerPage
		if end > len(employees) {
			end = len(employees)
		}
		for j, emp := range employees[i:end] {
			fPos := frontPositions[j]
			bPos := backPositions[j]
			drawCardFront(pdf, fPos[0], fPos[1], cardW, cardH, emp, fontName)
			if pdf.Error() != nil {
				log.Printf("[IDCard] drawCardFront error at emp=%s: %v", emp.EmployeeID, pdf.Error())
			}
			drawCardBack(pdf, bPos[0], bPos[1], cardW, cardH, emp, fontName)
			if pdf.Error() != nil {
				log.Printf("[IDCard] drawCardBack error at emp=%s: %v", emp.EmployeeID, pdf.Error())
			}
		}
	}

	if pdf.Error() != nil {
		log.Printf("[IDCard] final PDF error: %v", pdf.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF generation failed: " + pdf.Error().Error()})
		return
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		log.Printf("[IDCard] PDF output error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF: " + err.Error()})
		return
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	filename := "id_cards_" + strings.ReplaceAll(time.Now().Format("2006-01-02"), "-", "_") + ".pdf"
	c.JSON(http.StatusOK, gin.H{"data": encoded, "filename": filename})
}

func drawCardFront(pdf *gofpdf.Fpdf, x, y, w, h float64, emp models.Employee, font string) {
	// Card border
	pdf.SetDrawColor(180, 180, 180)
	pdf.SetLineWidth(0.3)
	pdf.Rect(x, y, w, h, "D")

	// Header bar
	headerH := 7.0
	pdf.SetFillColor(29, 78, 137)
	pdf.Rect(x+0.5, y+0.5, w-1, headerH, "F")

	companyName := ""
	if emp.Company.CompanyNameBn != "" {
		companyName = emp.Company.CompanyNameBn
	} else if emp.Company.CompanyNameEn != "" {
		companyName = emp.Company.CompanyNameEn
	} else {
		companyName = "Company"
	}
	pdf.SetFont(font, "B", 6.5)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(x, y+0.8)
	pdf.CellFormat(w, 3, companyName, "", 0, "C", false, 0, "")

	pdf.SetFont(font, "", 4.5)
	pdf.SetXY(x, y+4)
	pdf.CellFormat(w, 3, "কর্মচারী পরিচয় পত্র", "", 0, "C", false, 0, "")

	// Photo area
	photoX := x + 2.5
	photoY := y + headerH + 2
	photoW := 18.0
	photoH := 24.0

	pdf.SetDrawColor(200, 200, 200)
	pdf.SetLineWidth(0.3)
	pdf.Rect(photoX, photoY, photoW, photoH, "D")

	if emp.ImageURL != "" {
		imgPath := resolveImagePath(emp.ImageURL)
		if _, err := os.Stat(imgPath); err == nil {
			pdf.ImageOptions(imgPath, photoX+1, photoY+1, photoW-2, 0, false, gofpdf.ImageOptions{ImageType: filepath.Ext(imgPath)}, 0, "")
		}
	}

	// Employee info right of photo
	textX := photoX + photoW + 2.5
	textY := photoY

	pdf.SetTextColor(0, 0, 0)

	name := emp.NameBn
	if name == "" {
		name = emp.NameEn
	}
	pdf.SetFont(font, "B", 7)
	pdf.SetXY(textX, textY)
	pdf.CellFormat(w-textX+x-1, 4, name, "", 0, "L", false, 0, "")

	designation := ""
	if emp.DesignationRef != nil {
		designation = emp.DesignationRef.NameBn
		if designation == "" {
			designation = emp.DesignationRef.Name
		}
	}
	pdf.SetFont(font, "", 5)
	pdf.SetTextColor(80, 80, 80)
	pdf.SetXY(textX, textY+4.2)
	pdf.CellFormat(w-textX+x-1, 3.5, designation, "", 0, "L", false, 0, "")

	dept := ""
	if emp.Department != nil {
		dept = emp.Department.NameBn
		if dept == "" {
			dept = emp.Department.Name
		}
	}
	pdf.SetXY(textX, textY+7.8)
	pdf.CellFormat(w-textX+x-1, 3.5, dept, "", 0, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)

	// Info grid below photo
	infoY := photoY + photoH + 2
	labelW := 18.0
	valW := 22.0

	drawInfo := func(label, value string, rowY float64, colX float64) {
		pdf.SetFont(font, "B", 4.5)
		pdf.SetTextColor(100, 100, 100)
		pdf.SetXY(colX, rowY)
		pdf.CellFormat(labelW, 3, label, "", 0, "L", false, 0, "")
		pdf.SetFont(font, "", 5)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetXY(colX+labelW, rowY)
		pdf.CellFormat(valW, 3, value, "", 0, "L", false, 0, "")
	}

	drawInfo("কর্মচারী আইডি", emp.EmployeeID, infoY, x+2.5)
	drawInfo("রক্তের গ্রুপ", emp.BloodGroup, infoY+3.3, x+2.5)
	drawInfo("ফোন", emp.Phone, infoY+6.6, x+2.5)
	joinDate := ""
	if !emp.JoiningDate.IsZero() {
		joinDate = emp.JoiningDate.Format("02-01-2006")
	}
	drawInfo("যোগদান", joinDate, infoY+9.9, x+2.5)

	section := ""
	if emp.SectionRef != nil {
		section = emp.SectionRef.NameBn
		if section == "" {
			section = emp.SectionRef.Name
		}
	}
	rightColX := x + w/2 + 1
	drawInfo("সেকশন", section, infoY, rightColX)
	drawInfo("গ্রেড", emp.Grade, infoY+3.3, rightColX)
	drawInfo("লিঙ্গ", emp.Gender, infoY+6.6, rightColX)
	drawInfo("জাতীয় পরিচয়", emp.NID, infoY+9.9, rightColX)

	// Footer
	footerY := y + h - 3.5
	pdf.SetDrawColor(29, 78, 137)
	pdf.SetLineWidth(0.4)
	pdf.Line(x+2, footerY, x+w-2, footerY)
	pdf.SetFont(font, "", 4)
	pdf.SetTextColor(150, 150, 150)
	pdf.SetXY(x, footerY+0.8)
	pdf.CellFormat(w, 2.5, "এই কার্ডটি কোম্পানির সম্পত্তি। পাওয়া গেলে ফেরত দিন।", "", 0, "C", false, 0, "")
}

func drawCardBack(pdf *gofpdf.Fpdf, x, y, w, h float64, emp models.Employee, font string) {
	pdf.SetDrawColor(180, 180, 180)
	pdf.SetLineWidth(0.3)
	pdf.Rect(x, y, w, h, "D")

	// Header bar
	headerH := 7.0
	pdf.SetFillColor(29, 78, 137)
	pdf.Rect(x+0.5, y+0.5, w-1, headerH, "F")

	companyName := ""
	if emp.Company.CompanyNameBn != "" {
		companyName = emp.Company.CompanyNameBn
	} else if emp.Company.CompanyNameEn != "" {
		companyName = emp.Company.CompanyNameEn
	} else {
		companyName = "Company"
	}
	pdf.SetFont(font, "B", 6.5)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(x, y+0.8)
	pdf.CellFormat(w, 3, companyName, "", 0, "C", false, 0, "")

	pdf.SetFont(font, "", 4.5)
	pdf.SetXY(x, y+4)
	pdf.CellFormat(w, 3, "কর্মচারী পরিচয় পত্র — পেছনের অংশ", "", 0, "C", false, 0, "")

	contentY := y + headerH + 2.5
	m := 3.0

	pdf.SetTextColor(0, 0, 0)

	// Company address
	pdf.SetFont(font, "B", 5)
	pdf.SetXY(x+m, contentY)
	pdf.CellFormat(w-2*m, 3.5, "কোম্পানির ঠিকানা:", "", 0, "L", false, 0, "")
	pdf.SetFont(font, "", 4.5)
	address := emp.Company.AddressBn
	if address == "" {
		address = emp.Company.AddressEn
	}
	pdf.SetXY(x+m, contentY+3.5)
	pdf.MultiCell(w-2*m, 3, address, "", "L", false)

	// Contact
	pdf.SetFont(font, "B", 5)
	pdf.SetXY(x+m, contentY+10)
	pdf.CellFormat(w-2*m, 3.5, "ফোন / ইমেইল:", "", 0, "L", false, 0, "")
	pdf.SetFont(font, "", 4.5)
	contact := emp.Company.Phone
	if emp.Company.Email != "" {
		contact += " | " + emp.Company.Email
	}
	pdf.SetXY(x+m, contentY+13.5)
	pdf.CellFormat(w-2*m, 3, contact, "", 0, "L", false, 0, "")

	// Employee details section
	detailsY := contentY + 19
	pdf.SetFont(font, "B", 5)
	pdf.SetXY(x+m, detailsY)
	pdf.CellFormat(w-2*m, 3.5, "কর্মচারীর তথ্য:", "", 0, "L", false, 0, "")

	backInfo := func(label, value string, rowY float64, colX float64) {
		pdf.SetFont(font, "B", 4.5)
		pdf.SetTextColor(100, 100, 100)
		pdf.SetXY(colX, rowY)
		pdf.CellFormat(16, 3, label, "", 0, "L", false, 0, "")
		pdf.SetFont(font, "", 4.5)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetXY(colX+16, rowY)
		pdf.CellFormat(w/2-18, 3, value, "", 0, "L", false, 0, "")
	}

	name := emp.NameBn
	if name == "" {
		name = emp.NameEn
	}
	backInfo("নাম", name, detailsY+3.5, x+m)

	desig := ""
	if emp.DesignationRef != nil {
		desig = emp.DesignationRef.NameBn
		if desig == "" {
			desig = emp.DesignationRef.Name
		}
	}
	backInfo("পদবি", desig, detailsY+7, x+m)

	addr := emp.PresentAddress
	if addr == "" {
		addr = emp.PermanentAddress
	}
	backInfo("ঠিকানা", truncateString(addr, 32), detailsY+10.5, x+m)

	rightX := x + w/2 + 1
	backInfo("রক্তের গ্রুপ", emp.BloodGroup, detailsY+3.5, rightX)
	backInfo("জরুরী যোগাযোগ", emp.EmergencyContact, detailsY+7, rightX)
	backInfo("জরুরী ফোন", emp.EmergencyPhone, detailsY+10.5, rightX)

	// Signature
	sigY := y + h - 8
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.2)
	pdf.Line(x+20, sigY, x+w-4, sigY)
	pdf.SetFont(font, "", 4.5)
	pdf.SetTextColor(130, 130, 130)
	pdf.SetXY(x+20, sigY+1)
	pdf.CellFormat(w-24, 2.5, "অনুমোদিত স্বাক্ষর", "", 0, "C", false, 0, "")

	// Footer
	pdf.SetFont(font, "", 4)
	pdf.SetXY(x, y+h-3.5)
	pdf.CellFormat(w, 2.5, "এই কার্ডটি পাওয়া গেলে উপরের ঠিকানায় ফেরত দিন।", "", 0, "C", false, 0, "")
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
