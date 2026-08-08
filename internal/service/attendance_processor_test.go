package service

import (
	"testing"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/models"
)

// helper: build a DataLog with a PunchTime from a date+HH:mm string.
func mkPunch(dateHHMM string) models.DataLog {
	t, err := time.Parse("2006-01-02 15:04", dateHHMM)
	if err != nil {
		panic(err)
	}
	return models.DataLog{PunchTime: t}
}

// shiftEnd helper: build a time.Time for shiftEnd on a given date.
func mkShiftEnd(dateHHMM string) time.Time {
	t, err := time.Parse("2006-01-02 15:04", dateHHMM)
	if err != nil {
		panic(err)
	}
	return t
}

// ─── Single-punch rules ───────────────────────────────────────────────────────
// Shift General: 08:00 – 17:00.  shiftEndDT = 17:00.  threshold = 12:00.

// Test 1: Single punch in inTime zone (07:52 < 12:00) → checkIn = 07:52, checkOut = nil.
func TestResolveInOut_SinglePunch_InTimeZone_IsInTime(t *testing.T) {
	shiftEnd := mkShiftEnd("2026-08-04 17:00")
	punches := []models.DataLog{mkPunch("2026-08-04 07:52")}

	in, out := resolveInOut(punches, shiftEnd)

	if in == nil || in.Hour() != 7 || in.Minute() != 52 {
		t.Errorf("checkIn: want 07:52, got %v", in)
	}
	if out != nil {
		t.Errorf("checkOut: want nil, got %v", *out)
	}
}

// Test 2: Single punch in outTime zone (18:30 >= 12:00) → checkIn = nil, checkOut = 18:30.
func TestResolveInOut_SinglePunch_AfterShiftEnd_IsOutTime(t *testing.T) {
	shiftEnd := mkShiftEnd("2026-08-04 17:00")
	punches := []models.DataLog{mkPunch("2026-08-04 18:30")} // 18:30 > 17:00

	in, out := resolveInOut(punches, shiftEnd)

	if in != nil {
		t.Errorf("checkIn: want nil, got %v", *in)
	}
	if out == nil || out.Hour() != 18 || out.Minute() != 30 {
		t.Errorf("checkOut: want 18:30, got %v", out)
	}
}

// Test 3: Single punch at next-day early morning (05-08 03:58 >= 04-08 12:00) → checkIn = nil, checkOut = 03:58.
func TestResolveInOut_SinglePunch_NextDayEarlyMorning_IsOutTime(t *testing.T) {
	shiftEnd := mkShiftEnd("2026-08-04 17:00")
	punches := []models.DataLog{mkPunch("2026-08-05 03:58")}

	in, out := resolveInOut(punches, shiftEnd)

	if in != nil {
		t.Errorf("checkIn: want nil, got %v", *in)
	}
	if out == nil || out.Day() != 5 || out.Hour() != 3 || out.Minute() != 58 {
		t.Errorf("checkOut: want 05-08 03:58, got %v", out)
	}
}

// Test 4: Single punch exactly at threshold (12:00) → outTime zone (12:00 >= 12:00) → checkIn = nil, checkOut = 12:00.
func TestResolveInOut_SinglePunch_ExactlyAtThreshold_IsOutTime(t *testing.T) {
	shiftEnd := mkShiftEnd("2026-08-04 17:00")
	punches := []models.DataLog{mkPunch("2026-08-04 12:00")}

	in, out := resolveInOut(punches, shiftEnd)

	if in != nil {
		t.Errorf("checkIn: want nil, got %v", *in)
	}
	if out == nil || out.Hour() != 12 || out.Minute() != 0 {
		t.Errorf("checkOut: want 12:00, got %v", out)
	}
}

// Test 5: Single punch, no shift configured → treated as checkIn.
func TestResolveInOut_SinglePunch_NoShift_IsInTime(t *testing.T) {
	punches := []models.DataLog{mkPunch("2026-08-04 07:52")}

	in, out := resolveInOut(punches, time.Time{})

	if in == nil || in.Hour() != 7 || in.Minute() != 52 {
		t.Errorf("checkIn: want 07:52, got %v", in)
	}
	if out != nil {
		t.Errorf("checkOut: want nil, got %v", *out)
	}
}

// ─── Multi-punch rules ────────────────────────────────────────────────────────

// Test 6: Multiple punches, ALL in outTime zone (all >= 12:00) → last = checkOut, checkIn = nil.
func TestResolveInOut_MultiPunch_AllAfterThreshold_IsOutTimeOnly(t *testing.T) {
	shiftEnd := mkShiftEnd("2026-08-04 17:00")
	punches := []models.DataLog{
		mkPunch("2026-08-04 17:30"),
		mkPunch("2026-08-04 18:00"),
		mkPunch("2026-08-04 18:45"),
	}

	in, out := resolveInOut(punches, shiftEnd)

	if in != nil {
		t.Errorf("checkIn: want nil, got %v", *in)
	}
	if out == nil || out.Hour() != 18 || out.Minute() != 45 {
		t.Errorf("checkOut: want 18:45, got %v", out)
	}
}

// Test 7: Normal punches (first in inTime zone, last after 25m) → first=in, last=out.
func TestResolveInOut_Normal_FirstInLastOut(t *testing.T) {
	shiftEnd := mkShiftEnd("2026-08-04 17:00")
	punches := []models.DataLog{
		mkPunch("2026-08-04 07:52"), // In (inTime zone < 12:00)
		mkPunch("2026-08-04 08:05"), // within 25-min debounce → ignored
		mkPunch("2026-08-04 17:30"), // Out (outTime zone)
	}

	in, out := resolveInOut(punches, shiftEnd)

	if in == nil || in.Hour() != 7 || in.Minute() != 52 {
		t.Errorf("checkIn: want 07:52, got %v", in)
	}
	if out == nil || out.Hour() != 17 || out.Minute() != 30 {
		t.Errorf("checkOut: want 17:30, got %v", out)
	}
}

// Test 8: Overnight punch scenario — 04-08 07:52 and 05-08 03:58 in same window.
func TestResolveInOut_Overnight_NextDayPunch(t *testing.T) {
	shiftEnd := mkShiftEnd("2026-08-04 17:00")
	punches := []models.DataLog{
		mkPunch("2026-08-04 07:52"), // In
		mkPunch("2026-08-05 03:58"), // Out
	}

	in, out := resolveInOut(punches, shiftEnd)

	if in == nil || in.Hour() != 7 || in.Minute() != 52 {
		t.Errorf("checkIn: want 07:52, got %v", in)
	}
	if out == nil || out.Day() != 5 || out.Hour() != 3 || out.Minute() != 58 {
		t.Errorf("checkOut: want 05-08 03:58, got %v", out)
	}
}

// Test 9: 25-min debounce — all punches within 25 min of first → checkIn set, no checkOut.
func TestResolveInOut_Debounce_AllWithin25Min(t *testing.T) {
	shiftEnd := mkShiftEnd("2026-08-04 17:00")
	punches := []models.DataLog{
		mkPunch("2026-08-04 08:00"),
		mkPunch("2026-08-04 08:10"),
		mkPunch("2026-08-04 08:20"),
		mkPunch("2026-08-04 08:24"), // 24 min after first → within debounce
	}

	in, out := resolveInOut(punches, shiftEnd)

	if in == nil || in.Hour() != 8 || in.Minute() != 0 {
		t.Errorf("checkIn: want 08:00, got %v", in)
	}
	if out != nil {
		t.Errorf("checkOut: want nil (all within debounce), got %v", *out)
	}
}

// Test 10: No punches → both nil.
func TestResolveInOut_NoPunches(t *testing.T) {
	shiftEnd := mkShiftEnd("2026-08-04 17:00")
	in, out := resolveInOut(nil, shiftEnd)
	if in != nil || out != nil {
		t.Errorf("want (nil, nil), got (%v, %v)", in, out)
	}
}
