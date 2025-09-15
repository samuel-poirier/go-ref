package app

import "context"

type Consumer interface {
	StartConsuming(ctx context.Context) error
}
