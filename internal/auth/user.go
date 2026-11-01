package auth

import (
	"errors"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

type User struct {
	claims  *jwt.UserClaims
	keyPair nkeys.KeyPair
	name    string
	pubKey  string
}

func NewUser(claims *jwt.UserClaims, keyPair nkeys.KeyPair) (*User, error) {
	pubKey, err := keyPair.PublicKey()
	if err != nil {
		return nil, err
	}

	if pubKey != claims.Subject {
		return nil, errors.New("JWT and key mismatch")
	}

	return &User{
		claims:  claims,
		keyPair: keyPair,
		name:    claims.Name,
		pubKey:  pubKey,
	}, nil
}

func GenerateUser(name string, perms jwt.Permissions) (*User, error) {
	var err error

	keyPair, err := nkeys.CreateUser()
	if err != nil {
		return nil, err
	}

	pubKey, err := keyPair.PublicKey()
	if err != nil {
		return nil, err
	}

	claims := jwt.NewUserClaims(pubKey)
	claims.Name = name
	claims.Permissions = perms

	return &User{
		claims:  claims,
		keyPair: keyPair,
		name:    name,
		pubKey:  pubKey,
	}, nil
}

func (u *User) Name() string {
	return u.name
}

func (u *User) PublicKey() string {
	return u.pubKey
}

func (u *User) Claims() *jwt.UserClaims {
	return u.claims
}

func (u *User) KeyPair() nkeys.KeyPair {
	return u.keyPair
}

func SetDummyUserJWT(claims *jwt.UserClaims) {
	claims.Tags.Add("fj:dummy")
}

func IsDummyUserJWT(claims *jwt.UserClaims) bool {
	return claims.Tags.Contains("fj:dummy")
}

var (
	inactive = Status("Inactive")
	active   = Status("Active")
	expired  = Status("Expired")
)

type Status string

func (s Status) String() string {
	return string(s)
}

func GetJWTStatus(claims *jwt.UserClaims) Status {
	if IsDummyUserJWT(claims) {
		return inactive
	}
	if claims.NotBefore > 0 && time.Now().Before(time.Unix(claims.NotBefore, 0)) {
		return inactive
	}
	if claims.Expires > 0 && time.Now().After(time.Unix(claims.Expires, 0)) {
		return expired
	}
	return active
}
