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
	TabNameTokenPublicKeys          = "token_public_keys"
	ColNameTokenPublicKeysID        = "id"
	ColNameTokenPublicKeysPublicKey = "public_key"
)

type TokenPublicKey struct {
	ID        uint64 `db:"id" goqu:"skipinsert,skipupdate"`
	PublicKey []byte `db:"public_key"`
}
type TokenPublicKeyDataAccessor interface {
	CreatePublicKey(ctx context.Context, tokenPublicKey TokenPublicKey) (uint64, error)
	GetPublicKey(ctx context.Context, id uint64) (TokenPublicKey, error)
	WithDatabase(database Database) TokenPublicKeyDataAccessor
}
type tokenPublicKeyDataAccessor struct {
	database Database
}

func NewTokenPublicKeyDataAccessor(database *goqu.Database) TokenPublicKeyDataAccessor {
	return &tokenPublicKeyDataAccessor{
		database: database,
	}
}
func (a tokenPublicKeyDataAccessor) CreatePublicKey(ctx context.Context, tokenPublicKey TokenPublicKey) (uint64, error) {
	result, err := a.database.
		Insert(TabNameTokenPublicKeys).
		Rows(goqu.Record{
			ColNameTokenPublicKeysPublicKey: tokenPublicKey.PublicKey,
		}).
		Executor().
		ExecContext(ctx)
	if err != nil {
		log.Printf("failed to create token public key")
		return 0, status.Error(codes.Internal, "failed to create token public key")
	}
	lastInsertedID, err := result.LastInsertId()
	if err != nil {
		log.Printf("failed to get last inserted id")
		return 0, status.Error(codes.Internal, "failed to get last inserted id")
	}
	return uint64(lastInsertedID), nil
}
func (a tokenPublicKeyDataAccessor) GetPublicKey(ctx context.Context, id uint64) (TokenPublicKey, error) {
	tokenPublicKey := TokenPublicKey{}
	found, err := a.database.
		Select().
		From(TabNameTokenPublicKeys).
		Where(goqu.Ex{
			ColNameTokenPublicKeysID: id,
		}).
		Executor().
		ScanStructContext(ctx, &tokenPublicKey)
	if err != nil {
		log.Printf("failed to get public key")
		return TokenPublicKey{}, status.Error(codes.Internal, "failed to get public key")
	}
	if !found {
		log.Printf("public key not found")
		return TokenPublicKey{}, sql.ErrNoRows
	}
	return tokenPublicKey, nil
}
func (a tokenPublicKeyDataAccessor) WithDatabase(database Database) TokenPublicKeyDataAccessor {
	a.database = database
	return a
}
