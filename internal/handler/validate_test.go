package handler_test

import (
	"testing"

	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/directory"
	"github.com/foohq/foojank/internal/handler"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

func testUserID(t *testing.T) string {
	t.Helper()
	kp, err := nkeys.CreateUser()
	require.NoError(t, err)
	id, err := kp.PublicKey()
	require.NoError(t, err)
	return id
}

func TestValidateCreateUserRequest(t *testing.T) {
	id := testUserID(t)

	tests := []struct {
		name    string
		req     protodaemon.CreateUserRequest
		wantErr string
		wantIs  error
	}{
		{
			name: "valid",
			req: protodaemon.CreateUserRequest{
				ID:   id,
				Name: "alice",
			},
		},
		{
			name: "invalid user id",
			req: protodaemon.CreateUserRequest{
				ID:   "alice",
				Name: "alice",
			},
			wantErr: "invalid user id",
		},
		{
			name: "empty user id",
			req: protodaemon.CreateUserRequest{
				Name: "alice",
			},
			wantErr: "invalid user id",
		},
		{
			name: "empty name",
			req: protodaemon.CreateUserRequest{
				ID: id,
			},
			wantErr: "name is required",
		},
		{
			name: "invalid name",
			req: protodaemon.CreateUserRequest{
				ID:   id,
				Name: "alice bob",
			},
			wantIs: directory.ErrNameInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateCreateUserRequest(tt.req)
			if tt.wantErr == "" && tt.wantIs == nil {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			}
			if tt.wantIs != nil {
				require.ErrorIs(t, err, tt.wantIs)
			}
		})
	}
}

func TestValidateGetUserRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     protodaemon.GetUserRequest
		wantErr string
		wantIs  error
	}{
		{
			name: "valid name",
			req:  protodaemon.GetUserRequest{Name: "alice"},
		},
		{
			name: "valid id",
			req:  protodaemon.GetUserRequest{Name: testUserID(t)},
		},
		{
			name:    "empty name",
			req:     protodaemon.GetUserRequest{},
			wantErr: "name is required",
		},
		{
			name:   "invalid name",
			req:    protodaemon.GetUserRequest{Name: "alice bob"},
			wantIs: directory.ErrNameInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateGetUserRequest(tt.req)
			if tt.wantErr == "" && tt.wantIs == nil {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			}
			if tt.wantIs != nil {
				require.ErrorIs(t, err, tt.wantIs)
			}
		})
	}
}

func TestValidateCreateAgentRequest(t *testing.T) {
	valid := protodaemon.CreateAgentRequest{
		Name:    "agent",
		Gateway: "gw",
		Config: protodaemon.AgentConfig{
			OS:   "linux",
			Arch: "amd64",
		},
	}

	tests := []struct {
		name    string
		req     protodaemon.CreateAgentRequest
		wantErr string
		wantIs  error
	}{
		{
			name: "valid",
			req:  valid,
		},
		{
			name: "empty name",
			req: func() protodaemon.CreateAgentRequest {
				req := valid
				req.Name = ""
				return req
			}(),
			wantErr: "name is required",
		},
		{
			name: "invalid name",
			req: func() protodaemon.CreateAgentRequest {
				req := valid
				req.Name = "bad name"
				return req
			}(),
			wantIs: directory.ErrNameInvalid,
		},
		{
			name: "empty gateway",
			req: func() protodaemon.CreateAgentRequest {
				req := valid
				req.Gateway = ""
				return req
			}(),
			wantErr: "gateway is required",
		},
		{
			name: "empty os",
			req: func() protodaemon.CreateAgentRequest {
				req := valid
				req.Config.OS = ""
				return req
			}(),
			wantErr: "os is required",
		},
		{
			name: "empty arch",
			req: func() protodaemon.CreateAgentRequest {
				req := valid
				req.Config.Arch = ""
				return req
			}(),
			wantErr: "arch is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateCreateAgentRequest(tt.req)
			if tt.wantErr == "" && tt.wantIs == nil {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			}
			if tt.wantIs != nil {
				require.ErrorIs(t, err, tt.wantIs)
			}
		})
	}
}

func TestValidateGetAgentRequest(t *testing.T) {
	err := handler.ValidateGetAgentRequest(protodaemon.GetAgentRequest{Name: "agent"})
	require.NoError(t, err)

	err = handler.ValidateGetAgentRequest(protodaemon.GetAgentRequest{})
	require.EqualError(t, err, "name is required")

	err = handler.ValidateGetAgentRequest(protodaemon.GetAgentRequest{Name: "bad name"})
	require.ErrorIs(t, err, directory.ErrNameInvalid)
}
