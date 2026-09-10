package handler

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nkeys"
	"github.com/nats-io/nuid"

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/directory"
	"github.com/foohq/foojank/internal/log"
	"github.com/foohq/foojank/internal/privilege"
	protogw "github.com/foohq/foojank/proto/gateway"

	"github.com/foohq/foojank/internal/message"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

type NATSHandlerFn func(context.Context, map[string]string, message.Msg) any

type NATSHandlerConfig struct {
	Connection       jetstream.JetStream
	AgentDirectory   *directory.AgentDirectory
	GatewayDirectory *directory.GatewayDirectory
	UserDirectory    *directory.UserDirectory
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

		protodaemon.CreateAgentSubject(): h.CreateAgent,
		protodaemon.GetAgentSubject():    h.GetAgent,
		protodaemon.ListAgentsSubject():  h.ListAgents,

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

	// TODO: add request validation

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
		Kind:        directory.UserKindClient,
		Privileges:  privilege.FormatPrivileges(userPrivs.Privileges()),
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   time.Time{}, // TODO
	})
	if err != nil {
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

	// TODO: add request validation

	userName := req.Name

	user, err := h.conf.UserDirectory.Get(ctx, userName)
	if err != nil {
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

	subj, err := privilege.JWTIssue{UserID: user.ID}.Subject()
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot create client permissions: %v", err)
		return protodaemon.GetUserResponse{
			Error: err,
		}
	}

	userPerms, err := auth.NewClientPermissions(user.ID, []string{subj})
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

	userJWT, err := claims.Encode(h.conf.AccountKey)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot encode user JWT: %v", err)
		return protodaemon.IssueJWTResponse{
			Error: err,
		}
	}

	return protodaemon.GetUserResponse{
		User: protodaemon.User{
			ID:          user.ID,
			Name:        user.Name,
			Description: user.Description,
			Kind:        user.Kind,
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
			Kind:        entries[i].Kind,
			Privileges:  entries[i].Privileges,
			CreatedAt:   entries[i].CreatedAt.Unix(),
			ExpiresAt:   entries[i].ExpiresAt.Unix(),
		}
	}

	return protodaemon.ListUsersResponse{
		Users: users,
	}
}

func (h *NATSHandler) IssueJWT(ctx context.Context, params map[string]string, msg message.Msg) any {
	userName := params["<user>"]

	user, err := h.conf.UserDirectory.Get(ctx, userName)
	if err != nil {
		return protodaemon.IssueJWTResponse{
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

	// TODO: add request validation

	agentName := req.Name
	agentDesc := req.Description
	gatewayName := req.Gateway
	agentConf := req.Config

	gateway, err := h.conf.GatewayDirectory.Get(ctx, gatewayName)
	if err != nil {
		if errors.Is(err, directory.ErrKeyNotFound) {
			err = fmt.Errorf("%q not found", agentName)
		}
		h.logger.ErrorContext(ctx, "Cannot get gateway: %v", err)
		return protodaemon.CreateAgentResponse{
			Error: err,
		}
	}

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

	agentPerms := auth.NewAgentPermissions(gateway.ID, agentID)

	agentClaims, err := auth.NewUserJWT(agentName, agentPerms, agentKeyPair)
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
		err = h.conf.AgentDirectory.Delete(ctx, agent)
		if err != nil {
			h.logger.ErrorContext(ctx, "Cannot delete agent: %v", err)
		}
	}()

	// TODO: add timeout (all NATS calls should have a timeout!)
	props, err := h.RequestRegisterAgent(ctx, agent)
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
	maps.Copy(agent.Config.Extra, props)

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

	// TODO: add request validation

	agentName := req.Name

	agent, err := h.conf.AgentDirectory.Get(ctx, agentName)
	if err != nil {
		return protodaemon.GetAgentResponse{
			Error: err,
		}
	}

	return protodaemon.GetAgentResponse{
		Agent: protodaemon.Agent{
			ID:          agent.ID,
			Name:        agent.Name,
			Description: agent.Description,
			GatewayID:   agent.GatewayID,
			Config: protodaemon.AgentConfig{
				OS:    agent.Config.OS,
				Arch:  agent.Config.Arch,
				JWT:   agent.Config.UserJWT,
				Key:   agent.Config.UserKey,
				Extra: agent.Config.Extra,
			},
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

	agents := make([]protodaemon.Agent, len(entries))
	for i := range entries {
		agents[i] = protodaemon.Agent{
			ID:          entries[i].ID,
			Name:        entries[i].Name,
			Description: entries[i].Description,
			GatewayID:   entries[i].GatewayID,
			Config: protodaemon.AgentConfig{
				OS:    entries[i].Config.OS,
				Arch:  entries[i].Config.Arch,
				JWT:   entries[i].Config.UserJWT,
				Key:   entries[i].Config.UserKey,
				Extra: entries[i].Config.Extra,
			},
			CreatedAt: entries[i].CreatedAt.Unix(),
		}
	}

	return protodaemon.ListAgentsResponse{
		Agents: agents,
	}
}

func (h *NATSHandler) RequestRegisterAgent(ctx context.Context, agent directory.AgentDirectoryEntry) (map[string]string, error) {
	b, err := protogw.Marshal(protogw.RegisterAgentRequest{
		AgentID: agent.ID,
		OS:      agent.Config.OS,
		Arch:    agent.Config.Arch,
		Env:     agent.Config.Extra,
	})
	if err != nil {
		return nil, err
	}

	resp, err := h.request(ctx, &nats.Msg{
		Subject: protogw.RegisterAgentSubject(agent.GatewayID),
		Data:    b,
	})
	if err != nil {
		return nil, err
	}

	data, err := protogw.Unmarshal(resp.Data)
	if err != nil {
		return nil, err
	}

	v, ok := data.(protogw.RegisterAgentResponse)
	if !ok {
		return nil, fmt.Errorf("invalid response: %T", data)
	}

	if v.Error != nil {
		return nil, v.Error
	}

	return v.Env, nil
}

func (h *NATSHandler) request(ctx context.Context, msg *nats.Msg) (*nats.Msg, error) {
	if msg.Header == nil {
		msg.Header = make(nats.Header)
	}
	msg.Header.Set(nats.MsgIdHdr, nuid.Next())
	return h.conf.Connection.Conn().RequestMsgWithContext(ctx, msg)
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
