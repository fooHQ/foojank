package auth

import (
	"errors"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

type Account struct {
	claims  *jwt.AccountClaims
	keyPair nkeys.KeyPair
	name    string
	pubKey  string
}

func NewAccount(claims *jwt.AccountClaims, keyPair nkeys.KeyPair) (*Account, error) {
	pubKey, err := keyPair.PublicKey()
	if err != nil {
		return nil, err
	}

	if pubKey != claims.Subject {
		return nil, errors.New("JWT and key mismatch")
	}

	return &Account{
		claims:  claims,
		keyPair: keyPair,
		name:    claims.Name,
		pubKey:  pubKey,
	}, nil
}

func GenerateAccount(name string) (*Account, error) {
	var err error

	keyPair, err := nkeys.CreateAccount()
	if err != nil {
		return nil, err
	}

	pubKey, err := keyPair.PublicKey()
	if err != nil {
		return nil, err
	}

	claims := jwt.NewAccountClaims(pubKey)
	claims.Name = name
	claims.Limits.JetStreamLimits = jwt.JetStreamLimits{
		DiskStorage:   -1,
		MemoryStorage: -1,
	}

	return &Account{
		claims:  claims,
		keyPair: keyPair,
		name:    name,
		pubKey:  pubKey,
	}, nil
}

func (a *Account) Name() string {
	return a.name
}

func (a *Account) PublicKey() string {
	return a.pubKey
}

func (a *Account) Claims() *jwt.AccountClaims {
	return a.claims
}

func (a *Account) KeyPair() nkeys.KeyPair {
	return a.keyPair
}

func (a *Account) Encode(keyPair nkeys.KeyPair) (string, error) {
	return a.claims.Encode(keyPair)
}

func NewAccountKey() (nkeys.KeyPair, error) {
	return nkeys.CreateAccount()
}

func NewAccountJWT(name string, accountKey nkeys.KeyPair) (*jwt.AccountClaims, error) {
	accountPublicKey, err := accountKey.PublicKey()
	if err != nil {
		return nil, err
	}

	claims := jwt.NewAccountClaims(accountPublicKey)
	claims.Name = name
	claims.Limits.JetStreamLimits = jwt.JetStreamLimits{
		DiskStorage:   -1,
		MemoryStorage: -1,
	}
	return claims, nil
}
