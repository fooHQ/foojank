package proto_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/proto"
)

func TestMarshalUnmarshal(t *testing.T) {
	testError := errors.New("test error")

	tests := []struct {
		name        string
		input       any
		want        any
		wantMarshal bool
		wantErr     error
	}{
		{
			name: "StartWorkerRequest",
			input: proto.StartWorkerRequest{
				File: "/path/to/file",
				Args: []string{"arg1", "arg2"},
				Env:  []string{"KEY1=val1", "KEY2=val2"},
			},
			want: proto.StartWorkerRequest{
				File: "/path/to/file",
				Args: []string{"arg1", "arg2"},
				Env:  []string{"KEY1=val1", "KEY2=val2"},
			},
			wantMarshal: true,
		},
		{
			name: "StartWorkerRequest with special characters",
			input: proto.StartWorkerRequest{
				File: "/path/with spaces/file",
				Args: []string{"arg with spaces", "arg\nwith\nnewlines"},
				Env:  []string{"KEY=value with spaces", "VAR=value\nwith\nnewlines"},
			},
			want: proto.StartWorkerRequest{
				File: "/path/with spaces/file",
				Args: []string{"arg with spaces", "arg\nwith\nnewlines"},
				Env:  []string{"KEY=value with spaces", "VAR=value\nwith\nnewlines"},
			},
			wantMarshal: true,
		},
		{
			name: "StartWorkerRequest with empty slices",
			input: proto.StartWorkerRequest{
				File: "/path/to/file",
				Args: []string{},
				Env:  []string{},
			},
			want: proto.StartWorkerRequest{
				File: "/path/to/file",
				Args: nil,
				Env:  nil,
			},
			wantMarshal: true,
		},
		{
			name: "StartWorkerResponse without error",
			input: proto.StartWorkerResponse{
				Error: nil,
			},
			want: proto.StartWorkerResponse{
				Error: nil,
			},
			wantMarshal: true,
		},
		{
			name: "StartWorkerResponse with error",
			input: proto.StartWorkerResponse{
				Error: testError,
			},
			want: proto.StartWorkerResponse{
				Error: testError,
			},
			wantMarshal: true,
		},
		{
			name:        "StopWorkerRequest",
			input:       proto.StopWorkerRequest{},
			want:        proto.StopWorkerRequest{},
			wantMarshal: true,
		},
		{
			name: "StopWorkerResponse without error",
			input: proto.StopWorkerResponse{
				Error: nil,
			},
			want: proto.StopWorkerResponse{
				Error: nil,
			},
			wantMarshal: true,
		},
		{
			name: "StopWorkerResponse with error",
			input: proto.StopWorkerResponse{
				Error: testError,
			},
			want: proto.StopWorkerResponse{
				Error: testError,
			},
			wantMarshal: true,
		},
		{
			name: "UpdateWorkerStatus",
			input: proto.UpdateWorkerStatus{
				Status: 42,
			},
			want: proto.UpdateWorkerStatus{
				Status: 42,
			},
			wantMarshal: true,
		},
		{
			name: "UpdateWorkerStdio",
			input: proto.UpdateWorkerStdio{
				Data: []byte("Hello, World!"),
			},
			want: proto.UpdateWorkerStdio{
				Data: []byte("Hello, World!"),
			},
			wantMarshal: true,
		},
		{
			name: "UpdateClientInfo",
			input: proto.UpdateClientInfo{
				Username: "testuser",
				Hostname: "testhost",
				System:   "linux",
				Address:  "192.168.1.1",
			},
			want: proto.UpdateClientInfo{
				Username: "testuser",
				Hostname: "testhost",
				System:   "linux",
				Address:  "192.168.1.1",
			},
			wantMarshal: true,
		},
		{
			name: "UpdateClientInfo with empty fields",
			input: proto.UpdateClientInfo{
				Username: "",
				Hostname: "",
				System:   "",
				Address:  "",
			},
			want: proto.UpdateClientInfo{
				Username: "",
				Hostname: "",
				System:   "",
				Address:  "",
			},
			wantMarshal: true,
		},
		{
			name: "UpdateClientInfo with special characters",
			input: proto.UpdateClientInfo{
				Username: "user with spaces",
				Hostname: "host-name.local",
				System:   "Windows 10",
				Address:  "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			},
			want: proto.UpdateClientInfo{
				Username: "user with spaces",
				Hostname: "host-name.local",
				System:   "Windows 10",
				Address:  "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			},
			wantMarshal: true,
		},
		{
			name:    "Unsupported type",
			input:   struct{}{},
			wantErr: proto.ErrUnknownMessage,
		},
		{
			name:    "Nil input",
			input:   nil,
			wantErr: proto.ErrUnknownMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Marshal
			marshaled, err := proto.Marshal(tt.input)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, marshaled)

			// Test Unmarshal
			unmarshaled, err := proto.Unmarshal(marshaled)
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
			_, err := proto.Unmarshal(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
