package auth

import (
	"errors"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

type Operator struct {
	claims  *jwt.OperatorClaims
	keyPair nkeys.KeyPair
	name    string
	pubKey  string
}

func NewOperator(claims *jwt.OperatorClaims, keyPair nkeys.KeyPair) (*Operator, error) {
	pubKey, err := keyPair.PublicKey()
	if err != nil {
		return nil, err
	}

	if pubKey != claims.Subject {
		return nil, errors.New("JWT and key mismatch")
	}

	return &Operator{
		claims:  claims,
		keyPair: keyPair,
		name:    claims.Name,
		pubKey:  pubKey,
	}, nil
}

func GenerateOperator(name string) (*Operator, error) {
	var err error

	keyPair, err := nkeys.CreateOperator()
	if err != nil {
		return nil, err
	}

	pubKey, err := keyPair.PublicKey()
	if err != nil {
		return nil, err
	}

	claims := jwt.NewOperatorClaims(pubKey)
	claims.Name = name

	return &Operator{
		claims:  claims,
		keyPair: keyPair,
		name:    name,
		pubKey:  pubKey,
	}, nil
}

func (o *Operator) Name() string {
	return o.name
}

func (o *Operator) PublicKey() string {
	return o.pubKey
}

func (o *Operator) Claims() *jwt.OperatorClaims {
	return o.claims
}

func (o *Operator) KeyPair() nkeys.KeyPair {
	return o.keyPair
}

func (o *Operator) Encode(keyPair nkeys.KeyPair) (string, error) {
	return o.claims.Encode(keyPair)
}
