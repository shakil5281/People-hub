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
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

type GenerateIdCardRequest struct {
	EmployeeIDs []string `json:"employee_ids" binding:"required"`
	Lang        string   `json:"lang"`   // "en" or "bn"
	Format      string   `json:"format"` // "2x4" (default) or "sheet" (A4 6-up)
}

// GenerateIdCards godoc
//
//	@Summary      Generate ID cards PDF (2x4 inch portrait or A4 sheet)
//	@Description  Generate printable ID cards for Ekushe Fashions Ltd. in 2x4 inch portrait mode (Page 1 Front, Page 2 Back) or A4 sheet format in English or Bangla.
//	@Tags         ID Cards
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        request body GenerateIdCardRequest true "Employee IDs, format and language"
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

	isBn := strings.ToLower(req.Lang) == "bn" || strings.ToLower(req.Lang) == "bangla"

	var employees []models.Employee
	if err := database.DB.
		Preload("Company").
		Preload("Department").
		Preload("DesignationRef").
		Preload("SectionRef").
		Preload("LineRef").
		Preload("GroupRef").
		Where("employee_id IN ? AND deleted_at IS NULL", req.EmployeeIDs).
		Order("LENGTH(employee_id) ASC, employee_id ASC").
		Find(&employees).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var pdf *gofpdf.Fpdf
	var font string

	// Unified A4 Sheet Printing Mode (6 Employees / 12 Cards per A4 Sheet)
	// Front Part: 55.0 mm Wide x 86.0 mm High
	// Back Part: 50.0 mm Wide x 86.0 mm High
	// Margins & Gaps: All 2.0 mm (startX: 2.0 mm, startY: 2.0 mm, gapPairX: 2.0 mm, gapY: 2.0 mm)
	// Front & Back side-by-side with ZERO center gap (0.0 mm)
	pdf = gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	if isBn {
		font = loadBanglaFont(pdf)
	} else {
		font = "Arial"
	}

	const (
		employeesPerPage = 6
		frontW           = 54.0
		backW            = 44.0
		cardH            = 84.0
		startX           = 2.0
		startY           = 5.0
		gapPairX         = 6.0
		gapY             = 6.0
	)

	for i := 0; i < len(employees); i += employeesPerPage {
		pdf.AddPage()
		if pdf.Error() != nil {
			service.WriteErrorLog("idcard", "AddPage error: "+pdf.Error().Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF page error: " + pdf.Error().Error()})
			return
		}

		end := i + employeesPerPage
		if end > len(employees) {
			end = len(employees)
		}

		pageEmps := employees[i:end]
		for idx, emp := range pageEmps {
			row := idx / 2
			colInRow := idx % 2

			pairW := frontW + backW // 90.0 mm total pair width
			frontX := startX + float64(colInRow)*(pairW+gapPairX)
			backX := frontX + frontW // ZERO center gap between Front (52mm) and Back (38mm)!
			cardY := startY + float64(row)*(cardH+gapY)

			drawCardFront(pdf, frontX, cardY, frontW, cardH, emp, font, isBn)
			drawCardBack(pdf, backX, cardY, backW, cardH, emp, font, isBn)
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
	filename := "id_card_" + strings.ReplaceAll(time.Now().Format("2006-01-02"), "-", "_") + ".pdf"
	c.JSON(http.StatusOK, gin.H{"data": encoded, "filename": filename})
}

func companyDisplayName(company models.Company, isBn bool) string {
	if isBn {
		if company.CompanyNameBn != "" {
			return utils.UnicodeToBijoy(company.CompanyNameBn)
		}
		if company.CompanyNameEn != "" {
			return utils.UnicodeToBijoy(company.CompanyNameEn)
		}
		return utils.UnicodeToBijoy("একুশে ফ্যাশন লিঃ")
	}
	if company.CompanyNameEn != "" {
		return company.CompanyNameEn
	}
	if company.CompanyNameBn != "" {
		return company.CompanyNameBn
	}
	return "Ekushe Fashions Ltd."
}

func fitTextFontSize(pdf *gofpdf.Fpdf, font string, style string, text string, maxWidth float64, initialSize float64, minSize float64) float64 {
	size := initialSize
	for size > minSize {
		pdf.SetFont(font, style, size)
		if pdf.GetStringWidth(text) <= maxWidth {
			return size
		}
		size -= 0.2
	}
	return minSize
}

func getCategoryText(emp models.Employee, isBn bool) string {
	deptName := ""
	if emp.Department != nil {
		deptName = strings.ToLower(emp.Department.Name + " " + emp.Department.NameBn)
	}

	groupStr := ""
	if emp.GroupRef != nil {
		groupStr += " " + emp.GroupRef.Name
	}
	if emp.EmployeeType != "" {
		groupStr += " " + emp.EmployeeType
	}
	if emp.Grade != "" {
		groupStr += " " + emp.Grade
	}
	if emp.DesignationRef != nil {
		groupStr += " " + emp.DesignationRef.Name + " " + emp.DesignationRef.NameBn
	}
	groupStr = strings.ToLower(groupStr)

	isStaffExec := strings.Contains(groupStr, "staff") ||
		strings.Contains(groupStr, "executive") ||
		strings.Contains(groupStr, "officer") ||
		strings.Contains(groupStr, "manager") ||
		strings.Contains(groupStr, "স্টাফ") ||
		strings.Contains(groupStr, "এজিএম") ||
		strings.Contains(groupStr, "অফিসার") ||
		strings.Contains(groupStr, "ম্যানেজার")

	isAdmin := strings.Contains(deptName, "admin") ||
		strings.Contains(deptName, "এডমিন") ||
		strings.Contains(deptName, "hr") ||
		strings.Contains(deptName, "management")

	isProd := strings.Contains(deptName, "production") ||
		strings.Contains(deptName, "প্রোডাকশন") ||
		strings.Contains(deptName, "sewing") ||
		strings.Contains(deptName, "cutting") ||
		strings.Contains(deptName, "finishing")

	if isStaffExec {
		if isAdmin {
			if isBn {
				return utils.UnicodeToBijoy("অফিস স্টাফ")
			}
			return "Office Staff"
		}
		if isProd {
			if isBn {
				return utils.UnicodeToBijoy("প্রোডাকশন স্টাফ")
			}
			return "Production Staff"
		}
		if isBn {
			return utils.UnicodeToBijoy("অফিস স্টাফ")
		}
		return "Office Staff"
	}

	if isBn {
		return utils.UnicodeToBijoy("ওয়ার্কার")
	}
	return "Worker"
}

func drawCardFront(pdf *gofpdf.Fpdf, x, y, w, h float64, emp models.Employee, font string, isBn bool) {
	// Card Outer Border (Rounded corners)
	pdf.SetDrawColor(210, 215, 220)
	pdf.SetLineWidth(0.4)
	pdf.Rect(x, y, w, h, "D")

	companyName := companyDisplayName(emp.Company, isBn)

	// Top Header Bar (#123A63)
	pdf.SetFillColor(18, 58, 99)
	pdf.Rect(x, y, w, 13.0, "F")

	// Light Blue Accent Line (#3B82B6)
	pdf.SetDrawColor(59, 130, 182)
	pdf.SetLineWidth(0.6)
	pdf.Line(x, y+13.0, x+w, y+13.0)

	// Company Title inside Top Bar (Centered, Bold, dynamically fitted)
	fontSz := fitTextFontSize(pdf, font, "B", companyName, w-4.0, 10.5, 7.5)
	pdf.SetFont(font, "B", fontSz)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(x+2.0, y+2.5)
	pdf.CellFormat(w-4.0, 4.0, companyName, "", 0, "C", false, 0, "")

	// Employee Category Subtitle (Office Staff / Production Staff / Worker)
	categoryText := getCategoryText(emp, isBn)
	pdf.SetFont(font, "B", 6.2)
	pdf.SetTextColor(220, 235, 252)
	pdf.SetXY(x+2.0, y+7.8)
	pdf.CellFormat(w-4.0, 3.0, categoryText, "", 0, "C", false, 0, "")

	// --- Employee Photo (ROUNDED CIRCLE AVATAR: 18mm width & height, 0.3mm border width) ---
	photoCenterX := x + w/2.0
	photoCenterY := y + 25.0 // 3.0 mm gap after header bar
	photoR := 9.0            // 18mm diameter

	pdf.SetFillColor(245, 247, 250)
	pdf.SetDrawColor(18, 58, 99)
	pdf.SetLineWidth(0.3) // 0.3mm border width
	pdf.Circle(photoCenterX, photoCenterY, photoR, "FD")

	hasImage := false
	if emp.ImageURL != "" {
		imgPath := resolveImagePath(emp.ImageURL)
		if _, err := os.Stat(imgPath); err == nil {
			pdf.ClipCircle(photoCenterX, photoCenterY, photoR-0.3, true)
			pdf.ImageOptions(imgPath, photoCenterX-photoR, photoCenterY-photoR, photoR*2, photoR*2, false, gofpdf.ImageOptions{}, 0, "")
			pdf.ClipEnd()
			if pdf.Ok() {
				hasImage = true
			} else {
				pdf.SetError(nil)
			}
		}
	}

	if !hasImage {
		photoText := "PHOTO"
		if isBn {
			photoText = utils.UnicodeToBijoy("ছবি")
		}
		pdf.SetFont(font, "", 6.2)
		pdf.SetTextColor(150, 150, 150)
		pdf.SetXY(photoCenterX-photoR, photoCenterY-1.5)
		pdf.CellFormat(photoR*2, 3.0, photoText, "", 0, "C", false, 0, "")
	}

	// Outer Navy Ring Frame around Avatar (0.3mm border width)
	pdf.SetDrawColor(18, 58, 99)
	pdf.SetLineWidth(0.3)
	pdf.Circle(photoCenterX, photoCenterY, photoR, "D")

	// Employee Name & Designation
	nameY := photoCenterY + photoR + 2.5

	name := ""
	if isBn {
		if emp.NameBn != "" {
			name = utils.UnicodeToBijoy(emp.NameBn)
		} else {
			name = utils.UnicodeToBijoy(emp.NameEn)
		}
	} else {
		if emp.NameEn != "" {
			name = emp.NameEn
		} else {
			name = emp.EmployeeID
		}
		name = strings.ToUpper(name)
	}

	nameFontSz := fitTextFontSize(pdf, font, "B", name, w-4.0, 8.0, 6.2)
	pdf.SetFont(font, "B", nameFontSz)
	pdf.SetTextColor(18, 58, 99)
	pdf.SetXY(x+2.0, nameY)
	pdf.CellFormat(w-4.0, 3.2, truncateString(name, 24), "", 0, "C", false, 0, "")

	desig := ""
	if emp.DesignationRef != nil {
		if isBn {
			if emp.DesignationRef.NameBn != "" {
				desig = utils.UnicodeToBijoy(emp.DesignationRef.NameBn)
			} else {
				desig = utils.UnicodeToBijoy(emp.DesignationRef.Name)
			}
		} else {
			desig = emp.DesignationRef.Name
			if desig != "" {
				desig = strings.ToUpper(desig)
			}
		}
	}
	if desig == "" {
		if isBn {
			desig = utils.UnicodeToBijoy("কর্মকর্তা")
		} else {
			desig = "OFFICER"
		}
	}

	desigFontSz := fitTextFontSize(pdf, font, "", desig, w-4.0, 6.2, 5.2)
	pdf.SetFont(font, "", desigFontSz)
	pdf.SetTextColor(107, 114, 128)
	pdf.SetXY(x+2.0, nameY+3.4)
	pdf.CellFormat(w-4.0, 2.8, truncateString(desig, 26), "", 0, "C", false, 0, "")

	// Details List Rows
	infoY := nameY + 8.5
	lineH := 4.2

	dept := ""
	if emp.Department != nil {
		if isBn {
			if emp.Department.NameBn != "" {
				dept = utils.UnicodeToBijoy(emp.Department.NameBn)
			} else {
				dept = utils.UnicodeToBijoy(emp.Department.Name)
			}
		} else {
			dept = emp.Department.Name
		}
	}
	if dept == "" {
		if isBn {
			dept = utils.UnicodeToBijoy("এডমিন")
		} else {
			dept = "Admin"
		}
	}

	joinDate := ""
	if !emp.JoiningDate.IsZero() {
		if isBn {
			joinDate = emp.JoiningDate.Format("02/01/2006")
		} else {
			joinDate = emp.JoiningDate.Format("02 Jan 2006")
		}
	}
	if joinDate == "" {
		if isBn {
			joinDate = "০১/০৩/২০২০"
		} else {
			joinDate = "01 Mar 2020"
		}
	}

	blood := emp.BloodGroup
	if blood == "" {
		blood = "B+"
	} else if isBn {
		blood = utils.UnicodeToBijoy(blood)
	}

	lineNo := ""
	if emp.LineRef != nil {
		if isBn && emp.LineRef.NameBn != "" {
			lineNo = utils.UnicodeToBijoy(emp.LineRef.NameBn)
		} else {
			lineNo = emp.LineRef.Name
		}
	}
	if lineNo == "" {
		if isBn {
			lineNo = utils.UnicodeToBijoy("এডমিন")
		} else {
			lineNo = "Admin"
		}
	}

	var rows [][2]string
	if isBn {
		rows = [][2]string{
			{utils.UnicodeToBijoy("আইডি"), emp.EmployeeID},
			{utils.UnicodeToBijoy("বিভাগ"), dept},
			{utils.UnicodeToBijoy("লাইন"), lineNo},
			{utils.UnicodeToBijoy("যোগদান"), joinDate},
			{utils.UnicodeToBijoy("রক্তের গ্রুপ"), blood},
		}
	} else {
		rows = [][2]string{
			{"ID", emp.EmployeeID},
			{"Department", dept},
			{"Line No", lineNo},
			{"Joining", joinDate},
			{"Blood Group", blood},
		}
	}

	for i, r := range rows {
		ry := infoY + float64(i)*lineH

		// Label
		pdf.SetFont(font, "", 6.0)
		pdf.SetTextColor(107, 114, 128)
		pdf.SetXY(x+3.0, ry)
		pdf.CellFormat(16.0, 3.0, r[0], "", 0, "L", false, 0, "")

		// Colon
		pdf.SetXY(x+22.0, ry)
		pdf.CellFormat(2.0, 3.0, ":", "", 0, "C", false, 0, "")

		// Value
		pdf.SetFont(font, "B", 6.5)
		pdf.SetTextColor(38, 50, 56)
		pdf.SetXY(x+24.5, ry)
		pdf.CellFormat(w-26.5, 3.0, truncateString(r[1], 20), "", 0, "L", false, 0, "")

		// Divider line
		pdf.SetDrawColor(235, 238, 242)
		pdf.SetLineWidth(0.2)
		pdf.Line(x+0.075*w, ry+3.8, x+0.925*w, ry+3.8)
	}

	// Front Signatures
	sigY := y + h - 8.0
	lineWidth := 18.0
	leftX := x + 3.0
	rightX := x + w - lineWidth - 3.0

	// Draw Signature Images above lines if available
	if emp.SignatureURL != "" {
		sigPath := resolveImagePath(emp.SignatureURL)
		if _, err := os.Stat(sigPath); err == nil {
			pdf.ImageOptions(sigPath, rightX+1.0, sigY-5.5, lineWidth-2.0, 5.0, false, gofpdf.ImageOptions{ReadDpi: true}, 0, "")
			if !pdf.Ok() {
				pdf.SetError(nil)
			}
		}
	}
	if emp.Company.Signature != "" {
		authPath := resolveImagePath(emp.Company.Signature)
		if _, err := os.Stat(authPath); err == nil {
			pdf.ImageOptions(authPath, leftX+1.0, sigY-5.5, lineWidth-2.0, 5.0, false, gofpdf.ImageOptions{ReadDpi: true}, 0, "")
			if !pdf.Ok() {
				pdf.SetError(nil)
			}
		}
	}

	pdf.SetDrawColor(120, 120, 120)
	pdf.SetLineWidth(0.3)
	pdf.Line(leftX, sigY, leftX+lineWidth, sigY)
	pdf.Line(rightX, sigY, rightX+lineWidth, sigY)

	pdf.SetFont(font, "", 5.6)
	pdf.SetTextColor(107, 114, 128)

	authText := "Authorisation"
	sigText := "Holder Signature"
	if isBn {
		authText = utils.UnicodeToBijoy("অনুমোদন")
		sigText = utils.UnicodeToBijoy("কার্ডধারীর স্বাক্ষর")
	}

	pdf.SetXY(leftX, sigY+0.6)
	pdf.CellFormat(lineWidth, 2.2, authText, "", 0, "C", false, 0, "")
	pdf.SetXY(rightX, sigY+0.6)
	pdf.CellFormat(lineWidth, 2.2, sigText, "", 0, "C", false, 0, "")

	// Bottom Curved Dark Navy Wave Banner (#123A63)
	pdf.SetFillColor(18, 58, 99)
	pdf.Polygon([]gofpdf.PointType{
		{X: x, Y: y + h - 3.5},
		{X: x + w/2, Y: y + h - 5.5},
		{X: x + w, Y: y + h - 3.5},
		{X: x + w, Y: y + h},
		{X: x, Y: y + h},
	}, "F")

	// Light Blue Accent Line (#3B82B6)
	pdf.SetDrawColor(59, 130, 182)
	pdf.SetLineWidth(0.4)
	pdf.Line(x, y+h-3.5, x+w/2, y+h-5.5)
	pdf.Line(x+w/2, y+h-5.5, x+w, y+h-3.5)

	// Bottom Footer Brand Text (Full Company Name)
	footerText := companyDisplayName(emp.Company, isBn)
	pdf.SetFont(font, "B", 5.8)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(x, y+h-3.2)
	pdf.CellFormat(w, 3.0, footerText, "", 0, "C", false, 0, "")
}

func convertENToBNDigits(input string) string {
	r := strings.NewReplacer(
		"0", "০", "1", "১", "2", "২", "3", "৩", "4", "৪",
		"5", "৫", "6", "৬", "7", "৭", "8", "৮", "9", "৯",
	)
	return r.Replace(input)
}

func drawCardBack(pdf *gofpdf.Fpdf, x, y, w, h float64, emp models.Employee, font string, isBn bool) {
	// Card Outer Border
	pdf.SetDrawColor(200, 205, 210)
	pdf.SetLineWidth(0.3)
	pdf.Rect(x, y, w, h, "D")

	m := 2.0 // Fixed 2.0 mm left & right margin inside backpart card
	textW := w - 2*m
	cy := y + 10.0 // Top header gap adjusted for +1pt font size

	if isBn {
		// 1. Top validity note & Company Header (Centered, +1pt font size)
		pdf.SetFont(font, "", 6.0)
		pdf.SetTextColor(28, 28, 28)
		pdf.SetXY(x+m, cy)
		validityNote := utils.UnicodeToBijoy("এই কার্ডের মেয়াদ অব্যাহতি/অবসরপর্যন্ত")
		pdf.CellFormat(textW, 3.0, validityNote, "", 0, "C", false, 0, "")
		cy += 3.8

		companyName := companyDisplayName(emp.Company, isBn)
		pdf.SetFont(font, "B", 8.0)
		pdf.SetTextColor(28, 28, 28)
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(textW, 3.6, companyName, "", 0, "C", false, 0, "")
		cy += 6.5 // Gap after CompanyName

		// 2. Factory Address & Phone (Centered, +1pt font size -> 6.0pt)
		pdf.SetFont(font, "", 6.0)
		pdf.SetTextColor(28, 28, 28)
		factoryAddr := utils.UnicodeToBijoy("ফ্যাক্টরির ঠিকানাঃ মাষ্টারবাড়ী, গাজীপুর সদর, গাজীপুর।")
		pdf.SetXY(x+m, cy)
		pdf.MultiCell(textW, 3.0, factoryAddr, "", "C", false)
		cy = pdf.GetY() + 0.5

		phoneStr := emp.Company.Phone
		if phoneStr == "" {
			phoneStr = "01844001141"
		}
		factoryPhone := utils.UnicodeToBijoy("ফোনঃ- ") + utils.UnicodeToBijoy(convertENToBNDigits(phoneStr))
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(textW, 3.0, factoryPhone, "", 0, "C", false, 0, "")
		cy += 5.5 // Clean spacing gap

		// 3. Employee Personal Info (Left aligned, +1pt font size -> 6.0pt)
		permAddr := emp.PermanentAddress
		if permAddr == "" {
			permAddr = "ভাওয়াল মির্জাপুর,গাজীপুর সদর গাজীপুর।"
		}
		permText := utils.UnicodeToBijoy("স্থায়ী ঠিকানাঃ " + permAddr)
		pdf.SetXY(x+m, cy)
		pdf.MultiCell(textW, 3.0, permText, "", "L", false)
		cy = pdf.GetY() + 1.2

		blood := emp.BloodGroup
		if blood == "" {
			blood = "B+"
		}
		bloodText := utils.UnicodeToBijoy("রক্তের গ্রুপঃ  ") + blood
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(textW, 3.0, bloodText, "", 0, "L", false, 0, "")
		cy += 3.6

		emPhone := emp.EmergencyPhone
		if emPhone == "" {
			emPhone = emp.Phone
		}
		if emPhone == "" {
			emPhone = "01632706439"
		}
		emText := utils.UnicodeToBijoy("ফোন নংঃ ") + utils.UnicodeToBijoy(convertENToBNDigits(emPhone))
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(textW, 3.0, emText, "", 0, "L", false, 0, "")
		cy += 3.6

		nid := emp.NID
		if nid == "" {
			nid = "19983313067000302"
		}
		nidLabel := utils.UnicodeToBijoy("জাতীয় পরিচয় পত্র নংঃ")
		nidVal := utils.UnicodeToBijoy(convertENToBNDigits(nid))
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(textW, 3.0, nidLabel, "", 0, "L", false, 0, "")
		cy += 3.2
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(textW, 3.0, nidVal, "", 0, "L", false, 0, "")
		cy += 5.5 // Clean spacing gap

		// 4. Bottom Instruction Loss Note (Centered, +1pt font size -> 6.0pt)
		pdf.SetFont(font, "", 6.0)
		pdf.SetTextColor(28, 28, 28)
		lossLine1 := utils.UnicodeToBijoy("উক্ত পরিচয়পত্র হারাইয়া গেলে তাৎক্ষনিক")
		lossLine2 := utils.UnicodeToBijoy("ব্যবস্থাপনা কর্তৃপক্ষকে জানাইতে হইবে।")
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(textW, 3.0, lossLine1, "", 0, "C", false, 0, "")
		cy += 3.2
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(textW, 3.0, lossLine2, "", 0, "C", false, 0, "")

	} else {
		// English Mode (+1pt font sizes)
		pdf.SetFont(font, "", 6.0)
		pdf.SetTextColor(28, 28, 28)
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(textW, 3.0, "This card is valid until resignation/retirement", "", 0, "C", false, 0, "")
		cy += 3.8

		companyName := companyDisplayName(emp.Company, false)
		pdf.SetFont(font, "B", 8.0)
		pdf.SetTextColor(28, 28, 28)
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(textW, 3.6, companyName, "", 0, "C", false, 0, "")
		cy += 6.5

		pdf.SetFont(font, "", 6.0)
		pdf.SetTextColor(28, 28, 28)
		factoryAddr := "Factory Address: Masterbari, Gazipur Sadar, Gazipur."
		pdf.SetXY(x+m, cy)
		pdf.MultiCell(textW, 3.0, factoryAddr, "", "C", false)
		cy = pdf.GetY() + 0.5

		phoneStr := emp.Company.Phone
		if phoneStr == "" {
			phoneStr = "01844001141"
		}
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(textW, 3.0, "Phone: "+phoneStr, "", 0, "C", false, 0, "")
		cy += 5.5

		permAddr := emp.PermanentAddress
		if permAddr == "" {
			permAddr = "Bhawal Mirzapur, Gazipur Sadar, Gazipur."
		}
		pdf.SetXY(x+m, cy)
		pdf.MultiCell(textW, 3.0, "Permanent Address: "+permAddr, "", "L", false)
		cy = pdf.GetY() + 1.2

		blood := emp.BloodGroup
		if blood == "" {
			blood = "B+"
		}
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(textW, 3.0, "Blood Group: "+blood, "", 0, "L", false, 0, "")
		cy += 3.6

		emPhone := emp.EmergencyPhone
		if emPhone == "" {
			emPhone = emp.Phone
		}
		if emPhone == "" {
			emPhone = "01632706439"
		}
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(textW, 3.0, "Phone No: "+emPhone, "", 0, "L", false, 0, "")
		cy += 3.6

		nid := emp.NID
		if nid == "" {
			nid = "19983313067000302"
		}
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(textW, 3.0, "National ID No:", "", 0, "L", false, 0, "")
		cy += 3.2
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(textW, 3.0, nid, "", 0, "L", false, 0, "")
		cy += 5.5

		pdf.SetXY(x+m, cy)
		pdf.CellFormat(textW, 3.0, "If this ID card is lost, report immediately", "", 0, "C", false, 0, "")
		cy += 3.2
		pdf.SetXY(x+m, cy)
		pdf.CellFormat(textW, 3.0, "to management authority.", "", 0, "C", false, 0, "")
	}

	// Bottom Curved Dark Navy Wave Banner (#123A63)
	pdf.SetFillColor(18, 58, 99)
	pdf.Polygon([]gofpdf.PointType{
		{X: x, Y: y + h - 3.5},
		{X: x + w/2, Y: y + h - 5.5},
		{X: x + w, Y: y + h - 3.5},
		{X: x + w, Y: y + h},
		{X: x, Y: y + h},
	}, "F")

	// Light Blue Accent Line (#3B82B6)
	pdf.SetDrawColor(59, 130, 182)
	pdf.SetLineWidth(0.4)
	pdf.Line(x, y+h-3.5, x+w/2, y+h-5.5)
	pdf.Line(x+w/2, y+h-5.5, x+w, y+h-3.5)

	// Bottom Footer Brand Text (Full Company Name)
	footerTextBack := companyDisplayName(emp.Company, isBn)
	pdf.SetFont(font, "B", 5.8)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(x, y+h-3.2)
	pdf.CellFormat(w, 3.0, footerTextBack, "", 0, "C", false, 0, "")
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
