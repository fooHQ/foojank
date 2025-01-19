package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"github.com/pelletier/go-toml/v2"
)

var ErrParserError = errors.New("parser error")

var (
	DefaultClientConfigPath = func() string {
		confDir, err := os.UserConfigDir()
		if err != nil {
			confDir = "."
		}
		return filepath.Join(confDir, "foojank", "foojank.conf")
	}
	DefaultServerConfigPath = func() string {
		confDir, err := os.UserConfigDir()
		if err != nil {
			confDir = "."
		}
		return filepath.Join(confDir, "foojank", "server.conf")
	}
	DefaultServerStoreDirPath = func() string {
		return filepath.Join(os.TempDir(), "foojank")
	}
	DefaultHost           = "0.0.0.0"
	DefaultPort     int64 = 443
	DefaultServers        = []string{fmt.Sprintf("ws://localhost:%d", DefaultPort)}
	DefaultLogLevel       = "info"
	DefaultNoColor        = false
)

func setDefaults(conf *Config) *Config {
	if conf == nil {
		conf = &Config{}
	}

	if conf.Host == nil {
		conf.Host = &DefaultHost
	}

	if conf.Port == nil {
		conf.Port = &DefaultPort
	}

	if conf.StoreDir == nil {
		v := DefaultServerStoreDirPath()
		conf.StoreDir = &v
	}

	if conf.Servers == nil {
		conf.Servers = DefaultServers
	}

	if conf.LogLevel == nil {
		conf.LogLevel = &DefaultLogLevel
	}

	if conf.NoColor == nil {
		conf.NoColor = &DefaultNoColor
	}

	return conf
}

type Service struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
}

type Entity struct {
	JWT            string `toml:"jwt"`
	KeySeed        string `toml:"key_seed,omitempty"`
	SigningKeySeed string `toml:"signing_key_seed,omitempty"`
}

type Config struct {
	Host          *string  `toml:"host,omitempty"`
	Port          *int64   `toml:"port,omitempty"`
	Servers       []string `toml:"servers,omitempty"`
	Operator      *Entity  `toml:"operator,omitempty"`
	Account       *Entity  `toml:"account,omitempty"`
	SystemAccount *Entity  `toml:"system_account,omitempty"`
	User          *Entity  `toml:"user,omitempty"`
	LogLevel      *string  `toml:"log_level,omitempty"`
	NoColor       *bool    `toml:"no_color,omitempty"`
	Codebase      *string  `toml:"codebase,omitempty"`
	StoreDir      *string  `toml:"store_dir,omitempty"`
}

func (s *Config) String() string {
	b, _ := toml.Marshal(s)
	return string(b)
}

func ParseFile(file string, mustExist bool) (*Config, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) && !mustExist {
			return setDefaults(nil), nil
		}
		return nil, err
	}

	var conf Config
	err = toml.Unmarshal(b, &conf)
	if err != nil {
		return nil, errors.Join(ErrParserError, err)
	}

	conf = *setDefaults(&conf)

	return &conf, nil
}

func NewOperator(name string) (*Entity, error) {
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

	return &Entity{
		JWT:            claimsEnc,
		KeySeed:        string(keySeed),
		SigningKeySeed: string(signKeySeed),
	}, nil
}

func NewAccount(name string, operatorSignKey []byte, enableJetstream bool) (*Entity, error) {
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
		claims.Limits.JetStreamLimits.DiskStorage = -1
		claims.Limits.JetStreamLimits.MemoryStorage = -1
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

	return &Entity{
		JWT:            claimsEnc,
		KeySeed:        string(keySeed),
		SigningKeySeed: string(signKeySeed),
	}, nil
}

func NewUserManager(name, accountPubKey string, accountSigningKey []byte) (*Entity, error) {
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

	return &Entity{
		JWT:     claimsEnc,
		KeySeed: string(keySeed),
	}, nil
}

func NewUserAgent(name, accountPubKey string, accountSigningKey []byte) (*Entity, error) {
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
			fmt.Sprintf("$JS.API.STREAM.INFO.OBJ_%s", name),
			fmt.Sprintf("$JS.API.DIRECT.GET.OBJ_%s.$O.%s.M.*", name, name),
			fmt.Sprintf("$JS.API.CONSUMER.CREATE.OBJ_%s.*.$O.%s.C.*", name, name),
			fmt.Sprintf("$JS.API.CONSUMER.DELETE.OBJ_%s.*", name),
			fmt.Sprintf("%s.*.STDOUT", name),
			"_INBOX.>",
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

	return &Entity{
		JWT:     claimsEnc,
		KeySeed: string(keySeed),
	}, nil
}
