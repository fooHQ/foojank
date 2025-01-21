package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/testutils"
)

func TestMergeClient(t *testing.T) {
	defaults, err := config.NewDefaultClient()
	require.NoError(t, err)

	merged := config.MergeClient(defaults, &config.Client{
		Server: []string{
			"ws://example.com",
			"wss://example.com",
		},
		UserJWT:    testutils.NewString("JWT_PLACEHOLDER"),
		UserKey:    testutils.NewString("KEY_PLACEHOLDER"),
		AccountJWT: testutils.NewString("JWT_PLACEHOLDER"),
		AccountKey: testutils.NewString("KEY_PLACEHOLDER"),
	})
	require.Equal(t, &config.Client{
		Server: []string{
			"ws://example.com",
			"wss://example.com",
		},
		UserJWT:    testutils.NewString("JWT_PLACEHOLDER"),
		UserKey:    testutils.NewString("KEY_PLACEHOLDER"),
		AccountJWT: testutils.NewString("JWT_PLACEHOLDER"),
		AccountKey: testutils.NewString("KEY_PLACEHOLDER"),
	}, merged)
}

func TestParseClientFlags(t *testing.T) {
	flags := map[string]any{
		"server":      []string{"ws://example.com"},
		"user-jwt":    "JWT_PLACEHOLDER",
		"user-key":    "KEY_PLACEHOLDER",
		"account-jwt": "JWT_PLACEHOLDER",
		"account-key": "KEY_PLACEHOLDER",
	}
	getFlag := func(name string) (any, bool) {
		v, ok := flags[name]
		return v, ok
	}

	conf, err := config.ParseClientFlags(getFlag)
	require.NoError(t, err)

	require.Equal(t, &config.Client{
		Server: []string{
			"ws://example.com",
		},
		UserJWT:    testutils.NewString("JWT_PLACEHOLDER"),
		UserKey:    testutils.NewString("KEY_PLACEHOLDER"),
		AccountJWT: testutils.NewString("JWT_PLACEHOLDER"),
		AccountKey: testutils.NewString("KEY_PLACEHOLDER"),
	}, conf)
}
