package domain

import (
	"context"

	"github.com/samuel-poirier/go-pubsub-demo/events"
)

type Publisher interface {
	Initialize(context.Context) error
	Publish(events.Message) error
	Close()
}

type BackgroundWorker interface {
	Start(context.Context) error
}
