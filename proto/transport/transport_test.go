package transport_test

import (
	"testing"

	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/proto/transport"
)

func TestMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		input transport.Envelope
		want  transport.Envelope
	}{
		{
			name: "full envelope",
			input: transport.Envelope{
				JWT:       "header.payload.sig",
				WorkerID:  "worker1",
				Payload:   []byte("payload"),
				Signature: []byte("signature"),
			},
			want: transport.Envelope{
				JWT:       "header.payload.sig",
				WorkerID:  "worker1",
				Payload:   []byte("payload"),
				Signature: []byte("signature"),
			},
		},
		{
			name: "empty fields",
			input: transport.Envelope{
				JWT:       "",
				WorkerID:  "",
				Payload:   []byte{},
				Signature: []byte{},
			},
			want: transport.Envelope{
				JWT:       "",
				WorkerID:  "",
				Payload:   []byte{},
				Signature: []byte{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := transport.Marshal(tt.input)
			require.NoError(t, err)
			require.NotEmpty(t, marshaled)

			unmarshaled, err := transport.Unmarshal(marshaled)
			require.NoError(t, err)
			require.Equal(t, tt.want, unmarshaled)
		})
	}
}

func TestUnmarshalInvalidData(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "Empty input",
			input: []byte{},
		},
		{
			name:  "Invalid data",
			input: []byte("invalid data"),
		},
		{
			name:  "Nil input",
			input: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := transport.Unmarshal(tt.input)
			require.Error(t, err)
		})
	}
}

func TestUnmarshalInvalidArrayLength(t *testing.T) {
	// CBOR array of length 1 (0x81), which is not a valid envelope.
	_, err := transport.Unmarshal([]byte{0x81, 0x00})
	require.Error(t, err)
}

func TestUnmarshalTrailingBytes(t *testing.T) {
	marshaled, err := transport.Marshal(transport.Envelope{WorkerID: "worker1"})
	require.NoError(t, err)

	_, err = transport.Unmarshal(append(marshaled, 0xFF))
	require.Error(t, err)
}

func TestSignVerify(t *testing.T) {
	kp, err := nkeys.CreateUser()
	require.NoError(t, err)

	pub, err := kp.PublicKey()
	require.NoError(t, err)

	env := transport.Envelope{
		WorkerID: "worker1",
		Payload:  []byte("payload"),
	}

	err = env.Sign(kp)
	require.NoError(t, err)
	require.NotEmpty(t, env.Signature)

	// The signature survives a marshal/unmarshal round-trip.
	marshaled, err := transport.Marshal(env)
	require.NoError(t, err)

	unmarshaled, err := transport.Unmarshal(marshaled)
	require.NoError(t, err)

	err = unmarshaled.Verify(pub)
	require.NoError(t, err)
}

func TestVerifyTamper(t *testing.T) {
	kp, err := nkeys.CreateUser()
	require.NoError(t, err)

	pub, err := kp.PublicKey()
	require.NoError(t, err)

	base := transport.Envelope{
		WorkerID: "worker1",
		Payload:  []byte("payload"),
	}
	err = base.Sign(kp)
	require.NoError(t, err)

	tests := []struct {
		name   string
		mutate func(e *transport.Envelope)
	}{
		{
			name:   "tampered worker",
			mutate: func(e *transport.Envelope) { e.WorkerID = "worker2" },
		},
		{
			name:   "tampered payload",
			mutate: func(e *transport.Envelope) { e.Payload = []byte("other") },
		},
		{
			name:   "tampered signature",
			mutate: func(e *transport.Envelope) { e.Signature[0] ^= 0xFF },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := base
			env.Signature = append([]byte(nil), base.Signature...)
			tt.mutate(&env)

			err := env.Verify(pub)
			require.ErrorIs(t, err, transport.ErrInvalidSignature)
		})
	}
}

func TestVerifyWrongKey(t *testing.T) {
	kp, err := nkeys.CreateUser()
	require.NoError(t, err)

	other, err := nkeys.CreateUser()
	require.NoError(t, err)
	otherPub, err := other.PublicKey()
	require.NoError(t, err)

	env := transport.Envelope{WorkerID: "worker1", Payload: []byte("payload")}
	err = env.Sign(kp)
	require.NoError(t, err)

	err = env.Verify(otherPub)
	require.ErrorIs(t, err, transport.ErrInvalidSignature)
}

func TestVerifyInvalidPublicKey(t *testing.T) {
	env := transport.Envelope{WorkerID: "worker1", Signature: []byte("sig")}

	err := env.Verify("not-a-public-key")
	require.Error(t, err)
	require.NotErrorIs(t, err, transport.ErrInvalidSignature)
}
