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
		opts map[string]string
		want map[string]string
	}{
		{
			name: "empty options",
			opts: map[string]string{},
			want: map[string]string{},
		},
		{
			name: "with simple options",
			opts: map[string]string{
				"test-key": "test-value",
				"flag-1":   "true",
			},
			want: map[string]string{
				"test_key": "test-value",
				"flag_1":   "true",
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
	cfg := config.NewWithOptions(map[string]string{
		"string-key": "value",
		"other-key":  "42",
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
			name:      "existing int key as string",
			key:       "other-key",
			wantValue: "42",
			wantOk:    true,
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
	cfg := config.NewWithOptions(map[string]string{
		"bool-key":  "true",
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
		wantData map[string]string
	}{
		{
			name:     "valid json",
			content:  `{"key": "value", "bool_key": "true"}`,
			wantErr:  false,
			wantData: map[string]string{"key": "value", "bool_key": "true"},
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
		want  map[string]string
	}{
		{
			name: "merge two configs",
			confs: []*config.Config{
				config.NewWithOptions(map[string]string{"key1": "value1"}),
				config.NewWithOptions(map[string]string{"key2": "value2"}),
			},
			want: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name: "override values",
			confs: []*config.Config{
				config.NewWithOptions(map[string]string{"key": "value1"}),
				config.NewWithOptions(map[string]string{"key": "value2"}),
			},
			want: map[string]string{"key": "value2"},
		},
		{
			name: "with nil config",
			confs: []*config.Config{
				config.NewWithOptions(map[string]string{"key": "value"}),
				nil,
			},
			want: map[string]string{"key": "value"},
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

func TestParseKVPairsJSON(t *testing.T) {
	tests := []struct {
		name  string
		pairs []string
		want  map[string]any
	}{
		{
			name:  "simple string",
			pairs: []string{"key1=value1"},
			want:  map[string]any{"key1": "value1"},
		},
		{
			name:  "number",
			pairs: []string{"key1=123"},
			want:  map[string]any{"key1": 123.0},
		},
		{
			name:  "boolean",
			pairs: []string{"key1=true"},
			want:  map[string]any{"key1": true},
		},
		{
			name:  "json array",
			pairs: []string{"key1=[1, 2, 3]"},
			want:  map[string]any{"key1": []any{1.0, 2.0, 3.0}},
		},
		{
			name:  "json object",
			pairs: []string{`key1={"a": 1, "b": "c"}`},
			want:  map[string]any{"key1": map[string]any{"a": 1.0, "b": "c"}},
		},
		{
			name:  "multiple values",
			pairs: []string{"key1=value1", "key2=123", "key3=true", "key4=[1, 2, 3]", `key5={"a": 1}`},
			want: map[string]any{
				"key1": "value1",
				"key2": 123.0,
				"key3": true,
				"key4": []any{1.0, 2.0, 3.0},
				"key5": map[string]any{"a": 1.0},
			},
		},
		{
			name:  "invalid json falls back to string",
			pairs: []string{"key1={invalid"},
			want:  map[string]any{"key1": "{invalid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.ParseKVPairsJSON(tt.pairs)
			require.Equal(t, tt.want, got)
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
