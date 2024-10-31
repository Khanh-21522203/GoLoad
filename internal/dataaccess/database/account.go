package database

import (
	"context"
	"log"

	"github.com/doug-martin/goqu/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	TabNameAccounts    = goqu.T("accounts")
	ErrAccountNotFound = status.Error(codes.NotFound, "account not found")
)

const (
	ColNameAccountsID          = "id"
	ColNameAccountsAccountName = "account_name"
)

type Account struct {
	ID          uint64 `db:"id" goqu:"skipinsert,skipupdate"`
	AccountName string `db:"account_name"`
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
		return 0, status.Error(codes.Internal, "failed to create account")
	}
	lastInsertedID, err := result.LastInsertId()
	if err != nil {
		log.Printf("failed to get last inserted id, err=%+v\n", err)
		return 0, status.Error(codes.Internal, "failed to get last inserted id")
	}
	return uint64(lastInsertedID), nil
}

// GetAccountByID implements AccountDataAccessor.
func (a *accountDataAccessor) GetAccountByID(ctx context.Context, id uint64) (Account, error) {
	account := Account{}
	found, err := a.database.
		From(TabNameAccounts).
		Where(goqu.C(ColNameAccountsID).Eq(id)).
		ScanStructContext(ctx, &account)
	if err != nil {
		log.Printf("failed to get account by id, err=%+v\n", err)
		return Account{}, status.Error(codes.Internal, "failed to get account by id")
	}
	if !found {
		log.Printf("cannot find account by id, err=%+v\n", err)
		return Account{}, ErrAccountNotFound
	}
	return account, nil
}

// GetAccountByAccountName implements AccountDataAccessor.
func (a *accountDataAccessor) GetAccountByAccountName(ctx context.Context, accountName string) (Account, error) {
	account := Account{}
	found, err := a.database.
		From(TabNameAccounts).
		Where(goqu.C(ColNameAccountsAccountName).Eq(accountName)).
		ScanStructContext(ctx, &account)
	if err != nil {
		log.Printf("failed to get account by account name, err=%+v\n", err)
		return Account{}, status.Error(codes.Internal, "failed to get account by name")
	}
	if !found {
		log.Printf("cannot find account by account name, err=%+v\n", err)
		return Account{}, ErrAccountNotFound
	}
	return account, nil
}

// WithDatabase implements AccountDataAccessor.
func (a *accountDataAccessor) WithDatabase(database Database) AccountDataAccessor {
	return &accountDataAccessor{
		database: database,
	}
}
