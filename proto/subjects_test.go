package proto_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/proto"
)

func TestStartWorkerSubject(t *testing.T) {
	tests := []struct {
		name     string
		agentID  string
		workerID string
		want     string
	}{
		{
			name:     "basic replacement",
			agentID:  "agent1",
			workerID: "worker1",
			want:     "FJ.API.WORKER.START.agent1.worker1",
		},
		{
			name:     "empty agent ID",
			agentID:  "",
			workerID: "worker1",
			want:     "FJ.API.WORKER.START..worker1",
		},
		{
			name:     "empty worker ID",
			agentID:  "agent1",
			workerID: "",
			want:     "FJ.API.WORKER.START.agent1.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := proto.StartWorkerSubject(tt.agentID, tt.workerID)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestStopWorkerSubject(t *testing.T) {
	tests := []struct {
		name     string
		agentID  string
		workerID string
		want     string
	}{
		{
			name:     "basic replacement",
			agentID:  "agent1",
			workerID: "worker1",
			want:     "FJ.API.WORKER.STOP.agent1.worker1",
		},
		{
			name:     "empty agent ID",
			agentID:  "",
			workerID: "worker1",
			want:     "FJ.API.WORKER.STOP..worker1",
		},
		{
			name:     "empty worker ID",
			agentID:  "agent1",
			workerID: "",
			want:     "FJ.API.WORKER.STOP.agent1.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := proto.StopWorkerSubject(tt.agentID, tt.workerID)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestWriteWorkerStdinSubject(t *testing.T) {
	tests := []struct {
		name     string
		agentID  string
		workerID string
		want     string
	}{
		{
			name:     "basic replacement",
			agentID:  "agent1",
			workerID: "worker1",
			want:     "FJ.API.WORKER.WRITE.STDIN.agent1.worker1",
		},
		{
			name:     "empty agent ID",
			agentID:  "",
			workerID: "worker1",
			want:     "FJ.API.WORKER.WRITE.STDIN..worker1",
		},
		{
			name:     "empty worker ID",
			agentID:  "agent1",
			workerID: "",
			want:     "FJ.API.WORKER.WRITE.STDIN.agent1.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := proto.WriteWorkerStdinSubject(tt.agentID, tt.workerID)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestWriteWorkerStdoutSubject(t *testing.T) {
	tests := []struct {
		name     string
		agentID  string
		workerID string
		want     string
	}{
		{
			name:     "basic replacement",
			agentID:  "agent1",
			workerID: "worker1",
			want:     "FJ.API.WORKER.WRITE.STDOUT.agent1.worker1",
		},
		{
			name:     "empty agent ID",
			agentID:  "",
			workerID: "worker1",
			want:     "FJ.API.WORKER.WRITE.STDOUT..worker1",
		},
		{
			name:     "empty worker ID",
			agentID:  "agent1",
			workerID: "",
			want:     "FJ.API.WORKER.WRITE.STDOUT.agent1.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := proto.WriteWorkerStdoutSubject(tt.agentID, tt.workerID)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUpdateWorkerStatusSubject(t *testing.T) {
	tests := []struct {
		name     string
		agentID  string
		workerID string
		want     string
	}{
		{
			name:     "basic replacement",
			agentID:  "agent1",
			workerID: "worker1",
			want:     "FJ.API.WORKER.UPDATE.STATUS.agent1.worker1",
		},
		{
			name:     "empty agent ID",
			agentID:  "",
			workerID: "worker1",
			want:     "FJ.API.WORKER.UPDATE.STATUS..worker1",
		},
		{
			name:     "empty worker ID",
			agentID:  "agent1",
			workerID: "",
			want:     "FJ.API.WORKER.UPDATE.STATUS.agent1.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := proto.UpdateWorkerStatusSubject(tt.agentID, tt.workerID)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUpdateClientInfoSubject(t *testing.T) {
	tests := []struct {
		name    string
		agentID string
		want    string
	}{
		{
			name:    "basic replacement",
			agentID: "agent1",
			want:    "FJ.API.CLIENT.UPDATE.INFO.agent1",
		},
		{
			name:    "empty agent ID",
			agentID: "",
			want:    "FJ.API.CLIENT.UPDATE.INFO.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := proto.UpdateClientInfoSubject(tt.agentID)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestReplyMessageSubject(t *testing.T) {
	tests := []struct {
		name    string
		agentID string
		msgID   string
		want    string
	}{
		{
			name:    "basic replacement",
			agentID: "agent1",
			msgID:   "msg1",
			want:    "FJ.API.MESSAGE.REPLY.agent1.msg1",
		},
		{
			name:    "empty agent ID",
			agentID: "",
			msgID:   "msg1",
			want:    "FJ.API.MESSAGE.REPLY..msg1",
		},
		{
			name:    "empty message ID",
			agentID: "agent1",
			msgID:   "",
			want:    "FJ.API.MESSAGE.REPLY.agent1.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := proto.ReplyMessageSubject(tt.agentID, tt.msgID)
			require.Equal(t, tt.want, got)
		})
	}
}
