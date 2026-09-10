package directory

import (
	"errors"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
)

func TestTranslate(t *testing.T) {
	errOther := errors.New("other")

	tests := []struct {
		name string
		in   error
		want error
	}{
		{name: "nil", in: nil, want: nil},
		{name: "bucket exists", in: jetstream.ErrBucketExists, want: ErrStorageExists},
		{name: "no responders", in: nats.ErrNoResponders, want: ErrServiceUnavailable},
		{name: "no stream response", in: jetstream.ErrNoStreamResponse, want: ErrServiceUnavailable},
		{name: "key not found", in: jetstream.ErrKeyNotFound, want: ErrKeyNotFound},
		{name: "key exists", in: jetstream.ErrKeyExists, want: ErrKeyExists},
		{name: "invalid key", in: jetstream.ErrInvalidKey, want: ErrNameInvalid},
		{name: "passthrough", in: errOther, want: errOther},
		{name: "object not found is not mapped", in: jetstream.ErrObjectNotFound, want: jetstream.ErrObjectNotFound},
		{name: "bucket not found is not mapped", in: jetstream.ErrBucketNotFound, want: jetstream.ErrBucketNotFound},
		{name: "stream not found is not mapped", in: jetstream.ErrStreamNotFound, want: jetstream.ErrStreamNotFound},
		{name: "invalid store name is not mapped", in: jetstream.ErrInvalidStoreName, want: jetstream.ErrInvalidStoreName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translate(tt.in)
			if tt.want == nil {
				require.NoError(t, got)
				return
			}
			require.ErrorIs(t, got, tt.want)
		})
	}
}
