package profile_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/profile"
)

func TestVar(t *testing.T) {
	t.Run("NewVar sets value", func(t *testing.T) {
		v := profile.NewVar("foo")
		require.Equal(t, "foo", v.Value())
	})

	t.Run("SetValue updates value", func(t *testing.T) {
		v := profile.NewVar("foo")
		v.SetValue("bar")
		require.Equal(t, "bar", v.Value())
	})

	t.Run("JSON Marshaling", func(t *testing.T) {
		v := profile.NewVar("test-val")
		data, err := json.Marshal(v)
		require.NoError(t, err)

		var v2 profile.Var
		err = json.Unmarshal(data, &v2)
		require.NoError(t, err)
		require.Equal(t, "test-val", v2.Value())
	})

	t.Run("Description", func(t *testing.T) {
		v := profile.NewVar("foo")
		require.Empty(t, v.Description())

		v2 := profile.NewVar("bar")
		err := json.Unmarshal([]byte(`{"value":"bar", "description":"test desc"}`), &v2)
		require.NoError(t, err)
		require.Equal(t, "test desc", v2.Description())
	})
}

func TestProfile_Environment(t *testing.T) {
	p := profile.NewProfile()
	key := "TEST_VAR"
	val := "test_value"

	// Test Set and Get
	p.Set(key, profile.NewVar(val))
	got := p.Get(key)
	require.Equal(t, val, got.Value())

	// Test Get non-existent
	missing := p.Get("MISSING")
	require.NotNil(t, missing)
	require.Empty(t, missing.Value())

	// Test List
	list := p.ToEnv()
	require.Equal(t, val, list[key])

	// Test Delete
	p.Delete(key)
	require.Empty(t, p.Get(key).Value())
}

func TestProfile_Attributes(t *testing.T) {
	p := profile.NewProfile()

	require.Empty(t, p.OS())
	require.Empty(t, p.Arch())
	require.Empty(t, p.Target())
	require.Nil(t, p.Features())

	err := json.Unmarshal([]byte(`{"os":"linux", "arch":"amd64", "target":"prod", "features":["f1", "f2"]}`), &p)
	require.NoError(t, err)

	require.Equal(t, "linux", p.OS())
	require.Equal(t, "amd64", p.Arch())
	require.Equal(t, "prod", p.Target())
	require.Equal(t, []string{"f1", "f2"}, p.Features())
}

func TestProfile_JSON(t *testing.T) {
	p := profile.NewProfile()
	p.Set("FOO", profile.NewVar("bar"))

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var p2 profile.Profile
	err = json.Unmarshal(data, &p2)
	require.NoError(t, err)

	require.Equal(t, "bar", p2.Get("FOO").Value())
}

func TestProfiles_CRUD(t *testing.T) {
	var profiles profile.Profiles
	err := json.Unmarshal([]byte("{}"), &profiles)
	require.NoError(t, err)

	profName := "dev"
	prof := profile.NewProfile()

	t.Run("Add Profile", func(t *testing.T) {
		err = profiles.Add(profName, prof)
		require.NoError(t, err)
	})

	t.Run("Add Duplicate Profile", func(t *testing.T) {
		err = profiles.Add(profName, prof)
		require.ErrorIs(t, err, profile.ErrProfileExists)
	})

	t.Run("Get Profile", func(t *testing.T) {
		got, err := profiles.Get(profName)
		require.NoError(t, err)
		require.NotNil(t, got)
	})

	t.Run("Get Non-existent Profile", func(t *testing.T) {
		_, err = profiles.Get("missing")
		require.ErrorIs(t, err, profile.ErrProfileNotFound)
	})

	t.Run("Update Profile", func(t *testing.T) {
		profUpdate := profile.NewProfile()
		err = profiles.Update(profName, profUpdate)
		require.NoError(t, err)

		gotUpdated, _ := profiles.Get(profName)
		require.NotNil(t, gotUpdated)
	})

	t.Run("Update Non-existent Profile", func(t *testing.T) {
		profUpdate := profile.NewProfile()
		err = profiles.Update("missing", profUpdate)
		require.ErrorIs(t, err, profile.ErrProfileNotFound)
	})

	t.Run("List Profiles", func(t *testing.T) {
		profs := profiles.List()
		require.Contains(t, profs, profName)
	})

	t.Run("Delete Profile", func(t *testing.T) {
		err = profiles.Delete(profName)
		require.NoError(t, err)

		_, err = profiles.Get(profName)
		require.ErrorIs(t, err, profile.ErrProfileNotFound)
	})

	t.Run("Delete Non-existent Profile", func(t *testing.T) {
		err = profiles.Delete(profName)
		require.ErrorIs(t, err, profile.ErrProfileNotFound)
	})
}

func TestMerge(t *testing.T) {
	prof1 := profile.NewProfile()
	prof1.Set("VAR1", profile.NewVar("val1"))
	prof1.Set("VAR2", profile.NewVar("val2_base"))

	prof2 := profile.NewProfile()
	prof2.Set("VAR2", profile.NewVar("val2_override"))
	prof2.Set("VAR3", profile.NewVar("val3"))

	merged := profile.Merge(prof1, prof2)

	require.Equal(t, "val1", merged.Get("VAR1").Value())
	require.Equal(t, "val2_override", merged.Get("VAR2").Value())
	require.Equal(t, "val3", merged.Get("VAR3").Value())
}

func TestParseKVPairs(t *testing.T) {
	tests := []struct {
		name     string
		pairs    []string
		expected map[string]string
	}{
		{
			name:     "nil input",
			pairs:    nil,
			expected: map[string]string{},
		},
		{
			name:     "empty input",
			pairs:    []string{},
			expected: map[string]string{},
		},
		{
			name:  "simple key value",
			pairs: []string{"KEY=value"},
			expected: map[string]string{
				"KEY": "value",
			},
		},
		{
			name:  "multiple pairs",
			pairs: []string{"KEY1=value1", "KEY2=value2"},
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name:  "value containing equals sign",
			pairs: []string{"CONN_STR=host=localhost;port=5432"},
			expected: map[string]string{
				"CONN_STR": "host=localhost;port=5432",
			},
		},
		{
			name:  "key without value",
			pairs: []string{"FLAG_ENABLED"},
			expected: map[string]string{
				"FLAG_ENABLED": "",
			},
		},
		{
			name:  "empty value with equals",
			pairs: []string{"EMPTY_VAL="},
			expected: map[string]string{
				"EMPTY_VAL": "",
			},
		},
		{
			name:  "key with surrounding spaces is trimmed",
			pairs: []string{"  SPACED_KEY  =value"},
			expected: map[string]string{
				"SPACED_KEY": "value",
			},
		},
		{
			name:  "value with spaces is preserved",
			pairs: []string{"GREETING= Hello World "},
			expected: map[string]string{
				"GREETING": " Hello World ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profile.ParseKVPairs(tt.pairs)

			require.Len(t, got, len(tt.expected))

			for k, wantVal := range tt.expected {
				gotVar, ok := got[k]
				require.True(t, ok, "ParseKVPairs() missing key %q", k)
				require.Equal(t, wantVal, gotVar.Value(), "ParseKVPairs()[%q] value mismatch", k)
			}
		})
	}
}
