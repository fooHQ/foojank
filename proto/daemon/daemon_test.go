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
		JWT:         "header.payload.sig",
		Privileges:  []string{"pub", "sub"},
		CreatedAt:   1700000000,
		ExpiresAt:   1800000000,
	}
	testAgent := daemon.Agent{
		ID:          "UDUMMYAGENTPUBLICKEY",
		Name:        "web",
		Description: "web agent",
		GatewayID:   "UDUMMYGATEWAYPUBLICKEY",
		Config: daemon.AgentConfig{
			OS:    "linux",
			Arch:  "amd64",
			JWT:   "header.payload.sig",
			Key:   "SUDUMMYAGENTSEED",
			Extra: map[string]string{"key1": "val1", "key2": "val2"},
		},
		CreatedAt: 1700000000,
	}
	testGateway := daemon.Gateway{
		ID:          "UDUMMYGATEWAYPUBLICKEY",
		Name:        "gw1",
		Description: "edge gateway",
		Config: daemon.GatewayConfig{
			JWT:   "header.payload.sig",
			Key:   "SUDUMMYGATEWAYSEED",
			Extra: map[string]string{"key1": "val1", "key2": "val2"},
		},
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
				ID:          "UDUMMYUSERPUBLICKEY",
				Name:        "ops",
				Description: "operations user",
				Privileges:  []string{"pub", "sub"},
			},
			want: daemon.CreateUserRequest{
				ID:          "UDUMMYUSERPUBLICKEY",
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
			name:  "CreateUserResponse",
			input: daemon.CreateUserResponse{},
			want:  daemon.CreateUserResponse{},
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
			name:  "IssueJWTRequest",
			input: daemon.IssueJWTRequest{},
			want:  daemon.IssueJWTRequest{},
		},
		{
			name: "IssueJWTResponse",
			input: daemon.IssueJWTResponse{
				JWT: "header.payload.sig",
			},
			want: daemon.IssueJWTResponse{
				JWT: "header.payload.sig",
			},
		},
		{
			name:  "IssueJWTResponse with error",
			input: daemon.IssueJWTResponse{Error: testError},
			want:  daemon.IssueJWTResponse{Error: testError},
		},
		{
			name: "CreateAgentRequest",
			input: daemon.CreateAgentRequest{
				Name:        "web",
				Description: "web agent",
				Gateway:     "gw1",
				Config: daemon.AgentConfig{
					OS:    "linux",
					Arch:  "amd64",
					JWT:   "header.payload.sig",
					Key:   "SUDUMMYAGENTSEED",
					Extra: map[string]string{"key1": "val1", "key2": "val2"},
				},
			},
			want: daemon.CreateAgentRequest{
				Name:        "web",
				Description: "web agent",
				Gateway:     "gw1",
				Config: daemon.AgentConfig{
					OS:    "linux",
					Arch:  "amd64",
					JWT:   "header.payload.sig",
					Key:   "SUDUMMYAGENTSEED",
					Extra: map[string]string{"key1": "val1", "key2": "val2"},
				},
			},
		},
		{
			name: "CreateAgentRequest with empty fields",
			input: daemon.CreateAgentRequest{
				Config: daemon.AgentConfig{
					Extra: map[string]string{},
				},
			},
			want: daemon.CreateAgentRequest{
				Config: daemon.AgentConfig{
					Extra: nil,
				},
			},
		},
		{
			name:  "CreateAgentResponse",
			input: daemon.CreateAgentResponse{},
			want:  daemon.CreateAgentResponse{},
		},
		{
			name:  "CreateAgentResponse with error",
			input: daemon.CreateAgentResponse{Error: testError},
			want:  daemon.CreateAgentResponse{Error: testError},
		},
		{
			name: "GetAgentRequest",
			input: daemon.GetAgentRequest{
				Name: "web",
			},
			want: daemon.GetAgentRequest{
				Name: "web",
			},
		},
		{
			name: "GetAgentResponse",
			input: daemon.GetAgentResponse{
				Agent: testAgent,
			},
			want: daemon.GetAgentResponse{
				Agent: testAgent,
			},
		},
		{
			name:  "GetAgentResponse with error",
			input: daemon.GetAgentResponse{Error: testError},
			want:  daemon.GetAgentResponse{Error: testError},
		},
		{
			name:  "ListAgentsRequest",
			input: daemon.ListAgentsRequest{},
			want:  daemon.ListAgentsRequest{},
		},
		{
			name: "ListAgentsResponse",
			input: daemon.ListAgentsResponse{
				Agents: []daemon.Agent{testAgent},
			},
			want: daemon.ListAgentsResponse{
				Agents: []daemon.Agent{testAgent},
			},
		},
		{
			name: "ListAgentsResponse with empty slice",
			input: daemon.ListAgentsResponse{
				Agents: []daemon.Agent{},
			},
			want: daemon.ListAgentsResponse{
				Agents: nil,
			},
		},
		{
			name:  "ListAgentsResponse with error",
			input: daemon.ListAgentsResponse{Error: testError},
			want:  daemon.ListAgentsResponse{Error: testError},
		},
		{
			name: "CreateGatewayRequest",
			input: daemon.CreateGatewayRequest{
				Name:        "gw1",
				Description: "edge gateway",
				Config: daemon.GatewayConfig{
					JWT:   "header.payload.sig",
					Key:   "SUDUMMYGATEWAYSEED",
					Extra: map[string]string{"key1": "val1", "key2": "val2"},
				},
			},
			want: daemon.CreateGatewayRequest{
				Name:        "gw1",
				Description: "edge gateway",
				Config: daemon.GatewayConfig{
					JWT:   "header.payload.sig",
					Key:   "SUDUMMYGATEWAYSEED",
					Extra: map[string]string{"key1": "val1", "key2": "val2"},
				},
			},
		},
		{
			name: "CreateGatewayRequest with empty fields",
			input: daemon.CreateGatewayRequest{
				Config: daemon.GatewayConfig{
					Extra: map[string]string{},
				},
			},
			want: daemon.CreateGatewayRequest{
				Config: daemon.GatewayConfig{
					Extra: nil,
				},
			},
		},
		{
			name:  "CreateGatewayResponse",
			input: daemon.CreateGatewayResponse{},
			want:  daemon.CreateGatewayResponse{},
		},
		{
			name:  "CreateGatewayResponse with error",
			input: daemon.CreateGatewayResponse{Error: testError},
			want:  daemon.CreateGatewayResponse{Error: testError},
		},
		{
			name: "GetGatewayRequest",
			input: daemon.GetGatewayRequest{
				Name: "gw1",
			},
			want: daemon.GetGatewayRequest{
				Name: "gw1",
			},
		},
		{
			name: "GetGatewayResponse",
			input: daemon.GetGatewayResponse{
				Gateway: testGateway,
			},
			want: daemon.GetGatewayResponse{
				Gateway: testGateway,
			},
		},
		{
			name:  "GetGatewayResponse with error",
			input: daemon.GetGatewayResponse{Error: testError},
			want:  daemon.GetGatewayResponse{Error: testError},
		},
		{
			name:  "ListGatewaysRequest",
			input: daemon.ListGatewaysRequest{},
			want:  daemon.ListGatewaysRequest{},
		},
		{
			name: "ListGatewaysResponse",
			input: daemon.ListGatewaysResponse{
				Gateways: []daemon.Gateway{testGateway},
			},
			want: daemon.ListGatewaysResponse{
				Gateways: []daemon.Gateway{testGateway},
			},
		},
		{
			name: "ListGatewaysResponse with empty slice",
			input: daemon.ListGatewaysResponse{
				Gateways: []daemon.Gateway{},
			},
			want: daemon.ListGatewaysResponse{
				Gateways: nil,
			},
		},
		{
			name:  "ListGatewaysResponse with error",
			input: daemon.ListGatewaysResponse{Error: testError},
			want:  daemon.ListGatewaysResponse{Error: testError},
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

func TestIssueJWTSubject(t *testing.T) {
	got := daemon.IssueJWTSubject("UDUMMYUSERPUBLICKEY")
	require.Equal(t, "FJ.DAEMON.RPC.JWT.UDUMMYUSERPUBLICKEY.ISSUE", got)
}

func TestCreateAgentSubject(t *testing.T) {
	got := daemon.CreateAgentSubject()
	require.Equal(t, "FJ.DAEMON.RPC.AGENT.CREATE", got)
}

func TestGetAgentSubject(t *testing.T) {
	got := daemon.GetAgentSubject()
	require.Equal(t, "FJ.DAEMON.RPC.AGENT.GET", got)
}

func TestListAgentsSubject(t *testing.T) {
	got := daemon.ListAgentsSubject()
	require.Equal(t, "FJ.DAEMON.RPC.AGENT.LIST", got)
}

func TestCreateGatewaySubject(t *testing.T) {
	got := daemon.CreateGatewaySubject()
	require.Equal(t, "FJ.DAEMON.RPC.GATEWAY.CREATE", got)
}

func TestGetGatewaySubject(t *testing.T) {
	got := daemon.GetGatewaySubject()
	require.Equal(t, "FJ.DAEMON.RPC.GATEWAY.GET", got)
}

func TestListGatewaysSubject(t *testing.T) {
	got := daemon.ListGatewaysSubject()
	require.Equal(t, "FJ.DAEMON.RPC.GATEWAY.LIST", got)
}

func TestParseIssueJWTSubject(t *testing.T) {
	userID, ok := daemon.ParseIssueJWTSubject(daemon.IssueJWTSubject("UDUMMYUSERPUBLICKEY"))
	require.True(t, ok)
	require.Equal(t, "UDUMMYUSERPUBLICKEY", userID)
}

func TestParseIssueJWTSubjectInvalid(t *testing.T) {
	tests := []struct {
		name    string
		subject string
	}{
		{
			name:    "empty",
			subject: "",
		},
		{
			name:    "too short",
			subject: "FJ.DAEMON.RPC.JWT.ISSUE",
		},
		{
			name:    "wrong prefix",
			subject: "XX.DAEMON.RPC.JWT.UDUMMYUSERPUBLICKEY.ISSUE",
		},
		{
			name:    "user subject",
			subject: daemon.CreateUserSubject(),
		},
		{
			name:    "empty user id",
			subject: "FJ.DAEMON.RPC.JWT..ISSUE",
		},
		{
			name:    "wrong verb",
			subject: "FJ.DAEMON.RPC.JWT.UDUMMYUSERPUBLICKEY.GET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := daemon.ParseIssueJWTSubject(tt.subject)
			require.False(t, ok)
		})
	}
}
