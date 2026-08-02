package agent_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/proto/agent"
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
			name: "StartWorker",
			input: agent.StartWorker{
				Command: "cmd",
				Args:    []string{"arg1", "arg2"},
				Env:     []string{"KEY1=val1", "KEY2=val2"},
			},
			want: agent.StartWorker{
				Command: "cmd",
				Args:    []string{"arg1", "arg2"},
				Env:     []string{"KEY1=val1", "KEY2=val2"},
			},
		},
		{
			name: "StartWorker with empty slices",
			input: agent.StartWorker{
				Command: "cmd",
				Args:    []string{},
				Env:     []string{},
			},
			want: agent.StartWorker{
				Command: "cmd",
				Args:    nil,
				Env:     nil,
			},
		},
		{
			name: "WorkerStarted without error",
			input: agent.WorkerStarted{
				Error: nil,
			},
			want: agent.WorkerStarted{
				Error: nil,
			},
		},
		{
			name: "WorkerStarted with error",
			input: agent.WorkerStarted{
				Error: testError,
			},
			want: agent.WorkerStarted{
				Error: testError,
			},
		},
		{
			name:  "StopWorker",
			input: agent.StopWorker{},
			want:  agent.StopWorker{},
		},
		{
			name: "WorkerStopped without error",
			input: agent.WorkerStopped{
				Error: nil,
			},
			want: agent.WorkerStopped{
				Error: nil,
			},
		},
		{
			name: "WorkerStopped with error",
			input: agent.WorkerStopped{
				Error: testError,
			},
			want: agent.WorkerStopped{
				Error: testError,
			},
		},
		{
			name: "WorkerStatus",
			input: agent.WorkerStatus{
				Status: 42,
			},
			want: agent.WorkerStatus{
				Status: 42,
			},
		},
		{
			name: "WorkerStatus with error",
			input: agent.WorkerStatus{
				Status: agent.ExitFailure,
				Error:  testError,
			},
			want: agent.WorkerStatus{
				Status: agent.ExitFailure,
				Error:  testError,
			},
		},
		{
			name: "pointer input",
			input: &agent.StartWorker{
				Command: "cmd",
			},
			want: agent.StartWorker{
				Command: "cmd",
			},
		},
		{
			name: "WorkerInput",
			input: agent.WorkerInput{
				Data: []byte("Hello, World!"),
			},
			want: agent.WorkerInput{
				Data: []byte("Hello, World!"),
			},
		},
		{
			name: "WorkerOutput",
			input: agent.WorkerOutput{
				Data: []byte("Hello, World!"),
			},
			want: agent.WorkerOutput{
				Data: []byte("Hello, World!"),
			},
		},
		{
			name: "AgentInfo",
			input: agent.AgentInfo{
				Username: "testuser",
				Hostname: "testhost",
				System:   "linux",
				Address:  "192.168.1.1",
			},
			want: agent.AgentInfo{
				Username: "testuser",
				Hostname: "testhost",
				System:   "linux",
				Address:  "192.168.1.1",
			},
		},
		{
			name: "AgentInfo with empty fields",
			input: agent.AgentInfo{
				Username: "",
				Hostname: "",
				System:   "",
				Address:  "",
			},
			want: agent.AgentInfo{
				Username: "",
				Hostname: "",
				System:   "",
				Address:  "",
			},
		},
		{
			name:    "Unsupported type",
			input:   struct{}{},
			wantErr: agent.ErrUnknownType,
		},
		{
			name:    "Nil input",
			input:   nil,
			wantErr: agent.ErrUnknownType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Marshal
			marshaled, err := agent.Marshal(tt.input)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, marshaled)

			// Test Unmarshal
			unmarshaled, err := agent.Unmarshal(marshaled)
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
			name:    "Corrupt Cap'n Proto message",
			input:   []byte{0, 0, 0, 0, 0, 0, 0, 0},
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
			_, err := agent.Unmarshal(tt.input)
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
	_, err := agent.Unmarshal([]byte{0x82, 0x00, 0x00})
	require.ErrorIs(t, err, agent.ErrUnknownTag)
}

func TestUnmarshalInvalidArrayLength(t *testing.T) {
	// CBOR array of length 1 (0x81), which is not a valid message envelope.
	_, err := agent.Unmarshal([]byte{0x81, 0x00})
	require.Error(t, err)
}

func TestUnmarshalTrailingBytes(t *testing.T) {
	marshaled, err := agent.Marshal(agent.StopWorker{})
	require.NoError(t, err)

	_, err = agent.Unmarshal(append(marshaled, 0xFF))
	require.Error(t, err)
}

func TestCmdStartWorkerSubject(t *testing.T) {
	got := agent.CmdStartWorkerSubject("gateway1", "agent1", "worker1")
	require.Equal(t, "FJ.GATEWAY.gateway1.AGENT.agent1.CMD.WORKER.worker1.START", got)
}

func TestCmdStopWorkerSubject(t *testing.T) {
	got := agent.CmdStopWorkerSubject("gateway1", "agent1", "worker1")
	require.Equal(t, "FJ.GATEWAY.gateway1.AGENT.agent1.CMD.WORKER.worker1.STOP", got)
}

func TestCmdWriteStdinSubject(t *testing.T) {
	got := agent.CmdWriteStdinSubject("gateway1", "agent1", "worker1")
	require.Equal(t, "FJ.GATEWAY.gateway1.AGENT.agent1.CMD.WORKER.worker1.STDIN", got)
}

func TestEvtStartWorkerSubject(t *testing.T) {
	got := agent.EvtStartWorkerSubject("gateway1", "agent1", "worker1")
	require.Equal(t, "FJ.GATEWAY.gateway1.AGENT.agent1.EVT.WORKER.worker1.START", got)
}

func TestEvtStopWorkerSubject(t *testing.T) {
	got := agent.EvtStopWorkerSubject("gateway1", "agent1", "worker1")
	require.Equal(t, "FJ.GATEWAY.gateway1.AGENT.agent1.EVT.WORKER.worker1.STOP", got)
}

func TestEvtWorkerStatusSubject(t *testing.T) {
	got := agent.EvtWorkerStatusSubject("gateway1", "agent1", "worker1")
	require.Equal(t, "FJ.GATEWAY.gateway1.AGENT.agent1.EVT.WORKER.worker1.STATUS", got)
}

func TestEvtWorkerStdoutSubject(t *testing.T) {
	got := agent.EvtWorkerStdoutSubject("gateway1", "agent1", "worker1")
	require.Equal(t, "FJ.GATEWAY.gateway1.AGENT.agent1.EVT.WORKER.worker1.STDOUT", got)
}

func TestEvtAgentInfoSubject(t *testing.T) {
	got := agent.EvtAgentInfoSubject("gateway1", "agent1")
	require.Equal(t, "FJ.GATEWAY.gateway1.AGENT.agent1.EVT.INFO", got)
}
