package outbox

import (
	"time"

	"github.com/google/uuid"
)

type OutboxMessage struct {
	ID             uuid.UUID
	CreatedAt      time.Time
	ScheduledAt    time.Time
	Metadata       []byte
	Payload        []byte
	TimesAttempted int32
}

func NewOutboxMessage(metadata, payload []byte) OutboxMessage {
	now := time.Now()
	return OutboxMessage{
		ID:             uuid.New(),
		CreatedAt:      now,
		ScheduledAt:    now,
		Metadata:       metadata,
		Payload:        payload,
		TimesAttempted: 0,
	}
}
