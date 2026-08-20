package service

import (
	"testing"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
)

func TestComputeNightBill_Hourly45MinThreshold(t *testing.T) {
	dateStr := "2026-08-18"

	parseTime := func(tStr string) *time.Time {
		tm, err := time.Parse("2006-01-02 15:04:05", dateStr+" "+tStr)
		if err != nil {
			t.Fatalf("failed to parse time: %v", err)
		}
		return &tm
	}

	tests := []struct {
		name          string
		checkOutStr   string
		expectedHours float64
		expectedQual  bool
	}{
		{
			name:          "Out at 20:44 (44 mins past 20:00) -> 0 hrs, does not qualify",
			checkOutStr:   "20:44:00",
			expectedHours: 0,
			expectedQual:  false,
		},
		{
			name:          "Out at 20:45 (45 mins past 20:00) -> 1 hr, qualifies",
			checkOutStr:   "20:45:00",
			expectedHours: 1.0,
			expectedQual:  true,
		},
		{
			name:          "Out at 20:50 (User example: 50 mins past 20:00) -> 1 hr, qualifies",
			checkOutStr:   "20:50:00",
			expectedHours: 1.0,
			expectedQual:  true,
		},
		{
			name:          "Out at 21:44 (1 hr 44 mins past 20:00) -> 1 hr, qualifies",
			checkOutStr:   "21:44:00",
			expectedHours: 1.0,
			expectedQual:  true,
		},
		{
			name:          "Out at 21:45 (1 hr 45 mins past 20:00) -> 2 hrs, qualifies",
			checkOutStr:   "21:45:00",
			expectedHours: 2.0,
			expectedQual:  true,
		},
	}

	checkIn := parseTime("08:00:00")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkOut := parseTime(tt.checkOutStr)
			att := models.Attendance{
				CheckIn:  checkIn,
				CheckOut: checkOut,
			}

			eligibleHours, rate, amount, qualifies := computeNightBill(att, dateStr, "hourly", 0, 20.0)
			if qualifies != tt.expectedQual {
				t.Errorf("qualifies = %v, want %v", qualifies, tt.expectedQual)
			}
			if eligibleHours != tt.expectedHours {
				t.Errorf("eligibleHours = %v, want %v", eligibleHours, tt.expectedHours)
			}
			if tt.expectedQual {
				expectedAmount := tt.expectedHours * 20.0
				if amount != expectedAmount {
					t.Errorf("amount = %v, want %v", amount, expectedAmount)
				}
				if rate != 20.0 {
					t.Errorf("rate = %v, want 20.0", rate)
				}
			}
		})
	}
}
