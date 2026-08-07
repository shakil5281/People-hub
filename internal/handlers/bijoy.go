package handlers

import (
	"github.com/shakil5281/peoplehub-api/internal/utils"
)

func toBijoy(s string) string {
	return utils.UnicodeToBijoy(s)
}

func bijoyText(s, lang string) string {
	if lang != "bn" {
		return s
	}
	return utils.UnicodeToBijoy(s)
}
