package configdir

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/foohq/foojank/internal/config"
	"github.com/foohq/foojank/internal/profile"
)

func Init(dir string) error {
	err := os.MkdirAll(filepath.Join(dir, ".foojank"), 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	err = InitConfigJSON(dir)
	if err != nil {
		return err
	}

	err = InitProfilesJSON(dir)
	if err != nil {
		return err
	}

	return nil
}

func Search(dir string) (string, error) {
	for range 128 {
		isConfigDir, err := IsConfigDir(dir)
		if err != nil {
			return "", err
		}

		if !isConfigDir {
			dir = dir + "/../"
			continue
		}

		return dir, nil
	}

	return "", errors.New("configuration directory not found")
}

func IsConfigDir(dir string) (bool, error) {
	info, err := os.Stat(filepath.Join(dir, ".foojank"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	if !info.IsDir() {
		return false, nil
	}

	_, err = ParseConfigJSON(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func InitConfigJSON(dir string) error {
	pth := filepath.Join(dir, ".foojank", "config.json")
	_, err := os.Open(pth)
	if err == nil {
		// File already exists
		return nil
	}
	return os.WriteFile(pth, []byte("{}"), 0644)
}

func UpdateConfigJSON(dir string, conf *config.Config) error {
	b, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	pth := filepath.Join(dir, ".foojank", "config.json")
	f, err := os.CreateTemp(filepath.Dir(pth), "config*.json")
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()

	_, err = f.Write(b)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	err = os.Rename(f.Name(), pth)
	if err != nil {
		return err
	}

	return nil
}

func ParseConfigJSON(dir string) (*config.Config, error) {
	pth := filepath.Join(dir, ".foojank", "config.json")
	return config.ParseFile(pth)
}

func InitProfilesJSON(dir string) error {
	pth := filepath.Join(dir, ".foojank", "profiles.json")
	_, err := os.Open(pth)
	if err == nil {
		// File already exists
		return nil
	}
	return os.WriteFile(pth, []byte("{}"), 0644)
}

func UpdateProfilesJSON(dir string, profs *profile.Profiles) error {
	b, err := json.Marshal(profs)
	if err != nil {
		return err
	}

	pth := filepath.Join(dir, ".foojank", "profiles.json")
	f, err := os.CreateTemp(filepath.Dir(pth), "profiles*.json")
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()

	_, err = f.Write(b)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	err = os.Rename(f.Name(), pth)
	if err != nil {
		return err
	}

	return nil
}

func ParseProfilesJSON(dir string) (*profile.Profiles, error) {
	pth := filepath.Join(dir, ".foojank", "profiles.json")
	return profile.ParseFile(pth)
}
