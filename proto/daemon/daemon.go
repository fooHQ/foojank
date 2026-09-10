// Package daemon provides functions for marshaling and unmarshaling messages.
package daemon

import (
	"bytes"
	"errors"
	"io"

	"github.com/dtn7/cboring"
)

var (
	// ErrUnknownType is returned by Marshal when the given value is not a
	// known message type.
	ErrUnknownType = errors.New("unknown message type")
	// ErrUnknownTag is returned by Unmarshal when the encoded type tag does
	// not correspond to a known message type.
	ErrUnknownTag = errors.New("unknown message tag")
)

// Payload type tags used to discriminate the message type on the wire.
const (
	tagCreateUserRequest uint64 = iota + 1
	tagCreateUserResponse
	tagGetUserRequest
	tagGetUserResponse
	tagListUsersRequest
	tagListUsersResponse
)

// Marshal serializes the given message into a CBOR-encoded byte slice. It
// accepts either a value or a pointer of a known message type.
//
// The message is encoded as a CBOR array of two elements:
//
//	[type tag (uint), payload]
func Marshal(message any) ([]byte, error) {
	var tag uint64
	var payload cboring.CborMarshaler

	switch v := message.(type) {
	case CreateUserRequest:
		tag, payload = tagCreateUserRequest, &v
	case *CreateUserRequest:
		tag, payload = tagCreateUserRequest, v
	case CreateUserResponse:
		tag, payload = tagCreateUserResponse, &v
	case *CreateUserResponse:
		tag, payload = tagCreateUserResponse, v
	case GetUserRequest:
		tag, payload = tagGetUserRequest, &v
	case *GetUserRequest:
		tag, payload = tagGetUserRequest, v
	case GetUserResponse:
		tag, payload = tagGetUserResponse, &v
	case *GetUserResponse:
		tag, payload = tagGetUserResponse, v
	case ListUsersRequest:
		tag, payload = tagListUsersRequest, &v
	case *ListUsersRequest:
		tag, payload = tagListUsersRequest, v
	case ListUsersResponse:
		tag, payload = tagListUsersResponse, &v
	case *ListUsersResponse:
		tag, payload = tagListUsersResponse, v
	default:
		return nil, ErrUnknownType
	}

	var buf bytes.Buffer

	err := cboring.WriteArrayLength(2, &buf)
	if err != nil {
		return nil, err
	}

	err = cboring.WriteUInt(tag, &buf)
	if err != nil {
		return nil, err
	}

	err = payload.MarshalCbor(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal deserializes the given CBOR-encoded byte slice into a message.
// The returned value can be type-asserted to the specific message type.
func Unmarshal(b []byte) (any, error) {
	r := bytes.NewReader(b)

	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return nil, err
	}
	if l != 2 {
		return nil, errors.New("invalid message array length")
	}

	tag, err := cboring.ReadUInt(r)
	if err != nil {
		return nil, err
	}

	var payload any
	switch tag {
	case tagCreateUserRequest:
		var m CreateUserRequest
		err = m.UnmarshalCbor(r)
		payload = m
	case tagCreateUserResponse:
		var m CreateUserResponse
		err = m.UnmarshalCbor(r)
		payload = m
	case tagGetUserRequest:
		var m GetUserRequest
		err = m.UnmarshalCbor(r)
		payload = m
	case tagGetUserResponse:
		var m GetUserResponse
		err = m.UnmarshalCbor(r)
		payload = m
	case tagListUsersRequest:
		var m ListUsersRequest
		err = m.UnmarshalCbor(r)
		payload = m
	case tagListUsersResponse:
		var m ListUsersResponse
		err = m.UnmarshalCbor(r)
		payload = m
	default:
		return nil, ErrUnknownTag
	}
	if err != nil {
		return nil, err
	}

	if r.Len() != 0 {
		return nil, errors.New("unexpected trailing bytes")
	}

	return payload, nil
}

// CreateUserSubject returns the NATS subject for creating a user.
func CreateUserSubject() string {
	return "FJ.DAEMON.RPC.USER.CREATE"
}

// GetUserSubject returns the NATS subject for retrieving a user.
func GetUserSubject() string {
	return "FJ.DAEMON.RPC.USER.GET"
}

// ListUsersSubject returns the NATS subject for listing users.
func ListUsersSubject() string {
	return "FJ.DAEMON.RPC.USER.LIST"
}

// writeStringSlice writes a slice of strings as a definite-length CBOR array.
func writeStringSlice(ss []string, w io.Writer) error {
	err := cboring.WriteArrayLength(uint64(len(ss)), w)
	if err != nil {
		return err
	}

	for _, s := range ss {
		err := cboring.WriteTextString(s, w)
		if err != nil {
			return err
		}
	}

	return nil
}

// readStringSlice reads a definite-length CBOR array of strings. An empty
// array is decoded as a nil slice.
func readStringSlice(r io.Reader) ([]string, error) {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return nil, err
	}

	if l == 0 {
		return nil, nil
	}

	ss := make([]string, l)
	for i := range ss {
		s, err := cboring.ReadTextString(r)
		if err != nil {
			return nil, err
		}
		ss[i] = s
	}

	return ss, nil
}

// writeUserSlice writes a slice of users as a definite-length CBOR array.
// Each user is encoded with User.MarshalCbor.
func writeUserSlice(users []User, w io.Writer) error {
	err := cboring.WriteArrayLength(uint64(len(users)), w)
	if err != nil {
		return err
	}

	for i := range users {
		err := users[i].MarshalCbor(w)
		if err != nil {
			return err
		}
	}

	return nil
}

// readUserSlice reads a definite-length CBOR array of users. An empty array
// is decoded as a nil slice.
func readUserSlice(r io.Reader) ([]User, error) {
	l, err := cboring.ReadArrayLength(r)
	if err != nil {
		return nil, err
	}

	if l == 0 {
		return nil, nil
	}

	users := make([]User, l)
	for i := range users {
		err := users[i].UnmarshalCbor(r)
		if err != nil {
			return nil, err
		}
	}

	return users, nil
}

// writeError writes an error as a CBOR text string. A nil error is encoded as
// an empty string.
func writeError(e error, w io.Writer) error {
	var msg string
	if e != nil {
		msg = e.Error()
	}
	return cboring.WriteTextString(msg, w)
}

// readError reads an error from a CBOR text string. An empty string is decoded
// as a nil error.
func readError(r io.Reader) (error, error) {
	msg, err := cboring.ReadTextString(r)
	if err != nil {
		return nil, err
	}

	if msg == "" {
		return nil, nil
	}

	return errors.New(msg), nil
}
