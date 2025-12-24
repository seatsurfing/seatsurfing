package repository

import (
	"sync"
	"time"
)

type MailLogRepository struct {
}

type MailLog struct {
	ID             string
	Timestamp      time.Time
	Subject        string
	Recipient      string
	OrganizationID string
}

type MailLogBySubject struct {
	Subject string
	Count   int
}

type MailLogByOrganization struct {
	OrganizationID string
	Count          int
}

var mailLogRepository *MailLogRepository
var mailLogRepositoryOnce sync.Once

func GetMailLogRepository() *MailLogRepository {
	mailLogRepositoryOnce.Do(func() {
		mailLogRepository = &MailLogRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS mail_logs (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"timestamp TIMESTAMP NOT NULL, " +
			"subject VARCHAR NOT NULL, " +
			"recipient VARCHAR NOT NULL, " +
			"organization_id uuid NULL, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_mail_logs_timestamp ON mail_logs(timestamp)")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_mail_logs_recipient ON mail_logs(recipient)")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_mail_logs_organization_id ON mail_logs(organization_id)")
		if err != nil {
			panic(err)
		}
	})
	return mailLogRepository
}

func (r *MailLogRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	// No updates yet
}

func (r *MailLogRepository) Create(e *MailLog) error {
	var id string
	var orgID interface{}
	if e.OrganizationID == "" {
		orgID = nil
	} else {
		orgID = e.OrganizationID
	}
	err := GetDatabase().DB().QueryRow("INSERT INTO mail_logs "+
		"(timestamp, subject, recipient, organization_id) "+
		"VALUES ($1, $2, $3, $4) "+
		"RETURNING id",
		e.Timestamp, e.Subject, e.Recipient, orgID).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *MailLogRepository) LogEmail(subject, recipient, organizationID string) error {
	e := &MailLog{
		Timestamp:      time.Now(),
		Subject:        subject,
		Recipient:      recipient,
		OrganizationID: organizationID,
	}
	return r.Create(e)
}

func (r *MailLogRepository) GetCountByDate(date time.Time) (int, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var count int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(id) FROM mail_logs "+
		"WHERE timestamp >= $1 AND timestamp < $2",
		startOfDay, endOfDay).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *MailLogRepository) GetCountBySubjectAndDate(date time.Time) ([]*MailLogBySubject, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	rows, err := GetDatabase().DB().Query("SELECT subject, COUNT(id) as count FROM mail_logs "+
		"WHERE timestamp >= $1 AND timestamp < $2 "+
		"GROUP BY subject "+
		"ORDER BY count DESC, subject ASC",
		startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []*MailLogBySubject{}
	for rows.Next() {
		e := &MailLogBySubject{}
		if err := rows.Scan(&e.Subject, &e.Count); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *MailLogRepository) GetCountByOrganizationAndDate(date time.Time) ([]*MailLogByOrganization, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	rows, err := GetDatabase().DB().Query("SELECT organization_id, COUNT(id) as count FROM mail_logs "+
		"WHERE timestamp >= $1 AND timestamp < $2 AND organization_id IS NOT NULL "+
		"GROUP BY organization_id "+
		"ORDER BY count DESC, organization_id ASC",
		startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []*MailLogByOrganization{}
	for rows.Next() {
		e := &MailLogByOrganization{}
		if err := rows.Scan(&e.OrganizationID, &e.Count); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *MailLogRepository) AnonymizeAll(organizationID string) error {
	if _, err := GetDatabase().DB().Exec("UPDATE mail_logs SET recipient = '' "+
		"WHERE organization_id = $1", organizationID); err != nil {
		return err
	}
	return nil
}
