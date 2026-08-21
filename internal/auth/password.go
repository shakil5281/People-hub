package auth

import (
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateOTP() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	h := hex.EncodeToString(b)
	if len(h) > 6 {
		h = h[:6]
	}
	return h
}

func GenerateBackupCodes() []string {
	codes := make([]string, 8)
	for i := range codes {
		b := make([]byte, 4)
		_, _ = rand.Read(b)
		codes[i] = hex.EncodeToString(b)
	}
	return codes
}
