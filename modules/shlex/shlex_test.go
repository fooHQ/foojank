package shlex_test

import (
	"context"
	"testing"

	"github.com/risor-io/risor/object"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/modules/shlex"
)

func TestArgv(t *testing.T) {
	actual := shlex.Argv(context.Background(), object.NewString("start --append=\"foobar foobaz\" --nogood 'food'"))
	require.IsType(t, &object.List{}, actual)

	expected := []string{"start", "--append=foobar foobaz", "--nogood", "food"}
	items := actual.(*object.List).Value()
	for i, item := range items {
		require.IsType(t, &object.String{}, item)
		require.Equal(t, expected[i], item.(*object.String).Value())
	}
}
