package utils

import (
	"fmt"
	"math"
	"strings"
)

var ones = []string{"", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine", "Ten", "Eleven", "Twelve", "Thirteen", "Fourteen", "Fifteen", "Sixteen", "Seventeen", "Eighteen", "Nineteen"}
var tens = []string{"", "", "Twenty", "Thirty", "Forty", "Fifty", "Sixty", "Seventy", "Eighty", "Ninety"}

func convertLessThanThousand(n int) string {
	if n == 0 {
		return ""
	}
	var res string
	if n >= 100 {
		res += ones[n/100] + " Hundred "
		n %= 100
	}
	if n >= 20 {
		res += tens[n/10] + " "
		n %= 10
	}
	if n > 0 {
		res += ones[n] + " "
	}
	return strings.TrimSpace(res)
}

// NumberToWords converts a non-negative amount to words (South Asian numbering: Crore, Lakh, Thousand, Hundred).
func NumberToWords(amount float64) string {
	intVal := int(math.Round(amount))
	if intVal <= 0 {
		return "Zero"
	}

	crore := intVal / 10000000
	intVal %= 10000000

	lakh := intVal / 100000
	intVal %= 100000

	thousand := intVal / 1000
	intVal %= 1000

	hundreds := intVal

	var parts []string
	if crore > 0 {
		parts = append(parts, convertLessThanThousand(crore)+" Crore")
	}
	if lakh > 0 {
		parts = append(parts, convertLessThanThousand(lakh)+" Lakh")
	}
	if thousand > 0 {
		parts = append(parts, convertLessThanThousand(thousand)+" Thousand")
	}
	if hundreds > 0 {
		parts = append(parts, convertLessThanThousand(hundreds))
	}

	return strings.Join(parts, " ")
}

// FormatTakaInWords returns sentence case words formatted as:
// "In words taka: [Words] taka only." (e.g. "In words taka: Three hundred taka only.")
func FormatTakaInWords(amount float64) string {
	words := strings.TrimSpace(NumberToWords(amount))
	if words == "" {
		words = "Zero"
	}
	lower := strings.ToLower(words)
	sentenceCase := strings.ToUpper(lower[:1]) + lower[1:]
	return fmt.Sprintf("In words taka: %s taka only.", sentenceCase)
}
