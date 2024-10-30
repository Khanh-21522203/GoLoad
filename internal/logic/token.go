package logic

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt"

	"GoLoad/internal/configs"
	"GoLoad/internal/dataaccess/database"
)

const (
	rs512KeyPairBitCount = 2048
)

type Token interface {
	GetToken(ctx context.Context, accountID uint64) (string, time.Time, error)
	GetAccountIDAndExpireTime(ctx context.Context, token string) (uint64, time.Time, error)
	WithDatabase(database database.Database) Token
}

func generateRSAKeyPair(bits int) (*rsa.PrivateKey, error) {
	privateKeyPair, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	return privateKeyPair, nil
}

type token struct {
	accountDataAccessor        database.AccountDataAccessor
	tokenPublicKeyDataAccessor database.TokenPublicKeyDataAccessor
	expiresIn                  time.Duration
	privateKey                 *rsa.PrivateKey
	tokenPublicKeyID           uint64
	authConfig                 configs.Auth
}

func pemEncodePublicKey(pubKey *rsa.PublicKey) ([]byte, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return nil, err
	}
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	return pem.EncodeToMemory(block), nil
}
func NewToken(accountDataAccessor database.AccountDataAccessor, tokenPublicKeyDataAccessor database.TokenPublicKeyDataAccessor, authConfig configs.Auth) (Token, error) {
	expiresIn, err := authConfig.Token.GetExpiresInDuration()
	if err != nil {
		log.Printf("failed to parse expires_in")
		return nil, err
	}
	rsaKeyPair, err := generateRSAKeyPair(rs512KeyPairBitCount)
	if err != nil {
		log.Printf("failed to generate rsa key pair")
		return nil, err
	}
	publicKeyBytes, err := pemEncodePublicKey(&rsaKeyPair.PublicKey)
	if err != nil {
		log.Printf("failed to encode public key in pem format")
		return nil, err
	}
	tokenPublicKeyID, err := tokenPublicKeyDataAccessor.CreatePublicKey(
		context.Background(),
		database.TokenPublicKey{PublicKey: publicKeyBytes},
	)
	if err != nil {
		log.Printf("failed to create public key entry in database")
		return nil, err
	}
	return &token{
		accountDataAccessor:        accountDataAccessor,
		tokenPublicKeyDataAccessor: tokenPublicKeyDataAccessor,
		expiresIn:                  expiresIn,
		privateKey:                 rsaKeyPair,
		tokenPublicKeyID:           tokenPublicKeyID,
		authConfig:                 authConfig,
	}, nil
}
func (t token) getJWTPublicKey(ctx context.Context, id uint64) (*rsa.PublicKey, error) {
	tokenPublicKey, err := t.tokenPublicKeyDataAccessor.GetPublicKey(ctx, id)
	if err != nil {
		log.Printf("cannot get token's public key from database")
		return nil, err
	}
	return jwt.ParseRSAPublicKeyFromPEM(tokenPublicKey.PublicKey)
}
func (t token) GetAccountIDAndExpireTime(ctx context.Context, tokenString string) (uint64, time.Time, error) {
	parsedToken, err := jwt.Parse(tokenString, func(parsedToken *jwt.Token) (interface{}, error) {
		if _, ok := parsedToken.Method.(*jwt.SigningMethodRSA); !ok {
			log.Printf("unexpected signing method")
			return nil, errors.New("unexpected signing method")
		}
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			log.Printf("cannot get token's claims")
			return nil, errors.New("cannot get token's claims")
		}
		tokenPublicKeyID, ok := claims["kid"].(float64)
		if !ok {
			log.Printf("cannot get token's kid claim")
			return nil, errors.New("cannot get token's kid claim")
		}
		return t.getJWTPublicKey(ctx, uint64(tokenPublicKeyID))
	})
	if err != nil {
		log.Printf("failed to parse token")
		return 0, time.Time{}, err
	}
	if !parsedToken.Valid {
		log.Printf("invalid token")
		return 0, time.Time{}, errors.New("invalid token")
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("cannot get token's claims")
		return 0, time.Time{}, errors.New("cannot get token's claims")
	}
	accountID, ok := claims["sub"].(float64)
	if !ok {
		log.Printf("cannot get token's sub claim")
		return 0, time.Time{}, errors.New("cannot get token's sub claim")
	}
	expireTimeUnix, ok := claims["exp"].(float64)
	if !ok {
		log.Printf("cannot get token's exp claim")
		return 0, time.Time{}, errors.New("cannot get token's exp claim")
	}
	return uint64(accountID), time.Unix(int64(expireTimeUnix), 0), nil
}
func (t token) GetToken(ctx context.Context, accountID uint64) (string, time.Time, error) {
	expireTime := time.Now().Add(t.expiresIn)
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"sub": accountID,
		"exp": expireTime.Unix(),
		"kid": t.tokenPublicKeyID,
	})
	tokenString, err := token.SignedString(t.privateKey)
	if err != nil {
		log.Printf("failed to sign token")
		return "", time.Time{}, err
	}
	return tokenString, expireTime, nil
}
func (t token) WithDatabase(database database.Database) Token {
	t.accountDataAccessor = t.accountDataAccessor.WithDatabase(database)
	return t
}
