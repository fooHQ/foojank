package config_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/config/v2"
	"github.com/foohq/foojank/internal/testutils"
)

func TestMergeClient(t *testing.T) {
	defaults, err := config.NewDefaultClient()
	require.NoError(t, err)

	merged := config.MergeClient(defaults, &config.Client{
		Config: &config.Config{
			NoColor: testutils.NewBool(true),
		},
		Server: []string{
			"ws://example.com",
			"wss://example.com",
		},
		UserJWT:    testutils.NewString("JWT_PLACEHOLDER"),
		UserKey:    testutils.NewString("KEY_PLACEHOLDER"),
		AccountKey: testutils.NewString("KEY_PLACEHOLDER"),
	})
	require.Equal(t, &config.Client{
		Config: &config.Config{
			DataDir:  testutils.NewString(filepath.Join(testutils.GetUserHomeDir(t), "foojank")),
			LogLevel: testutils.NewString("info"),
			NoColor:  testutils.NewBool(true),
		},
		Server: []string{
			"ws://example.com",
			"wss://example.com",
		},
		UserJWT:    testutils.NewString("JWT_PLACEHOLDER"),
		UserKey:    testutils.NewString("KEY_PLACEHOLDER"),
		AccountKey: testutils.NewString("KEY_PLACEHOLDER"),
	}, merged)
}

func TestParseClientFlags(t *testing.T) {
	flags := map[string]any{
		"data-dir":    filepath.Join(testutils.GetUserHomeDir(t), "foojank"),
		"log-level":   "info",
		"no-color":    false,
		"server":      []string{"ws://example.com"},
		"user-jwt":    "JWT_PLACEHOLDER",
		"user-key":    "KEY_PLACEHOLDER",
		"account-key": "KEY_PLACEHOLDER",
	}
	getFlag := func(name string) (any, bool) {
		v, ok := flags[name]
		return v, ok
	}

	conf, err := config.ParseClientFlags(getFlag)
	require.NoError(t, err)

	require.Equal(t, &config.Client{
		Config: &config.Config{
			DataDir:  testutils.NewString(filepath.Join(testutils.GetUserHomeDir(t), "foojank")),
			LogLevel: testutils.NewString("info"),
			NoColor:  testutils.NewBool(false),
		},
		Server: []string{
			"ws://example.com",
		},
		UserJWT:    testutils.NewString("JWT_PLACEHOLDER"),
		UserKey:    testutils.NewString("KEY_PLACEHOLDER"),
		AccountKey: testutils.NewString("KEY_PLACEHOLDER"),
	}, conf)
}
