package domain

import (
	"context"

	"github.com/sam9291/go-pubsub-demo/events"
)

type Publisher interface {
	Initialize(context.Context) error
	Publish(events.Message) error
}

type BackgroundWorker interface {
	Start(context.Context) error
}
