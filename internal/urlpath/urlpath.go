package urlpath

import (
	"net/url"
	"path"
)

func Abs(pth, wd string) (string, error) {
	if wd == "" {
		wd = "/"
	}
	u, err := fromString(pth)
	if err != nil {
		return "", err
	}
	wdu, err := fromString(wd)
	if err != nil {
		return "", err
	}
	return normalize(u, wdu), nil
}

func Base(pth string) (string, error) {
	u, err := fromString(pth)
	if err != nil {
		return "", err
	}
	return path.Base(u.Path), nil
}

func Dir(pth string) (string, error) {
	u, err := fromString(pth)
	if err != nil {
		return "", err
	}
	u.Path = path.Dir(u.Path)
	return toString(u), nil
}

func Ext(pth string) (string, error) {
	u, err := fromString(pth)
	if err != nil {
		return "", err
	}
	return path.Ext(u.Path), nil
}

func Clean(pth string) (string, error) {
	u, err := fromString(pth)
	if err != nil {
		return "", err
	}
	u.Path = path.Clean(u.Path)
	return toString(u), nil
}

func IsAbs(pth string) (bool, error) {
	u, err := fromString(pth)
	if err != nil {
		return false, err
	}
	return isAbsURL(u), nil
}

func Join(elem ...string) (string, error) {
	var uf *url.URL
	for _, e := range elem {
		u, err := fromString(e)
		if err != nil {
			return "", err
		}
		if uf == nil {
			uf = u
			continue
		}
		uf.Path = path.Join(uf.Path, u.Path)
	}
	return toString(uf), nil
}

func Split(pth string) (string, string, error) {
	u, err := fromString(pth)
	if err != nil {
		return "", "", err
	}
	pthDir, pthFile := path.Split(u.Path)
	u.Path = pthDir
	return toString(u), pthFile, nil
}

func Match(pattern, name string) (bool, error) {
	u, err := fromString(name)
	if err != nil {
		return false, err
	}
	name = toString(u)
	return path.Match(pattern, name)
}

func isFileURL(u *url.URL) bool {
	return u.Scheme == "file" || u.Scheme == ""
}
