package auth

import (
	"fmt"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

type Account struct {
	JWT        string
	Key        string
	SigningKey string
}

func NewAccount(name string, operatorSignKey []byte, enableJetstream bool) (*Account, error) {
	operatorSignKeyPair, err := nkeys.FromSeed(operatorSignKey)
	if err != nil {
		return nil, fmt.Errorf("cannot decode operator's signing key: %w", err)
	}

	keyPair, err := nkeys.CreateAccount()
	if err != nil {
		return nil, fmt.Errorf("cannot generate account's key: %w", err)
	}

	pubKey, err := keyPair.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("cannot get account's public key: %w", err)
	}

	signKeyPair, err := nkeys.CreateAccount()
	if err != nil {
		return nil, fmt.Errorf("cannot generate account's key: %w", err)
	}

	signPubKey, err := signKeyPair.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("cannot get account's signing public key: %w", err)
	}

	claims := jwt.NewAccountClaims(pubKey)
	claims.Name = name
	if enableJetstream {
		claims.Limits.JetStreamLimits = jwt.JetStreamLimits{
			DiskStorage:   -1,
			MemoryStorage: -1,
		}
	}
	claims.SigningKeys.Add(signPubKey)
	claimsEnc, err := claims.Encode(operatorSignKeyPair)
	if err != nil {
		return nil, fmt.Errorf("cannot encode and sign account claims: %w", err)
	}

	keySeed, err := keyPair.Seed()
	if err != nil {
		return nil, fmt.Errorf("cannot get a seed of account's key: %w", err)
	}

	signKeySeed, err := signKeyPair.Seed()
	if err != nil {
		return nil, fmt.Errorf("cannot get a seed of account's key: %w", err)
	}

	return &Account{
		JWT:        claimsEnc,
		Key:        string(keySeed),
		SigningKey: string(signKeySeed),
	}, nil
}
