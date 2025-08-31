package auth

import (
	"fmt"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"

	"github.com/foohq/foojank/clients/vessel"
)

type User struct {
	JWT string
	Key string
}

func NewUserManager(name, accountPubKey string, accountSigningKey []byte) (*User, error) {
	accountSignKeyPair, err := nkeys.FromSeed(accountSigningKey)
	if err != nil {
		return nil, fmt.Errorf("cannot decode account's signing key: %w", err)
	}

	keyPair, err := nkeys.CreateUser()
	if err != nil {
		return nil, fmt.Errorf("cannot generate user key: %w", err)
	}

	pubKey, err := keyPair.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("cannot get user's public key: %w", err)
	}

	claims := jwt.NewUserClaims(pubKey)
	claims.Name = name
	claims.IssuerAccount = accountPubKey
	claimsEnc, err := claims.Encode(accountSignKeyPair)
	if err != nil {
		return nil, fmt.Errorf("cannot encode and sign user claims: %w", err)
	}

	keySeed, err := keyPair.Seed()
	if err != nil {
		return nil, fmt.Errorf("cannot get a seed of user's key-pair: %w", err)
	}

	return &User{
		JWT: claimsEnc,
		Key: string(keySeed),
	}, nil
}

func NewUserAgent(name, accountPubKey string, accountSigningKey []byte) (*User, error) {
	accountSignKeyPair, err := nkeys.FromSeed(accountSigningKey)
	if err != nil {
		return nil, fmt.Errorf("cannot decode account's signing key: %w", err)
	}

	keyPair, err := nkeys.CreateUser()
	if err != nil {
		return nil, fmt.Errorf("cannot generate user key: %w", err)
	}

	pubKey, err := keyPair.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("cannot get user's public key: %w", err)
	}

	claims := jwt.NewUserClaims(pubKey)
	claims.Name = name
	claims.IssuerAccount = accountPubKey
	claims.Sub = jwt.Permission{
		Allow: []string{
			fmt.Sprintf("_INBOX_%s.>", name),
		},
	}
	claims.Pub = jwt.Permission{
		Allow: []string{
			fmt.Sprintf(vessel.SubjectApiWorkerWriteStdoutT, name, "*"),
			fmt.Sprintf(vessel.SubjectApiWorkerStatusT, name, "*"),
			fmt.Sprintf(vessel.SubjectApiReplyT, name, "*"),
			fmt.Sprintf(vessel.SubjectApiConnInfoT, name),

			fmt.Sprintf("$JS.ACK.FJ_%s.>", name),
			fmt.Sprintf("$JS.API.STREAM.INFO.FJ_%s", name),
			fmt.Sprintf("$JS.API.STREAM.INFO.OBJ_%s", name),
			fmt.Sprintf("$JS.API.STREAM.PURGE.OBJ_%s", name),
			fmt.Sprintf("$JS.API.CONSUMER.INFO.FJ_%s.*", name),
			fmt.Sprintf("$JS.API.CONSUMER.MSG.NEXT.FJ_%s.*", name),
			fmt.Sprintf("$JS.API.CONSUMER.CREATE.OBJ_%s.*.$O.%s.M.*", name, name),
			fmt.Sprintf("$JS.API.CONSUMER.CREATE.OBJ_%s.>", name),
			fmt.Sprintf("$JS.API.CONSUMER.DELETE.OBJ_%s.*", name),
			fmt.Sprintf("$JS.API.DIRECT.GET.OBJ_%s.>", name),
			fmt.Sprintf("$O.%s.M.*", name),
			fmt.Sprintf("$O.%s.C.*", name),
		},
	}
	claimsEnc, err := claims.Encode(accountSignKeyPair)
	if err != nil {
		return nil, fmt.Errorf("cannot encode and sign user claims: %w", err)
	}

	keySeed, err := keyPair.Seed()
	if err != nil {
		return nil, fmt.Errorf("cannot get a seed of user's key-pair: %w", err)
	}

	return &User{
		JWT: claimsEnc,
		Key: string(keySeed),
	}, nil
}
