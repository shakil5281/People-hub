package handlers

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
)

func TestRenderSeparationListExports(t *testing.T) {
	gin.SetMode(gin.TestMode)

	comp := models.Company{
		CompanyNameEn: "Ekushe Fashions Ltd.",
		AddressEn:     "Masterbari, Gazipur",
	}

	seps := []models.Separation{
		{
			EmployeeID: "EMP-00100",
			Employee:   "Rahim Uddin",
			Type:       "Resign",
			Date:       time.Now().Format("2006-01-02"),
			Status:     "Approved",
			Reason:     "Personal reasons",
			Department: models.Department{Name: "Software"},
		},
		{
			EmployeeID: "EMP-00101",
			Employee:   "Karim Ahmed",
			Type:       "Lefty",
			Date:       time.Now().Format("2006-01-02"),
			Status:     "Processed",
			Reason:     "Absenteeism",
			Department: models.Department{Name: "Quality"},
		},
	}

	empDesigMap := map[string]string{
		"EMP-00100": "Senior Software Engineer",
		"EMP-00101": "Quality Inspector",
	}

	// 1. Test Excel Render
	t.Run("ExcelRender", func(t *testing.T) {
		router := gin.New()
		router.GET("/export/excel", func(c *gin.Context) {
			renderSeparationExcel(c, comp, "Period: 2026-08-01 to 2026-08-31", seps, empDesigMap)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/export/excel", nil)
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("excel status code: want 200, got %d", w.Code)
		}
		if len(w.Body.Bytes()) == 0 {
			t.Fatal("excel export produced empty body")
		}
		t.Logf("excel export size=%d bytes", len(w.Body.Bytes()))
	})

	// 2. Test PDF Render (EN)
	t.Run("PDFRender_EN", func(t *testing.T) {
		router := gin.New()
		router.GET("/export/pdf", func(c *gin.Context) {
			renderSeparationPDF(c, comp, "Period: 2026-08-01 to 2026-08-31", seps, empDesigMap, false)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/export/pdf", nil)
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("pdf status code: want 200, got %d", w.Code)
		}
		if len(w.Body.Bytes()) == 0 {
			t.Fatal("pdf export produced empty body")
		}
		t.Logf("pdf export size=%d bytes", len(w.Body.Bytes()))
	})

	// 3. Test PDF Render (BN)
	t.Run("PDFRender_BN", func(t *testing.T) {
		router := gin.New()
		router.GET("/export/pdf", func(c *gin.Context) {
			renderSeparationPDF(c, comp, "Period: 2026-08-01 to 2026-08-31", seps, empDesigMap, true)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/export/pdf?lang=bn", nil)
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("pdf status code: want 200, got %d", w.Code)
		}
		if len(w.Body.Bytes()) == 0 {
			t.Fatal("pdf export produced empty body")
		}
		t.Logf("pdf bn export size=%d bytes", len(w.Body.Bytes()))
	})
}
