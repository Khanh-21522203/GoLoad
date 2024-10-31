package logic

import (
	"GoLoad/internal/dataaccess/cache"
	"GoLoad/internal/dataaccess/database"
	"GoLoad/internal/generated/grpc/go_load"
	"context"
	"errors"
	"log"

	"github.com/doug-martin/goqu/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
type CreateSessionOutput struct {
	Account *go_load.Account
	Token   string
}

type Account interface {
	CreateAccount(ctx context.Context, params CreateAccountParams) (CreateAccountOutput, error)
	CreateSession(ctx context.Context, params CreateSessionParams) (CreateSessionOutput, error)
}
type account struct {
	goquDatabase                *goqu.Database
	takenAccountNameCache       cache.TakenAccountName
	accountDataAccessor         database.AccountDataAccessor
	accountPasswordDataAccessor database.AccountPasswordDataAccessor
	hashLogic                   Hash
	tokenLogic                  Token
}

func NewAccount(goquDatabase *goqu.Database, takenAccountNameCache cache.TakenAccountName, accountDataAccessor database.AccountDataAccessor,
	accountPasswordDataAccessor database.AccountPasswordDataAccessor, hashLogic Hash, tokenLogic Token) Account {
	return &account{
		goquDatabase:                goquDatabase,
		takenAccountNameCache:       takenAccountNameCache,
		accountDataAccessor:         accountDataAccessor,
		accountPasswordDataAccessor: accountPasswordDataAccessor,
		hashLogic:                   hashLogic,
		tokenLogic:                  tokenLogic,
	}
}
func (a account) databaseAccountToProtoAccount(account database.Account) *go_load.Account {
	return &go_load.Account{
		Id:          account.ID,
		AccountName: account.AccountName,
	}
}

func (a *account) isAccountAccountNameTaken(ctx context.Context, accountName string) (bool, error) {
	accountNameTaken, err := a.takenAccountNameCache.Has(ctx, accountName)
	if err != nil {
		log.Printf("failed to get account name from taken set in cache, will fall back to database")
	} else if accountNameTaken {
		return true, nil
	}
	if _, err := a.accountDataAccessor.GetAccountByAccountName(ctx, accountName); err != nil {
		if errors.Is(err, database.ErrAccountNotFound) {
			return false, nil
		}
		return false, err
	}
	if err := a.takenAccountNameCache.Add(ctx, accountName); err != nil {
		log.Printf("failed to set account name into taken set in cache")
	}
	return true, nil
}

func (a *account) CreateAccount(ctx context.Context, params CreateAccountParams) (CreateAccountOutput, error) {
	accountNameTaken, err := a.isAccountAccountNameTaken(ctx, params.AccountName)
	if err != nil {
		return CreateAccountOutput{}, status.Error(codes.Internal, "failed to check if account name is taken")
	}
	if accountNameTaken {
		return CreateAccountOutput{}, status.Error(codes.AlreadyExists, "account name is already taken")
	}
	var accountID uint64
	txErr := a.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		accountID, err = a.accountDataAccessor.WithDatabase(td).CreateAccount(ctx, database.Account{
			AccountName: params.AccountName,
		})
		if err != nil {
			return err
		}
		hashedPassword, hashErr := a.hashLogic.Hash(ctx, params.Password)
		if hashErr != nil {
			return hashErr
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
func (a *account) CreateSession(ctx context.Context, params CreateSessionParams) (CreateSessionOutput, error) {
	existingAccount, err := a.accountDataAccessor.GetAccountByAccountName(ctx, params.AccountName)
	if err != nil {
		return CreateSessionOutput{}, err
	}
	existingAccountPassword, err := a.accountPasswordDataAccessor.GetAccountPassword(ctx, existingAccount.ID)
	if err != nil {
		return CreateSessionOutput{}, err
	}
	isHashEqual, err := a.hashLogic.IsHashEqual(ctx, params.Password, existingAccountPassword.Hash)
	if err != nil {
		return CreateSessionOutput{}, err
	}
	if !isHashEqual {
		return CreateSessionOutput{}, status.Error(codes.Unauthenticated, "incorrect password")
	}
	token, _, err := a.tokenLogic.GetToken(ctx, existingAccount.ID)
	if err != nil {
		return CreateSessionOutput{}, err
	}
	return CreateSessionOutput{
		Account: a.databaseAccountToProtoAccount(existingAccount),
		Token:   token,
	}, nil
}
