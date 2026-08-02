package transport

import (
	"bytes"
	"errors"
	"io"

	"github.com/dtn7/cboring"
	"github.com/nats-io/nkeys"
)

// ErrInvalidSignature is returned by Verify when the signature does not match
// the envelope contents and the given public key.
var ErrInvalidSignature = errors.New("invalid signature")

// Envelope wraps an agent-Marshaled message for transports that carry no
// subject. It supplies the routing the receiver cannot otherwise deduce, plus
// the credential and signature that authenticate the sender.
type Envelope struct {
	// JWT is the sender's user JWT, issued by the account key. Its subject is
	// the agent's public key, against which Signature is verified. The JWT
	// itself must be validated by the receiver against the account public key.
	JWT string

	// WorkerID identifies the worker the message concerns. It is the only
	// routing token that cannot be derived from the JWT (agent) or the
	// receiver's own identity (gateway).
	WorkerID string

	// Payload is the opaque output of agent.Marshal. It is not decoded during
	// transport; the receiver forwards it verbatim once the envelope is
	// authenticated.
	Payload []byte

	// Signature is an Ed25519 signature over signedBytes, produced by the key
	// whose public part is the subject of JWT. It does not cover JWT itself:
	// altering the JWT changes the public key the signature is verified
	// against, which fails verification.
	Signature []byte
}

// MarshalCbor encodes the envelope as a CBOR array:
// [jwt, workerID, payload, signature].
func (m *Envelope) MarshalCbor(w io.Writer) error {
	err := cboring.WriteArrayLength(4, w)
	if err != nil {
		return err
	}

	err = cboring.WriteTextString(m.JWT, w)
	if err != nil {
		return err
	}

	err = cboring.WriteTextString(m.WorkerID, w)
	if err != nil {
		return err
	}

	err = cboring.WriteByteString(m.Payload, w)
	if err != nil {
		return err
	}

	return cboring.WriteByteString(m.Signature, w)
}

// UnmarshalCbor decodes the envelope from its CBOR representation.
func (m *Envelope) UnmarshalCbor(r io.Reader) error {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return err
	}
	if l != 4 {
		return errors.New("invalid message array length")
	}

	jwt, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}

	workerID, err := cboring.ReadTextString(r)
	if err != nil {
		return err
	}

	payload, err := cboring.ReadByteString(r)
	if err != nil {
		return err
	}

	signature, err := cboring.ReadByteString(r)
	if err != nil {
		return err
	}

	m.JWT = jwt
	m.WorkerID = workerID
	m.Payload = payload
	m.Signature = signature

	return nil
}

// Marshal serializes the envelope into a CBOR-encoded byte slice.
func Marshal(e Envelope) ([]byte, error) {
	var buf bytes.Buffer

	err := e.MarshalCbor(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal deserializes a CBOR-encoded envelope. It reports an error if the
// input contains trailing bytes after the envelope.
func Unmarshal(b []byte) (Envelope, error) {
	r := bytes.NewReader(b)

	var e Envelope
	err := e.UnmarshalCbor(r)
	if err != nil {
		return Envelope{}, err
	}

	if r.Len() != 0 {
		return Envelope{}, errors.New("unexpected trailing bytes")
	}

	return e, nil
}

// Sign computes the signature over the envelope's WorkerID and Payload
// using the given key pair and stores it in Signature. The key pair's public
// part must be the subject of JWT for the signature to later verify.
func (m *Envelope) Sign(kp nkeys.KeyPair) error {
	b, err := m.signedBytes()
	if err != nil {
		return err
	}

	sig, err := kp.Sign(b)
	if err != nil {
		return err
	}

	m.Signature = sig

	return nil
}

// Verify checks Signature against the envelope contents and the given public
// key, which the caller obtains from the separately validated JWT subject. It
// returns ErrInvalidSignature if the signature does not match.
func (m *Envelope) Verify(publicKey string) error {
	kp, err := nkeys.FromPublicKey(publicKey)
	if err != nil {
		return err
	}

	b, err := m.signedBytes()
	if err != nil {
		return err
	}

	err = kp.Verify(b, m.Signature)
	if err != nil {
		return ErrInvalidSignature
	}

	return nil
}

// signedBytes returns the canonical byte sequence covered by Signature: a CBOR
// array [workerID, payload]. Encoding it as a single framed array keeps
// the boundaries between fields unambiguous, so distinct field values cannot
// produce identical signed bytes. JWT and Signature are deliberately excluded
// (see Envelope.Signature).
func (m *Envelope) signedBytes() ([]byte, error) {
	var buf bytes.Buffer

	err := cboring.WriteArrayLength(2, &buf)
	if err != nil {
		return nil, err
	}

	err = cboring.WriteTextString(m.WorkerID, &buf)
	if err != nil {
		return nil, err
	}

	err = cboring.WriteByteString(m.Payload, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
