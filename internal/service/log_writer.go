package service

import (
	"time"

	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
)

func WriteLog(level, source, message string) {
	if database.DB == nil {
		return
	}
	go database.DB.Create(&models.SystemLog{
		Level:     level,
		Source:    source,
		Message:   message,
		CreatedAt: time.Now(),
	})
}

func WriteErrorLog(source, message string, stackTrace ...string) {
	if database.DB == nil {
		return
	}
	log := models.SystemLog{
		Level:     "error",
		Source:    source,
		Message:   message,
		CreatedAt: time.Now(),
	}
	if len(stackTrace) > 0 {
		log.StackTrace = &stackTrace[0]
	}
	go database.DB.Create(&log)
}

func WriteRequestLog(level, source, message, ip, userAgent, path, method string, userID, companyID *string, statusCode int, latency int64) {
	if database.DB == nil {
		return
	}
	go database.DB.Create(&models.SystemLog{
		Level:      level,
		Source:     source,
		Message:    message,
		UserID:     userID,
		CompanyID:  companyID,
		IPAddress:  &ip,
		UserAgent:  &userAgent,
		Method:     &method,
		Path:       &path,
		StatusCode: &statusCode,
		Latency:    &latency,
		CreatedAt:  time.Now(),
	})
}
