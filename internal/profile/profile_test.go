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
}

func TestProfile_SourceDir(t *testing.T) {
	p := profile.New()
	require.Empty(t, p.SourceDir())

	expected := "/path/to/source"
	p.SetSourceDir(expected)
	require.Equal(t, expected, p.SourceDir())
}

func TestProfile_Environment(t *testing.T) {
	p := profile.New()
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
	list := p.List()
	require.Len(t, list, 1)
	require.Equal(t, val, list[key])

	// Test Delete
	p.Delete(key)
	require.Empty(t, p.Get(key).Value())
	require.Empty(t, p.List())
}

func TestProfile_JSON(t *testing.T) {
	p := profile.New()
	p.SetSourceDir("/src")
	p.Set("FOO", profile.NewVar("bar"))

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var p2 profile.Profile
	err = json.Unmarshal(data, &p2)
	require.NoError(t, err)

	require.Equal(t, "/src", p2.SourceDir())
	require.Equal(t, "bar", p2.Get("FOO").Value())
}

func TestProfiles_CRUD(t *testing.T) {
	var profiles profile.Profiles
	err := json.Unmarshal([]byte("{}"), &profiles)
	require.NoError(t, err)

	profName := "dev"
	prof := profile.New()
	prof.SetSourceDir("/dev/src")

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
		require.Equal(t, "/dev/src", got.SourceDir())
	})

	t.Run("Get Non-existent Profile", func(t *testing.T) {
		_, err = profiles.Get("missing")
		require.ErrorIs(t, err, profile.ErrProfileNotFound)
	})

	t.Run("Update Profile", func(t *testing.T) {
		profUpdate := profile.New()
		profUpdate.SetSourceDir("/new/src")
		err = profiles.Update(profName, profUpdate)
		require.NoError(t, err)

		gotUpdated, _ := profiles.Get(profName)
		require.Equal(t, "/new/src", gotUpdated.SourceDir())
	})

	t.Run("Update Non-existent Profile", func(t *testing.T) {
		profUpdate := profile.New()
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
	prof1 := profile.New()
	prof1.SetSourceDir("dir1")
	prof1.Set("VAR1", profile.NewVar("val1"))
	prof1.Set("VAR2", profile.NewVar("val2_base"))

	prof2 := profile.New()
	prof2.SetSourceDir("dir2") // Should override
	prof2.Set("VAR2", profile.NewVar("val2_override"))
	prof2.Set("VAR3", profile.NewVar("val3"))

	merged := profile.Merge(prof1, prof2)

	require.Equal(t, "dir2", merged.SourceDir())
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
