// Package outbox provides GORM database-backed OutboxStore implementation for gorp framework.
package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	"gorm.io/gorm"
)

// GormOutboxModel is the GORM DB table model for outbox_messages.
type GormOutboxModel struct {
	ID         string     `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Topic      string     `gorm:"type:varchar(191);index" json:"topic"`
	Payload    string     `gorm:"type:text" json:"payload"`
	Status     string     `gorm:"type:varchar(32);index" json:"status"`
	RetryCount int        `gorm:"default:0" json:"retry_count"`
	Error      string     `gorm:"type:text" json:"error"`
	CreatedAt  time.Time  `gorm:"index" json:"created_at"`
	SentAt     *time.Time `json:"sent_at,omitempty"`
}

// TableName specifies table name for GORM.
func (GormOutboxModel) TableName() string {
	return "outbox_messages"
}

// GormOutboxStore implements OutboxStore contract using GORM.
type GormOutboxStore struct {
	db *gorm.DB
}

// NewGormOutboxStore creates a new GormOutboxStore.
func NewGormOutboxStore(db *gorm.DB) *GormOutboxStore {
	return &GormOutboxStore{db: db}
}

// AutoMigrate migrates outbox_messages table schema.
func (s *GormOutboxStore) AutoMigrate() error {
	if s.db == nil {
		return fmt.Errorf("outbox gorm db is nil")
	}
	return s.db.AutoMigrate(&GormOutboxModel{})
}

// Save persists an outbox message inside a local DB transaction or standard handle.
func (s *GormOutboxStore) Save(ctx context.Context, msg *integrationcontract.OutboxMessage) error {
	if s.db == nil {
		return fmt.Errorf("outbox gorm db is nil")
	}
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		return fmt.Errorf("marshal outbox payload: %w", err)
	}

	model := GormOutboxModel{
		ID:         msg.ID,
		Topic:      msg.Topic,
		Payload:    string(payloadBytes),
		Status:     string(msg.Status),
		RetryCount: msg.RetryCount,
		Error:      msg.Error,
		CreatedAt:  msg.CreatedAt,
		SentAt:     msg.SentAt,
	}

	dbHandle := s.db.WithContext(ctx)
	return dbHandle.Create(&model).Error
}

// GetPending fetches pending or retrying messages up to limit.
func (s *GormOutboxStore) GetPending(ctx context.Context, limit int) ([]*integrationcontract.OutboxMessage, error) {
	if s.db == nil {
		return nil, fmt.Errorf("outbox gorm db is nil")
	}
	var models []GormOutboxModel
	err := s.db.WithContext(ctx).
		Where("status IN ?", []string{string(integrationcontract.OutboxStatusPending), string(integrationcontract.OutboxStatusRetrying)}).
		Order("created_at ASC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	res := make([]*integrationcontract.OutboxMessage, 0, len(models))
	for _, m := range models {
		var payload any
		_ = json.Unmarshal([]byte(m.Payload), &payload)
		res = append(res, &integrationcontract.OutboxMessage{
			ID:         m.ID,
			Topic:      m.Topic,
			Payload:    payload,
			Status:     integrationcontract.OutboxStatus(m.Status),
			RetryCount: m.RetryCount,
			CreatedAt:  m.CreatedAt,
			SentAt:     m.SentAt,
			Error:      m.Error,
		})
	}
	return res, nil
}

// MarkSent marks a message status as sent.
func (s *GormOutboxStore) MarkSent(ctx context.Context, id string) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&GormOutboxModel{}).Where("id = ?", id).Updates(map[string]any{
		"status":  string(integrationcontract.OutboxStatusSent),
		"sent_at": &now,
	}).Error
}

// MarkFailed marks a message status as failed or increments retry count.
func (s *GormOutboxStore) MarkFailed(ctx context.Context, id string, err error) error {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	return s.db.WithContext(ctx).Model(&GormOutboxModel{}).Where("id = ?", id).Updates(map[string]any{
		"status":      string(integrationcontract.OutboxStatusFailed),
		"retry_count": gorm.Expr("retry_count + 1"),
		"error":       errMsg,
	}).Error
}
