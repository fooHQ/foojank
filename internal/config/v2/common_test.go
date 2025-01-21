package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/config/v2"
	"github.com/foohq/foojank/internal/testutils"
)

func TestMergeCommon(t *testing.T) {
	defaults, err := config.NewDefaultCommon()
	require.NoError(t, err)

	merged := config.MergeCommon(defaults, &config.Common{
		DataDir:  testutils.NewString("/tmp/test"),
		LogLevel: testutils.NewString("error"),
		NoColor:  testutils.NewBool(true),
	})
	require.Equal(t, &config.Common{
		DataDir:  testutils.NewString("/tmp/test"),
		LogLevel: testutils.NewString("error"),
		NoColor:  testutils.NewBool(true),
	}, merged)
}

func TestParseCommonFlags(t *testing.T) {
	flags := map[string]any{
		"data-dir":  "/tmp/test",
		"log-level": "error",
		"no-color":  true,
	}
	getFlag := func(name string) (any, bool) {
		v, ok := flags[name]
		return v, ok
	}

	conf, err := config.ParseCommonFlags(getFlag)
	require.NoError(t, err)

	require.Equal(t, &config.Common{
		DataDir:  testutils.NewString("/tmp/test"),
		LogLevel: testutils.NewString("error"),
		NoColor:  testutils.NewBool(true),
	}, conf)
}
