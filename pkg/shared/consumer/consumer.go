package consumer

import "context"

type Consumer interface {
	Subscribe(queueName string, msgChan *chan<- Message, ctx context.Context) error
}

type ConsumerHandler interface {
	GetQueueName() string
	Handle(message Message)
}

type Message struct {
	Data        []byte
	Redelivered bool
	Headers     map[string]interface{}
	Ack         func() error
	Nack        func(requeue bool) error
	Context     context.Context // Trace context for log correlation
}
