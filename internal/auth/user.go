package auth

import (
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
