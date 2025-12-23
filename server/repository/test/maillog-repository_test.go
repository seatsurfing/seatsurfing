package test

import (
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestMailLogCreate(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	mailLog := &MailLog{
		Timestamp:      time.Now(),
		Subject:        "Test Email",
		Recipient:      "user@test.com",
		OrganizationID: org.ID,
	}

	err := GetMailLogRepository().Create(mailLog)
	CheckTestBool(t, true, err == nil)
	CheckTestBool(t, true, mailLog.ID != "")
}

func TestMailLogCreateWithoutOrganization(t *testing.T) {
	ClearTestDB()

	mailLog := &MailLog{
		Timestamp:      time.Now(),
		Subject:        "Test Email",
		Recipient:      "user@test.com",
		OrganizationID: "",
	}

	err := GetMailLogRepository().Create(mailLog)
	CheckTestBool(t, true, err == nil)
	CheckTestBool(t, true, mailLog.ID != "")
}

func TestMailLogLogEmail(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	err := GetMailLogRepository().LogEmail("Welcome Email", "user@test.com", org.ID)
	CheckTestBool(t, true, err == nil)
}

func TestMailLogGetCountByDate(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	// Create logs for today
	today := time.Now()
	GetMailLogRepository().LogEmail("Email 1", "user1@test.com", org.ID)
	GetMailLogRepository().LogEmail("Email 2", "user2@test.com", org.ID)
	GetMailLogRepository().LogEmail("Email 3", "user3@test.com", org.ID)

	// Create a log for yesterday
	yesterday := today.AddDate(0, 0, -1)
	mailLog := &MailLog{
		Timestamp:      yesterday,
		Subject:        "Yesterday Email",
		Recipient:      "user4@test.com",
		OrganizationID: org.ID,
	}
	GetMailLogRepository().Create(mailLog)

	// Check today's count
	count, err := GetMailLogRepository().GetCountByDate(today)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 3, count)

	// Check yesterday's count
	count, err = GetMailLogRepository().GetCountByDate(yesterday)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, count)

	// Check tomorrow's count (should be 0)
	tomorrow := today.AddDate(0, 0, 1)
	count, err = GetMailLogRepository().GetCountByDate(tomorrow)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 0, count)
}

func TestMailLogGetCountBySubjectAndDate(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	// Create logs for today with different subjects
	today := time.Now()
	GetMailLogRepository().LogEmail("Welcome Email", "user1@test.com", org.ID)
	GetMailLogRepository().LogEmail("Welcome Email", "user2@test.com", org.ID)
	GetMailLogRepository().LogEmail("Welcome Email", "user3@test.com", org.ID)
	GetMailLogRepository().LogEmail("Password Reset", "user4@test.com", org.ID)
	GetMailLogRepository().LogEmail("Password Reset", "user5@test.com", org.ID)
	GetMailLogRepository().LogEmail("Booking Confirmation", "user6@test.com", org.ID)

	// Create a log for yesterday
	yesterday := today.AddDate(0, 0, -1)
	mailLog := &MailLog{
		Timestamp:      yesterday,
		Subject:        "Welcome Email",
		Recipient:      "user7@test.com",
		OrganizationID: org.ID,
	}
	GetMailLogRepository().Create(mailLog)

	// Get counts grouped by subject for today
	results, err := GetMailLogRepository().GetCountBySubjectAndDate(today)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 3, len(results))

	// Results should be ordered by count DESC, subject ASC
	// Welcome Email: 3, Password Reset: 2, Booking Confirmation: 1
	CheckTestString(t, "Welcome Email", results[0].Subject)
	CheckTestInt(t, 3, results[0].Count)
	CheckTestString(t, "Password Reset", results[1].Subject)
	CheckTestInt(t, 2, results[1].Count)
	CheckTestString(t, "Booking Confirmation", results[2].Subject)
	CheckTestInt(t, 1, results[2].Count)
}

func TestMailLogGetCountByOrganizationAndDate(t *testing.T) {
	ClearTestDB()
	org1 := CreateTestOrg("test1.com")
	org2 := CreateTestOrg("test2.com")

	// Create logs for today with different organizations
	today := time.Now()
	GetMailLogRepository().LogEmail("Email 1", "user1@test.com", org1.ID)
	GetMailLogRepository().LogEmail("Email 2", "user2@test.com", org1.ID)
	GetMailLogRepository().LogEmail("Email 3", "user3@test.com", org1.ID)
	GetMailLogRepository().LogEmail("Email 4", "user4@test.com", org2.ID)
	GetMailLogRepository().LogEmail("Email 5", "user5@test.com", org2.ID)

	// Create a log without organization (should be excluded)
	GetMailLogRepository().LogEmail("Email 6", "user6@test.com", "")

	// Create a log for yesterday
	yesterday := today.AddDate(0, 0, -1)
	mailLog := &MailLog{
		Timestamp:      yesterday,
		Subject:        "Yesterday Email",
		Recipient:      "user7@test.com",
		OrganizationID: org1.ID,
	}
	GetMailLogRepository().Create(mailLog)

	// Get counts grouped by organization for today
	results, err := GetMailLogRepository().GetCountByOrganizationAndDate(today)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 2, len(results))

	// Results should be ordered by count DESC, organization_id ASC
	// org1: 3, org2: 2
	CheckTestString(t, org1.ID, results[0].OrganizationID)
	CheckTestInt(t, 3, results[0].Count)
	CheckTestString(t, org2.ID, results[1].OrganizationID)
	CheckTestInt(t, 2, results[1].Count)

	// Get counts for yesterday
	results, err = GetMailLogRepository().GetCountByOrganizationAndDate(yesterday)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(results))
	CheckTestString(t, org1.ID, results[0].OrganizationID)
	CheckTestInt(t, 1, results[0].Count)
}

func TestMailLogEmptyResults(t *testing.T) {
	ClearTestDB()

	today := time.Now()

	// Test GetCountByDate with no logs
	count, err := GetMailLogRepository().GetCountByDate(today)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 0, count)

	// Test GetCountBySubjectAndDate with no logs
	results, err := GetMailLogRepository().GetCountBySubjectAndDate(today)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 0, len(results))

	// Test GetCountByOrganizationAndDate with no logs
	results2, err := GetMailLogRepository().GetCountByOrganizationAndDate(today)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 0, len(results2))
}
