package seed

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"os"
)

type Entity struct {
	JWT        string `yaml:"jwt"`
	Key        string `yaml:"key"`
	SigningKey string `yaml:"signing_key,omitempty"`
}

type Output struct {
	Servers       []string `yaml:"servers"`
	Operator      Entity   `yaml:"operator"`
	Account       Entity   `yaml:"account"`
	SystemAccount Entity   `yaml:"system_account"`
}

func (s *Output) String() string {
	b, _ := yaml.Marshal(s)
	return string(b)
}

func NewOutput(servers []string, operatorName, accountName string) (*Output, error) {
	operatorJWT, operatorKey, operatorSignKey, err := generateOperator(operatorName)
	if err != nil {
		return nil, err
	}

	accountJWT, accountKey, accountSignKey, err := generateAccount(accountName, operatorSignKey)
	if err != nil {
		return nil, err
	}

	sysAccountJWT, sysAccountKey, sysAccountSignKey, err := generateAccount("SYS", operatorSignKey)
	if err != nil {
		return nil, err
	}

	return &Output{
		Servers: servers,
		Operator: Entity{
			JWT:        operatorJWT,
			Key:        string(operatorKey),
			SigningKey: string(operatorSignKey),
		},
		Account: Entity{
			JWT:        accountJWT,
			Key:        string(accountKey),
			SigningKey: string(accountSignKey),
		},
		SystemAccount: Entity{
			JWT:        sysAccountJWT,
			Key:        string(sysAccountKey),
			SigningKey: string(sysAccountSignKey),
		},
	}, nil
}

func ParseOutput(file string) (*Output, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var conf Output
	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func generateOperator(name string) (string, []byte, []byte, error) {
	keyPair, err := nkeys.CreateOperator()
	if err != nil {
		return "", nil, nil, fmt.Errorf("cannot generate operator key-pair: %v", err)
	}

	signKeyPair, err := nkeys.CreateOperator()
	if err != nil {
		return "", nil, nil, fmt.Errorf("cannot generate operator key-pair: %v", err)
	}

	pubKey, err := keyPair.PublicKey()
	if err != nil {
		return "", nil, nil, fmt.Errorf("cannot get operator's public key: %v", err)
	}

	signPubKey, err := signKeyPair.PublicKey()
	if err != nil {
		return "", nil, nil, fmt.Errorf("cannot get operator's signing public key: %v", err)
	}

	claims := jwt.NewOperatorClaims(pubKey)
	claims.Name = name
	claims.SigningKeys.Add(signPubKey)
	claimsEnc, err := claims.Encode(keyPair)
	if err != nil {
		return "", nil, nil, fmt.Errorf("cannot encode and sign operator claims: %v", err)
	}

	keyPairSeedEnc, err := keyPair.Seed()
	if err != nil {
		return "", nil, nil, fmt.Errorf("cannot get a seed of operator's key-pair: %v", err)
	}

	signKeyPairSeed, err := signKeyPair.Seed()
	if err != nil {
		return "", nil, nil, fmt.Errorf("cannot get a seed of operator's key-pair: %v", err)
	}

	return claimsEnc, keyPairSeedEnc, signKeyPairSeed, nil
}

func generateAccount(name string, operatorSignKey []byte) (string, []byte, []byte, error) {
	keyPair, err := nkeys.CreateAccount()
	if err != nil {
		return "", nil, nil, fmt.Errorf("cannot generate account key-pair: %v", err)
	}

	signKeyPair, err := nkeys.CreateAccount()
	if err != nil {
		return "", nil, nil, fmt.Errorf("cannot generate account key-pair: %v", err)
	}

	operatorSignKeyPair, err := nkeys.FromSeed(operatorSignKey)
	if err != nil {
		return "", nil, nil, fmt.Errorf("cannot decode operator's signing key-pair: %v", err)
	}

	pubKey, err := keyPair.PublicKey()
	if err != nil {
		return "", nil, nil, fmt.Errorf("cannot get account's public key: %v", err)
	}

	signPubKey, err := signKeyPair.PublicKey()
	if err != nil {
		return "", nil, nil, fmt.Errorf("cannot get account's signing public key: %v", err)
	}

	claims := jwt.NewAccountClaims(pubKey)
	claims.Name = name
	claims.SigningKeys.Add(signPubKey)
	claimsEnc, err := claims.Encode(operatorSignKeyPair)
	if err != nil {
		return "", nil, nil, fmt.Errorf("cannot encode and sign account claims: %v", err)
	}

	keyPairSeedEnc, err := keyPair.Seed()
	if err != nil {
		return "", nil, nil, fmt.Errorf("cannot get a seed of account's key-pair: %v", err)
	}

	signKeyPairSeed, err := signKeyPair.Seed()
	if err != nil {
		return "", nil, nil, fmt.Errorf("cannot get a seed of account's key-pair: %v", err)
	}

	return claimsEnc, keyPairSeedEnc, signKeyPairSeed, nil
}
