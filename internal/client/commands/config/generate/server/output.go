package server

import (
	"github.com/goccy/go-yaml"
	"github.com/nats-io/nkeys"

	"github.com/foohq/foojank/internal/client/commands/config/generate/seed"
)

type Output struct {
	Host          string        `yaml:"host"`
	Operators     []seed.Entity `yaml:"operators"`
	Accounts      []seed.Entity `yaml:"accounts"`
	SystemAccount seed.Entity   `yaml:"system_account"`
}

func (s *Output) String() string {
	b, _ := yaml.Marshal(s)
	return string(b)
}

func NewOutput(seedFile *seed.Output) (*Output, error) {
	operatorKeyPair, err := nkeys.FromSeed([]byte(seedFile.Operator.Key))
	if err != nil {
		return nil, err
	}

	operatorPubKey, err := operatorKeyPair.PublicKey()
	if err != nil {
		return nil, err
	}

	accountKeyPair, err := nkeys.FromSeed([]byte(seedFile.Account.Key))
	if err != nil {
		return nil, err
	}

	accountPubKey, err := accountKeyPair.PublicKey()
	if err != nil {
		return nil, err
	}

	systemAccountKeyPair, err := nkeys.FromSeed([]byte(seedFile.SystemAccount.Key))
	if err != nil {
		return nil, err
	}

	systemAccountPubKey, err := systemAccountKeyPair.PublicKey()
	if err != nil {
		return nil, err
	}

	return &Output{
		Host: "localhost:4222",
		Operators: []seed.Entity{
			{
				JWT: seedFile.Operator.JWT,
				Key: operatorPubKey,
			},
		},
		Accounts: []seed.Entity{
			{
				JWT: seedFile.Account.JWT,
				Key: accountPubKey,
			},
		},
		SystemAccount: seed.Entity{
			JWT: seedFile.SystemAccount.JWT,
			Key: systemAccountPubKey,
		},
	}, nil
}
