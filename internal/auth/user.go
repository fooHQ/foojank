package auth

import (
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

func NewUserKey() (nkeys.KeyPair, error) {
	return nkeys.CreateUser()
}

func NewUserJWT(name string, perms jwt.Permissions, userKey nkeys.KeyPair) (*jwt.UserClaims, error) {
	userPublicKey, err := userKey.PublicKey()
	if err != nil {
		return nil, err
	}

	claims := jwt.NewUserClaims(userPublicKey)
	claims.Name = name
	claims.Permissions = perms
	return claims, nil
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
