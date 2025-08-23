package workmanager

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/foohq/foojank/internal/testutils"
)

func TestService(t *testing.T) {

	workerID := rand.Text()
	streamName := fmt.Sprintf("TEST-STREAM-%s", rand.Text())
	stdinName := fmt.Sprintf("TEST-STDIN-%s", rand.Text())
	stdoutName := fmt.Sprintf("TEST-STDOUT-%s", rand.Text())
	updateName := fmt.Sprintf("TEST-UPDATE-%s", rand.Text())
	_, js := testutils.NewJetStreamConnection(t)
}
