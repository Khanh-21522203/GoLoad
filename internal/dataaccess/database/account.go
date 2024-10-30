package database

import (
	"context"
	"database/sql"
	"log"

	"github.com/doug-martin/goqu/v9"
)

const (
	TabNameAccounts            = "accounts"
	ColNameAccountsID          = "id"
	ColNameAccountsAccountName = "account_name"
)

type Account struct {
	ID          uint64 `sql:"id"`
	AccountName string `sql:"account_name"`
}
type AccountDataAccessor interface {
	CreateAccount(ctx context.Context, account Account) (uint64, error)
	GetAccountByID(ctx context.Context, id uint64) (Account, error)
	GetAccountByAccountName(ctx context.Context, accountName string) (Account, error)
	WithDatabase(database Database) AccountDataAccessor
}
type accountDataAccessor struct {
	database Database
}

func NewAccountDataAccessor(database *goqu.Database) AccountDataAccessor {
	return &accountDataAccessor{
		database: database,
	}
}

// CreateAccount implements AccountDataAccessor.
func (a accountDataAccessor) CreateAccount(ctx context.Context, account Account) (uint64, error) {
	result, err := a.database.
		Insert("TabNameAccounts").
		Rows(goqu.Record{
			ColNameAccountsAccountName: account.AccountName,
		}).
		Executor().
		ExecContext(ctx)
	if err != nil {
		log.Printf("failed to create account, err=%+v\n", err)
		return 0, err
	}
	lastInsertedID, err := result.LastInsertId()
	if err != nil {
		log.Printf("failed to get last inserted id, err=%+v\n", err)
		return 0, err
	}
	return uint64(lastInsertedID), nil
}

// GetAccountByID implements AccountDataAccessor.
func (a *accountDataAccessor) GetAccountByID(ctx context.Context, id uint64) (Account, error) {
	account := Account{}
	found, err := a.database.
		From(TabNameAccounts).
		Where(goqu.Ex{ColNameAccountsID: id}).
		ScanStructContext(ctx, &account)
	if err != nil {
		log.Printf("failed to get account by id, err=%+v\n", err)
		return Account{}, err
	}
	if !found {
		log.Printf("cannot find account by id, err=%+v\n", err)
		return Account{}, sql.ErrNoRows
	}
	return account, nil
}

// GetAccountByAccountName implements AccountDataAccessor.
func (a *accountDataAccessor) GetAccountByAccountName(ctx context.Context, accountName string) (Account, error) {
	account := Account{}
	found, err := a.database.
		From(TabNameAccounts).
		Where(goqu.Ex{ColNameAccountsAccountName: accountName}).
		ScanStructContext(ctx, &account)
	if err != nil {
		log.Printf("failed to get account by account name, err=%+v\n", err)
		return Account{}, err
	}
	if !found {
		log.Printf("cannot find account by account name, err=%+v\n", err)
		return Account{}, sql.ErrNoRows
	}
	return account, nil
}

// WithDatabase implements AccountDataAccessor.
func (a *accountDataAccessor) WithDatabase(database Database) AccountDataAccessor {
	return &accountDataAccessor{
		database: database,
	}
}
