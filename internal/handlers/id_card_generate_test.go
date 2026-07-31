package handlers

import (
	"os"
	"testing"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/shakil5281/peoplehub-api/internal/models"
)

func TestRenderIdCards(t *testing.T) {
	emp := &models.Employee{
		EmployeeID:   "EMP-00125",
		PunchNumber:  "125",
		NameEn:       "Shakil Hossen",
		NameBn:       "শাকিল হোসেন",
		Phone:        "01700000000",
		EmployeeType: "Permanent",
		Grade:        "A",
		JoiningDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EmergencyPhone: "01800000000",
		Department:   &models.Department{Name: "Software", NameBn: "সফটওয়্যার"},
		SectionRef:   &models.Section{Name: "Development", NameBn: "ডেভেলপমেন্ট"},
		DesignationRef: &models.Designation{Name: "Software Engineer", NameBn: "সফটওয়্যার ইঞ্জিনিয়ার"},
		LineRef:      &models.Line{Name: "Line-02"},
		Company:      models.Company{
			CompanyNameEn: "ABC Garments Ltd.",
			CompanyNameBn: "এবিসি গার্মেন্টস লিমিটেড",
			AddressEn:     "House # 15, Road # 08, Industrial Area, Gazipur, Bangladesh",
			AddressBn:     "বাসা # ১৫, রোড # ০৮, শিল্প এলাকা, গাজীপুর, বাংলাদেশ",
			Phone:         "+880-2-123456",
			Email:         "hr@abcgarments.com",
		},
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	font := loadBanglaFont(pdf)

	pdf.AddPage()
	drawCardFront(pdf, 8, 26.75, 62.67, 95, *emp, font)
	drawCardBack(pdf, 8, 175.25, 62.67, 95, *emp, font)

	if pdf.Error() != nil {
		t.Fatalf("pdf error: %v", pdf.Error())
	}

	out := "C:\\Users\\shaki\\AppData\\Local\\Temp\\opencode\\id_cards_test.pdf"
	if err := pdf.OutputFileAndClose(out); err != nil {
		t.Fatalf("output: %v", err)
	}
	info, _ := os.Stat(out)
	t.Logf("id card pdf size=%d", info.Size())
}
