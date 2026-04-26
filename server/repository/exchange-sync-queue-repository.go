package repository

import (
	"sync"
	"time"
)

type ExchangeSyncQueueRepository struct {
}

type ExchangeSyncQueueItem struct {
	ID          string
	BookingID   string
	Operation   string
	Status      string
	RetryCount  int
	NextRetryAt time.Time
	LastError   string
	Payload     string
	CreatedAt   time.Time
}

var exchangeSyncQueueRepository *ExchangeSyncQueueRepository
var exchangeSyncQueueRepositoryOnce sync.Once

func GetExchangeSyncQueueRepository() *ExchangeSyncQueueRepository {
	exchangeSyncQueueRepositoryOnce.Do(func() {
		exchangeSyncQueueRepository = &ExchangeSyncQueueRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS exchange_sync_queue (" +
			"id UUID NOT NULL DEFAULT uuid_generate_v4(), " +
			"booking_id UUID NOT NULL, " +
			"operation VARCHAR NOT NULL, " +
			"status VARCHAR NOT NULL DEFAULT 'pending', " +
			"retry_count INTEGER NOT NULL DEFAULT 0, " +
			"next_retry_at TIMESTAMP NOT NULL DEFAULT NOW(), " +
			"last_error VARCHAR NOT NULL DEFAULT '', " +
			"payload TEXT NOT NULL DEFAULT '', " +
			"created_at TIMESTAMP NOT NULL DEFAULT NOW(), " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_exchange_sync_queue_status_retry " +
			"ON exchange_sync_queue (status, next_retry_at) " +
			"WHERE status = 'pending'")
		if err != nil {
			panic(err)
		}
	})
	return exchangeSyncQueueRepository
}

func (r *ExchangeSyncQueueRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
}

func (r *ExchangeSyncQueueRepository) Enqueue(bookingID, operation, payload string) error {
	_, err := GetDatabase().DB().Exec("INSERT INTO exchange_sync_queue "+
		"(booking_id, operation, status, retry_count, next_retry_at, payload) "+
		"VALUES ($1, $2, 'pending', 0, NOW(), $3)",
		bookingID, operation, payload)
	return err
}

// ClaimBatch atomically claims up to maxJobs pending jobs ready for processing.
func (r *ExchangeSyncQueueRepository) ClaimBatch(maxJobs int) ([]*ExchangeSyncQueueItem, error) {
	rows, err := GetDatabase().DB().Query(
		"UPDATE exchange_sync_queue SET status = 'processing' "+
			"WHERE id IN ("+
			"  SELECT id FROM exchange_sync_queue "+
			"  WHERE status = 'pending' AND next_retry_at <= NOW() "+
			"  ORDER BY next_retry_at "+
			"  LIMIT $1 "+
			"  FOR UPDATE SKIP LOCKED"+
			") RETURNING id, booking_id, operation, retry_count, last_error, payload, created_at",
		maxJobs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*ExchangeSyncQueueItem
	for rows.Next() {
		e := &ExchangeSyncQueueItem{Status: "processing"}
		if err := rows.Scan(&e.ID, &e.BookingID, &e.Operation, &e.RetryCount, &e.LastError, &e.Payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *ExchangeSyncQueueRepository) Delete(id string) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM exchange_sync_queue WHERE id = $1", id)
	return err
}

func (r *ExchangeSyncQueueRepository) MarkFailed(id, lastError string) error {
	_, err := GetDatabase().DB().Exec("UPDATE exchange_sync_queue SET status = 'failed', last_error = $2 WHERE id = $1", id, lastError)
	return err
}

func (r *ExchangeSyncQueueRepository) ScheduleRetry(id string, retryCount int, lastError string) error {
	_, err := GetDatabase().DB().Exec(
		"UPDATE exchange_sync_queue SET status = 'pending', retry_count = $2, last_error = $3, "+
			"next_retry_at = NOW() + (INTERVAL '1 minute' * power(2, ($2 - 1)::float)) "+
			"WHERE id = $1",
		id, retryCount, lastError)
	return err
}

func (r *ExchangeSyncQueueRepository) GetFailed(orgID string) ([]*ExchangeSyncQueueItem, error) {
	// Filter by org through payload JSON - simpler: just return all failed items
	// (per-org filtering done in router via payload parsing or we store orgID column)
	// Since orgID is in the payload, we use a simpler approach: list all failed
	// and filter in router. For now, return all failed for this org via payload match.
	rows, err := GetDatabase().DB().Query(
		"SELECT id, booking_id, operation, status, retry_count, next_retry_at, last_error, payload, created_at "+
			"FROM exchange_sync_queue "+
			"WHERE status = 'failed' AND payload::jsonb->>'orgID' = $1 "+
			"ORDER BY created_at DESC",
		orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*ExchangeSyncQueueItem
	for rows.Next() {
		e := &ExchangeSyncQueueItem{}
		if err := rows.Scan(&e.ID, &e.BookingID, &e.Operation, &e.Status, &e.RetryCount, &e.NextRetryAt, &e.LastError, &e.Payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *ExchangeSyncQueueRepository) ResetToPending(id string) error {
	_, err := GetDatabase().DB().Exec(
		"UPDATE exchange_sync_queue SET status = 'pending', retry_count = 0, next_retry_at = NOW() WHERE id = $1 AND status = 'failed'",
		id)
	return err
}
