package privilege_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/privilege"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

func TestPrivilegeStringSubject(t *testing.T) {
	userID := "UDUMMYUSERPUBLICKEY"

	tests := []struct {
		name    string
		priv    privilege.Privilege
		str     string
		subject string
		wantErr bool
	}{
		{
			name:    "UserCreate",
			priv:    privilege.UserCreate{},
			str:     "USER.CREATE",
			subject: protodaemon.CreateUserSubject(),
		},
		{
			name:    "UserGet",
			priv:    privilege.UserGet{},
			str:     "USER.GET",
			subject: protodaemon.GetUserSubject(),
		},
		{
			name:    "UserList",
			priv:    privilege.UserList{},
			str:     "USER.LIST",
			subject: protodaemon.ListUsersSubject(),
		},
		{
			name: "JWTIssue",
			priv: privilege.JWTIssue{
				UserID: userID,
			},
			str:     "JWT.ISSUE." + userID,
			subject: protodaemon.IssueJWTSubject(userID),
		},
		{
			name:    "JWTIssue missing user ID",
			priv:    privilege.JWTIssue{},
			str:     "JWT.ISSUE.",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.str, tt.priv.String())

			subject, err := tt.priv.Subject()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.subject, subject)
		})
	}
}

func TestParsePrivileges(t *testing.T) {
	userID := "UDUMMYUSERPUBLICKEY"

	tests := []struct {
		name    string
		input   string
		want    privilege.Privilege
		wantErr bool
	}{
		{
			name:  "UserCreate",
			input: "USER.CREATE",
			want:  privilege.UserCreate{},
		},
		{
			name:  "UserCreate lowercase",
			input: "user.create",
			want:  privilege.UserCreate{},
		},
		{
			name:  "UserGet",
			input: "USER.GET",
			want:  privilege.UserGet{},
		},
		{
			name:  "UserList",
			input: "USER.LIST",
			want:  privilege.UserList{},
		},
		{
			name:  "JWTIssue",
			input: "JWT.ISSUE." + userID,
			want: privilege.JWTIssue{
				UserID: userID,
			},
		},
		{
			name:  "JWTIssue mixed case prefix",
			input: "jwt.issue." + userID,
			want: privilege.JWTIssue{
				UserID: userID,
			},
		},
		{
			name:    "JWTIssue without user ID",
			input:   "JWT.ISSUE",
			wantErr: true,
		},
		{
			name:    "JWTIssue empty user ID",
			input:   "JWT.ISSUE.",
			wantErr: true,
		},
		{
			name:    "JWTIssue extra argument",
			input:   "JWT.ISSUE." + userID + ".extra",
			wantErr: true,
		},
		{
			name:    "UserCreate extra argument",
			input:   "USER.CREATE.extra",
			wantErr: true,
		},
		{
			name:    "UserCreate extra arguments",
			input:   "USER.CREATE.a.b",
			wantErr: true,
		},
		{
			name:    "unknown",
			input:   "AGENT.CREATE",
			wantErr: true,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := privilege.ParsePrivileges([]string{tt.input})
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, []privilege.Privilege{tt.want}, got.Privileges())

			subject, err := tt.want.Subject()
			require.NoError(t, err)
			require.Equal(t, []string{subject}, got.Permissions())
		})
	}
}

func TestParseFormatPrivilegesRoundTrip(t *testing.T) {
	userID := "UDUMMYUSERPUBLICKEY"
	privileges := []privilege.Privilege{
		privilege.UserCreate{},
		privilege.JWTIssue{
			UserID: userID,
		},
		privilege.UserGet{},
		privilege.UserList{},
	}

	ss := privilege.FormatPrivileges(privileges)
	require.Equal(t, []string{
		"USER.CREATE",
		"JWT.ISSUE." + userID,
		"USER.GET",
		"USER.LIST",
	}, ss)

	got, err := privilege.ParsePrivileges(ss)
	require.NoError(t, err)
	require.Equal(t, privileges, got.Privileges())
	require.Equal(t, []string{
		protodaemon.CreateUserSubject(),
		protodaemon.IssueJWTSubject(userID),
		protodaemon.GetUserSubject(),
		protodaemon.ListUsersSubject(),
	}, got.Permissions())
}

func TestParseFormatPrivilegesEmpty(t *testing.T) {
	require.Nil(t, privilege.FormatPrivileges(nil))
	require.Nil(t, privilege.FormatPrivileges([]privilege.Privilege{}))

	got, err := privilege.ParsePrivileges(nil)
	require.NoError(t, err)
	require.Nil(t, got.Privileges())
	require.Nil(t, got.Permissions())

	got, err = privilege.ParsePrivileges([]string{})
	require.NoError(t, err)
	require.Nil(t, got.Privileges())
	require.Nil(t, got.Permissions())
}

func TestParsePrivilegesUnknown(t *testing.T) {
	_, err := privilege.ParsePrivileges([]string{"USER.CREATE", "NOPE"})
	require.Error(t, err)
}
