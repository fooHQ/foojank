package auth

import (
	"fmt"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

type Operator struct {
	JWT        string
	Key        string
	SigningKey string
}

func NewOperator(name string) (*Operator, error) {
	keyPair, err := nkeys.CreateOperator()
	if err != nil {
		return nil, fmt.Errorf("cannot generate operator's key: %w", err)
	}

	pubKey, err := keyPair.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("cannot get operator's public key: %w", err)
	}

	signKeyPair, err := nkeys.CreateOperator()
	if err != nil {
		return nil, fmt.Errorf("cannot generate operator's signing key: %w", err)
	}

	signPubKey, err := signKeyPair.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("cannot get operator's signing public key: %w", err)
	}

	claims := jwt.NewOperatorClaims(pubKey)
	claims.Name = name
	claims.SigningKeys.Add(signPubKey)
	claimsEnc, err := claims.Encode(keyPair)
	if err != nil {
		return nil, fmt.Errorf("cannot encode and sign operator claims: %w", err)
	}

	keySeed, err := keyPair.Seed()
	if err != nil {
		return nil, fmt.Errorf("cannot get seed of operator's key: %w", err)
	}

	signKeySeed, err := signKeyPair.Seed()
	if err != nil {
		return nil, fmt.Errorf("cannot get seed of operator's signing key: %w", err)
	}

	return &Operator{
		JWT:        claimsEnc,
		Key:        string(keySeed),
		SigningKey: string(signKeySeed),
	}, nil
}
