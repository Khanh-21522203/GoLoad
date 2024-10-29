package logic

import (
	"context"
	"database/sql"
	"errors"

	"GoLoad/internal/dataaccess/database"

	"github.com/doug-martin/goqu/v9"
)

type CreateAccountParams struct {
	AccountName string
	Password    string
}
type CreateAccountOutput struct {
	ID          uint64
	AccountName string
}
type CreateSessionParams struct {
	AccountName string
	Password    string
}
type Account interface {
	CreateAccount(ctx context.Context, params CreateAccountParams) (CreateAccountOutput, error)
	CreateSession(ctx context.Context, params CreateSessionParams) (account Account, token string, err error)
}
type account struct {
	goquDatabase                *goqu.Database
	accountDataAccessor         database.AccountDataAccessor
	accountPasswordDataAccessor database.AccountPasswordDataAccessor
	hashLogic                   Hash
}

func NewAccount(
	goquDatabase *goqu.Database,
	accountDataAccessor database.AccountDataAccessor,
	accountPasswordDataAccessor database.AccountPasswordDataAccessor,
	hashLogic Hash,
) Account {
	return &account{
		goquDatabase:                goquDatabase,
		accountDataAccessor:         accountDataAccessor,
		accountPasswordDataAccessor: accountPasswordDataAccessor,
		hashLogic:                   hashLogic,
	}
}
func (a account) isAccountAccountNameTaken(ctx context.Context, accountName string) (bool, error) {
	if _, err := a.accountDataAccessor.GetAccountByAccountName(ctx, accountName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
func (a account) CreateAccount(ctx context.Context, params CreateAccountParams) (CreateAccountOutput, error) {
	var accountID uint64
	txErr := a.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		accountNameTaken, err := a.isAccountAccountNameTaken(ctx, params.AccountName)
		if err != nil {
			return err
		}
		if accountNameTaken {
			return errors.New("account name is already taken")
		}
		accountID, err = a.accountDataAccessor.WithDatabase(td).CreateAccount(ctx, database.Account{
			AccountName: params.AccountName,
		})
		if err != nil {
			return err
		}
		hashedPassword, err := a.hashLogic.Hash(ctx, params.Password)
		if err != nil {
			return err
		}
		if err := a.accountPasswordDataAccessor.WithDatabase(td).CreateAccountPassword(ctx, database.AccountPassword{
			OfAccountID: accountID,
			Hash:        hashedPassword,
		}); err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		return CreateAccountOutput{}, txErr
	}
	return CreateAccountOutput{
		ID:          accountID,
		AccountName: params.AccountName,
	}, nil
}
func (a account) CreateSession(ctx context.Context, params CreateSessionParams) (account Account, token string, err error) {
	return
}
