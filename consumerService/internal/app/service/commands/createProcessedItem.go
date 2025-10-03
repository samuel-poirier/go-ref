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

	tx, err := c.db.Begin(ctx)

	if err != nil {
		return err
	}

	defer func() {
		if retErr != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	return c.repo.WithTx(tx).CreateProcessedItem(ctx, cmd.Data)
}
