package config_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/config"
)

func TestNewWithOptions(t *testing.T) {
	tests := []struct {
		name string
		opts map[string]any
		want map[string]any
	}{
		{
			name: "empty options",
			opts: map[string]any{},
			want: map[string]any{},
		},
		{
			name: "with simple options",
			opts: map[string]any{
				"test-key": "test-value",
				"flag-1":   true,
			},
			want: map[string]any{
				"test_key": "test-value",
				"flag_1":   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := config.NewWithOptions(tt.opts)

			actual, err := json.Marshal(conf)
			require.NoError(t, err)

			expected, err := json.Marshal(tt.want)
			require.NoError(t, err)

			require.JSONEq(t, string(expected), string(actual))
		})
	}
}

func TestConfig_String(t *testing.T) {
	cfg := config.NewWithOptions(map[string]any{
		"string-key": "value",
		"other-key":  42,
	})

	tests := []struct {
		name      string
		key       string
		wantValue string
		wantOk    bool
	}{
		{
			name:      "existing string key",
			key:       "string-key",
			wantValue: "value",
			wantOk:    true,
		},
		{
			name:      "wrong type key",
			key:       "other-key",
			wantValue: "",
			wantOk:    false,
		},
		{
			name:      "non-existing key",
			key:       "missing",
			wantValue: "",
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := cfg.String(tt.key)
			require.Equal(t, tt.wantValue, gotValue)
			require.Equal(t, tt.wantOk, gotOk)
		})
	}
}

func TestConfig_Bool(t *testing.T) {
	cfg := config.NewWithOptions(map[string]any{
		"bool-key":  true,
		"other-key": "not-a-bool",
	})

	tests := []struct {
		name      string
		key       string
		wantValue bool
		wantOk    bool
	}{
		{
			name:      "existing bool key",
			key:       "bool-key",
			wantValue: true,
			wantOk:    true,
		},
		{
			name:      "wrong type key",
			key:       "other-key",
			wantValue: false,
			wantOk:    false,
		},
		{
			name:      "non-existing key",
			key:       "missing",
			wantValue: false,
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := cfg.Bool(tt.key)
			require.Equal(t, tt.wantValue, gotValue)
			require.Equal(t, tt.wantOk, gotOk)
		})
	}
}

func TestParseFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErr  bool
		wantData map[string]any
	}{
		{
			name:     "valid json",
			content:  `{"key": "value", "bool_key": true}`,
			wantErr:  false,
			wantData: map[string]any{"key": "value", "bool_key": true},
		},
		{
			name:    "invalid json",
			content: `{invalid json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "config-*.json")
			require.NoError(t, err)
			defer func() {
				err := tmpFile.Close()
				require.NoError(t, err)
			}()

			_, err = tmpFile.Write([]byte(tt.content))
			require.NoError(t, err)

			conf, err := config.ParseFile(tmpFile.Name())
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			actual, err := json.Marshal(conf)
			require.NoError(t, err)

			expected, err := json.Marshal(tt.wantData)
			require.NoError(t, err)

			require.JSONEq(t, string(expected), string(actual))
		})
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name  string
		confs []*config.Config
		want  map[string]any
	}{
		{
			name: "merge two configs",
			confs: []*config.Config{
				config.NewWithOptions(map[string]any{"key1": "value1"}),
				config.NewWithOptions(map[string]any{"key2": "value2"}),
			},
			want: map[string]any{"key1": "value1", "key2": "value2"},
		},
		{
			name: "override values",
			confs: []*config.Config{
				config.NewWithOptions(map[string]any{"key": "value1"}),
				config.NewWithOptions(map[string]any{"key": "value2"}),
			},
			want: map[string]any{"key": "value2"},
		},
		{
			name: "with nil config",
			confs: []*config.Config{
				config.NewWithOptions(map[string]any{"key": "value"}),
				nil,
			},
			want: map[string]any{"key": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := config.Merge(tt.confs...)

			actual, err := json.Marshal(conf)
			require.NoError(t, err)

			expected, err := json.Marshal(tt.want)
			require.NoError(t, err)

			require.JSONEq(t, string(expected), string(actual))
		})
	}
}

func TestFlagToOption(t *testing.T) {
	tests := []struct {
		name string
		flag string
		want string
	}{
		{
			name: "simple flag",
			flag: "flag",
			want: "flag",
		},
		{
			name: "flag with single dash",
			flag: "flag-name",
			want: "flag_name",
		},
		{
			name: "flag with multiple dashes",
			flag: "flag-with-multiple-dashes",
			want: "flag_with_multiple_dashes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.FlagToOption(tt.flag)
			require.Equal(t, tt.want, got)
		})
	}
}
