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
	is2x4Format := strings.ToLower(req.Format) == "2x4"

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

	var pdf *gofpdf.Fpdf
	var font string

	if is2x4Format {
		// Single 2x4 Inch Portrait Card Layout (Width: 50.8 mm, Height: 101.6 mm)
		// Page 1 = Front side, Page 2 = Back side
		const (
			cardW = 50.8
			cardH = 101.6
		)

		pdf = gofpdf.NewCustom(&gofpdf.InitType{
			OrientationStr: "P",
			UnitStr:        "mm",
			Size: gofpdf.SizeType{
				Wd: cardW,
				Ht: cardH,
			},
		})
		pdf.SetMargins(0, 0, 0)
		pdf.SetAutoPageBreak(false, 0)
		if isBn {
			font = loadBanglaFont(pdf)
		} else {
			font = "Arial"
		}

		for _, emp := range employees {
			// Page 1: Front
			pdf.AddPage()
			drawCardFront(pdf, 0, 0, cardW, cardH, emp, font, isBn)

			// Page 2: Back
			pdf.AddPage()
			drawCardBack(pdf, 0, 0, cardW, cardH, emp, font, isBn)
		}
	} else {
		// Default: A4 Sheet Layout (6 employees per A4 page: Front + Back side-by-side)
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
			cardW            = 46.0
			cardH            = 86.0
			startX           = 6.0
			startY           = 9.0
			gapX             = 4.0
			gapY             = 6.5
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

				frontCol := colInRow * 2
				backCol := frontCol + 1

				frontX := startX + float64(frontCol)*(cardW+gapX)
				backX := startX + float64(backCol)*(cardW+gapX)
				cardY := startY + float64(row)*(cardH+gapY)

				drawCardFront(pdf, frontX, cardY, cardW, cardH, emp, font, isBn)
				drawCardBack(pdf, backX, cardY, cardW, cardH, emp, font, isBn)
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
		return utils.UnicodeToBijoy("একুশে ফ্যাশনস লিঃ")
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

func drawCardFront(pdf *gofpdf.Fpdf, x, y, w, h float64, emp models.Employee, font string, isBn bool) {
	// Card Outer Border (Rounded corners)
	pdf.SetDrawColor(210, 215, 220)
	pdf.SetLineWidth(0.4)
	pdf.Rect(x, y, w, h, "D")

	companyName := companyDisplayName(emp.Company, isBn)

	// Top Punch Slot Indicator (Pill shape)
	pdf.SetDrawColor(180, 185, 190)
	pdf.SetLineWidth(0.3)
	pdf.SetFillColor(245, 247, 250)
	slotW, slotH := 12.0, 2.5
	pdf.Rect(x+(w-slotW)/2, y+1.2, slotW, slotH, "FD")

	// Header: Circular Logo Badge + Company Name on ONE LINE (LARGE FONT)
	logoR := 3.8
	logoCenterX := x + 6.0
	logoCenterY := y + 8.2

	pdf.SetFillColor(18, 58, 99) // Primary Navy #123A63
	pdf.SetDrawColor(18, 58, 99)
	pdf.SetLineWidth(0.3)
	pdf.Circle(logoCenterX, logoCenterY, logoR, "FD")

	pdf.SetFont("Arial", "B", 3.2)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(logoCenterX-logoR, logoCenterY-1.5)
	pdf.CellFormat(logoR*2, 3, "LOGO", "", 0, "C", false, 0, "")

	// Fit Company Name — LARGE FONT (starting at 9.5pt)
	headerFontSz := fitTextFontSize(pdf, font, "B", companyName, w-12.5, 9.5, 6.5)
	pdf.SetFont(font, "B", headerFontSz)
	pdf.SetTextColor(18, 58, 99)
	pdf.SetXY(x+11.0, y+5.8)
	pdf.CellFormat(w-12.0, 5.0, companyName, "", 0, "L", false, 0, "")

	// Tagline
	tagline := "Quality • Commitment • Excellence"
	if isBn {
		tagline = utils.UnicodeToBijoy("গুণগত মান • প্রতিশ্রুতি • শ্রেষ্ঠত্ব")
	}
	pdf.SetFont(font, "", 3.6)
	pdf.SetTextColor(107, 114, 128)
	pdf.SetXY(x, y+13.5)
	pdf.CellFormat(w, 2.8, tagline, "", 0, "C", false, 0, "")

	pdf.SetDrawColor(210, 215, 220)
	pdf.SetLineWidth(0.3)
	pdf.Line(x+3.0, y+16.5, x+w-3.0, y+16.5)

	// --- Employee Photo (ROUNDED CIRCLE AVATAR) ---
	photoCenterX := x + w/2
	photoCenterY := y + 31.5
	photoR := 13.0

	pdf.SetFillColor(245, 247, 250)
	pdf.SetDrawColor(18, 58, 99)
	pdf.SetLineWidth(0.6)
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
		pdf.SetFont(font, "", 4.2)
		pdf.SetTextColor(150, 150, 150)
		pdf.SetXY(photoCenterX-photoR, photoCenterY-1.5)
		pdf.CellFormat(photoR*2, 3.0, photoText, "", 0, "C", false, 0, "")
	}

	// Outer Navy Ring Frame around Avatar
	pdf.SetDrawColor(18, 58, 99)
	pdf.SetLineWidth(0.6)
	pdf.Circle(photoCenterX, photoCenterY, photoR, "D")

	// Employee Name & Designation (below circular photo)
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

	nameFontSz := fitTextFontSize(pdf, font, "B", name, w-4.0, 6.0, 4.2)
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

	desigFontSz := fitTextFontSize(pdf, font, "", desig, w-4.0, 4.2, 3.4)
	pdf.SetFont(font, "", desigFontSz)
	pdf.SetTextColor(107, 114, 128)
	pdf.SetXY(x+2.0, nameY+3.4)
	pdf.CellFormat(w-4.0, 2.8, truncateString(desig, 26), "", 0, "C", false, 0, "")

	// Underline accent bar
	pdf.SetDrawColor(59, 130, 182)
	pdf.SetLineWidth(0.6)
	pdf.Line(x+w/2-7.0, nameY+7.2, x+w/2+7.0, nameY+7.2)

	// Details List Rows (Rounded Circular Icon Badge + Label : Value)
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
			dept = utils.UnicodeToBijoy("প্রোডাকশন")
		} else {
			dept = "Production"
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

	blood := emp.BloodGroup
	if blood == "" {
		blood = "-"
	} else if isBn {
		blood = utils.UnicodeToBijoy(blood)
	}

	var rows [][2]string
	if isBn {
		rows = [][2]string{
			{utils.UnicodeToBijoy("কর্মচারী আইডি"), emp.EmployeeID},
			{utils.UnicodeToBijoy("বিভাগ"), dept},
			{utils.UnicodeToBijoy("যোগদানের তারিখ"), joinDate},
			{utils.UnicodeToBijoy("রক্তের গ্রুপ"), blood},
		}
	} else {
		rows = [][2]string{
			{"Employee ID", emp.EmployeeID},
			{"Department", dept},
			{"Joining Date", joinDate},
			{"Blood Group", blood},
		}
	}

	for i, r := range rows {
		ry := infoY + float64(i)*lineH

		// Circular Icon Badge (No sharp corners)
		pdf.SetFillColor(18, 58, 99)
		pdf.Circle(x+3.6, ry+1.8, 1.6, "F")

		// Label
		pdf.SetFont(font, "", 4.0)
		pdf.SetTextColor(107, 114, 128)
		pdf.SetXY(x+6.0, ry)
		pdf.CellFormat(14.0, 3.0, r[0], "", 0, "L", false, 0, "")

		// Colon
		pdf.SetXY(x+20.0, ry)
		pdf.CellFormat(2.0, 3.0, ":", "", 0, "C", false, 0, "")

		// Value
		pdf.SetFont(font, "B", 4.5)
		pdf.SetTextColor(38, 50, 56)
		pdf.SetXY(x+22.5, ry)
		pdf.CellFormat(w-24.5, 3.0, truncateString(r[1], 20), "", 0, "L", false, 0, "")

		// Divider line
		pdf.SetDrawColor(235, 238, 242)
		pdf.SetLineWidth(0.2)
		pdf.Line(x+6.0, ry+3.8, x+w-2.0, ry+3.8)
	}

	// Front Signatures: TWO SEPARATE SIGNATURE SECTIONS (Authorisation & Signature)
	sigY := y + h - 12.0
	lineWidth := 16.0
	leftX := x + 2.5
	rightX := x + w - lineWidth - 2.5

	pdf.SetDrawColor(120, 120, 120)
	pdf.SetLineWidth(0.3)
	pdf.Line(leftX, sigY, leftX+lineWidth, sigY)
	pdf.Line(rightX, sigY, rightX+lineWidth, sigY)

	pdf.SetFont(font, "", 3.6)
	pdf.SetTextColor(107, 114, 128)

	authText := "Authorisation"
	sigText := "Signature"
	if isBn {
		authText = utils.UnicodeToBijoy("অনুমোদন")
		sigText = utils.UnicodeToBijoy("স্বাক্ষর")
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

	// Bottom Footer Brand Text
	footerText := "EKUSHE FASHIONS"
	if isBn {
		footerText = utils.UnicodeToBijoy("একুশে ফ্যাশনস")
	}
	pdf.SetFont(font, "B", 3.8)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(x, y+h-3.2)
	pdf.CellFormat(w, 3.0, footerText, "", 0, "C", false, 0, "")
}

func drawCardBack(pdf *gofpdf.Fpdf, x, y, w, h float64, emp models.Employee, font string, isBn bool) {
	// Card Outer Border
	pdf.SetDrawColor(210, 215, 220)
	pdf.SetLineWidth(0.4)
	pdf.Rect(x, y, w, h, "D")

	companyName := companyDisplayName(emp.Company, isBn)

	// Top Punch Slot Indicator
	pdf.SetDrawColor(180, 185, 190)
	pdf.SetLineWidth(0.3)
	pdf.SetFillColor(245, 247, 250)
	slotW, slotH := 12.0, 2.5
	pdf.Rect(x+(w-slotW)/2, y+1.2, slotW, slotH, "FD")

	// Header: Circular Logo Badge + Company Name on ONE LINE (LARGE FONT)
	logoR := 3.8
	logoCenterX := x + 6.0
	logoCenterY := y + 8.2

	pdf.SetFillColor(18, 58, 99)
	pdf.SetDrawColor(18, 58, 99)
	pdf.SetLineWidth(0.3)
	pdf.Circle(logoCenterX, logoCenterY, logoR, "FD")

	pdf.SetFont("Arial", "B", 3.2)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(logoCenterX-logoR, logoCenterY-1.5)
	pdf.CellFormat(logoR*2, 3, "LOGO", "", 0, "C", false, 0, "")

	headerFontSz := fitTextFontSize(pdf, font, "B", companyName, w-12.5, 9.5, 6.5)
	pdf.SetFont(font, "B", headerFontSz)
	pdf.SetTextColor(18, 58, 99)
	pdf.SetXY(x+11.0, y+5.8)
	pdf.CellFormat(w-12.0, 5.0, companyName, "", 0, "L", false, 0, "")

	pdf.SetDrawColor(210, 215, 220)
	pdf.SetLineWidth(0.3)
	pdf.Line(x+3.0, y+13.5, x+w-3.0, y+13.5)

	m := 2.5
	cy := y + 14.5

	// Terms & Conditions
	termsTitle := "TERMS & CONDITIONS"
	if isBn {
		termsTitle = utils.UnicodeToBijoy("শর্তাবলী")
	}
	pdf.SetFont(font, "B", 4.4)
	pdf.SetTextColor(18, 58, 99)
	pdf.SetXY(x+m, cy)
	pdf.CellFormat(w-2*m, 2.8, termsTitle, "", 0, "L", false, 0, "")
	cy += 3.2

	var terms []string
	if isBn {
		terms = []string{
			utils.UnicodeToBijoy("১. এই কার্ডটি একুশে ফ্যাশনস এর সম্পত্তি।"),
			utils.UnicodeToBijoy("২. এই কার্ডটি অ-হস্তান্তরযোগ্য।"),
			utils.UnicodeToBijoy("৩. ডিউটির সময় কার্ডটি সাথে রাখুন।"),
			utils.UnicodeToBijoy("৪. কার্ড হারালে সাথে সাথে এইচআর এ জানান।"),
			utils.UnicodeToBijoy("৫. পাওয়া গেলে এইচআর বিভাগে জমা দিন।"),
		}
	} else {
		terms = []string{
			"1. This card is company property.",
			"2. Non-transferable.",
			"3. Must carry during duty.",
			"4. Report lost card to HR.",
			"5. Return to HR if found.",
		}
	}

	pdf.SetFont(font, "", 3.6)
	pdf.SetTextColor(38, 50, 56)
	for _, term := range terms {
		pdf.SetXY(x+m, cy)
		pdf.MultiCell(w-2*m, 2.4, term, "", "L", false)
		cy = pdf.GetY() + 0.4
	}

	pdf.SetDrawColor(230, 235, 240)
	pdf.SetLineWidth(0.2)
	pdf.Line(x+m, cy+1.0, x+w-m, cy+1.0)
	cy += 2.0

	// Company Information / Contact
	infoTitle := "COMPANY INFORMATION"
	if isBn {
		infoTitle = utils.UnicodeToBijoy("যোগাযোগ")
	}
	pdf.SetFont(font, "B", 4.4)
	pdf.SetTextColor(18, 58, 99)
	pdf.SetXY(x+m, cy)
	pdf.CellFormat(w-2*m, 2.8, infoTitle, "", 0, "L", false, 0, "")
	cy += 3.2

	address := ""
	if isBn {
		if emp.Company.AddressBn != "" {
			address = utils.UnicodeToBijoy(emp.Company.AddressBn)
		} else {
			address = utils.UnicodeToBijoy(emp.Company.AddressEn)
		}
	} else {
		address = emp.Company.AddressEn
	}
	if address == "" {
		if isBn {
			address = utils.UnicodeToBijoy("গাজীপুর, বাংলাদেশ")
		} else {
			address = "Gazipur, Bangladesh"
		}
	}

	pdf.SetFont(font, "", 3.8)
	pdf.SetTextColor(38, 50, 56)
	pdf.SetXY(x+m, cy)
	pdf.MultiCell(w-2*m, 2.6, address, "", "L", false)
	cy = pdf.GetY() + 1.0

	phone := emp.Company.Phone
	if phone == "" {
		phone = "+880 1700 000000"
	}
	email := emp.Company.Email
	if email == "" {
		email = "hr@ekushefashions.com"
	}

	pdf.SetFont(font, "", 3.6)
	pdf.SetXY(x+m, cy)
	pdf.CellFormat(w-2*m, 2.4, "Phone: "+phone, "", 0, "L", false, 0, "")
	cy += 2.8
	pdf.SetXY(x+m, cy)
	pdf.CellFormat(w-2*m, 2.4, "Email: "+email, "", 0, "L", false, 0, "")
	cy += 3.5

	// QR Code Box Placeholder
	qrSize := 12.0
	qrX := x + (w-qrSize)/2
	qrY := cy + 1.0
	pdf.SetDrawColor(18, 58, 99)
	pdf.SetLineWidth(0.4)
	pdf.SetFillColor(255, 255, 255)
	pdf.Rect(qrX, qrY, qrSize, qrSize, "FD")

	// Draw stylized QR matrix code inside frame
	pdf.SetFillColor(18, 58, 99)
	pdf.Rect(qrX+1.2, qrY+1.2, 3.0, 3.0, "F")
	pdf.Rect(qrX+qrSize-4.2, qrY+1.2, 3.0, 3.0, "F")
	pdf.Rect(qrX+1.2, qrY+qrSize-4.2, 3.0, 3.0, "F")
	pdf.Rect(qrX+5.0, qrY+5.0, 2.0, 2.0, "F")
	pdf.Rect(qrX+7.5, qrY+7.5, 2.5, 2.5, "F")

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

	// Bottom Footer Brand Text
	footerTextBack := "EKUSHE FASHIONS LTD."
	if isBn {
		footerTextBack = utils.UnicodeToBijoy("একুশে ফ্যাশনস লিঃ")
	}
	pdf.SetFont(font, "B", 3.8)
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
