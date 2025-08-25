package app

import "context"

type Publisher interface {
	StartPublishing(context.Context) error
}
