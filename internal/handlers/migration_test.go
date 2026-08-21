package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
)

func TestRenderMigrationListExports(t *testing.T) {
	gin.SetMode(gin.TestMode)

	comp := models.Company{
		CompanyNameEn: "Ekushe Fashions Ltd.",
		AddressEn:     "Masterbari, Gazipur",
	}

	summary := &repository.FullMigrationSummary{
		TotalJoining: 15,
		TotalLeft:    5,
		TotalResign:  3,
		TotalActive:  120,
		ByDepartment: []repository.OrganizationalSummaryRow{
			{ID: "d1", Name: "Cutting", NewJoining: 5, TotalLeft: 2, TotalResign: 1, ActiveTotal: 45},
			{ID: "d2", Name: "Sewing", NewJoining: 10, TotalLeft: 3, TotalResign: 2, ActiveTotal: 75},
		},
		BySection: []repository.OrganizationalSummaryRow{
			{ID: "s1", Name: "Section A", NewJoining: 8, TotalLeft: 2, TotalResign: 1, ActiveTotal: 60},
		},
		ByDesignation: []repository.OrganizationalSummaryRow{
			{ID: "des1", Name: "Operator", NewJoining: 12, TotalLeft: 4, TotalResign: 2, ActiveTotal: 90},
		},
		ByLine: []repository.OrganizationalSummaryRow{
			{ID: "l1", Name: "Line 01", NewJoining: 6, TotalLeft: 1, TotalResign: 1, ActiveTotal: 30},
		},
	}

	// 1. Test Excel Render
	t.Run("ExcelRender", func(t *testing.T) {
		router := gin.New()
		router.GET("/migrations/export/excel", func(c *gin.Context) {
			renderFullMigrationExcel(c, comp, "Period: 2026-08-01 to 2026-08-31", summary)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/migrations/export/excel", nil)
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("excel status code: want 200, got %d", w.Code)
		}
		if len(w.Body.Bytes()) == 0 {
			t.Fatal("excel export produced empty body")
		}
		t.Logf("excel migration summary export size=%d bytes", len(w.Body.Bytes()))
	})

	// 2. Test PDF Render (EN)
	t.Run("PDFRender_EN", func(t *testing.T) {
		router := gin.New()
		router.GET("/migrations/export/pdf", func(c *gin.Context) {
			renderFullMigrationPDF(c, comp, "Period: 2026-08-01 to 2026-08-31", summary, false)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/migrations/export/pdf", nil)
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("pdf status code: want 200, got %d", w.Code)
		}
		if len(w.Body.Bytes()) == 0 {
			t.Fatal("pdf export produced empty body")
		}
		t.Logf("pdf migration summary export size=%d bytes", len(w.Body.Bytes()))
	})

	// 3. Test PDF Render (BN)
	t.Run("PDFRender_BN", func(t *testing.T) {
		router := gin.New()
		router.GET("/migrations/export/pdf", func(c *gin.Context) {
			renderFullMigrationPDF(c, comp, "Period: 2026-08-01 to 2026-08-31", summary, true)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/migrations/export/pdf?lang=bn", nil)
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("pdf status code: want 200, got %d", w.Code)
		}
		if len(w.Body.Bytes()) == 0 {
			t.Fatal("pdf export produced empty body")
		}
		t.Logf("pdf bn migration summary export size=%d bytes", len(w.Body.Bytes()))
	})
}
