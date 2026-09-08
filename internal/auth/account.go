package auth

import (
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

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
