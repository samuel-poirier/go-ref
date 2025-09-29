package commands

import (
	"context"

	"github.com/go-playground/validator/v10"
)

type CreateProcessedItemCommand struct {
	Data string `validate:"required"`
}

func (c commands) CreateProcessedItem(ctx context.Context, cmd CreateProcessedItemCommand) error {
	v := validator.New()
	err := v.Struct(cmd)

	if err != nil {
		return err
	}

	tx, err := c.db.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Commit(ctx)

	return c.repo.WithTx(tx).CreateProcessedItem(ctx, cmd.Data)
}
