package handlers

import (
	"os"
	"testing"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/shakil5281/peoplehub-api/internal/models"
)

func TestRenderLeaveFormPDF(t *testing.T) {
	emp := &models.Employee{
		EmployeeID:     "EMP-00125",
		PunchNumber:    "125",
		NameEn:         "Shakil Hossen",
		NameBn:         "শাকিল হোসেন",
		Phone:          "01700000000",
		EmployeeType:   "Permanent",
		Grade:          "A",
		JoiningDate:    time.Now(),
		PresentAddress: "Dhanmondi, Dhaka",
		EmergencyPhone: "01800000000",
		Department:     &models.Department{Name: "Software", NameBn: "সফটওয়্যার"},
		SectionRef:     &models.Section{Name: "Development", NameBn: "ডেভেলপমেন্ট"},
		DesignationRef: &models.Designation{Name: "Software Engineer", NameBn: "সফটওয়্যার ইঞ্জিনিয়ার"},
		Shift:          &models.Shift{Name: "General"},
		Manager:        &models.Employee{NameEn: "Md. Rahman", NameBn: "মো. রহমান"},
	}
	leave := &models.Leave{
		ID:         "3f2c1a0b-9d6e-4c5a-b8f7-2e1d0c9a8b76",
		EmployeeID: "EMP-00125",
		FromDate:   "2026-07-20",
		ToDate:     "2026-07-22",
		TotalDays:  3,
		Reason:     "Family emergency",
		Status:     "pending",
		Company: models.Company{
			CompanyNameEn: "ABC Garments Ltd.",
			CompanyNameBn: "এবিসি গার্মেন্টস লিমিটেড",
			AddressEn:     "Head Office, Dhaka",
			AddressBn:     "প্রধান কার্যালয়, ঢাকা",
			Phone:         "+8802-123456",
			Email:         "info@abc.com",
		},
		Employee:  *emp,
		LeaveType: models.LeaveType{Name: "Casual Leave"},
	}

	for _, lang := range []string{"en", "bn"} {
		labels := leaveEnLabels
		if lang == "bn" {
			labels = leaveBnLabels
		}
		allocs := []models.LeaveAllocation{
			{LeaveType: models.LeaveType{Name: "Annual Leave"}, TotalDays: 20, UsedDays: 8},
			{LeaveType: models.LeaveType{Name: "Casual Leave"}, TotalDays: 10, UsedDays: 3},
			{LeaveType: models.LeaveType{Name: "Sick Leave"}, TotalDays: 14, UsedDays: 2},
		}
		data := buildLeaveFormData(leave, lang, labels, allocs)

		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.SetMargins(0, 0, 0)
		pdf.SetAutoPageBreak(false, 0)
		pdf.AddPage()
		font := leaveFormFont(pdf, lang)
		renderLeaveFormPDFPage(pdf, font, lang, data, labels)

		out := "C:\\Users\\shaki\\AppData\\Local\\Temp\\opencode\\leave_form_" + lang + ".pdf"
		if err := pdf.OutputFileAndClose(out); err != nil {
			t.Fatalf("lang=%s Output: %v", lang, err)
		}
		info, _ := os.Stat(out)
		t.Logf("lang=%s size=%d", lang, info.Size())
	}
}
