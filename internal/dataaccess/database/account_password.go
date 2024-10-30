package database

import (
	"context"
	"database/sql"
	"log"

	"github.com/doug-martin/goqu/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	TabNameAccountPasswords            = "account_passwords"
	ColNameAccountPasswordsOfAccountID = "of_account_id"
	ColNameAccountPasswordsHash        = "hash"
)

type AccountPassword struct {
	OfAccountID uint64 `sql:"of_account_id"`
	Hash        string `sql:"hash"`
}
type AccountPasswordDataAccessor interface {
	CreateAccountPassword(ctx context.Context, accountPassword AccountPassword) error
	GetAccountPassword(ctx context.Context, ofAccountID uint64) (AccountPassword, error)
	UpdateAccountPassword(ctx context.Context, accountPassword AccountPassword) error
	WithDatabase(database Database) AccountPasswordDataAccessor
}
type accountPasswordDataAccessor struct {
	database Database
}

func NewAccountPasswordDataAccessor(database *goqu.Database) AccountPasswordDataAccessor {
	return &accountPasswordDataAccessor{
		database: database,
	}
}

// CreateAccountPassword implements AccountPasswordDataAccessor.
func (a *accountPasswordDataAccessor) CreateAccountPassword(ctx context.Context, accountPassword AccountPassword) error {
	_, err := a.database.
		Insert(TabNameAccountPasswords).
		Rows(goqu.Record{
			ColNameAccountPasswordsOfAccountID: accountPassword.OfAccountID,
			ColNameAccountPasswordsHash:        accountPassword.Hash,
		}).
		Executor().
		ExecContext(ctx)
	if err != nil {
		log.Printf("failed to create account password, err=%+v\n", err)
		return status.Errorf(codes.Internal, "failed to create account password: %+v", err)
	}
	return nil
}

func (a accountPasswordDataAccessor) GetAccountPassword(ctx context.Context, ofAccountID uint64) (AccountPassword, error) {
	accountPassword := AccountPassword{}
	found, err := a.database.
		From(TabNameAccounts).
		Where(goqu.Ex{ColNameAccountPasswordsOfAccountID: ofAccountID}).
		ScanStructContext(ctx, &accountPassword)
	if err != nil {
		log.Printf("failed to get account password by id, err=%+v\n", err)
		return AccountPassword{}, status.Errorf(codes.Internal, "failed to get account password by id: %+v", err)
	}
	if !found {
		log.Printf("cannot find account by id, err=%+v\n", err)
		return AccountPassword{}, sql.ErrNoRows
	}
	return accountPassword, nil
}

// UpdateAccountPassword implements AccountPasswordDataAccessor.
func (a *accountPasswordDataAccessor) UpdateAccountPassword(ctx context.Context, accountPassword AccountPassword) error {
	_, err := a.database.
		Update(TabNameAccountPasswords).
		Set(goqu.Record{ColNameAccountPasswordsHash: accountPassword.Hash}).
		Where(goqu.Ex{ColNameAccountPasswordsOfAccountID: accountPassword.OfAccountID}).
		Executor().
		ExecContext(ctx)
	if err != nil {
		log.Printf("failed to update account password, err=%+v\n", err)
		return status.Errorf(codes.Internal, "failed to update account password: %+v", err)
	}
	return nil
}
func (a accountPasswordDataAccessor) WithDatabase(database Database) AccountPasswordDataAccessor {
	return &accountPasswordDataAccessor{
		database: database,
	}
}
