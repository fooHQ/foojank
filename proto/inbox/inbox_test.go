package inbox_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/proto/inbox"
)

func TestPrefix(t *testing.T) {
	got := inbox.Prefix("user1")
	require.Equal(t, "_INBOX.user1", got)
}
