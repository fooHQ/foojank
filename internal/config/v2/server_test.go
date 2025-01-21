package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/config/v2"
	"github.com/foohq/foojank/internal/testutils"
)

func TestMergeServer(t *testing.T) {
	defaults, err := config.NewDefaultServer()
	require.NoError(t, err)

	merged := config.MergeServer(defaults, &config.Server{
		Host:             testutils.NewString("localhost"),
		Port:             testutils.NewInt64(8080),
		OperatorJWT:      testutils.NewString("JWT_PLACEHOLDER"),
		OperatorKey:      testutils.NewString("KEY_PLACEHOLDER"),
		AccountJWT:       testutils.NewString("JWT_PLACEHOLDER"),
		AccountKey:       testutils.NewString("KEY_PLACEHOLDER"),
		SystemAccountJWT: testutils.NewString("JWT_PLACEHOLDER"),
		SystemAccountKey: testutils.NewString("KEY_PLACEHOLDER"),
	})
	require.Equal(t, &config.Server{
		Host:             testutils.NewString("localhost"),
		Port:             testutils.NewInt64(8080),
		OperatorJWT:      testutils.NewString("JWT_PLACEHOLDER"),
		OperatorKey:      testutils.NewString("KEY_PLACEHOLDER"),
		AccountJWT:       testutils.NewString("JWT_PLACEHOLDER"),
		AccountKey:       testutils.NewString("KEY_PLACEHOLDER"),
		SystemAccountJWT: testutils.NewString("JWT_PLACEHOLDER"),
		SystemAccountKey: testutils.NewString("KEY_PLACEHOLDER"),
	}, merged)
}

func TestParseServerFlags(t *testing.T) {
	flags := map[string]any{
		"host":               "localhost",
		"port":               int64(8080),
		"operator-jwt":       "JWT_PLACEHOLDER",
		"operator-key":       "KEY_PLACEHOLDER",
		"account-jwt":        "JWT_PLACEHOLDER",
		"account-key":        "KEY_PLACEHOLDER",
		"system-account-jwt": "JWT_PLACEHOLDER",
		"system-account-key": "KEY_PLACEHOLDER",
	}
	getFlag := func(name string) (any, bool) {
		v, ok := flags[name]
		return v, ok
	}

	conf, err := config.ParseCommonFlags(getFlag)
	require.NoError(t, err)

	require.Equal(t, &config.Server{
		Host:             testutils.NewString("localhost"),
		Port:             testutils.NewInt64(8080),
		OperatorJWT:      testutils.NewString("JWT_PLACEHOLDER"),
		OperatorKey:      testutils.NewString("KEY_PLACEHOLDER"),
		AccountJWT:       testutils.NewString("JWT_PLACEHOLDER"),
		AccountKey:       testutils.NewString("KEY_PLACEHOLDER"),
		SystemAccountJWT: testutils.NewString("JWT_PLACEHOLDER"),
		SystemAccountKey: testutils.NewString("KEY_PLACEHOLDER"),
	}, conf)
}
