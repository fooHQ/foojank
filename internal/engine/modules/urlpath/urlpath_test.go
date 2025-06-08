package urlpath_test

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/risor-io/risor/object"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/engine/modules/urlpath"
)

func TestAbs(t *testing.T) {
	ctx := context.Background()
	wd, err := os.Getwd()
	require.NoError(t, err)
	wd = strings.ReplaceAll(wd, "\\", "/")
	abs := urlpath.Abs(ctx, object.NewString("foo"))
	require.IsType(t, &object.String{}, abs)
	require.Equal(t, path.Join(wd, "foo"), abs.(*object.String).Value())
}

func TestBase(t *testing.T) {
	ctx := context.Background()
	base := urlpath.Base(ctx, object.NewString("/foo/bar.txt"))
	require.IsType(t, &object.String{}, base)
	require.Equal(t, "bar.txt", base.(*object.String).Value())
}

func TestClean(t *testing.T) {
	ctx := context.Background()
	clean := urlpath.Clean(ctx, object.NewString("/foo/../foo/bar//baz"))
	require.IsType(t, &object.String{}, clean)
	require.Equal(t, "/foo/bar/baz", clean.(*object.String).Value())
}

func TestDir(t *testing.T) {
	ctx := context.Background()
	dir := urlpath.Dir(ctx, object.NewString("/foo/bar/baz.txt"))
	require.IsType(t, &object.String{}, dir)
	require.Equal(t, "/foo/bar", dir.(*object.String).Value())
}

func TestExt(t *testing.T) {
	ctx := context.Background()
	ext := urlpath.Ext(ctx, object.NewString("bar/baz.txt"))
	require.IsType(t, &object.String{}, ext)
	require.Equal(t, ".txt", ext.(*object.String).Value())
}

func TestIsAbs(t *testing.T) {
	ctx := context.Background()
	isAbsTrue := urlpath.IsAbs(ctx, object.NewString("/foo/bar"))
	require.IsType(t, &object.Bool{}, isAbsTrue)
	require.True(t, isAbsTrue.(*object.Bool).Value())

	isAbsFalse := urlpath.IsAbs(ctx, object.NewString("foo/bar"))
	require.IsType(t, &object.Bool{}, isAbsFalse)
	require.False(t, isAbsFalse.(*object.Bool).Value())
}

func TestJoin(t *testing.T) {
	ctx := context.Background()
	join := urlpath.Join(ctx, object.NewString("foo"), object.NewString("bar"), object.NewString("baz.txt"))
	require.IsType(t, &object.String{}, join)
	require.Equal(t, "foo/bar/baz.txt", join.(*object.String).Value())
}

func TestMatch(t *testing.T) {
	ctx := context.Background()
	result := urlpath.Match(ctx, object.NewString("*.txt"), object.NewString("file.txt"))
	require.IsType(t, &object.Bool{}, result)
	require.True(t, result.(*object.Bool).Value())

	result = urlpath.Match(ctx, object.NewString("*.txt"), object.NewString("file.jpg"))
	require.IsType(t, &object.Bool{}, result)
	require.False(t, result.(*object.Bool).Value())
}

func TestSplit(t *testing.T) {
	ctx := context.Background()
	split := urlpath.Split(ctx, object.NewString("/foo/bar/baz.txt"))
	require.IsType(t, &object.List{}, split)
	l := split.(*object.List)
	items := l.Value()
	require.Len(t, items, 2)
	require.Equal(t, "/foo/bar", items[0].(*object.String).Value())
	require.Equal(t, "baz.txt", items[1].(*object.String).Value())
}

func TestSplitList(t *testing.T) {
	ctx := context.Background()
	splitList := urlpath.SplitList(ctx, object.NewString(strings.Join([]string{"/foo", "/bar", "/baz"}, string(filepath.ListSeparator))))
	require.IsType(t, &object.List{}, splitList)
	items := splitList.(*object.List).Value()
	require.Len(t, items, 3)
	require.Equal(t, "/foo", items[0].(*object.String).Value())
	require.Equal(t, "/bar", items[1].(*object.String).Value())
	require.Equal(t, "/baz", items[2].(*object.String).Value())
}

func TestWalkDir(t *testing.T) {
	callFunc := func(ctx context.Context, fn *object.Function, args []object.Object) (object.Object, error) {
		require.FailNow(t, "callFunc should not be called")
		return nil, nil
	}
	ctx := context.Background()
	ctx = object.WithCallFunc(ctx, callFunc)

	var items []string
	result := urlpath.WalkDir(
		ctx,
		object.NewString("testdir"),
		object.NewBuiltin("test", func(ctx context.Context, args ...object.Object) object.Object {
			require.Len(t, args, 3)
			require.IsType(t, &object.String{}, args[0])
			items = append(items, args[0].(*object.String).Value())
			return nil
		}),
	)

	require.Equal(t, object.Nil, result)
	require.Equal(t, []string{
		"testdir",
		"testdir/a",
		"testdir/a/a.txt",
		"testdir/b",
		"testdir/b/b.txt",
	}, items)

	var goldenItems []string
	err := filepath.WalkDir("testdir", func(path string, info os.DirEntry, err error) error {
		path = strings.ReplaceAll(path, "\\", "/")
		goldenItems = append(goldenItems, path)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, goldenItems, items)
}
