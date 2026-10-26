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
			req:  protodaemon.GetUserRequest{UserRef: "alice"},
		},
		{
			name: "valid id",
			req:  protodaemon.GetUserRequest{UserRef: testUserID(t)},
		},
		{
			name:    "empty ref",
			req:     protodaemon.GetUserRequest{},
			wantErr: "ref is required",
		},
		{
			name:   "invalid name",
			req:    protodaemon.GetUserRequest{UserRef: "alice bob"},
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
		Name:       "agent",
		GatewayRef: "gw",
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
				req.GatewayRef = ""
				return req
			}(),
			wantErr: "ref is required",
		},
		{
			name: "invalid gateway",
			req: func() protodaemon.CreateAgentRequest {
				req := valid
				req.GatewayRef = "bad name"
				return req
			}(),
			wantIs: directory.ErrNameInvalid,
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
		{
			name: "valid with commands",
			req: func() protodaemon.CreateAgentRequest {
				req := valid
				req.Commands = map[string]protodaemon.AgentCommand{
					"ls": {
						Description: "List a directory",
						Flags: map[string]protodaemon.AgentCommandFlag{
							"recursive": {
								Description: "Recurse into directories",
								Default:     "false",
								Required:    true,
							},
						},
						Environment: map[string]protodaemon.AgentCommandVar{
							"DEPTH": {
								Value:       "2",
								Description: "Max recursion depth",
							},
						},
					},
				}
				return req
			}(),
		},
		{
			name: "valid with empty commands",
			req: func() protodaemon.CreateAgentRequest {
				req := valid
				req.Commands = map[string]protodaemon.AgentCommand{}
				return req
			}(),
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
	err := handler.ValidateGetAgentRequest(protodaemon.GetAgentRequest{AgentRef: "agent"})
	require.NoError(t, err)

	err = handler.ValidateGetAgentRequest(protodaemon.GetAgentRequest{})
	require.EqualError(t, err, "ref is required")

	err = handler.ValidateGetAgentRequest(protodaemon.GetAgentRequest{AgentRef: "bad name"})
	require.ErrorIs(t, err, directory.ErrNameInvalid)
}

func TestValidateGetCommandRequest(t *testing.T) {
	err := handler.ValidateGetCommandRequest(protodaemon.GetCommandRequest{
		AgentRef: "agent",
		Name:     "ls",
	})
	require.NoError(t, err)

	err = handler.ValidateGetCommandRequest(protodaemon.GetCommandRequest{Name: "ls"})
	require.EqualError(t, err, "ref is required")

	err = handler.ValidateGetCommandRequest(protodaemon.GetCommandRequest{AgentRef: "bad name", Name: "ls"})
	require.ErrorIs(t, err, directory.ErrNameInvalid)

	err = handler.ValidateGetCommandRequest(protodaemon.GetCommandRequest{AgentRef: "agent"})
	require.EqualError(t, err, "name is required")

	err = handler.ValidateGetCommandRequest(protodaemon.GetCommandRequest{AgentRef: "agent", Name: "bad name"})
	require.ErrorIs(t, err, directory.ErrNameInvalid)
}

func TestValidateListCommandsRequest(t *testing.T) {
	err := handler.ValidateListCommandsRequest(protodaemon.ListCommandsRequest{})
	require.EqualError(t, err, "ref is required")

	err = handler.ValidateListCommandsRequest(protodaemon.ListCommandsRequest{AgentRef: "agent"})
	require.NoError(t, err)

	err = handler.ValidateListCommandsRequest(protodaemon.ListCommandsRequest{AgentRef: "bad name"})
	require.ErrorIs(t, err, directory.ErrNameInvalid)
}

func TestValidateDeleteAgentRequest(t *testing.T) {
	err := handler.ValidateDeleteAgentRequest(protodaemon.DeleteAgentRequest{AgentRef: "agent"})
	require.NoError(t, err)

	err = handler.ValidateDeleteAgentRequest(protodaemon.DeleteAgentRequest{})
	require.EqualError(t, err, "ref is required")

	err = handler.ValidateDeleteAgentRequest(protodaemon.DeleteAgentRequest{AgentRef: "bad name"})
	require.ErrorIs(t, err, directory.ErrNameInvalid)
}

func TestValidateGetGatewayRequest(t *testing.T) {
	err := handler.ValidateGetGatewayRequest(protodaemon.GetGatewayRequest{GatewayRef: "gw1"})
	require.NoError(t, err)

	err = handler.ValidateGetGatewayRequest(protodaemon.GetGatewayRequest{})
	require.EqualError(t, err, "ref is required")

	err = handler.ValidateGetGatewayRequest(protodaemon.GetGatewayRequest{GatewayRef: "bad name"})
	require.ErrorIs(t, err, directory.ErrNameInvalid)
}

func TestValidateDeleteGatewayRequest(t *testing.T) {
	err := handler.ValidateDeleteGatewayRequest(protodaemon.DeleteGatewayRequest{GatewayRef: "gw1"})
	require.NoError(t, err)

	err = handler.ValidateDeleteGatewayRequest(protodaemon.DeleteGatewayRequest{})
	require.EqualError(t, err, "ref is required")

	err = handler.ValidateDeleteGatewayRequest(protodaemon.DeleteGatewayRequest{GatewayRef: "bad name"})
	require.ErrorIs(t, err, directory.ErrNameInvalid)
}

func TestValidateCreateGatewayRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     protodaemon.CreateGatewayRequest
		wantErr string
		wantIs  error
	}{
		{
			name: "valid",
			req:  protodaemon.CreateGatewayRequest{Name: "gw1"},
		},
		{
			name:    "empty name",
			req:     protodaemon.CreateGatewayRequest{},
			wantErr: "name is required",
		},
		{
			name:   "invalid name",
			req:    protodaemon.CreateGatewayRequest{Name: "bad name"},
			wantIs: directory.ErrNameInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateCreateGatewayRequest(tt.req)
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

func TestValidateCreateRoleRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     protodaemon.CreateRoleRequest
		wantErr string
		wantIs  error
	}{
		{
			name: "valid",
			req:  protodaemon.CreateRoleRequest{Name: "operator"},
		},
		{
			name:    "empty name",
			req:     protodaemon.CreateRoleRequest{},
			wantErr: "name is required",
		},
		{
			name:   "invalid name",
			req:    protodaemon.CreateRoleRequest{Name: "bad name"},
			wantIs: directory.ErrNameInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateCreateRoleRequest(tt.req)
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

func TestValidateGetRoleRequest(t *testing.T) {
	err := handler.ValidateGetRoleRequest(protodaemon.GetRoleRequest{RoleRef: "operator"})
	require.NoError(t, err)

	err = handler.ValidateGetRoleRequest(protodaemon.GetRoleRequest{})
	require.EqualError(t, err, "ref is required")

	err = handler.ValidateGetRoleRequest(protodaemon.GetRoleRequest{RoleRef: "bad name"})
	require.ErrorIs(t, err, directory.ErrNameInvalid)
}

func TestValidateUpdateUserRequest(t *testing.T) {
	err := handler.ValidateUpdateUserRequest(protodaemon.UpdateUserRequest{UserRef: "alice"})
	require.NoError(t, err)

	err = handler.ValidateUpdateUserRequest(protodaemon.UpdateUserRequest{})
	require.EqualError(t, err, "ref is required")

	err = handler.ValidateUpdateUserRequest(protodaemon.UpdateUserRequest{UserRef: "bad name"})
	require.ErrorIs(t, err, directory.ErrNameInvalid)
}

func TestValidateDeleteUserRequest(t *testing.T) {
	err := handler.ValidateDeleteUserRequest(protodaemon.DeleteUserRequest{UserRef: "alice"})
	require.NoError(t, err)

	err = handler.ValidateDeleteUserRequest(protodaemon.DeleteUserRequest{})
	require.EqualError(t, err, "ref is required")

	err = handler.ValidateDeleteUserRequest(protodaemon.DeleteUserRequest{UserRef: "bad name"})
	require.ErrorIs(t, err, directory.ErrNameInvalid)
}

func TestValidateDeleteRoleRequest(t *testing.T) {
	err := handler.ValidateDeleteRoleRequest(protodaemon.DeleteRoleRequest{RoleRef: "operator"})
	require.NoError(t, err)

	err = handler.ValidateDeleteRoleRequest(protodaemon.DeleteRoleRequest{})
	require.EqualError(t, err, "ref is required")

	err = handler.ValidateDeleteRoleRequest(protodaemon.DeleteRoleRequest{RoleRef: "bad name"})
	require.ErrorIs(t, err, directory.ErrNameInvalid)
}

func TestValidateCreateJobRequest(t *testing.T) {
	valid := protodaemon.CreateJobRequest{
		AgentRef: "agent",
		Command:  "echo",
	}

	tests := []struct {
		name    string
		req     protodaemon.CreateJobRequest
		wantErr string
	}{
		{
			name: "valid",
			req:  valid,
		},
		{
			name: "valid with args and env",
			req: protodaemon.CreateJobRequest{
				AgentRef: "agent",
				Command:  "echo",
				Args:     []string{"hello"},
				Env:      []string{"A=b"},
			},
		},
		{
			name: "invalid agent ref",
			req: protodaemon.CreateJobRequest{
				AgentRef: "bad name",
				Command:  "echo",
			},
			wantErr: "invalid name",
		},
		{
			name: "empty agent ref",
			req: protodaemon.CreateJobRequest{
				Command: "echo",
			},
			wantErr: "ref is required",
		},
		{
			name: "empty command",
			req: protodaemon.CreateJobRequest{
				AgentRef: "agent",
			},
			wantErr: "command is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateCreateJobRequest(tt.req)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestValidateGetJobRequest(t *testing.T) {
	err := handler.ValidateGetJobRequest(protodaemon.GetJobRequest{JobRef: "job1"})
	require.NoError(t, err)

	err = handler.ValidateGetJobRequest(protodaemon.GetJobRequest{})
	require.EqualError(t, err, "ref is required")

	err = handler.ValidateGetJobRequest(protodaemon.GetJobRequest{JobRef: "bad id"})
	require.ErrorIs(t, err, directory.ErrNameInvalid)
}

func TestValidateCancelJobRequest(t *testing.T) {
	err := handler.ValidateCancelJobRequest(protodaemon.CancelJobRequest{JobRef: "job1"})
	require.NoError(t, err)

	err = handler.ValidateCancelJobRequest(protodaemon.CancelJobRequest{})
	require.EqualError(t, err, "ref is required")

	err = handler.ValidateCancelJobRequest(protodaemon.CancelJobRequest{JobRef: "bad id"})
	require.ErrorIs(t, err, directory.ErrNameInvalid)
}

func TestValidateDeleteJobRequest(t *testing.T) {
	err := handler.ValidateDeleteJobRequest(protodaemon.DeleteJobRequest{JobRef: "job1"})
	require.NoError(t, err)

	err = handler.ValidateDeleteJobRequest(protodaemon.DeleteJobRequest{})
	require.EqualError(t, err, "ref is required")

	err = handler.ValidateDeleteJobRequest(protodaemon.DeleteJobRequest{JobRef: "bad id"})
	require.ErrorIs(t, err, directory.ErrNameInvalid)
}

func TestValidateGetJobOutputRequest(t *testing.T) {
	err := handler.ValidateGetJobOutputRequest(protodaemon.GetJobOutputRequest{JobRef: "job1", Offset: 12})
	require.NoError(t, err)

	err = handler.ValidateGetJobOutputRequest(protodaemon.GetJobOutputRequest{})
	require.EqualError(t, err, "ref is required")

	err = handler.ValidateGetJobOutputRequest(protodaemon.GetJobOutputRequest{JobRef: "bad id"})
	require.ErrorIs(t, err, directory.ErrNameInvalid)
}

func TestValidateListJobsRequest(t *testing.T) {
	err := handler.ValidateListJobsRequest(protodaemon.ListJobsRequest{})
	require.NoError(t, err)

	err = handler.ValidateListJobsRequest(protodaemon.ListJobsRequest{AgentRef: "agent"})
	require.NoError(t, err)

	err = handler.ValidateListJobsRequest(protodaemon.ListJobsRequest{AgentRef: "bad name"})
	require.ErrorIs(t, err, directory.ErrNameInvalid)
}
