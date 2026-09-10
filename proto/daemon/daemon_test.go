package daemon_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/proto/daemon"
)

func TestMarshalUnmarshal(t *testing.T) {
	testError := errors.New("test error")
	testUser := daemon.User{
		Name:        "ops",
		ID:          "UDUMMYUSERPUBLICKEY",
		Description: "operations user",
		Kind:        "client",
		IssuedAt:    1700000000,
		ExpiresAt:   1800000000,
	}

	tests := []struct {
		name    string
		input   any
		want    any
		wantErr error
	}{
		{
			name: "CreateUserRequest",
			input: daemon.CreateUserRequest{
				Name:        "ops",
				Description: "operations user",
				Privileges:  []string{"pub", "sub"},
			},
			want: daemon.CreateUserRequest{
				Name:        "ops",
				Description: "operations user",
				Privileges:  []string{"pub", "sub"},
			},
		},
		{
			name: "CreateUserRequest with empty fields",
			input: daemon.CreateUserRequest{
				Privileges: []string{},
			},
			want: daemon.CreateUserRequest{
				Privileges: nil,
			},
		},
		{
			name: "CreateUserResponse",
			input: daemon.CreateUserResponse{
				JWT: "header.payload.sig",
				Key: "SUDUMMYUSERSEED",
			},
			want: daemon.CreateUserResponse{
				JWT: "header.payload.sig",
				Key: "SUDUMMYUSERSEED",
			},
		},
		{
			name:  "CreateUserResponse with error",
			input: daemon.CreateUserResponse{Error: testError},
			want:  daemon.CreateUserResponse{Error: testError},
		},
		{
			name: "GetUserRequest",
			input: daemon.GetUserRequest{
				Name: "ops",
			},
			want: daemon.GetUserRequest{
				Name: "ops",
			},
		},
		{
			name: "GetUserResponse",
			input: daemon.GetUserResponse{
				User: testUser,
			},
			want: daemon.GetUserResponse{
				User: testUser,
			},
		},
		{
			name:  "GetUserResponse with error",
			input: daemon.GetUserResponse{Error: testError},
			want:  daemon.GetUserResponse{Error: testError},
		},
		{
			name:  "ListUsersRequest",
			input: daemon.ListUsersRequest{},
			want:  daemon.ListUsersRequest{},
		},
		{
			name: "ListUsersResponse",
			input: daemon.ListUsersResponse{
				Users: []daemon.User{testUser},
			},
			want: daemon.ListUsersResponse{
				Users: []daemon.User{testUser},
			},
		},
		{
			name: "ListUsersResponse with empty slice",
			input: daemon.ListUsersResponse{
				Users: []daemon.User{},
			},
			want: daemon.ListUsersResponse{
				Users: nil,
			},
		},
		{
			name:  "ListUsersResponse with error",
			input: daemon.ListUsersResponse{Error: testError},
			want:  daemon.ListUsersResponse{Error: testError},
		},
		{
			name: "pointer input",
			input: &daemon.CreateUserRequest{
				Name: "ops",
			},
			want: daemon.CreateUserRequest{
				Name: "ops",
			},
		},
		{
			name:    "Unsupported type",
			input:   struct{}{},
			wantErr: daemon.ErrUnknownType,
		},
		{
			name:    "Nil input",
			input:   nil,
			wantErr: daemon.ErrUnknownType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := daemon.Marshal(tt.input)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, marshaled)

			unmarshaled, err := daemon.Unmarshal(marshaled)
			require.NoError(t, err)
			require.Equal(t, tt.want, unmarshaled)
		})
	}
}

func TestUnmarshalInvalidData(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "Empty input",
			input:   []byte{},
			wantErr: true,
		},
		{
			name:    "Invalid data",
			input:   []byte("invalid data"),
			wantErr: true,
		},
		{
			name:    "Nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := daemon.Unmarshal(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUnmarshalUnknownTag(t *testing.T) {
	// CBOR array of length 2 (0x82) with type tag 0 (0x00) and an empty
	// payload placeholder (0x00); tag 0 is not a known message type.
	_, err := daemon.Unmarshal([]byte{0x82, 0x00, 0x00})
	require.ErrorIs(t, err, daemon.ErrUnknownTag)
}

func TestUnmarshalInvalidArrayLength(t *testing.T) {
	// CBOR array of length 1 (0x81), which is not a valid message envelope.
	_, err := daemon.Unmarshal([]byte{0x81, 0x00})
	require.Error(t, err)
}

func TestUnmarshalTrailingBytes(t *testing.T) {
	marshaled, err := daemon.Marshal(daemon.ListUsersRequest{})
	require.NoError(t, err)

	_, err = daemon.Unmarshal(append(marshaled, 0xFF))
	require.Error(t, err)
}

func TestCreateUserSubject(t *testing.T) {
	got := daemon.CreateUserSubject()
	require.Equal(t, "FJ.DAEMON.RPC.USER.CREATE", got)
}

func TestGetUserSubject(t *testing.T) {
	got := daemon.GetUserSubject()
	require.Equal(t, "FJ.DAEMON.RPC.USER.GET", got)
}

func TestListUsersSubject(t *testing.T) {
	got := daemon.ListUsersSubject()
	require.Equal(t, "FJ.DAEMON.RPC.USER.LIST", got)
}
