package client

import (
	"fmt"
	"github.com/foohq/foojank/internal/application/commands/config/generate/seed"
	"github.com/goccy/go-yaml"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

type Output struct {
	Servers []string    `yaml:"servers"`
	User    seed.Entity `yaml:"user"`
}

func (c *Output) String() string {
	b, _ := yaml.Marshal(c)
	return string(b)
}

func NewOutput(seedFile *seed.Output, username string) (*Output, error) {
	userJWT, userKey, err := generateUser(username, []byte(seedFile.Account.Key), []byte(seedFile.Account.SigningKey))
	if err != nil {
		return nil, err
	}

	return &Output{
		Servers: seedFile.Servers,
		User: seed.Entity{
			JWT: userJWT,
			Key: string(userKey),
		},
	}, nil
}

func generateUser(name string, accountKey, accountSignKey []byte) (string, []byte, error) {
	keyPair, err := nkeys.CreateUser()
	if err != nil {
		return "", nil, fmt.Errorf("cannot generate user key-pair: %v", err)
	}

	accountKeyPair, err := nkeys.FromSeed(accountKey)
	if err != nil {
		return "", nil, fmt.Errorf("cannot decode account's key-pair: %v", err)
	}

	accountSignKeyPair, err := nkeys.FromSeed(accountSignKey)
	if err != nil {
		return "", nil, fmt.Errorf("cannot decode account's signing key-pair: %v", err)
	}

	pubKey, err := keyPair.PublicKey()
	if err != nil {
		return "", nil, fmt.Errorf("cannot get user's public key: %v", err)
	}

	accountPubKey, err := accountKeyPair.PublicKey()
	if err != nil {
		return "", nil, fmt.Errorf("cannot get account's public key: %v", err)
	}

	claims := jwt.NewUserClaims(pubKey)
	claims.Name = name
	claims.IssuerAccount = accountPubKey
	// TODO: define permissions
	claimsEnc, err := claims.Encode(accountSignKeyPair)
	if err != nil {
		return "", nil, fmt.Errorf("cannot encode and sign user claims: %v", err)
	}

	keyPairSeedEnc, err := keyPair.Seed()
	if err != nil {
		return "", nil, fmt.Errorf("cannot get a seed of user's key-pair: %v", err)
	}

	return claimsEnc, keyPairSeedEnc, nil
}
