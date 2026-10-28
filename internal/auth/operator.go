package auth

import (
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

func NewOperatorKey() (nkeys.KeyPair, error) {
	return nkeys.CreateOperator()
}

func NewOperatorJWT(name string, operatorKey nkeys.KeyPair) (*jwt.OperatorClaims, error) {
	operatorPublicKey, err := operatorKey.PublicKey()
	if err != nil {
		return nil, err
	}

	claims := jwt.NewOperatorClaims(operatorPublicKey)
	claims.Name = name

	return claims, nil
}
