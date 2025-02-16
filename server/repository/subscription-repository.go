package repository

import (
	"sync"
	"time"
)

type SubscriptionRepository struct {
}

type SubscriptionEventType string

const (
	SubscriptionEventActivate   SubscriptionEventType = "activate"
	SubscriptionEventDeactivate SubscriptionEventType = "deactivate"
	SubscriptionEventUpdate     SubscriptionEventType = "update"
)

type SubscriptionEvent struct {
	ID                   string
	OrganizationID       string
	EventType            SubscriptionEventType
	EventTime            time.Time
	ActivationTime       time.Time
	MaxUsers             int
	Price                float32
	BrokerSubscriptionID string
	BrokerCustomerID     string
	BrokerEventID        string
	Processed            bool
}

var subscriptionRepository *SubscriptionRepository
var subscriptionRepositoryOnce sync.Once

func GetSubscriptionRepository() *SubscriptionRepository {
	subscriptionRepositoryOnce.Do(func() {
		subscriptionRepository = &SubscriptionRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS subscription_events (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"organization_id uuid NOT NULL, " +
			"event_type VARCHAR NOT NULL, " +
			"event_time TIMESTAMP NOT NULL, " +
			"activation_time TIMESTAMP NOT NULL, " +
			"max_users INT, " +
			"price NUMERIC(8, 2), " +
			"broker_subscription_id VARCHAR NOT NULL, " +
			"broker_customer_id VARCHAR NOT NULL, " +
			"broker_event_id VARCHAR NOT NULL, " +
			"processed BOOLEAN, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_subscription_events_organization_id ON subscription_events(organization_id)")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_subscription_events_broker_event_id ON subscription_events(broker_event_id)")
		if err != nil {
			panic(err)
		}
	})
	return subscriptionRepository
}

func (r *SubscriptionRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	// No updates yet
}

func (r *SubscriptionRepository) Create(e *SubscriptionEvent) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO subscription_events "+
		"(organization_id, event_type, event_time, activation_time, max_users, price, broker_subscription_id, broker_customer_id, broker_event_id, processed) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) "+
		"RETURNING id",
		e.OrganizationID, e.EventType, e.EventTime, e.ActivationTime, e.MaxUsers, e.Price, e.BrokerSubscriptionID, e.BrokerCustomerID, e.BrokerEventID, e.Processed).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *SubscriptionRepository) GetLatest(organizationID string, maxResults int) ([]*SubscriptionEvent, error) {
	var result []*SubscriptionEvent
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, event_type, event_time, activation_time, max_users, price, broker_subscription_id, broker_customer_id, broker_event_id, processed "+
		"FROM subscription_events "+
		"WHERE organization_id = $1 "+
		"ORDER BY event_time DESC "+
		"LIMIT $2", organizationID, maxResults)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &SubscriptionEvent{}
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.EventType, &e.EventTime, &e.ActivationTime, &e.MaxUsers, &e.Price, &e.BrokerSubscriptionID, &e.BrokerCustomerID, &e.BrokerEventID, &e.Processed)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *SubscriptionRepository) GetProcessedByBrokerEventID(brokerEventID string) (*SubscriptionEvent, error) {
	e := &SubscriptionEvent{}
	err := GetDatabase().DB().QueryRow("SELECT id, organization_id, event_type, event_time, activation_time, max_users, price, broker_subscription_id, broker_customer_id, broker_event_id, processed "+
		"FROM subscription_events "+
		"WHERE processed = TRUE AND broker_event_id = $1",
		brokerEventID).Scan(&e.ID, &e.OrganizationID, &e.EventType, &e.EventTime, &e.ActivationTime, &e.MaxUsers, &e.Price, &e.BrokerSubscriptionID, &e.BrokerCustomerID, &e.BrokerEventID, &e.Processed)
	if err != nil {
		return nil, err
	}
	return e, nil
}
