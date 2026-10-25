package handler

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nkeys"
	"github.com/nats-io/nuid"

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/clients/gateway"
	"github.com/foohq/foojank/internal/directory"
	"github.com/foohq/foojank/internal/log"
	"github.com/foohq/foojank/internal/privilege"
	protogw "github.com/foohq/foojank/proto/gateway"

	"github.com/foohq/foojank/internal/message"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

type NATSHandlerFn func(context.Context, map[string]string, message.Msg) any

type NATSHandlerConfig struct {
	Gateway          *gateway.Client
	AgentDirectory   *directory.AgentDirectory
	GatewayDirectory *directory.GatewayDirectory
	UserDirectory    *directory.UserDirectory
	RoleDirectory    *directory.RoleDirectory
	JobDirectory     *directory.JobDirectory
	AccountKey       nkeys.KeyPair
}

type NATSHandler struct {
	conf   NATSHandlerConfig
	logger *log.Logger
}

func NewNATSHandler(logger *log.Logger, conf NATSHandlerConfig) *NATSHandler {
	return &NATSHandler{
		conf:   conf,
		logger: logger,
	}
}

func (h *NATSHandler) Match(msg message.Msg) (func(context.Context) message.Msg, bool) {
	routes := map[string]NATSHandlerFn{
		protodaemon.CreateUserSubject(): h.CreateUser,
		protodaemon.GetUserSubject():    h.GetUser,
		protodaemon.ListUsersSubject():  h.ListUsers,
		protodaemon.UpdateUserSubject(): h.UpdateUser,
		protodaemon.DeleteUserSubject(): h.DeleteUser,

		protodaemon.CreateAgentSubject():  h.CreateAgent,
		protodaemon.GetAgentSubject():     h.GetAgent,
		protodaemon.ListAgentsSubject():   h.ListAgents,
		protodaemon.DeleteAgentSubject():  h.DeleteAgent,
		protodaemon.GetCommandSubject():   h.GetCommand,
		protodaemon.ListCommandsSubject(): h.ListCommands,

		protodaemon.CreateGatewaySubject(): h.CreateGateway,
		protodaemon.GetGatewaySubject():    h.GetGateway,
		protodaemon.ListGatewaysSubject():  h.ListGateways,
		protodaemon.DeleteGatewaySubject(): h.DeleteGateway,

		protodaemon.CreateRoleSubject(): h.CreateRole,
		protodaemon.GetRoleSubject():    h.GetRole,
		protodaemon.ListRolesSubject():  h.ListRoles,
		protodaemon.DeleteRoleSubject(): h.DeleteRole,

		protodaemon.CreateJobSubject():    h.CreateJob,
		protodaemon.GetJobSubject():       h.GetJob,
		protodaemon.ListJobsSubject():     h.ListJobs,
		protodaemon.CancelJobSubject():    h.CancelJob,
		protodaemon.GetJobOutputSubject(): h.GetJobOutput,

		protodaemon.IssueJWTSubject("<user>"): h.IssueJWT,
	}
	for route, handler := range routes {
		params, ok := match(route, msg.Subject())
		if !ok {
			continue
		}
		return func(ctx context.Context) message.Msg {
			v := handler(ctx, params, msg)
			return NATSHandlerMessage{
				msg:     msg,
				subject: msg.ReplySubject(),
				data:    v,
			}
		}, true
	}
	h.logger.WarnContext(context.Background(), "Cannot match message with subject %q to route", msg.Subject())
	return nil, false
}

func (h *NATSHandler) CreateUser(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.CreateUserRequest)
	if !ok {
		return protodaemon.CreateUserResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateCreateUserRequest(req)
	if err != nil {
		return protodaemon.CreateUserResponse{
			Error: err,
		}
	}

	userID := req.ID
	userName := req.Name
	userDesc := req.Description

	userPrivs, err := privilege.ParsePrivileges(req.Privileges)
	if err != nil {
		return protodaemon.CreateUserResponse{
			Error: err,
		}
	}

	_, err = h.conf.UserDirectory.Create(ctx, directory.UserDirectoryEntry{
		ID:          userID,
		Name:        userName,
		Description: userDesc,
		Privileges:  privilege.FormatPrivileges(userPrivs.Privileges()),
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   time.Time{}, // TODO
	})
	if err != nil {
		if errors.Is(err, directory.ErrKeyExists) {
			err = fmt.Errorf("%q already exists", userName)
		}
		h.logger.ErrorContext(ctx, "Cannot create user: %v", err)
		return protodaemon.CreateUserResponse{
			Error: err,
		}
	}

	return protodaemon.CreateUserResponse{}
}

func (h *NATSHandler) GetUser(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.GetUserRequest)
	if !ok {
		return protodaemon.GetUserResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateGetUserRequest(req)
	if err != nil {
		return protodaemon.GetUserResponse{
			Error: err,
		}
	}

	userRef := req.UserRef

	user, err := h.conf.UserDirectory.Get(ctx, userRef)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", userRef)
		}
		return protodaemon.GetUserResponse{
			Error: err,
		}
	}

	userKeyPair, err := nkeys.FromPublicKey(user.ID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot create user keypair: %v", err)
		return protodaemon.GetUserResponse{
			Error: err,
		}
	}

	userPerms, err := auth.NewRefreshTokenPermissions(user.ID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot create client permissions: %v", err)
		return protodaemon.GetUserResponse{
			Error: err,
		}
	}

	claims, err := auth.NewUserJWT(userRef, userPerms, userKeyPair)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot generate a user JWT: %v", err)
		return protodaemon.GetUserResponse{
			Error: err,
		}
	}

	claims.Expires = user.ExpiresAt.Unix()

	userJWT, err := claims.Encode(h.conf.AccountKey)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot encode user JWT: %v", err)
		return protodaemon.GetUserResponse{
			Error: err,
		}
	}

	return protodaemon.GetUserResponse{
		User: protodaemon.User{
			ID:          user.ID,
			Name:        user.Name,
			Description: user.Description,
			JWT:         userJWT,
			Privileges:  user.Privileges,
			CreatedAt:   user.CreatedAt.Unix(),
			ExpiresAt:   user.ExpiresAt.Unix(),
		},
	}
}

func (h *NATSHandler) ListUsers(ctx context.Context, params map[string]string, msg message.Msg) any {
	_, ok := msg.Data().(protodaemon.ListUsersRequest)
	if !ok {
		return protodaemon.ListUsersResponse{
			Error: errors.New("invalid request data"),
		}
	}

	entries, err := h.conf.UserDirectory.List(ctx)
	if err != nil {
		return protodaemon.ListUsersResponse{
			Error: err,
		}
	}

	users := make([]protodaemon.User, len(entries))
	for i := range entries {
		users[i] = protodaemon.User{
			ID:          entries[i].ID,
			Name:        entries[i].Name,
			Description: entries[i].Description,
			Privileges:  entries[i].Privileges,
			CreatedAt:   entries[i].CreatedAt.Unix(),
			ExpiresAt:   entries[i].ExpiresAt.Unix(),
		}
	}

	return protodaemon.ListUsersResponse{
		Users: users,
	}
}

func (h *NATSHandler) UpdateUser(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.UpdateUserRequest)
	if !ok {
		return protodaemon.UpdateUserResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateUpdateUserRequest(req)
	if err != nil {
		return protodaemon.UpdateUserResponse{
			Error: err,
		}
	}

	user, err := h.conf.UserDirectory.Get(ctx, req.UserRef)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", req.UserRef)
		}
		return protodaemon.UpdateUserResponse{
			Error: err,
		}
	}

	if req.IsDescription {
		user.Description = req.Description
	}

	privs, err := privilege.UpdatePrivileges(user.Privileges, req.SetPrivileges, req.UnsetPrivileges)
	if err != nil {
		return protodaemon.UpdateUserResponse{
			Error: err,
		}
	}

	user.Privileges = privs

	_, err = h.conf.UserDirectory.Update(ctx, user)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot update user: %v", err)
		return protodaemon.UpdateUserResponse{
			Error: err,
		}
	}

	return protodaemon.UpdateUserResponse{}
}

func (h *NATSHandler) DeleteUser(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.DeleteUserRequest)
	if !ok {
		return protodaemon.DeleteUserResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateDeleteUserRequest(req)
	if err != nil {
		return protodaemon.DeleteUserResponse{
			Error: err,
		}
	}

	user, err := h.conf.UserDirectory.Get(ctx, req.UserRef)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", req.UserRef)
		}
		return protodaemon.DeleteUserResponse{
			Error: err,
		}
	}

	err = h.conf.UserDirectory.Delete(ctx, user)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot delete user: %v", err)
		return protodaemon.DeleteUserResponse{
			Error: err,
		}
	}

	return protodaemon.DeleteUserResponse{}
}

func (h *NATSHandler) CreateRole(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.CreateRoleRequest)
	if !ok {
		return protodaemon.CreateRoleResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateCreateRoleRequest(req)
	if err != nil {
		return protodaemon.CreateRoleResponse{
			Error: err,
		}
	}

	roleName := req.Name
	roleDesc := req.Description

	rolePrivs, err := privilege.ParsePrivileges(req.Privileges)
	if err != nil {
		return protodaemon.CreateRoleResponse{
			Error: err,
		}
	}

	_, err = h.conf.RoleDirectory.Create(ctx, directory.RoleDirectoryEntry{
		Name:        roleName,
		Description: roleDesc,
		Privileges:  privilege.FormatPrivileges(rolePrivs.Privileges()),
		CreatedAt:   time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, directory.ErrKeyExists) {
			err = fmt.Errorf("%q already exists", roleName)
		}
		h.logger.ErrorContext(ctx, "Cannot create role: %v", err)
		return protodaemon.CreateRoleResponse{
			Error: err,
		}
	}

	return protodaemon.CreateRoleResponse{}
}

func (h *NATSHandler) GetRole(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.GetRoleRequest)
	if !ok {
		return protodaemon.GetRoleResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateGetRoleRequest(req)
	if err != nil {
		return protodaemon.GetRoleResponse{
			Error: err,
		}
	}

	role, err := h.conf.RoleDirectory.Get(ctx, req.RoleRef)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", req.RoleRef)
		}
		return protodaemon.GetRoleResponse{
			Error: err,
		}
	}

	return protodaemon.GetRoleResponse{
		Role: protodaemon.Role{
			Name:        role.Name,
			Description: role.Description,
			Privileges:  role.Privileges,
			CreatedAt:   role.CreatedAt.Unix(),
		},
	}
}

func (h *NATSHandler) ListRoles(ctx context.Context, params map[string]string, msg message.Msg) any {
	_, ok := msg.Data().(protodaemon.ListRolesRequest)
	if !ok {
		return protodaemon.ListRolesResponse{
			Error: errors.New("invalid request data"),
		}
	}

	entries, err := h.conf.RoleDirectory.List(ctx)
	if err != nil {
		return protodaemon.ListRolesResponse{
			Error: err,
		}
	}

	roles := make([]protodaemon.Role, len(entries))
	for i := range entries {
		roles[i] = protodaemon.Role{
			Name:        entries[i].Name,
			Description: entries[i].Description,
			Privileges:  entries[i].Privileges,
			CreatedAt:   entries[i].CreatedAt.Unix(),
		}
	}

	return protodaemon.ListRolesResponse{
		Roles: roles,
	}
}

func (h *NATSHandler) DeleteRole(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.DeleteRoleRequest)
	if !ok {
		return protodaemon.DeleteRoleResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateDeleteRoleRequest(req)
	if err != nil {
		return protodaemon.DeleteRoleResponse{
			Error: err,
		}
	}

	role, err := h.conf.RoleDirectory.Get(ctx, req.RoleRef)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", req.RoleRef)
		}
		return protodaemon.DeleteRoleResponse{
			Error: err,
		}
	}

	err = h.conf.RoleDirectory.Delete(ctx, role)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot delete role: %v", err)
		return protodaemon.DeleteRoleResponse{
			Error: err,
		}
	}

	return protodaemon.DeleteRoleResponse{}
}

func (h *NATSHandler) IssueJWT(ctx context.Context, params map[string]string, msg message.Msg) any {
	userName := params["user"]

	user, err := h.conf.UserDirectory.Get(ctx, userName)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", userName)
		}
		return protodaemon.IssueJWTResponse{
			Error: err,
		}
	}

	userKeyPair, err := nkeys.FromPublicKey(user.ID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot create user keypair: %v", err)
		return protodaemon.IssueJWTResponse{
			Error: err,
		}
	}

	userPrivs, err := privilege.ParsePrivileges(user.Privileges)
	if err != nil {
		return protodaemon.IssueJWTResponse{
			Error: err,
		}
	}

	userPerms, err := auth.NewClientPermissions(user.ID, userPrivs.Permissions())
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot create client permissions: %v", err)
		return protodaemon.IssueJWTResponse{
			Error: err,
		}
	}

	claims, err := auth.NewUserJWT(userName, userPerms, userKeyPair)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot generate a user JWT: %v", err)
		return protodaemon.IssueJWTResponse{
			Error: err,
		}
	}

	claims.Expires = time.Now().Add(15 * time.Minute).Unix()

	userJWT, err := claims.Encode(h.conf.AccountKey)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot encode user JWT: %v", err)
		return protodaemon.IssueJWTResponse{
			Error: err,
		}
	}

	return protodaemon.IssueJWTResponse{
		JWT: userJWT,
	}
}

func (h *NATSHandler) CreateAgent(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.CreateAgentRequest)
	if !ok {
		return protodaemon.CreateAgentResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateCreateAgentRequest(req)
	if err != nil {
		return protodaemon.CreateAgentResponse{
			Error: err,
		}
	}

	agentName := req.Name
	agentDesc := req.Description
	gatewayRef := req.GatewayRef
	agentConf := req.Config

	gateway, err := h.conf.GatewayDirectory.Get(ctx, gatewayRef)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", gatewayRef)
		}
		h.logger.ErrorContext(ctx, "Cannot get gateway: %v", err)
		return protodaemon.CreateAgentResponse{
			Error: err,
		}
	}

	extra := make(map[string]string, len(gateway.Config.Extra)+len(agentConf.Extra))
	maps.Copy(extra, gateway.Config.Extra)
	maps.Copy(extra, agentConf.Extra)
	agentConf.Extra = extra

	agentKeyPair, err := auth.NewUserKey()
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot generate a user key: %v", err)
		return protodaemon.CreateAgentResponse{
			Error: err,
		}
	}

	agentID, err := agentKeyPair.PublicKey()
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot create agent ID: %v", err)
		return protodaemon.CreateAgentResponse{
			Error: err,
		}
	}

	agentClaims, err := auth.NewUserJWT(agentName, jwt.Permissions{}, agentKeyPair)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot generate a user JWT: %v", err)
		return protodaemon.CreateAgentResponse{
			Error: err,
		}
	}

	agentJWT, err := agentClaims.Encode(h.conf.AccountKey)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot encode user JWT: %v", err)
		return protodaemon.CreateAgentResponse{
			Error: err,
		}
	}

	agentSeed, err := agentKeyPair.Seed()
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot encode user seed: %v", err)
		return protodaemon.CreateAgentResponse{
			Error: err,
		}
	}

	agent, err := h.conf.AgentDirectory.Create(ctx, directory.AgentDirectoryEntry{
		ID:          agentID,
		Name:        agentName,
		Description: agentDesc,
		GatewayID:   gateway.ID,
		Config: directory.AgentBuildConfig{
			OS:      agentConf.OS,
			Arch:    agentConf.Arch,
			UserJWT: agentJWT,
			UserKey: string(agentSeed),
			Extra:   agentConf.Extra,
		},
		Commands:  copyProtoCommands(req.Commands),
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, directory.ErrKeyExists) {
			err = fmt.Errorf("%q already exists", agentName)
		}
		h.logger.ErrorContext(ctx, "Cannot create agent: %v", err)
		return protodaemon.CreateAgentResponse{
			Error: err,
		}
	}
	defer func() {
		if err == nil {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = h.conf.AgentDirectory.Delete(cleanupCtx, agent)
		if err != nil {
			h.logger.ErrorContext(ctx, "Cannot delete agent: %v", err)
		}
		cancel()
	}()

	resp, err := h.conf.Gateway.RegisterAgent(ctx, gateway.ID, protogw.RegisterAgentRequest{
		AgentID: agent.ID,
		OS:      agent.Config.OS,
		Arch:    agent.Config.Arch,
		Env:     agent.Config.Extra,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot register agent: %v", err)
		return protodaemon.CreateAgentResponse{
			Error: err,
		}
	}

	// Merge all properties together.
	if agent.Config.Extra == nil {
		agent.Config.Extra = make(map[string]string)
	}
	maps.Copy(agent.Config.Extra, resp.Env)

	_, err = h.conf.AgentDirectory.Update(ctx, agent)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot create agent: %v", err)
		return protodaemon.CreateAgentResponse{
			Error: err,
		}
	}

	return protodaemon.CreateAgentResponse{}
}

func (h *NATSHandler) GetAgent(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.GetAgentRequest)
	if !ok {
		return protodaemon.GetAgentResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateGetAgentRequest(req)
	if err != nil {
		return protodaemon.GetAgentResponse{
			Error: err,
		}
	}

	agentRef := req.AgentRef

	agent, err := h.conf.AgentDirectory.Get(ctx, agentRef)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", agentRef)
		}
		return protodaemon.GetAgentResponse{
			Error: err,
		}
	}

	var gatewayName string
	gateway, err := h.conf.GatewayDirectory.Get(ctx, agent.GatewayID)
	if err != nil {
		if !errors.Is(err, directory.ErrKeyNotFound) {
			h.logger.ErrorContext(ctx, "Cannot get gateway: %v", err)
		}
	} else {
		gatewayName = gateway.Name
	}

	return protodaemon.GetAgentResponse{
		Agent: protodaemon.Agent{
			ID:          agent.ID,
			Name:        agent.Name,
			Description: agent.Description,
			GatewayID:   agent.GatewayID,
			GatewayName: gatewayName,
			Config: protodaemon.AgentConfig{
				OS:    agent.Config.OS,
				Arch:  agent.Config.Arch,
				JWT:   agent.Config.UserJWT,
				Key:   agent.Config.UserKey,
				Extra: agent.Config.Extra,
			},
			Commands:  copyDirectoryCommands(agent.Commands),
			CreatedAt: agent.CreatedAt.Unix(),
		},
	}
}

func (h *NATSHandler) ListAgents(ctx context.Context, params map[string]string, msg message.Msg) any {
	_, ok := msg.Data().(protodaemon.ListAgentsRequest)
	if !ok {
		return protodaemon.ListAgentsResponse{
			Error: errors.New("invalid request data"),
		}
	}

	entries, err := h.conf.AgentDirectory.List(ctx)
	if err != nil {
		return protodaemon.ListAgentsResponse{
			Error: err,
		}
	}

	gateways, err := h.conf.GatewayDirectory.List(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot list gateways: %v", err)
	}

	names := make(map[string]string, len(gateways))
	for _, gateway := range gateways {
		names[gateway.ID] = gateway.Name
	}

	agents := make([]protodaemon.Agent, len(entries))
	for i, entry := range entries {
		agents[i] = protodaemon.Agent{
			ID:          entry.ID,
			Name:        entry.Name,
			Description: entry.Description,
			GatewayID:   entry.GatewayID,
			GatewayName: names[entry.GatewayID],
			Config: protodaemon.AgentConfig{
				OS:    entry.Config.OS,
				Arch:  entry.Config.Arch,
				JWT:   entry.Config.UserJWT,
				Key:   entry.Config.UserKey,
				Extra: entry.Config.Extra,
			},
			Commands:  copyDirectoryCommands(entry.Commands),
			CreatedAt: entry.CreatedAt.Unix(),
		}
	}

	return protodaemon.ListAgentsResponse{
		Agents: agents,
	}
}

func (h *NATSHandler) DeleteAgent(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.DeleteAgentRequest)
	if !ok {
		return protodaemon.DeleteAgentResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateDeleteAgentRequest(req)
	if err != nil {
		return protodaemon.DeleteAgentResponse{
			Error: err,
		}
	}

	agent, err := h.conf.AgentDirectory.Get(ctx, req.AgentRef)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", req.AgentRef)
		}
		return protodaemon.DeleteAgentResponse{
			Error: err,
		}
	}

	_, err = h.conf.Gateway.UnregisterAgent(ctx, agent.GatewayID, protogw.UnregisterAgentRequest{
		AgentID: agent.ID,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot unregister agent: %v", err)
		return protodaemon.DeleteAgentResponse{
			Error: err,
		}
	}

	err = h.conf.AgentDirectory.Delete(ctx, agent)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot delete agent: %v", err)
		return protodaemon.DeleteAgentResponse{
			Error: err,
		}
	}

	return protodaemon.DeleteAgentResponse{}
}

func (h *NATSHandler) GetCommand(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.GetCommandRequest)
	if !ok {
		return protodaemon.GetCommandResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateGetCommandRequest(req)
	if err != nil {
		return protodaemon.GetCommandResponse{
			Error: err,
		}
	}

	agent, err := h.conf.AgentDirectory.Get(ctx, req.AgentRef)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", req.AgentRef)
		}
		return protodaemon.GetCommandResponse{
			Error: err,
		}
	}

	command, ok := agent.Commands[req.Name]
	if !ok {
		return protodaemon.GetCommandResponse{
			Error: fmt.Errorf("%q not found", req.Name),
		}
	}

	copied := copyDirectoryCommands(map[string]directory.AgentCommand{
		req.Name: command,
	})

	return protodaemon.GetCommandResponse{
		Command: copied[req.Name],
	}
}

func (h *NATSHandler) ListCommands(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.ListCommandsRequest)
	if !ok {
		return protodaemon.ListCommandsResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateListCommandsRequest(req)
	if err != nil {
		return protodaemon.ListCommandsResponse{
			Error: err,
		}
	}

	agent, err := h.conf.AgentDirectory.Get(ctx, req.AgentRef)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", req.AgentRef)
		}
		return protodaemon.ListCommandsResponse{
			Error: err,
		}
	}

	return protodaemon.ListCommandsResponse{
		Commands: copyDirectoryCommands(agent.Commands),
	}
}

func (h *NATSHandler) CreateGateway(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.CreateGatewayRequest)
	if !ok {
		return protodaemon.CreateGatewayResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateCreateGatewayRequest(req)
	if err != nil {
		return protodaemon.CreateGatewayResponse{
			Error: err,
		}
	}

	gatewayName := req.Name
	gatewayDesc := req.Description
	gatewayConf := req.Config

	gatewayKeyPair, err := auth.NewUserKey()
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot generate a user key: %v", err)
		return protodaemon.CreateGatewayResponse{
			Error: err,
		}
	}

	gatewayID, err := gatewayKeyPair.PublicKey()
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot create gateway ID: %v", err)
		return protodaemon.CreateGatewayResponse{
			Error: err,
		}
	}

	gatewayPerms := auth.NewGatewayPermissions(gatewayID)

	gatewayClaims, err := auth.NewUserJWT(gatewayName, gatewayPerms, gatewayKeyPair)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot generate a user JWT: %v", err)
		return protodaemon.CreateGatewayResponse{
			Error: err,
		}
	}

	gatewayJWT, err := gatewayClaims.Encode(h.conf.AccountKey)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot encode user JWT: %v", err)
		return protodaemon.CreateGatewayResponse{
			Error: err,
		}
	}

	gatewaySeed, err := gatewayKeyPair.Seed()
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot encode user seed: %v", err)
		return protodaemon.CreateGatewayResponse{
			Error: err,
		}
	}

	gateway, err := h.conf.GatewayDirectory.Create(ctx, directory.GatewayDirectoryEntry{
		ID:          gatewayID,
		Name:        gatewayName,
		Description: gatewayDesc,
		Config: directory.GatewayConfig{
			UserJWT: gatewayJWT,
			UserKey: string(gatewaySeed),
			Extra:   gatewayConf.Extra,
		},
	})
	if err != nil {
		if errors.Is(err, directory.ErrKeyExists) {
			err = fmt.Errorf("%q already exists", gatewayName)
		}
		h.logger.ErrorContext(ctx, "Cannot create gateway: %v", err)
		return protodaemon.CreateGatewayResponse{
			Error: err,
		}
	}
	defer func() {
		if err == nil {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = h.conf.GatewayDirectory.Delete(cleanupCtx, gateway)
		if err != nil {
			h.logger.ErrorContext(ctx, "Cannot delete gateway: %v", err)
		}
		cancel()
	}()

	err = h.conf.Gateway.RegisterGateway(ctx, gateway.ID)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot register gateway: %v", err)
		return protodaemon.CreateGatewayResponse{
			Error: err,
		}
	}

	return protodaemon.CreateGatewayResponse{}
}

func (h *NATSHandler) GetGateway(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.GetGatewayRequest)
	if !ok {
		return protodaemon.GetGatewayResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateGetGatewayRequest(req)
	if err != nil {
		return protodaemon.GetGatewayResponse{
			Error: err,
		}
	}

	gatewayRef := req.GatewayRef

	gateway, err := h.conf.GatewayDirectory.Get(ctx, gatewayRef)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", gatewayRef)
		}
		return protodaemon.GetGatewayResponse{
			Error: err,
		}
	}

	return protodaemon.GetGatewayResponse{
		Gateway: protodaemon.Gateway{
			ID:          gateway.ID,
			Name:        gateway.Name,
			Description: gateway.Description,
			Config: protodaemon.GatewayConfig{
				JWT:   gateway.Config.UserJWT,
				Key:   gateway.Config.UserKey,
				Extra: gateway.Config.Extra,
			},
		},
	}
}

func (h *NATSHandler) ListGateways(ctx context.Context, params map[string]string, msg message.Msg) any {
	_, ok := msg.Data().(protodaemon.ListGatewaysRequest)
	if !ok {
		return protodaemon.ListGatewaysResponse{
			Error: errors.New("invalid request data"),
		}
	}

	entries, err := h.conf.GatewayDirectory.List(ctx)
	if err != nil {
		return protodaemon.ListGatewaysResponse{
			Error: err,
		}
	}

	gateways := make([]protodaemon.Gateway, len(entries))
	for i := range entries {
		gateways[i] = protodaemon.Gateway{
			ID:          entries[i].ID,
			Name:        entries[i].Name,
			Description: entries[i].Description,
			Config: protodaemon.GatewayConfig{
				JWT:   entries[i].Config.UserJWT,
				Key:   entries[i].Config.UserKey,
				Extra: entries[i].Config.Extra,
			},
		}
	}

	return protodaemon.ListGatewaysResponse{
		Gateways: gateways,
	}
}

func (h *NATSHandler) DeleteGateway(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.DeleteGatewayRequest)
	if !ok {
		return protodaemon.DeleteGatewayResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateDeleteGatewayRequest(req)
	if err != nil {
		return protodaemon.DeleteGatewayResponse{
			Error: err,
		}
	}

	gateway, err := h.conf.GatewayDirectory.Get(ctx, req.GatewayRef)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", req.GatewayRef)
		}
		return protodaemon.DeleteGatewayResponse{
			Error: err,
		}
	}

	err = h.conf.Gateway.UnregisterGateway(ctx, gateway.ID)
	if err != nil && !errors.Is(err, jetstream.ErrConsumerNotFound) {
		h.logger.ErrorContext(ctx, "Cannot unregister gateway: %v", err)
		return protodaemon.DeleteGatewayResponse{
			Error: err,
		}
	}

	err = h.conf.GatewayDirectory.Delete(ctx, gateway)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot delete gateway: %v", err)
		return protodaemon.DeleteGatewayResponse{
			Error: err,
		}
	}

	return protodaemon.DeleteGatewayResponse{}
}

func (h *NATSHandler) CreateJob(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.CreateJobRequest)
	if !ok {
		return protodaemon.CreateJobResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateCreateJobRequest(req)
	if err != nil {
		return protodaemon.CreateJobResponse{
			Error: err,
		}
	}

	agentRef := req.AgentRef

	agent, err := h.conf.AgentDirectory.Get(ctx, agentRef)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", agentRef)
		}
		h.logger.ErrorContext(ctx, "Cannot get agent: %v", err)
		return protodaemon.CreateJobResponse{
			Error: err,
		}
	}

	jobID := nuid.Next()

	job, err := h.conf.JobDirectory.Create(ctx, directory.JobDirectoryEntry{
		ID:        jobID,
		AgentID:   agent.ID,
		WorkerID:  "",
		GatewayID: agent.GatewayID,
		Config: directory.JobConfig{
			Command: req.Command,
			Args:    req.Args,
			Env:     req.Env,
		},
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, directory.ErrKeyExists) {
			err = fmt.Errorf("%q already exists", jobID)
		}
		h.logger.ErrorContext(ctx, "Cannot create job: %v", err)
		return protodaemon.CreateJobResponse{
			Error: err,
		}
	}
	defer func() {
		if err == nil {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = h.conf.JobDirectory.Delete(cleanupCtx, job)
		if err != nil {
			h.logger.ErrorContext(ctx, "Cannot delete job: %v", err)
		}
		cancel()
	}()

	// TODO: create job with the gateway!
	// TODO: update job directory entry!

	return protodaemon.CreateJobResponse{
		JobID: job.ID,
	}
}

func (h *NATSHandler) GetJob(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.GetJobRequest)
	if !ok {
		return protodaemon.GetJobResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateGetJobRequest(req)
	if err != nil {
		return protodaemon.GetJobResponse{
			Error: err,
		}
	}

	job, err := h.conf.JobDirectory.Get(ctx, req.JobRef)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", req.JobRef)
		}
		return protodaemon.GetJobResponse{
			Error: err,
		}
	}

	var gatewayName string
	gateway, err := h.conf.GatewayDirectory.Get(ctx, job.GatewayID)
	if err != nil {
		if !errors.Is(err, directory.ErrKeyNotFound) {
			h.logger.ErrorContext(ctx, "Cannot get gateway: %v", err)
		}
	} else {
		gatewayName = gateway.Name
	}

	var agentName string
	agent, err := h.conf.AgentDirectory.Get(ctx, job.AgentID)
	if err != nil {
		if !errors.Is(err, directory.ErrKeyNotFound) {
			h.logger.ErrorContext(ctx, "Cannot get agent: %v", err)
		}
	} else {
		agentName = agent.Name
	}

	return protodaemon.GetJobResponse{
		Job: protodaemon.Job{
			ID:          job.ID,
			GatewayID:   job.GatewayID,
			GatewayName: gatewayName,
			AgentID:     job.AgentID,
			AgentName:   agentName,
			WorkerID:    job.WorkerID,
			Command:     job.Config.Command,
			Args:        job.Config.Args,
			Env:         job.Config.Env,
			Status: protodaemon.JobStatus{
				Status:    job.State.Status,
				Error:     job.State.Error,
				UpdatedAt: job.State.UpdatedAt.Unix(),
			},
			CreatedAt: job.CreatedAt.Unix(),
		},
	}
}

func (h *NATSHandler) ListJobs(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.ListJobsRequest)
	if !ok {
		return protodaemon.ListJobsResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateListJobsRequest(req)
	if err != nil {
		return protodaemon.ListJobsResponse{
			Error: err,
		}
	}

	var entries []directory.JobDirectoryEntry
	if req.AgentRef == "" {
		entries, err = h.conf.JobDirectory.List(ctx)
	} else {
		entries, err = h.conf.JobDirectory.ListByAgentID(ctx, req.AgentRef)
	}
	if err != nil {
		return protodaemon.ListJobsResponse{
			Error: err,
		}
	}

	gateways, err := h.conf.GatewayDirectory.List(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot list gateways: %v", err)
	}

	gatewayNames := make(map[string]string, len(gateways))
	for _, gateway := range gateways {
		gatewayNames[gateway.ID] = gateway.Name
	}

	agents, err := h.conf.AgentDirectory.List(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot list agents: %v", err)
	}
	agentNames := make(map[string]string, len(agents))
	for _, agent := range agents {
		agentNames[agent.ID] = agent.Name
	}

	jobs := make([]protodaemon.Job, len(entries))
	for i := range entries {
		jobs[i] = protodaemon.Job{
			ID:          entries[i].ID,
			GatewayID:   entries[i].GatewayID,
			GatewayName: gatewayNames[entries[i].GatewayID],
			AgentID:     entries[i].AgentID,
			AgentName:   agentNames[entries[i].AgentID],
			WorkerID:    entries[i].WorkerID,
			Command:     entries[i].Config.Command,
			Args:        entries[i].Config.Args,
			Env:         entries[i].Config.Env,
			Status: protodaemon.JobStatus{
				Status:    entries[i].State.Status,
				Error:     entries[i].State.Error,
				UpdatedAt: entries[i].State.UpdatedAt.Unix(),
			},
			CreatedAt: entries[i].CreatedAt.Unix(),
		}
	}

	return protodaemon.ListJobsResponse{
		Jobs: jobs,
	}
}

func (h *NATSHandler) CancelJob(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.CancelJobRequest)
	if !ok {
		return protodaemon.CancelJobResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateCancelJobRequest(req)
	if err != nil {
		return protodaemon.CancelJobResponse{
			Error: err,
		}
	}

	job, err := h.conf.JobDirectory.Get(ctx, req.JobRef)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", req.JobRef)
		}
		return protodaemon.CancelJobResponse{
			Error: err,
		}
	}

	_, err = h.conf.Gateway.CancelJob(ctx, job.GatewayID, protogw.CancelJobRequest{
		AgentID: job.AgentID,
		JobID:   job.ID,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot cancel job: %v", err)
		return protodaemon.CancelJobResponse{
			Error: err,
		}
	}

	return protodaemon.CancelJobResponse{}
}

func (h *NATSHandler) GetJobOutput(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.GetJobOutputRequest)
	if !ok {
		return protodaemon.GetJobOutputResponse{
			Error: errors.New("invalid request data"),
		}
	}

	err := ValidateGetJobOutputRequest(req)
	if err != nil {
		return protodaemon.GetJobOutputResponse{
			Error: err,
		}
	}

	job, err := h.conf.JobDirectory.Get(ctx, req.JobRef)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", req.JobRef)
		}
		return protodaemon.GetJobOutputResponse{
			Error: err,
		}
	}

	resp, err := h.conf.Gateway.ReadJobOutput(ctx, job.GatewayID, protogw.JobOutputRequest{
		AgentID: job.AgentID,
		JobID:   job.ID,
		Offset:  req.Offset,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot get job output: %v", err)
		return protodaemon.GetJobOutputResponse{
			Error: err,
		}
	}

	return protodaemon.GetJobOutputResponse{
		Data: resp.Data,
	}
}

type NATSHandlerMessage struct {
	msg     message.Msg
	subject string
	data    any
}

func (m NATSHandlerMessage) ID() string {
	return m.msg.ID()
}

func (m NATSHandlerMessage) Subject() string {
	return m.subject
}

func (m NATSHandlerMessage) ReplySubject() string {
	return ""
}

func (m NATSHandlerMessage) Data() any {
	return m.data
}

func (m NATSHandlerMessage) Ack() error {
	return m.msg.Ack()
}

func match(route, subject string) (map[string]string, bool) {
	// Split route and subject into segments
	routeParts := strings.Split(route, ".")
	subjectParts := strings.Split(subject, ".")

	// Check if lengths match
	if len(routeParts) != len(subjectParts) {
		return nil, false
	}

	// Initialize result map for variables
	params := make(map[string]string)

	// Compare each segment
	for i, routePart := range routeParts {
		// Check if routePart is a variable (starts with < and ends with >)
		if strings.HasPrefix(routePart, "<") && strings.HasSuffix(routePart, ">") {
			// Extract variable name (remove < and >)
			varName := strings.TrimPrefix(strings.TrimSuffix(routePart, ">"), "<")
			// Validate variable name doesn't contain < or >
			if strings.Contains(varName, "<") || strings.Contains(varName, ">") {
				return nil, false
			}
			// Store variable name and corresponding subject value
			params[varName] = subjectParts[i]
		} else {
			// If not a variable, segments must match exactly
			if routePart != subjectParts[i] {
				return nil, false
			}
		}
	}

	return params, true
}

func copyProtoCommands(in map[string]protodaemon.AgentCommand) map[string]directory.AgentCommand {
	out := make(map[string]directory.AgentCommand, len(in))
	for name, command := range in {
		flags := make(map[string]directory.AgentCommandFlag, len(command.Flags))
		for flagName, flag := range command.Flags {
			flags[flagName] = directory.AgentCommandFlag{
				Description: flag.Description,
				Default:     flag.Default,
				Required:    flag.Required,
			}
		}

		env := make(map[string]directory.AgentCommandVar, len(command.Environment))
		for envName, variable := range command.Environment {
			env[envName] = directory.AgentCommandVar{
				Value:       variable.Value,
				Description: variable.Description,
			}
		}

		out[name] = directory.AgentCommand{
			Description: command.Description,
			Flags:       flags,
			Environment: env,
		}
	}
	return out
}

func copyDirectoryCommands(in map[string]directory.AgentCommand) map[string]protodaemon.AgentCommand {
	out := make(map[string]protodaemon.AgentCommand, len(in))
	for name, command := range in {
		flags := make(map[string]protodaemon.AgentCommandFlag, len(command.Flags))
		for flagName, flag := range command.Flags {
			flags[flagName] = protodaemon.AgentCommandFlag{
				Description: flag.Description,
				Default:     flag.Default,
				Required:    flag.Required,
			}
		}

		env := make(map[string]protodaemon.AgentCommandVar, len(command.Environment))
		for envName, variable := range command.Environment {
			env[envName] = protodaemon.AgentCommandVar{
				Value:       variable.Value,
				Description: variable.Description,
			}
		}

		out[name] = protodaemon.AgentCommand{
			Description: command.Description,
			Flags:       flags,
			Environment: env,
		}
	}
	return out
}
