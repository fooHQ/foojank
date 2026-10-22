package profile_test

import (
	"encoding/json"
	"sort"
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
	list := p.Env()
	require.Equal(t, val, list[key])

	// Test Delete
	p.Delete(key)
	require.Empty(t, p.Get(key).Value())
}

func TestProfile_Attributes(t *testing.T) {
	p := profile.NewProfile()

	require.Empty(t, p.OS())
	require.Empty(t, p.Arch())

	err := json.Unmarshal([]byte(`{"os":"linux", "arch":"amd64"}`), &p)
	require.NoError(t, err)

	require.Equal(t, "linux", p.OS())
	require.Equal(t, "amd64", p.Arch())
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

func TestProfile_Commands(t *testing.T) {
	p := profile.NewProfile()

	ls := profile.NewCommand("ls")
	ls.SetDescription("List a directory")
	recursive := profile.NewCommandFlag("recursive")
	recursive.SetDescription("Recurse into directories")
	ls.SetFlags(map[string]*profile.CommandFlag{
		"recursive": recursive,
	})
	depth := profile.NewVar("1")
	ls.SetEnvironment(map[string]*profile.Var{
		"DEPTH": depth,
	})

	reg := profile.NewCommand("reg")
	reg.SetDescription("Read the registry")

	p.SetCommand(ls)
	p.SetCommand(reg)

	ls.SetDescription("changed")

	got := p.GetCommand("ls")
	require.Equal(t, "ls", got.Name())
	require.Equal(t, "List a directory", got.Description())
	require.Equal(t, "recursive", got.Flags()["recursive"].Name())
	require.Equal(t, "Recurse into directories", got.Flags()["recursive"].Description())
	require.Equal(t, "1", got.Environment()["DEPTH"].Value())

	names := commandNames(p.Commands())
	require.Equal(t, []string{"ls", "reg"}, names)

	updated := profile.NewCommand("ls")
	updated.SetDescription("List files")
	p.SetCommand(updated)
	require.Equal(t, []string{"ls", "reg"}, commandNames(p.Commands()))
	require.Equal(t, "List files", p.GetCommand("ls").Description())
	require.Empty(t, p.GetCommand("ls").Flags())

	p.DeleteCommand("ls")
	require.Equal(t, []string{"reg"}, commandNames(p.Commands()))
	require.Empty(t, p.GetCommand("ls").Name())

	p.SetCommand(nil)
	require.Equal(t, []string{"reg"}, commandNames(p.Commands()))

	missing := p.GetCommand("missing")
	require.Empty(t, missing.Name())
	require.Empty(t, missing.Flags())
	require.Empty(t, missing.Environment())
}

func TestProfile_CommandJSON(t *testing.T) {
	raw := []byte(`{
		"commands": {
			"ls": {
				"description": "List a directory",
				"flags": {
					"recursive": {"description": "Recurse into directories"},
					"depth": {"description": "Max recursion depth", "default": "1"}
				},
				"environment": {
					"DEPTH": {"description": "Max recursion depth", "default": "1"}
				}
			}
		}
	}`)

	var p profile.Profile
	err := json.Unmarshal(raw, &p)
	require.NoError(t, err)

	cmd := p.GetCommand("ls")
	require.Equal(t, "ls", cmd.Name())
	require.Equal(t, "List a directory", cmd.Description())
	require.Equal(t, "recursive", cmd.Flags()["recursive"].Name())
	require.Equal(t, "1", cmd.Flags()["depth"].Default())
	require.Equal(t, "Max recursion depth", cmd.Environment()["DEPTH"].Description())
	require.Equal(t, "1", cmd.Environment()["DEPTH"].Value())

	data, err := json.Marshal(&p)
	require.NoError(t, err)

	var again profile.Profile
	err = json.Unmarshal(data, &again)
	require.NoError(t, err)
	require.Equal(t, cmd.Description(), again.GetCommand("ls").Description())
	require.Equal(t, "recursive", again.GetCommand("ls").Flags()["recursive"].Name())
	require.Equal(t, "Recurse into directories", again.GetCommand("ls").Flags()["recursive"].Description())
}

func TestMergeCommands(t *testing.T) {
	raw := []byte(`{
		"defaults": {
			"commands": {
				"ls": {
					"description": "List a directory",
					"flags": {
						"all": {"description": "Include hidden entries"}
					}
				},
				"ps": {"description": "List processes"}
			}
		},
		"profiles": {
			"linux-amd64": {"os": "linux", "arch": "amd64"},
			"windows-amd64": {
				"os": "windows",
				"arch": "amd64",
				"commands": {
					"reg": {
						"description": "Read the registry",
						"flags": {
							"key": {"description": "Registry key"}
						}
					},
					"ls": {
						"description": "List a directory",
						"flags": {
							"force": {}
						}
					}
				}
			}
		}
	}`)

	var profiles profile.Profiles
	err := json.Unmarshal(raw, &profiles)
	require.NoError(t, err)

	linux, err := profiles.Get("linux-amd64")
	require.NoError(t, err)
	require.Equal(t, []string{"ls", "ps"}, commandNames(linux.Commands()))
	require.Equal(t, "all", linux.GetCommand("ls").Flags()["all"].Name())

	windows, err := profiles.Get("windows-amd64")
	require.NoError(t, err)
	require.Equal(t, []string{"ls", "ps", "reg"}, commandNames(windows.Commands()))
	require.Equal(t, "List processes", windows.GetCommand("ps").Description())
	require.Equal(t, "force", windows.GetCommand("ls").Flags()["force"].Name())
	require.Equal(t, "key", windows.GetCommand("reg").Flags()["key"].Name())

	data, err := json.Marshal(&profiles)
	require.NoError(t, err)
	var stored struct {
		Profiles map[string]struct {
			Commands map[string]struct {
				Description string `json:"description"`
			} `json:"commands"`
		} `json:"profiles"`
	}
	err = json.Unmarshal(data, &stored)
	require.NoError(t, err)
	require.Empty(t, stored.Profiles["linux-amd64"].Commands)
	require.Equal(t, []string{"ls", "reg"}, commandKeys(stored.Profiles["windows-amd64"].Commands))
	require.Equal(t, "Read the registry", stored.Profiles["windows-amd64"].Commands["reg"].Description)
}

func commandNames(cmds map[string]*profile.Command) []string {
	names := make([]string, 0, len(cmds))
	for name := range cmds {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func commandKeys[T any](cmds map[string]T) []string {
	names := make([]string, 0, len(cmds))
	for name := range cmds {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
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
