package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

var ErrParserError = errors.New("parser error")

type Entity struct {
	JWT            string `yaml:"jwt"`
	PublicKey      string `yaml:"public_key"`
	KeySeed        string `yaml:"key_seed,omitempty"`
	SigningKeySeed string `yaml:"signing_key_seed,omitempty"`
}

type Config struct {
	Host          string   `yaml:"host,omitempty"`
	Servers       []string `yaml:"servers,omitempty"`
	Operator      *Entity  `yaml:"operator,omitempty"`
	Account       *Entity  `yaml:"account,omitempty"`
	SystemAccount *Entity  `yaml:"system_account,omitempty"`
	User          *Entity  `yaml:"user,omitempty"`
	LogLevel      *int64   `yaml:"log_level,omitempty"`
	NoColor       *bool    `yaml:"no_color,omitempty"`
}

func (s *Config) String() string {
	b, _ := yaml.Marshal(s)
	return string(b)
}

func ParseFile(file string) (*Config, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var conf Config
	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		return nil, errors.Join(ErrParserError, err)
	}

	return &conf, nil
}

func NewOperator(name string) (*Entity, error) {
	keyPair, err := nkeys.CreateOperator()
	if err != nil {
		return nil, fmt.Errorf("cannot generate operator's key: %v", err)
	}

	pubKey, err := keyPair.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("cannot get operator's public key: %v", err)
	}

	signKeyPair, err := nkeys.CreateOperator()
	if err != nil {
		return nil, fmt.Errorf("cannot generate operator's signing key: %v", err)
	}

	signPubKey, err := signKeyPair.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("cannot get operator's signing public key: %v", err)
	}

	claims := jwt.NewOperatorClaims(pubKey)
	claims.Name = name
	claims.SigningKeys.Add(signPubKey)
	claimsEnc, err := claims.Encode(keyPair)
	if err != nil {
		return nil, fmt.Errorf("cannot encode and sign operator claims: %v", err)
	}

	keySeed, err := keyPair.Seed()
	if err != nil {
		return nil, fmt.Errorf("cannot get seed of operator's key: %v", err)
	}

	signKeySeed, err := signKeyPair.Seed()
	if err != nil {
		return nil, fmt.Errorf("cannot get seed of operator's signing key: %v", err)
	}

	return &Entity{
		JWT:            claimsEnc,
		PublicKey:      pubKey,
		KeySeed:        string(keySeed),
		SigningKeySeed: string(signKeySeed),
	}, nil
}

func NewAccount(name string, operatorSignKey []byte) (*Entity, error) {
	operatorSignKeyPair, err := nkeys.FromSeed(operatorSignKey)
	if err != nil {
		return nil, fmt.Errorf("cannot decode operator's signing key: %v", err)
	}

	keyPair, err := nkeys.CreateAccount()
	if err != nil {
		return nil, fmt.Errorf("cannot generate account's key: %v", err)
	}

	pubKey, err := keyPair.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("cannot get account's public key: %v", err)
	}

	signKeyPair, err := nkeys.CreateAccount()
	if err != nil {
		return nil, fmt.Errorf("cannot generate account's key: %v", err)
	}

	signPubKey, err := signKeyPair.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("cannot get account's signing public key: %v", err)
	}

	claims := jwt.NewAccountClaims(pubKey)
	claims.Name = name
	claims.SigningKeys.Add(signPubKey)
	claimsEnc, err := claims.Encode(operatorSignKeyPair)
	if err != nil {
		return nil, fmt.Errorf("cannot encode and sign account claims: %v", err)
	}

	keySeed, err := keyPair.Seed()
	if err != nil {
		return nil, fmt.Errorf("cannot get a seed of account's key: %v", err)
	}

	signKeySeed, err := signKeyPair.Seed()
	if err != nil {
		return nil, fmt.Errorf("cannot get a seed of account's key: %v", err)
	}

	return &Entity{
		JWT:            claimsEnc,
		PublicKey:      pubKey,
		KeySeed:        string(keySeed),
		SigningKeySeed: string(signKeySeed),
	}, nil
}

func NewUserManager(name, accountPubKey string, accountSigningKey []byte) (*Entity, error) {
	accountSignKeyPair, err := nkeys.FromSeed(accountSigningKey)
	if err != nil {
		return nil, fmt.Errorf("cannot decode account's signing key: %v", err)
	}

	keyPair, err := nkeys.CreateUser()
	if err != nil {
		return nil, fmt.Errorf("cannot generate user key: %v", err)
	}

	pubKey, err := keyPair.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("cannot get user's public key: %v", err)
	}

	claims := jwt.NewUserClaims(pubKey)
	claims.Name = name
	claims.IssuerAccount = accountPubKey
	// TODO: define permissions
	claimsEnc, err := claims.Encode(accountSignKeyPair)
	if err != nil {
		return nil, fmt.Errorf("cannot encode and sign user claims: %v", err)
	}

	keySeed, err := keyPair.Seed()
	if err != nil {
		return nil, fmt.Errorf("cannot get a seed of user's key-pair: %v", err)
	}

	return &Entity{
		JWT:       claimsEnc,
		PublicKey: pubKey,
		KeySeed:   string(keySeed),
	}, nil
}

func NewUserAgent(name, accountPubKey string, accountSigningKey []byte) (*Entity, error) {
	accountSignKeyPair, err := nkeys.FromSeed(accountSigningKey)
	if err != nil {
		return nil, fmt.Errorf("cannot decode account's signing key: %v", err)
	}

	keyPair, err := nkeys.CreateUser()
	if err != nil {
		return nil, fmt.Errorf("cannot generate user key: %v", err)
	}

	pubKey, err := keyPair.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("cannot get user's public key: %v", err)
	}

	claims := jwt.NewUserClaims(pubKey)
	claims.Name = name
	claims.IssuerAccount = accountPubKey
	claims.Sub = jwt.Permission{
		Allow: []string{
			fmt.Sprintf("_INBOX_%s.>", name),
			"$SRV.PING",
			fmt.Sprintf("$SRV.PING.%s", name),
			fmt.Sprintf("$SRV.PING.%s.*", name),
			"$SRV.INFO",
			fmt.Sprintf("$SRV.INFO.%s", name),
			fmt.Sprintf("$SRV.INFO.%s.*", name),
			"$SRV.STATS",
			fmt.Sprintf("$SRV.STATS.%s", name),
			fmt.Sprintf("$SRV.STATS.%s.*", name),
			fmt.Sprintf("%s.RPC", name),
			fmt.Sprintf("%s.*.DATA", name),
			fmt.Sprintf("%s.*.STDIN", name),
		},
	}
	claims.Pub = jwt.Permission{
		Allow: []string{
			fmt.Sprintf("_INBOX_%s.>", name),
			fmt.Sprintf("$JS.API.STREAM.INFO.OBJ_%s", name),
			fmt.Sprintf("$JS.API.DIRECT.GET.OBJ_%s.$O.%s.M.*", name, name),
			fmt.Sprintf("$JS.API.CONSUMER.CREATE.OBJ_%s.*.$O.%s.C.*", name, name),
			fmt.Sprintf("$JS.API.CONSUMER.DELETE.OBJ_%s.*", name),
			fmt.Sprintf("%s.*.STDOUT", name),
		},
	}
	claims.Resp = &jwt.ResponsePermission{
		MaxMsgs: 1,
	}
	claimsEnc, err := claims.Encode(accountSignKeyPair)
	if err != nil {
		return nil, fmt.Errorf("cannot encode and sign user claims: %v", err)
	}

	keySeed, err := keyPair.Seed()
	if err != nil {
		return nil, fmt.Errorf("cannot get a seed of user's key-pair: %v", err)
	}

	return &Entity{
		JWT:       claimsEnc,
		PublicKey: pubKey,
		KeySeed:   string(keySeed),
	}, nil
}
