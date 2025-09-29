package app

import (
	"context"

	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/app/service"
)

type Consumer interface {
	StartConsuming(ctx context.Context, service *service.Service) error
}
