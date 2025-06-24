//go:build !module_urlpath_stub

package urlpath

import (
	"context"
	"io/fs"
	"strings"

	"github.com/risor-io/risor/arg"
	"github.com/risor-io/risor/object"
	risoros "github.com/risor-io/risor/os"

	"github.com/foohq/urlpath"
)

func Abs(ctx context.Context, args ...object.Object) object.Object {
	if err := arg.Require("urlpath.abs", 1, args); err != nil {
		return err
	}
	path, rerr := object.AsString(args[0])
	if rerr != nil {
		return rerr
	}
	osObj := risoros.GetDefaultOS(ctx)
	wd, err := osObj.Getwd()
	if err != nil {
		return object.NewError(err)
	}
	abs, err := urlpath.Abs(path, wd)
	if err != nil {
		return object.NewError(err)
	}
	return object.NewString(abs)
}

func Base(ctx context.Context, args ...object.Object) object.Object {
	if err := arg.Require("urlpath.base", 1, args); err != nil {
		return err
	}
	path, rerr := object.AsString(args[0])
	if rerr != nil {
		return rerr
	}
	base, err := urlpath.Base(path)
	if err != nil {
		return object.NewError(err)
	}
	return object.NewString(base)
}

func Clean(ctx context.Context, args ...object.Object) object.Object {
	if err := arg.Require("urlpath.clean", 1, args); err != nil {
		return err
	}
	path, rerr := object.AsString(args[0])
	if rerr != nil {
		return rerr
	}
	cleanPath, err := urlpath.Clean(path)
	if err != nil {
		return object.NewError(err)
	}
	return object.NewString(cleanPath)
}

func Dir(ctx context.Context, args ...object.Object) object.Object {
	if rerr := arg.Require("urlpath.dir", 1, args); rerr != nil {
		return rerr
	}
	path, rerr := object.AsString(args[0])
	if rerr != nil {
		return rerr
	}
	dirPath, err := urlpath.Dir(path)
	if err != nil {
		return object.NewError(err)
	}
	return object.NewString(dirPath)
}

func Ext(ctx context.Context, args ...object.Object) object.Object {
	if rerr := arg.Require("urlpath.ext", 1, args); rerr != nil {
		return rerr
	}
	path, rerr := object.AsString(args[0])
	if rerr != nil {
		return rerr
	}
	extension, err := urlpath.Ext(path)
	if err != nil {
		return object.NewError(err)
	}
	return object.NewString(extension)
}

func IsAbs(ctx context.Context, args ...object.Object) object.Object {
	if rerr := arg.Require("urlpath.is_abs", 1, args); rerr != nil {
		return rerr
	}
	path, rerr := object.AsString(args[0])
	if rerr != nil {
		return rerr
	}
	isAbs, err := urlpath.IsAbs(path)
	if err != nil {
		return object.NewError(err)
	}
	return object.NewBool(isAbs)
}

func Join(ctx context.Context, args ...object.Object) object.Object {
	paths := make([]string, len(args))
	for i, arg := range args {
		path, rerr := object.AsString(arg)
		if rerr != nil {
			return rerr
		}
		paths[i] = path
	}
	res, err := urlpath.Join(paths...)
	if err != nil {
		return object.NewError(err)
	}
	return object.NewString(res)
}

func Match(ctx context.Context, args ...object.Object) object.Object {
	if rerr := arg.Require("urlpath.match", 2, args); rerr != nil {
		return rerr
	}
	pattern, rerr := object.AsString(args[0])
	if rerr != nil {
		return rerr
	}
	name, rerr := object.AsString(args[1])
	if rerr != nil {
		return rerr
	}
	matched, err := urlpath.Match(pattern, name)
	if err != nil {
		return object.NewError(err)
	}
	return object.NewBool(matched)
}

func Split(ctx context.Context, args ...object.Object) object.Object {
	if rerr := arg.Require("urlpath.split", 1, args); rerr != nil {
		return rerr
	}
	path, rerr := object.AsString(args[0])
	if rerr != nil {
		return rerr
	}
	dir, file, err := urlpath.Split(path)
	if err != nil {
		return object.NewError(err)
	}
	return object.NewList([]object.Object{
		object.NewString(dir),
		object.NewString(file),
	})
}

func SplitList(ctx context.Context, args ...object.Object) object.Object {
	if rerr := arg.Require("filepath.split_list", 1, args); rerr != nil {
		return rerr
	}
	pathList, rerr := object.AsString(args[0])
	if rerr != nil {
		return rerr
	}
	if pathList == "" {
		return object.NewStringList([]string{})
	}
	osObj := risoros.GetDefaultOS(ctx)
	paths := strings.Split(pathList, string(osObj.PathListSeparator()))
	pathObjs := make([]object.Object, 0, len(paths))
	for _, path := range paths {
		pathObjs = append(pathObjs, object.NewString(path))
	}
	return object.NewList(pathObjs)
}

func WalkDir(ctx context.Context, args ...object.Object) object.Object {
	if rerr := arg.Require("urlpath.walk_dir", 2, args); rerr != nil {
		return rerr
	}
	root, rerr := object.AsString(args[0])
	if rerr != nil {
		return rerr
	}
	callFunc, found := object.GetCallFunc(ctx)
	if !found {
		return object.Errorf("eval error: urlpath.walk() context did not contain a call function")
	}
	osObj := risoros.GetDefaultOS(ctx)

	type callable func(path, info, err object.Object) object.Object
	var callback callable

	switch obj := args[1].(type) {
	case *object.Builtin:
		callback = func(path, info, err object.Object) object.Object {
			return obj.Call(ctx, path, info, err)
		}
	case *object.Function:
		callback = func(path, info, err object.Object) object.Object {
			args := []object.Object{path, info, err}
			result, resultErr := callFunc(ctx, obj, args)
			if resultErr != nil {
				return object.NewError(resultErr)
			}
			return result
		}
	default:
		return object.TypeErrorf("type error: urlpath.walk() expected a function (%s given)", obj.Type())
	}

	walkFn := func(path string, d fs.DirEntry, err error) error {
		var errObj object.Object
		if err != nil {
			errObj = object.NewError(err)
		} else {
			errObj = object.Nil
		}
		path, err = urlpath.Clean(path)
		if err != nil {
			errObj = object.NewError(err)
		}
		wrapper := risoros.DirEntryWrapper{DirEntry: d}
		result := callback(object.NewString(path), object.NewDirEntry(&wrapper), errObj)
		switch result := result.(type) {
		case *object.Error:
			return result.Value()
		default:
			return nil
		}
	}
	walkErr := osObj.WalkDir(root, walkFn)
	if walkErr != nil {
		return object.NewError(walkErr)
	}
	return object.Nil
}

func Module() *object.Module {
	return object.NewBuiltinsModule("urlpath", map[string]object.Object{
		"abs":        object.NewBuiltin("abs", Abs),
		"base":       object.NewBuiltin("base", Base),
		"clean":      object.NewBuiltin("clean", Clean),
		"dir":        object.NewBuiltin("dir", Dir),
		"ext":        object.NewBuiltin("ext", Ext),
		"is_abs":     object.NewBuiltin("is_abs", IsAbs),
		"join":       object.NewBuiltin("join", Join),
		"match":      object.NewBuiltin("match", Match),
		"split_list": object.NewBuiltin("split_list", SplitList),
		"split":      object.NewBuiltin("split", Split),
		"walk_dir":   object.NewBuiltin("walk_dir", WalkDir),
	})
}

func Builtins() map[string]object.Object {
	return nil
}
