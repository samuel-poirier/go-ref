package domain

import (
	"context"

	"github.com/samuel-poirier/go-ref/events"
)

type Publisher interface {
	Initialize(context.Context) error
	Publish(events.Message) error
	Close()
}

type BackgroundWorker interface {
	Start(context.Context) error
}
