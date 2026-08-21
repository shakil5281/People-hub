package utils

import (
	"time"
)

// AttendanceWindow defines the inclusive 24-hour punch window for one attendance date.
// Start = shiftStart - 1 hour
// End   = Start + 24h - 1s
//
// Example: shift 08:00 → window 07:00:00 … next-day 06:59:59
type AttendanceWindow struct {
	Start time.Time
	End   time.Time
}

// CalculateAttendanceWindow returns the 24-hour punch window for the given attendance
// date and shift start time.
//
//	windowStart = (attendanceDate @ shiftStartHH:mm) - 1 hour
//	windowEnd   = windowStart + 24h - 1s
//
// The shiftStart parameter must already carry the correct HH:mm (the date portion is
// ignored; only the time-of-day is used).
func CalculateAttendanceWindow(attendanceDate time.Time, shiftStart time.Time) AttendanceWindow {
	// Anchor the shift start to the attendance date (keep same clock time).
	shiftOnDate := time.Date(
		attendanceDate.Year(),
		attendanceDate.Month(),
		attendanceDate.Day(),
		shiftStart.Hour(),
		shiftStart.Minute(),
		0, 0,
		attendanceDate.Location(),
	)

	windowStart := shiftOnDate.Add(-1 * time.Hour)
	windowEnd := windowStart.Add(24*time.Hour - time.Second)

	return AttendanceWindow{Start: windowStart, End: windowEnd}
}

// ParseShiftHHMM parses a "HH:mm" string into hour/minute integers.
// Returns 0,0 and false on error.
func ParseShiftHHMM(hhmm string) (hour, minute int, ok bool) {
	if len(hhmm) != 5 || hhmm[2] != ':' {
		return 0, 0, false
	}
	h := int(hhmm[0]-'0')*10 + int(hhmm[1]-'0')
	m := int(hhmm[3]-'0')*10 + int(hhmm[4]-'0')
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, 0, false
	}
	return h, m, true
}

// ShiftStartOnDate returns the full time.Time for a shift's start on a given date.
// hhmm is expected to be in "HH:mm" format.
func ShiftStartOnDate(hhmm string, onDate time.Time) time.Time {
	h, m, ok := ParseShiftHHMM(hhmm)
	if !ok {
		return time.Time{}
	}
	return time.Date(onDate.Year(), onDate.Month(), onDate.Day(), h, m, 0, 0, onDate.Location())
}

// IsOvernightShift returns true when the shift spans midnight
// (i.e. start time is numerically after end time, e.g. 22:00–06:00).
func IsOvernightShift(startHHMM, endHHMM string) bool {
	sh, sm, sok := ParseShiftHHMM(startHHMM)
	eh, em, eok := ParseShiftHHMM(endHHMM)
	if !sok || !eok {
		return false
	}
	startMins := sh*60 + sm
	endMins := eh*60 + em
	return startMins > endMins
}

// BuildShiftEndDatetime returns the full datetime of the shift's end for an attendance
// date, correctly placed on the next day for overnight shifts.
//
// For a normal day shift (08:00–17:00) the end is on the same day.
// For an overnight shift (22:00–06:00) the end is on attendanceDate + 1 day.
func BuildShiftEndDatetime(attendanceDate time.Time, startHHMM, endHHMM string) time.Time {
	eh, em, ok := ParseShiftHHMM(endHHMM)
	if !ok {
		return time.Time{}
	}
	endOnSameDay := time.Date(
		attendanceDate.Year(), attendanceDate.Month(), attendanceDate.Day(),
		eh, em, 0, 0,
		attendanceDate.Location(),
	)
	if IsOvernightShift(startHHMM, endHHMM) {
		return endOnSameDay.AddDate(0, 0, 1)
	}
	return endOnSameDay
}

// CalculateOvertime returns the number of whole OT hours an employee earns.
//
// Rules:
//   - If overtimeEnabled is false → always 0.
//   - If checkOut is not after shiftEnd → 0.
//   - OT minutes = checkOut - shiftEnd (minutes).
//   - If OT minutes < 45 → 0.
//   - Otherwise: 1 + floor((otMinutes - 45) / 60).
//
// Examples:
//
//	30 min  →  0
//	44 min  →  0
//	45 min  →  1
//	60 min  →  1
//	104 min →  1
//	105 min →  2
//	120 min →  2
//	165 min →  3
func CalculateOvertime(checkOut, shiftEnd time.Time, overtimeEnabled bool) int {
	if !overtimeEnabled {
		return 0
	}
	if !checkOut.After(shiftEnd) {
		return 0
	}
	otMinutes := int(checkOut.Sub(shiftEnd).Minutes())
	if otMinutes < 45 {
		return 0
	}
	hours := 1 + (otMinutes-45)/60
	if hours == 7 {
		hours = 8
	}
	return hours
}

// CalcTotalHoursStr computes net worked duration as an "HH:MM" string,
// deducting lunch break duration overlapping with 13:00 to 14:00.
func CalcTotalHoursStr(checkIn, checkOut *time.Time) *string {
	if checkIn == nil || checkOut == nil {
		return nil
	}
	if checkOut.Before(*checkIn) || checkOut.Equal(*checkIn) {
		return nil
	}

	netHours := CalcNetWorkHours(*checkIn, *checkOut)
	if netHours <= 0 {
		return nil
	}
	h := int(netHours)
	m := int((netHours - float64(h)) * 60 + 0.5)
	if m >= 60 {
		h += 1
		m = 0
	}
	s := formatHHMM(h, m)
	return &s
}

// CalcNetWorkHours calculates actual working hours deducting fixed lunch break (13:00 - 14:00).
func CalcNetWorkHours(checkIn, checkOut time.Time) float64 {
	if checkOut.Before(checkIn) || checkOut.Equal(checkIn) {
		return 0
	}
	totalSpan := checkOut.Sub(checkIn).Hours()

	lunchStart := time.Date(checkIn.Year(), checkIn.Month(), checkIn.Day(), 13, 0, 0, 0, checkIn.Location())
	lunchEnd := time.Date(checkIn.Year(), checkIn.Month(), checkIn.Day(), 14, 0, 0, 0, checkIn.Location())

	overlapStart := checkIn
	if overlapStart.Before(lunchStart) {
		overlapStart = lunchStart
	}
	overlapEnd := checkOut
	if overlapEnd.After(lunchEnd) {
		overlapEnd = lunchEnd
	}

	lunchOverlap := 0.0
	if overlapStart.Before(overlapEnd) {
		lunchOverlap = overlapEnd.Sub(overlapStart).Hours()
	}

	netHours := totalSpan - lunchOverlap
	if netHours < 0 {
		netHours = 0
	}
	return netHours
}

func formatHHMM(h, m int) string {
	buf := make([]byte, 5)
	buf[0] = byte('0' + h/10)
	buf[1] = byte('0' + h%10)
	buf[2] = ':'
	buf[3] = byte('0' + m/10)
	buf[4] = byte('0' + m%10)
	return string(buf)
}

// ParseHHMMToMinutes converts "HH:mm" to total minutes.
func ParseHHMMToMinutes(hhmm string) (int, bool) {
	h, m, ok := ParseShiftHHMM(hhmm)
	if !ok {
		return 0, false
	}
	return h*60 + m, true
}
