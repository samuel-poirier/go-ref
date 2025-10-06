package domain

import (
	"context"
)

type BackgroundWorker interface {
	Start(context.Context) error
}
