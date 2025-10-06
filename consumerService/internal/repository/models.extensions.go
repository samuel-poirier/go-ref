package repository

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/google/uuid"
)

func (o Outbox) GetID() uuid.UUID {
	return o.ID
}

func (o Outbox) GetPayload() []byte {
	return o.Payload
}

func (o Outbox) GetMetadata() []byte {
	return o.Metadata
}

func NewCreateOutboxMessageParams[T any](payload T) (CreateOutboxMessageParams, error) {
	queueName := reflect.TypeOf(payload).Name()
	payloadJson, err := json.Marshal(payload)

	metadata := make(map[string]string)
	metadata["queueName"] = queueName
	metadataJson, _ := json.Marshal(metadata)

	return CreateOutboxMessageParams{
		ID:          uuid.New(),
		Payload:     payloadJson,
		ScheduledAt: time.Now(),
		Metadata:    metadataJson,
	}, err
}
