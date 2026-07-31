package handlers

import (
	"bytes"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/shakil5281/peoplehub-api/internal/models"
)

func sampleJobCardSections() []jobCardSection {
	emp := &models.Employee{
		EmployeeID:   "EMP-00125",
		PunchNumber:  "125",
		NameEn:       "Shakil Hossen",
		NameBn:       "শাকিল হোসেন",
		Phone:        "01700000000",
		JoiningDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Department:   &models.Department{Name: "Software", NameBn: "সফটওয়্যার"},
		SectionRef:   &models.Section{Name: "Development", NameBn: "ডেভেলপমেন্ট"},
		DesignationRef: &models.Designation{Name: "Software Engineer", NameBn: "সফটওয়্যার ইঞ্জিনিয়ার"},
		Company:      models.Company{
			CompanyNameEn: "ABC Garments Ltd.",
			CompanyNameBn: "এবিসি গার্মেন্টস লিমিটেড",
			AddressEn:     "Gazipur, Bangladesh",
			Phone:         "+880-2-123456",
			Email:         "hr@abcgarments.com",
		},
	}

	var rows []models.Attendance
	for d := 1; d <= 10; d++ {
		status := "present"
		switch d {
		case 3:
			status = "late"
		case 6:
			status = "absent"
		case 8:
			status = "on_leave"
		case 9:
			status = "half_day"
		case 10:
			status = "weekend"
		}
		in := time.Date(2026, 7, d, 9, 0, 0, 0, time.UTC)
		out := time.Date(2026, 7, d, 18, 0, 0, 0, time.UTC)
		total := "8.0"
		rows = append(rows, models.Attendance{
			EmployeeID:  emp.EmployeeID,
			CompanyID:   emp.CompanyID,
			Date:        time.Date(2026, 7, d, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
			CheckIn:     &in,
			CheckOut:    &out,
			TotalHours:  &total,
			Status:      status,
			LateMinutes: 12,
			Employee:    *emp,
			Shift:       &models.Shift{Name: "General"},
		})
	}

	return []jobCardSection{{Employee: *emp, Rows: rows}}
}

func TestRenderJobCardPDF(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(false, 0)
	font := loadBanglaFont(pdf)
	labels := jobCardEnLabels
	company := models.Company{CompanyNameEn: "ABC Garments Ltd."}
	period := jobCardPeriod("2026-07-01", "2026-07-10")

	for _, sec := range sampleJobCardSections() {
		pdf.AddPage()
		drawJobCardPage(pdf, font, labels, company, period, sec)
	}
	if pdf.Error() != nil {
		t.Fatalf("pdf error: %v", pdf.Error())
	}
	out := "C:\\Users\\shaki\\AppData\\Local\\Temp\\opencode\\job_card_en.pdf"
	if err := pdf.OutputFileAndClose(out); err != nil {
		t.Fatalf("output: %v", err)
	}
	info, _ := os.Stat(out)
	t.Logf("job card en pdf size=%d", info.Size())

	// BN
	pdf2 := gofpdf.New("P", "mm", "A4", "")
	pdf2.SetMargins(10, 10, 10)
	pdf2.SetAutoPageBreak(false, 0)
	font2 := loadBanglaFont(pdf2)
	for _, sec := range sampleJobCardSections() {
		pdf2.AddPage()
		drawJobCardPage(pdf2, font2, jobCardBnLabels, company, period, sec)
	}
	out2 := "C:\\Users\\shaki\\AppData\\Local\\Temp\\opencode\\job_card_bn.pdf"
	if err := pdf2.OutputFileAndClose(out2); err != nil {
		t.Fatalf("output bn: %v", err)
	}
	info2, _ := os.Stat(out2)
	t.Logf("job card bn pdf size=%d", info2.Size())
}

func TestRenderJobCardExcel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/export", func(c *gin.Context) {
		renderJobCardExcel(c, "2026-07-01", "2026-07-10", sampleJobCardSections())
	})

	req := httptest.NewRequest("GET", "/export", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
	if w.Body.Len() == 0 {
		t.Fatal("empty body")
	}
	out := "C:\\Users\\shaki\\AppData\\Local\\Temp\\opencode\\job_card.xlsx"
	if err := os.WriteFile(out, w.Body.Bytes(), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	info, _ := os.Stat(out)
	t.Logf("job card xlsx size=%d", info.Size())
}

func TestJobCardHelpers(t *testing.T) {
	if jobCardStatusCode("absent") != "A" {
		t.Fatal("absent status code wrong")
	}
	if jobCardStatusCode("half_day") != "HD" {
		t.Fatal("half_day status code wrong")
	}
	if jobCardDateDisplay("2026-07-01") != "01 Jul 2026" {
		t.Fatal("date display wrong: " + jobCardDateDisplay("2026-07-01"))
	}
	secs := groupJobCard(sampleJobCardSections()[0].Rows)
	if len(secs) != 1 || len(secs[0].Rows) != 10 {
		t.Fatalf("group failed: %d sections, %d rows", len(secs), len(secs[0].Rows))
	}
	st := jobCardStatsFor(secs[0].Rows)
	if st.Total != 10 || st.Absent != 1 || st.Leave != 1 {
		t.Fatalf("stats wrong: %+v", st)
	}
	var buf bytes.Buffer
	if buf.Len() != 0 {
		t.Fatal("buf should be empty")
	}
}
