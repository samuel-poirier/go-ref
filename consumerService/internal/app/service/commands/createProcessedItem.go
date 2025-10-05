package commands

import (
	"context"

	"github.com/go-playground/validator/v10"
)

type CreateProcessedItemCommand struct {
	Data string `validate:"required"`
}

func (c commands) CreateProcessedItem(ctx context.Context, cmd CreateProcessedItemCommand) (retErr error) {
	v := validator.New()
	err := v.Struct(cmd)

	if err != nil {
		return err
	}

	return c.repo.CreateProcessedItem(ctx, cmd.Data)
}
