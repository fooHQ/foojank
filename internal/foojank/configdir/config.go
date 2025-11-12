package configdir

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/foohq/foojank/internal/config"
)

func Init(dir string) error {
	err := os.MkdirAll(filepath.Join(dir, ".foojank"), 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	err = InitConfigJson(dir)
	if err != nil {
		return err
	}

	return nil
}

func Search(dir string) (string, error) {
	for i := 0; i < 128; i++ {
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

	_, err = ParseConfigJson(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func InitConfigJson(dir string) error {
	pth := filepath.Join(dir, ".foojank", "config.json")
	return os.WriteFile(pth, []byte("{}"), 0644)
}

func UpdateConfigJson(dir string, conf *config.Config) error {
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

func ParseConfigJson(dir string) (*config.Config, error) {
	pth := filepath.Join(dir, ".foojank", "config.json")
	return config.ParseFile(pth)
}
