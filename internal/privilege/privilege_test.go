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
			name:    "UserUpdate",
			priv:    privilege.UserUpdate{},
			str:     "USER.UPDATE",
			subject: protodaemon.UpdateUserSubject(),
		},
		{
			name:    "UserDelete",
			priv:    privilege.UserDelete{},
			str:     "USER.DELETE",
			subject: protodaemon.DeleteUserSubject(),
		},
		{
			name:    "AgentCreate",
			priv:    privilege.AgentCreate{},
			str:     "AGENT.CREATE",
			subject: protodaemon.CreateAgentSubject(),
		},
		{
			name:    "AgentGet",
			priv:    privilege.AgentGet{},
			str:     "AGENT.GET",
			subject: protodaemon.GetAgentSubject(),
		},
		{
			name:    "AgentList",
			priv:    privilege.AgentList{},
			str:     "AGENT.LIST",
			subject: protodaemon.ListAgentsSubject(),
		},
		{
			name:    "AgentDelete",
			priv:    privilege.AgentDelete{},
			str:     "AGENT.DELETE",
			subject: protodaemon.DeleteAgentSubject(),
		},
		{
			name:    "GatewayCreate",
			priv:    privilege.GatewayCreate{},
			str:     "GATEWAY.CREATE",
			subject: protodaemon.CreateGatewaySubject(),
		},
		{
			name:    "GatewayGet",
			priv:    privilege.GatewayGet{},
			str:     "GATEWAY.GET",
			subject: protodaemon.GetGatewaySubject(),
		},
		{
			name:    "GatewayList",
			priv:    privilege.GatewayList{},
			str:     "GATEWAY.LIST",
			subject: protodaemon.ListGatewaysSubject(),
		},
		{
			name:    "GatewayDelete",
			priv:    privilege.GatewayDelete{},
			str:     "GATEWAY.DELETE",
			subject: protodaemon.DeleteGatewaySubject(),
		},
		{
			name:    "RoleCreate",
			priv:    privilege.RoleCreate{},
			str:     "ROLE.CREATE",
			subject: protodaemon.CreateRoleSubject(),
		},
		{
			name:    "RoleGet",
			priv:    privilege.RoleGet{},
			str:     "ROLE.GET",
			subject: protodaemon.GetRoleSubject(),
		},
		{
			name:    "RoleList",
			priv:    privilege.RoleList{},
			str:     "ROLE.LIST",
			subject: protodaemon.ListRolesSubject(),
		},
		{
			name:    "RoleDelete",
			priv:    privilege.RoleDelete{},
			str:     "ROLE.DELETE",
			subject: protodaemon.DeleteRoleSubject(),
		},
		{
			name:    "JobCreate",
			priv:    privilege.JobCreate{},
			str:     "JOB.CREATE",
			subject: protodaemon.CreateJobSubject(),
		},
		{
			name:    "JobGet",
			priv:    privilege.JobGet{},
			str:     "JOB.GET",
			subject: protodaemon.GetJobSubject(),
		},
		{
			name:    "JobList",
			priv:    privilege.JobList{},
			str:     "JOB.LIST",
			subject: protodaemon.ListJobsSubject(),
		},
		{
			name:    "JobCancel",
			priv:    privilege.JobCancel{},
			str:     "JOB.CANCEL",
			subject: protodaemon.CancelJobSubject(),
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
			name:  "UserUpdate",
			input: "USER.UPDATE",
			want:  privilege.UserUpdate{},
		},
		{
			name:  "UserUpdate lowercase",
			input: "user.update",
			want:  privilege.UserUpdate{},
		},
		{
			name:  "UserDelete",
			input: "USER.DELETE",
			want:  privilege.UserDelete{},
		},
		{
			name:  "AgentCreate",
			input: "AGENT.CREATE",
			want:  privilege.AgentCreate{},
		},
		{
			name:  "AgentCreate lowercase",
			input: "agent.create",
			want:  privilege.AgentCreate{},
		},
		{
			name:  "AgentGet",
			input: "AGENT.GET",
			want:  privilege.AgentGet{},
		},
		{
			name:  "AgentList",
			input: "AGENT.LIST",
			want:  privilege.AgentList{},
		},
		{
			name:  "AgentDelete",
			input: "AGENT.DELETE",
			want:  privilege.AgentDelete{},
		},
		{
			name:  "GatewayCreate",
			input: "GATEWAY.CREATE",
			want:  privilege.GatewayCreate{},
		},
		{
			name:  "GatewayGet",
			input: "GATEWAY.GET",
			want:  privilege.GatewayGet{},
		},
		{
			name:  "GatewayList",
			input: "GATEWAY.LIST",
			want:  privilege.GatewayList{},
		},
		{
			name:  "GatewayDelete",
			input: "GATEWAY.DELETE",
			want:  privilege.GatewayDelete{},
		},
		{
			name:  "RoleCreate",
			input: "ROLE.CREATE",
			want:  privilege.RoleCreate{},
		},
		{
			name:  "RoleGet",
			input: "ROLE.GET",
			want:  privilege.RoleGet{},
		},
		{
			name:  "RoleList",
			input: "ROLE.LIST",
			want:  privilege.RoleList{},
		},
		{
			name:  "RoleDelete",
			input: "ROLE.DELETE",
			want:  privilege.RoleDelete{},
		},
		{
			name:  "JobCreate",
			input: "JOB.CREATE",
			want:  privilege.JobCreate{},
		},
		{
			name:  "JobGet",
			input: "JOB.GET",
			want:  privilege.JobGet{},
		},
		{
			name:  "JobList",
			input: "JOB.LIST",
			want:  privilege.JobList{},
		},
		{
			name:  "JobCancel",
			input: "JOB.CANCEL",
			want:  privilege.JobCancel{},
		},
		{
			name:  "JobCancel lowercase",
			input: "job.cancel",
			want:  privilege.JobCancel{},
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
			name:    "AgentCreate extra argument",
			input:   "AGENT.CREATE.extra",
			wantErr: true,
		},
		{
			name:    "JobUpdate is not a privilege",
			input:   "JOB.UPDATE",
			wantErr: true,
		},
		{
			name:    "JobCancel extra argument",
			input:   "JOB.CANCEL.extra",
			wantErr: true,
		},
		{
			name:    "JobDelete is not a privilege",
			input:   "JOB.DELETE",
			wantErr: true,
		},
		{
			name:    "unknown",
			input:   "NOPE",
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
		privilege.UserUpdate{},
		privilege.UserDelete{},
		privilege.AgentCreate{},
		privilege.AgentGet{},
		privilege.AgentList{},
		privilege.AgentDelete{},
		privilege.GatewayCreate{},
		privilege.GatewayGet{},
		privilege.GatewayList{},
		privilege.GatewayDelete{},
		privilege.RoleCreate{},
		privilege.RoleGet{},
		privilege.RoleList{},
		privilege.RoleDelete{},
		privilege.JobCreate{},
		privilege.JobGet{},
		privilege.JobList{},
		privilege.JobCancel{},
	}

	ss := privilege.FormatPrivileges(privileges)
	require.Equal(t, []string{
		"USER.CREATE",
		"JWT.ISSUE." + userID,
		"USER.GET",
		"USER.LIST",
		"USER.UPDATE",
		"USER.DELETE",
		"AGENT.CREATE",
		"AGENT.GET",
		"AGENT.LIST",
		"AGENT.DELETE",
		"GATEWAY.CREATE",
		"GATEWAY.GET",
		"GATEWAY.LIST",
		"GATEWAY.DELETE",
		"ROLE.CREATE",
		"ROLE.GET",
		"ROLE.LIST",
		"ROLE.DELETE",
		"JOB.CREATE",
		"JOB.GET",
		"JOB.LIST",
		"JOB.CANCEL",
	}, ss)

	got, err := privilege.ParsePrivileges(ss)
	require.NoError(t, err)
	require.Equal(t, privileges, got.Privileges())
	require.Equal(t, []string{
		protodaemon.CreateUserSubject(),
		protodaemon.IssueJWTSubject(userID),
		protodaemon.GetUserSubject(),
		protodaemon.ListUsersSubject(),
		protodaemon.UpdateUserSubject(),
		protodaemon.DeleteUserSubject(),
		protodaemon.CreateAgentSubject(),
		protodaemon.GetAgentSubject(),
		protodaemon.ListAgentsSubject(),
		protodaemon.DeleteAgentSubject(),
		protodaemon.CreateGatewaySubject(),
		protodaemon.GetGatewaySubject(),
		protodaemon.ListGatewaysSubject(),
		protodaemon.DeleteGatewaySubject(),
		protodaemon.CreateRoleSubject(),
		protodaemon.GetRoleSubject(),
		protodaemon.ListRolesSubject(),
		protodaemon.DeleteRoleSubject(),
		protodaemon.CreateJobSubject(),
		protodaemon.GetJobSubject(),
		protodaemon.ListJobsSubject(),
		protodaemon.CancelJobSubject(),
	}, got.Permissions())
}

func TestUpdatePrivileges(t *testing.T) {
	tests := []struct {
		name    string
		current []string
		add     []string
		without []string
		want    []string
		wantErr bool
	}{
		{
			name:    "add and remove",
			current: []string{"USER.CREATE", "USER.GET"},
			add:     []string{"USER.LIST"},
			without: []string{"user.create"},
			want:    []string{"USER.GET", "USER.LIST"},
		},
		{
			name:    "add existing",
			current: []string{"USER.GET"},
			add:     []string{"user.get", "USER.GET"},
			want:    []string{"USER.GET"},
		},
		{
			name:    "remove missing",
			current: []string{"USER.GET"},
			without: []string{"USER.LIST"},
			want:    []string{"USER.GET"},
		},
		{
			name:    "unset wins",
			current: []string{"USER.GET"},
			add:     []string{"USER.LIST"},
			without: []string{"USER.GET", "USER.LIST"},
			want:    []string{},
		},
		{
			name:    "sorted and deduplicated",
			current: []string{"USER.LIST", "user.get", "USER.GET"},
			add:     []string{"USER.CREATE"},
			want:    []string{"USER.CREATE", "USER.GET", "USER.LIST"},
		},
		{
			name:    "empty",
			current: nil,
			want:    nil,
		},
		{
			name:    "unknown add",
			add:     []string{"NOPE"},
			wantErr: true,
		},
		{
			name:    "unknown without",
			without: []string{"NOPE"},
			wantErr: true,
		},
		{
			name:    "unknown current",
			current: []string{"NOPE"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := privilege.UpdatePrivileges(tt.current, tt.add, tt.without)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
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
