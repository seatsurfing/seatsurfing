package repository

import (
	"sync"
	"time"
)

type DebugTimeIssuesRepository struct {
}

type DebugTimeIssueItem struct {
	ID      string
	Created time.Time
}

var debugTimeIssuesRepository *DebugTimeIssuesRepository
var debugTimeIssuesRepositoryOnce sync.Once

func GetDebugTimeIssuesRepository() *DebugTimeIssuesRepository {
	debugTimeIssuesRepositoryOnce.Do(func() {
		debugTimeIssuesRepository = &DebugTimeIssuesRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS debug_time_issues (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"created TIMESTAMP NOT NULL, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
	})
	return debugTimeIssuesRepository
}

func (r *DebugTimeIssuesRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	// No updates yet
}

func (r *DebugTimeIssuesRepository) Create(e *DebugTimeIssueItem) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO debug_time_issues "+
		"(created) "+
		"VALUES ($1) "+
		"RETURNING id",
		e.Created).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *DebugTimeIssuesRepository) GetOne(id string) (*DebugTimeIssueItem, error) {
	e := &DebugTimeIssueItem{}
	err := GetDatabase().DB().QueryRow("SELECT id, created "+
		"FROM debug_time_issues "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.Created)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *DebugTimeIssuesRepository) Delete(e *DebugTimeIssueItem) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM debug_time_issues WHERE id = $1", e.ID)
	return err
}
