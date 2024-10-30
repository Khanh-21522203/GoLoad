package cache

import (
	"context"
)

const (
	setKeyNameTakenAccountName = "taken_account_name_set"
)

type TakenAccountName interface {
	Add(ctx context.Context, accountName string) error
	Has(ctx context.Context, accountName string) (bool, error)
}
type takenAccountName struct {
	client Client
}

func NewTakenAccountName(client Client) TakenAccountName {
	return &takenAccountName{
		client: client,
	}
}
func (c takenAccountName) Add(ctx context.Context, accountName string) error {
	if err := c.client.AddToSet(ctx, setKeyNameTakenAccountName, accountName); err != nil {
		return err
	}
	return nil
}
func (c takenAccountName) Has(ctx context.Context, accountName string) (bool, error) {
	return c.client.IsDataInSet(ctx, setKeyNameTakenAccountName, accountName)
}
