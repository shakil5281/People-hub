package utils_test

import (
	"testing"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/utils"
)

// helper: parse "HH:mm" on a fixed reference date for window tests
func mustTime(date, hhmm string) time.Time {
	t, err := time.Parse("2006-01-02 15:04", date+" "+hhmm)
	if err != nil {
		panic(err)
	}
	return t
}

func mustDate(date string) time.Time {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		panic(err)
	}
	return t
}

// ─── Test 1: Window calculation — normal day shift ───────────────────────────

func TestCalculateAttendanceWindow_NormalShift(t *testing.T) {
	// Shift 08:00 → window 07:00 .. next-day 06:59:59
	attendanceDate := mustDate("2026-08-08")
	shiftStart := mustTime("2026-08-08", "08:00")

	w := utils.CalculateAttendanceWindow(attendanceDate, shiftStart)

	expectedStart := mustTime("2026-08-08", "07:00")
	expectedEnd := mustTime("2026-08-09", "06:59").Add(59 * time.Second)

	if !w.Start.Equal(expectedStart) {
		t.Errorf("window start: want %v, got %v", expectedStart, w.Start)
	}
	if !w.End.Equal(expectedEnd) {
		t.Errorf("window end: want %v, got %v", expectedEnd, w.End)
	}
}

// ─── Test 2: Window calculation — overnight shift ────────────────────────────

func TestCalculateAttendanceWindow_OvernightShift(t *testing.T) {
	// Shift 22:00 → window 21:00 .. next-day 20:59:59
	attendanceDate := mustDate("2026-08-08")
	shiftStart := mustTime("2026-08-08", "22:00")

	w := utils.CalculateAttendanceWindow(attendanceDate, shiftStart)

	expectedStart := mustTime("2026-08-08", "21:00")
	expectedEnd := mustTime("2026-08-09", "20:59").Add(59 * time.Second)

	if !w.Start.Equal(expectedStart) {
		t.Errorf("window start: want %v, got %v", expectedStart, w.Start)
	}
	if !w.End.Equal(expectedEnd) {
		t.Errorf("window end: want %v, got %v", expectedEnd, w.End)
	}
}

// ─── Test 3: IsOvernightShift ─────────────────────────────────────────────────

func TestIsOvernightShift(t *testing.T) {
	if utils.IsOvernightShift("22:00", "06:00") == false {
		t.Error("22:00-06:00 should be overnight")
	}
	if utils.IsOvernightShift("08:00", "17:00") == true {
		t.Error("08:00-17:00 should NOT be overnight")
	}
	if utils.IsOvernightShift("00:00", "23:59") == true {
		t.Error("00:00-23:59 should NOT be overnight")
	}
}

// ─── Test 4: BuildShiftEndDatetime — normal shift ─────────────────────────────

func TestBuildShiftEndDatetime_NormalShift(t *testing.T) {
	d := mustDate("2026-08-08")
	end := utils.BuildShiftEndDatetime(d, "08:00", "17:00")
	expected := mustTime("2026-08-08", "17:00")
	if !end.Equal(expected) {
		t.Errorf("want %v, got %v", expected, end)
	}
}

// ─── Test 5: BuildShiftEndDatetime — overnight shift ─────────────────────────

func TestBuildShiftEndDatetime_OvernightShift(t *testing.T) {
	d := mustDate("2026-08-08")
	end := utils.BuildShiftEndDatetime(d, "22:00", "06:00")
	expected := mustTime("2026-08-09", "06:00")
	if !end.Equal(expected) {
		t.Errorf("want %v, got %v", expected, end)
	}
}

// ─── Tests 6-14: CalculateOvertime ───────────────────────────────────────────

func TestCalculateOvertime_Disabled(t *testing.T) {
	shiftEnd := mustTime("2026-08-08", "17:00")
	checkOut := mustTime("2026-08-08", "20:00")
	got := utils.CalculateOvertime(checkOut, shiftEnd, false)
	if got != 0 {
		t.Errorf("overtime_status=false: want 0, got %d", got)
	}
}

func TestCalculateOvertime_NotAfterShiftEnd(t *testing.T) {
	shiftEnd := mustTime("2026-08-08", "17:00")
	checkOut := mustTime("2026-08-08", "16:50")
	got := utils.CalculateOvertime(checkOut, shiftEnd, true)
	if got != 0 {
		t.Errorf("early out: want 0, got %d", got)
	}
}

func TestCalculateOvertime_30Minutes(t *testing.T) {
	// 30 min < 45 → 0
	shiftEnd := mustTime("2026-08-08", "17:00")
	checkOut := mustTime("2026-08-08", "17:30")
	got := utils.CalculateOvertime(checkOut, shiftEnd, true)
	if got != 0 {
		t.Errorf("30 min OT: want 0, got %d", got)
	}
}

func TestCalculateOvertime_44Minutes(t *testing.T) {
	// 44 min < 45 → 0
	shiftEnd := mustTime("2026-08-08", "17:00")
	checkOut := mustTime("2026-08-08", "17:44")
	got := utils.CalculateOvertime(checkOut, shiftEnd, true)
	if got != 0 {
		t.Errorf("44 min OT: want 0, got %d", got)
	}
}

func TestCalculateOvertime_45Minutes(t *testing.T) {
	// 45 min → 1
	shiftEnd := mustTime("2026-08-08", "17:00")
	checkOut := mustTime("2026-08-08", "17:45")
	got := utils.CalculateOvertime(checkOut, shiftEnd, true)
	if got != 1 {
		t.Errorf("45 min OT: want 1, got %d", got)
	}
}

func TestCalculateOvertime_60Minutes(t *testing.T) {
	// 60 min → 1
	shiftEnd := mustTime("2026-08-08", "17:00")
	checkOut := mustTime("2026-08-08", "18:00")
	got := utils.CalculateOvertime(checkOut, shiftEnd, true)
	if got != 1 {
		t.Errorf("60 min OT: want 1, got %d", got)
	}
}

func TestCalculateOvertime_104Minutes(t *testing.T) {
	// 104 min → 1
	shiftEnd := mustTime("2026-08-08", "17:00")
	checkOut := mustTime("2026-08-08", "18:44")
	got := utils.CalculateOvertime(checkOut, shiftEnd, true)
	if got != 1 {
		t.Errorf("104 min OT: want 1, got %d", got)
	}
}

func TestCalculateOvertime_105Minutes(t *testing.T) {
	// 105 min → 2
	shiftEnd := mustTime("2026-08-08", "17:00")
	checkOut := mustTime("2026-08-08", "18:45")
	got := utils.CalculateOvertime(checkOut, shiftEnd, true)
	if got != 2 {
		t.Errorf("105 min OT: want 2, got %d", got)
	}
}

func TestCalculateOvertime_165Minutes(t *testing.T) {
	// 165 min → 3
	shiftEnd := mustTime("2026-08-08", "17:00")
	checkOut := mustTime("2026-08-08", "19:45")
	got := utils.CalculateOvertime(checkOut, shiftEnd, true)
	if got != 3 {
		t.Errorf("165 min OT: want 3, got %d", got)
	}
}

// ─── Test 15-16: CalcTotalHoursStr ───────────────────────────────────────────

func TestCalcTotalHoursStr_Normal(t *testing.T) {
	in := mustTime("2026-08-08", "08:00")
	out := mustTime("2026-08-08", "17:00")
	result := utils.CalcTotalHoursStr(&in, &out)
	if result == nil || *result != "08:00" {
		t.Errorf("want 08:00, got %v", result)
	}
}

func TestCalcTotalHoursStr_NilPointers(t *testing.T) {
	result := utils.CalcTotalHoursStr(nil, nil)
	if result != nil {
		t.Errorf("expected nil for nil inputs, got %v", result)
	}
}

func TestCalcTotalHoursStr_Overnight(t *testing.T) {
	in := mustTime("2026-08-08", "22:00")
	out := mustTime("2026-08-09", "06:00") // 8 hours across midnight (no 13:00-14:00 overlap)
	result := utils.CalcTotalHoursStr(&in, &out)
	if result == nil || *result != "08:00" {
		t.Errorf("want 08:00, got %v", result)
	}
}

// ─── Test 17: ParseShiftHHMM ──────────────────────────────────────────────────

func TestParseShiftHHMM(t *testing.T) {
	tests := []struct {
		input   string
		wantH   int
		wantM   int
		wantOK  bool
	}{
		{"08:00", 8, 0, true},
		{"22:30", 22, 30, true},
		{"00:00", 0, 0, true},
		{"23:59", 23, 59, true},
		{"invalid", 0, 0, false},
		{"8:00", 0, 0, false},
		{"25:00", 0, 0, false},
	}
	for _, tt := range tests {
		h, m, ok := utils.ParseShiftHHMM(tt.input)
		if ok != tt.wantOK || (ok && (h != tt.wantH || m != tt.wantM)) {
			t.Errorf("ParseShiftHHMM(%q): want h=%d m=%d ok=%v, got h=%d m=%d ok=%v",
				tt.input, tt.wantH, tt.wantM, tt.wantOK, h, m, ok)
		}
	}
}

// ─── Test 18: Window boundary — punch outside window is excluded ──────────────

func TestWindowBoundary_PunchOutsideIgnored(t *testing.T) {
	// Shift 08:00 → window 07:00:00 .. next-day 06:59:59
	attendanceDate := mustDate("2026-08-08")
	shiftStart := mustTime("2026-08-08", "08:00")
	w := utils.CalculateAttendanceWindow(attendanceDate, shiftStart)

	// 06:55 today → BEFORE window start (07:00:00) → should be excluded
	punchBefore := mustTime("2026-08-08", "06:55")
	// 07:05 today → inside window → included
	punchInside := mustTime("2026-08-08", "07:05")
	// 07:00 next day → AFTER window end (06:59:59 next day) → excluded
	punchAfter := mustTime("2026-08-09", "07:00")
	// 06:59:59 next day → exactly at window end → included
	punchAtEnd := w.End

	if !punchBefore.Before(w.Start) {
		t.Error("punch 06:55 should be BEFORE window start")
	}
	if punchInside.Before(w.Start) || punchInside.After(w.End) {
		t.Error("punch 07:05 should be INSIDE window")
	}
	if !punchAfter.After(w.End) {
		t.Error("punch 07:00 next-day should be AFTER window end")
	}
	if punchAtEnd.After(w.End) {
		t.Error("punch at window.End should be AT or INSIDE window end")
	}
}

// ─── Test 19: ParseHHMMToMinutes ─────────────────────────────────────────────

func TestParseHHMMToMinutes(t *testing.T) {
	m, ok := utils.ParseHHMMToMinutes("09:30")
	if !ok || m != 570 {
		t.Errorf("want 570, got %d (ok=%v)", m, ok)
	}
	_, ok2 := utils.ParseHHMMToMinutes("invalid")
	if ok2 {
		t.Error("expected ok=false for invalid input")
	}
}

// ─── Test 20: OT formula exhaustive table ─────────────────────────────────────

func TestCalculateOvertime_TableDriven(t *testing.T) {
	shiftEnd := mustTime("2026-08-08", "17:00")
	cases := []struct {
		extraMinutes int
		wantOT       int
	}{
		{0, 0},
		{30, 0},
		{44, 0},
		{45, 1},
		{60, 1},
		{104, 1},
		{105, 2},
		{120, 2},
		{345, 6}, // 6h -> 6h
		{405, 8}, // 7h -> 8h (7+1 bonus rule!)
		{465, 8}, // 8h -> 8h
		{525, 9}, // 9h -> 9h
	}
	for _, c := range cases {
		checkOut := shiftEnd.Add(time.Duration(c.extraMinutes) * time.Minute)
		got := utils.CalculateOvertime(checkOut, shiftEnd, true)
		if got != c.wantOT {
			t.Errorf("OT %d min: want %d hours, got %d hours", c.extraMinutes, c.wantOT, got)
		}
	}
}

// ─── Test 21: Overnight OT — shift end on next day ────────────────────────────

func TestCalculateOvertime_OvernightShiftOT(t *testing.T) {
	// Shift 22:00 – 06:00 next day
	// Employee out: 07:00 next day → 60 min OT → 1 OT hour
	shiftEnd := mustTime("2026-08-09", "06:00") // next-day end
	checkOut := mustTime("2026-08-09", "07:00")
	got := utils.CalculateOvertime(checkOut, shiftEnd, true)
	if got != 1 {
		t.Errorf("overnight OT 60 min: want 1, got %d", got)
	}
}

// ─── Test 22: ShiftStartOnDate ────────────────────────────────────────────────

func TestShiftStartOnDate(t *testing.T) {
	d := mustDate("2026-08-08")
	result := utils.ShiftStartOnDate("08:00", d)
	expected := mustTime("2026-08-08", "08:00")
	if !result.Equal(expected) {
		t.Errorf("want %v, got %v", expected, result)
	}
}

func TestShiftStartOnDate_Invalid(t *testing.T) {
	d := mustDate("2026-08-08")
	result := utils.ShiftStartOnDate("bad", d)
	if !result.IsZero() {
		t.Error("invalid hhmm should return zero time")
	}
}

// ─── Test 23: Half-day — worked < 4 hours ─────────────────────────────────────

func TestHalfDayDetection(t *testing.T) {
	// worked 3h30m < 4h
	in := mustTime("2026-08-08", "08:00")
	out := mustTime("2026-08-08", "11:30")
	result := utils.CalcTotalHoursStr(&in, &out)
	if result == nil {
		t.Fatal("expected non-nil total hours")
	}
	m, ok := utils.ParseHHMMToMinutes(*result)
	if !ok {
		t.Fatalf("parse failed: %s", *result)
	}
	if m >= 4*60 {
		t.Errorf("3h30m: expected < 240 minutes, got %d", m)
	}
}

// ─── Test 24: Window is exactly 24h - 1s ──────────────────────────────────────

func TestWindowDurationIs24HoursMinusOneSecond(t *testing.T) {
	attendanceDate := mustDate("2026-08-08")
	shiftStart := mustTime("2026-08-08", "08:00")
	w := utils.CalculateAttendanceWindow(attendanceDate, shiftStart)

	duration := w.End.Sub(w.Start)
	expected := 24*time.Hour - time.Second

	if duration != expected {
		t.Errorf("window duration: want %v, got %v", expected, duration)
	}
}

// ─── Test 25: OT disabled even with large overtime ────────────────────────────

func TestCalculateOvertime_DisabledWith3HoursWorked(t *testing.T) {
	shiftEnd := mustTime("2026-08-08", "17:00")
	checkOut := mustTime("2026-08-08", "20:00") // 3h overtime
	got := utils.CalculateOvertime(checkOut, shiftEnd, false)
	if got != 0 {
		t.Errorf("OT disabled 3h extra: want 0, got %d", got)
	}
}
