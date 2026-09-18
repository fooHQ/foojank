package auth_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/foohq/foojank/internal/auth"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

func TestPrivilegeStringSubject(t *testing.T) {
	userID := "UDUMMYUSERPUBLICKEY"

	tests := []struct {
		name    string
		priv    auth.Privilege
		str     string
		subject string
		wantErr bool
	}{
		{
			name:    "UserCreate",
			priv:    auth.UserCreate{},
			str:     "USER.CREATE",
			subject: protodaemon.CreateUserSubject(),
		},
		{
			name:    "UserGet",
			priv:    auth.UserGet{},
			str:     "USER.GET",
			subject: protodaemon.GetUserSubject(),
		},
		{
			name:    "UserList",
			priv:    auth.UserList{},
			str:     "USER.LIST",
			subject: protodaemon.ListUsersSubject(),
		},
		{
			name:    "JWTIssue",
			priv:    auth.JWTIssue{UserID: userID},
			str:     "JWT.ISSUE." + userID,
			subject: protodaemon.IssueJWTSubject(userID),
		},
		{
			name:    "JWTIssue missing user ID",
			priv:    auth.JWTIssue{},
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

func TestParsePrivilege(t *testing.T) {
	userID := "UDUMMYUSERPUBLICKEY"

	tests := []struct {
		name    string
		input   string
		want    auth.Privilege
		wantErr bool
	}{
		{
			name:  "UserCreate",
			input: "USER.CREATE",
			want:  auth.UserCreate{},
		},
		{
			name:  "UserCreate lowercase",
			input: "user.create",
			want:  auth.UserCreate{},
		},
		{
			name:  "UserGet",
			input: "USER.GET",
			want:  auth.UserGet{},
		},
		{
			name:  "UserList",
			input: "USER.LIST",
			want:  auth.UserList{},
		},
		{
			name:  "JWTIssue",
			input: "JWT.ISSUE." + userID,
			want:  auth.JWTIssue{UserID: userID},
		},
		{
			name:  "JWTIssue mixed case prefix",
			input: "jwt.issue." + userID,
			want:  auth.JWTIssue{UserID: userID},
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
			got, err := auth.ParsePrivilege(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseFormatPrivilegesRoundTrip(t *testing.T) {
	userID := "UDUMMYUSERPUBLICKEY"
	privileges := []auth.Privilege{
		auth.UserCreate{},
		auth.JWTIssue{UserID: userID},
		auth.UserGet{},
		auth.UserList{},
	}

	ss := auth.FormatPrivileges(privileges)
	require.Equal(t, []string{
		"USER.CREATE",
		"JWT.ISSUE." + userID,
		"USER.GET",
		"USER.LIST",
	}, ss)

	got, err := auth.ParsePrivileges(ss)
	require.NoError(t, err)
	require.Equal(t, privileges, got)
}

func TestParseFormatPrivilegesEmpty(t *testing.T) {
	require.Nil(t, auth.FormatPrivileges(nil))
	require.Nil(t, auth.FormatPrivileges([]auth.Privilege{}))

	got, err := auth.ParsePrivileges(nil)
	require.NoError(t, err)
	require.Nil(t, got)

	got, err = auth.ParsePrivileges([]string{})
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestParsePrivilegesUnknown(t *testing.T) {
	_, err := auth.ParsePrivileges([]string{"USER.CREATE", "NOPE"})
	require.Error(t, err)
}

func TestNewClientPermissions(t *testing.T) {
	userID := "UDUMMYUSERPUBLICKEY"

	perms, err := auth.NewClientPermissions([]auth.Privilege{
		auth.UserCreate{},
		auth.UserGet{},
		auth.UserList{},
		auth.JWTIssue{UserID: userID},
	})
	require.NoError(t, err)
	require.EqualValues(t, []string{
		protodaemon.CreateUserSubject(),
		protodaemon.GetUserSubject(),
		protodaemon.ListUsersSubject(),
		protodaemon.IssueJWTSubject(userID),
	}, perms.Pub.Allow)
}

func TestNewClientPermissionsEmpty(t *testing.T) {
	perms, err := auth.NewClientPermissions(nil)
	require.NoError(t, err)
	require.Nil(t, perms.Pub.Allow)
}

func TestNewClientPermissionsErrors(t *testing.T) {
	_, err := auth.NewClientPermissions([]auth.Privilege{auth.JWTIssue{}})
	require.Error(t, err)

	_, err = auth.NewClientPermissions([]auth.Privilege{nil})
	require.Error(t, err)
}
