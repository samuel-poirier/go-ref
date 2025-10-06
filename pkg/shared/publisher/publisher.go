package publisher

import (
	"context"
	"encoding/json"
	"reflect"
)

type Publisher interface {
	Initialize(context.Context) error
	Publish(message MessageEnvelope) error
	Close()
}

type MessageEnvelope struct {
	QueueName string
	Metadata  []byte
	Message   []byte
}

func NewMessageEnvelope[T any](message T) (MessageEnvelope, error) {
	queueName := reflect.TypeOf(message).Name()

	metadata := make(map[string]string)
	metadata["queueName"] = queueName

	metadataJson, _ := json.Marshal(metadata)

	payloadJson, err := json.Marshal(message)

	return MessageEnvelope{
		Message:   payloadJson,
		QueueName: queueName,
		Metadata:  metadataJson,
	}, err
}
