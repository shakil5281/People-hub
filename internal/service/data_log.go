package service

import (
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
)

type DataLogService struct {
	dataLogRepo *repository.DataLogRepository
	mdbReader   *MDBReader
}

func NewDataLogService(dataLogRepo *repository.DataLogRepository, mdbReader *MDBReader) *DataLogService {
	return &DataLogService{
		dataLogRepo: dataLogRepo,
		mdbReader:   mdbReader,
	}
}

type ImportResult struct {
	Imported int
	Skipped  int
}

// ImportFromMDB reads punch records from a ZKTeco MDB file, filters out duplicates,
// and inserts new records into the data_logs table.
func (s *DataLogService) ImportFromMDB(filePath, startDate, endDate string) (*ImportResult, error) {
	if filePath == "" {
		filePath = s.mdbReader.DefaultPath
	}

	records, err := s.mdbReader.ReadPunches(filePath, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var logs []models.DataLog
	skipped := 0

	for _, rec := range records {
		punchTime := rec.PunchTimeParsed()
		if s.dataLogRepo.ExistsByBadgeAndPunchTime(rec.BadgeNumber, punchTime) {
			skipped++
			continue
		}
		log := models.DataLog{
			UserID:       rec.UserID,
			BadgeNumber:  rec.BadgeNumber,
			EmployeeName: rec.Name,
			PunchTime:    punchTime,
			PunchType:    rec.PunchType,
			DeviceID:     rec.DeviceID,
			DeviceSN:     rec.DeviceSN,
			Date:         punchTime.Format("2006-01-02"),
		}
		logs = append(logs, log)
	}

	if len(logs) == 0 {
		return &ImportResult{Imported: 0, Skipped: skipped}, nil
	}

	if err := s.dataLogRepo.BatchCreate(logs); err != nil {
		return nil, err
	}

	return &ImportResult{Imported: len(logs), Skipped: skipped}, nil
}
