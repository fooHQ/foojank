package gateway_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/proto/gateway"
)

func TestMarshalUnmarshal(t *testing.T) {
	testError := errors.New("test error")

	tests := []struct {
		name    string
		input   any
		want    any
		wantErr error
	}{
		{
			name: "RegisterAgentRequest",
			input: gateway.RegisterAgentRequest{
				AgentID: "agent1",
				OS:      "linux",
				Arch:    "amd64",
				Env:     map[string]string{"key1": "val1", "key2": "val2"},
			},
			want: gateway.RegisterAgentRequest{
				AgentID: "agent1",
				OS:      "linux",
				Arch:    "amd64",
				Env:     map[string]string{"key1": "val1", "key2": "val2"},
			},
		},
		{
			name: "RegisterAgentRequest with empty fields",
			input: gateway.RegisterAgentRequest{
				AgentID: "",
				Env:     map[string]string{},
			},
			want: gateway.RegisterAgentRequest{
				AgentID: "",
				Env:     nil,
			},
		},
		{
			name: "RegisterAgentResponse",
			input: gateway.RegisterAgentResponse{
				Env: map[string]string{"key1": "val1"},
			},
			want: gateway.RegisterAgentResponse{
				Env: map[string]string{"key1": "val1"},
			},
		},
		{
			name:  "RegisterAgentResponse with error",
			input: gateway.RegisterAgentResponse{Error: testError},
			want:  gateway.RegisterAgentResponse{Error: testError},
		},
		{
			name: "UnregisterAgentRequest",
			input: gateway.UnregisterAgentRequest{
				AgentID: "agent1",
			},
			want: gateway.UnregisterAgentRequest{
				AgentID: "agent1",
			},
		},
		{
			name:  "UnregisterAgentResponse",
			input: gateway.UnregisterAgentResponse{},
			want:  gateway.UnregisterAgentResponse{},
		},
		{
			name:  "UnregisterAgentResponse with error",
			input: gateway.UnregisterAgentResponse{Error: testError},
			want:  gateway.UnregisterAgentResponse{Error: testError},
		},
		{
			name: "CreateJobRequest",
			input: gateway.CreateJobRequest{
				AgentID: "agent1",
				Command: "echo",
				Args:    []string{"hello"},
				Env:     []string{"A=1", "B=2"},
			},
			want: gateway.CreateJobRequest{
				AgentID: "agent1",
				Command: "echo",
				Args:    []string{"hello"},
				Env:     []string{"A=1", "B=2"},
			},
		},
		{
			name: "CreateJobRequest with empty slices",
			input: gateway.CreateJobRequest{
				AgentID: "agent1",
				Args:    []string{},
				Env:     []string{},
			},
			want: gateway.CreateJobRequest{
				AgentID: "agent1",
			},
		},
		{
			name: "CreateJobResponse",
			input: gateway.CreateJobResponse{
				JobID: "job1",
			},
			want: gateway.CreateJobResponse{
				JobID: "job1",
			},
		},
		{
			name:  "CreateJobResponse with error",
			input: gateway.CreateJobResponse{Error: testError},
			want:  gateway.CreateJobResponse{Error: testError},
		},
		{
			name: "CancelJobRequest",
			input: gateway.CancelJobRequest{
				AgentID: "agent1",
				JobID:   "job1",
			},
			want: gateway.CancelJobRequest{
				AgentID: "agent1",
				JobID:   "job1",
			},
		},
		{
			name:  "CancelJobResponse",
			input: gateway.CancelJobResponse{},
			want:  gateway.CancelJobResponse{},
		},
		{
			name:  "CancelJobResponse with error",
			input: gateway.CancelJobResponse{Error: testError},
			want:  gateway.CancelJobResponse{Error: testError},
		},
		{
			name: "JobInputRequest",
			input: gateway.JobInputRequest{
				AgentID: "agent1",
				JobID:   "job1",
				Data:    []byte("hello"),
			},
			want: gateway.JobInputRequest{
				AgentID: "agent1",
				JobID:   "job1",
				Data:    []byte("hello"),
			},
		},
		{
			name: "JobInputRequest with empty data",
			input: gateway.JobInputRequest{
				AgentID: "agent1",
				JobID:   "job1",
			},
			want: gateway.JobInputRequest{
				AgentID: "agent1",
				JobID:   "job1",
				Data:    []byte{},
			},
		},
		{
			name:  "JobInputResponse",
			input: gateway.JobInputResponse{},
			want:  gateway.JobInputResponse{},
		},
		{
			name:  "JobInputResponse with error",
			input: gateway.JobInputResponse{Error: testError},
			want:  gateway.JobInputResponse{Error: testError},
		},
		{
			name: "JobOutputRequest",
			input: gateway.JobOutputRequest{
				AgentID: "agent1",
				JobID:   "job1",
				Offset:  12,
			},
			want: gateway.JobOutputRequest{
				AgentID: "agent1",
				JobID:   "job1",
				Offset:  12,
			},
		},
		{
			name: "JobOutputResponse",
			input: gateway.JobOutputResponse{
				Data: []byte("hello"),
			},
			want: gateway.JobOutputResponse{
				Data: []byte("hello"),
			},
		},
		{
			name:  "JobOutputResponse with empty data",
			input: gateway.JobOutputResponse{},
			want:  gateway.JobOutputResponse{Data: []byte{}},
		},
		{
			name: "JobOutputResponse with error",
			input: gateway.JobOutputResponse{
				Error: testError,
			},
			want: gateway.JobOutputResponse{
				Data:  []byte{},
				Error: testError,
			},
		},
		{
			name: "pointer input",
			input: &gateway.RegisterAgentRequest{
				AgentID: "agent1",
			},
			want: gateway.RegisterAgentRequest{
				AgentID: "agent1",
			},
		},
		{
			name:    "Unsupported type",
			input:   struct{}{},
			wantErr: gateway.ErrUnknownType,
		},
		{
			name:    "Nil input",
			input:   nil,
			wantErr: gateway.ErrUnknownType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Marshal
			marshaled, err := gateway.Marshal(tt.input)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, marshaled)

			// Test Unmarshal
			unmarshaled, err := gateway.Unmarshal(marshaled)
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
			_, err := gateway.Unmarshal(tt.input)
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
	_, err := gateway.Unmarshal([]byte{0x82, 0x00, 0x00})
	require.ErrorIs(t, err, gateway.ErrUnknownTag)
}

func TestUnmarshalInvalidArrayLength(t *testing.T) {
	// CBOR array of length 1 (0x81), which is not a valid message envelope.
	_, err := gateway.Unmarshal([]byte{0x81, 0x00})
	require.Error(t, err)
}

func TestUnmarshalTrailingBytes(t *testing.T) {
	marshaled, err := gateway.Marshal(gateway.RegisterAgentRequest{AgentID: "agent1"})
	require.NoError(t, err)

	_, err = gateway.Unmarshal(append(marshaled, 0xFF))
	require.Error(t, err)
}

func TestRegisterAgentSubject(t *testing.T) {
	got := gateway.RegisterAgentSubject("gateway1")
	require.Equal(t, "FJ.GATEWAY.gateway1.RPC.AGENT.REGISTER", got)
}

func TestUnregisterAgentSubject(t *testing.T) {
	got := gateway.UnregisterAgentSubject("gateway1")
	require.Equal(t, "FJ.GATEWAY.gateway1.RPC.AGENT.UNREGISTER", got)
}

func TestCreateJobSubject(t *testing.T) {
	got := gateway.CreateJobSubject("gateway1")
	require.Equal(t, "FJ.GATEWAY.gateway1.RPC.JOB.CREATE", got)
}

func TestCancelJobSubject(t *testing.T) {
	got := gateway.CancelJobSubject("gateway1")
	require.Equal(t, "FJ.GATEWAY.gateway1.RPC.JOB.CANCEL", got)
}

func TestJobInputSubject(t *testing.T) {
	got := gateway.JobInputSubject("gateway1")
	require.Equal(t, "FJ.GATEWAY.gateway1.RPC.JOB.INPUT", got)
}

func TestJobOutputSubject(t *testing.T) {
	got := gateway.JobOutputSubject("gateway1")
	require.Equal(t, "FJ.GATEWAY.gateway1.RPC.JOB.OUTPUT", got)
}
