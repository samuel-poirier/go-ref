package consumer

import "context"

type Consumer interface {
	Subscribe(queueName string, msgChan *chan<- Message, ctx context.Context) error
}

type ConsumerHandler interface {
	Handle(message Message)
}

type Message struct {
	Data        []byte
	Redelivered bool
	Ack         func() error
	Nack        func(requeue bool) error
}
