package handler

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nkeys"

	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/directory"
	"github.com/foohq/foojank/internal/log"

	"github.com/foohq/foojank/internal/message"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

type NATSHandlerFn func(context.Context, map[string]string, message.Msg) any

type NATSHandlerConfig struct {
	Connection     jetstream.JetStream
	AgentDirectory *directory.AgentDirectory
	UserDirectory  *directory.UserDirectory
	AccountKey     nkeys.KeyPair
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
		protodaemon.CreateUserSubject(): h.createUser,
		protodaemon.GetUserSubject():    h.getUser,
		protodaemon.ListUsersSubject():  h.listUsers,
	}
	for route, handler := range routes {
		params, ok := match(route, msg.Subject())
		if !ok {
			h.logger.WarnContext(context.Background(), "Cannot match route %q with subject %q", route, msg.Subject())
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
	return nil, false
}

func (h *NATSHandler) createUser(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.CreateUserRequest)
	if !ok {
		return protodaemon.CreateUserResponse{
			Error: errors.New("invalid request data"),
		}
	}

	// TODO: add request validation

	userName := req.Name
	userDesc := req.Description
	userPrivs := req.Privileges

	user, err := auth.NewUserKey()
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot generate a user key: %v", err)
		return protodaemon.CreateUserResponse{
			Error: err,
		}
	}

	userPerms, err := auth.NewClientPermissions(userPrivs)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot create client permissions: %v", err)
		return protodaemon.CreateUserResponse{
			Error: err,
		}
	}

	claims, err := auth.NewUserJWT(userName, userPerms, user)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot generate a user JWT: %v", err)
		return protodaemon.CreateUserResponse{
			Error: err,
		}
	}

	userJWT, err := claims.Encode(h.conf.AccountKey)
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot encode user JWT: %v", err)
		return protodaemon.CreateUserResponse{
			Error: err,
		}
	}

	userKey, err := user.Seed()
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot encode user seed: %v", err)
		return protodaemon.CreateUserResponse{
			Error: err,
		}
	}

	userID, err := user.PublicKey()
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot create agent ID: %v", err)
		return protodaemon.CreateUserResponse{
			Error: err,
		}
	}

	_, err = h.conf.UserDirectory.Create(ctx, directory.UserDirectoryEntry{
		ID:          userID,
		Name:        userName,
		Description: userDesc,
		Kind:        directory.UserKindClient,
		IssuedAt:    time.Unix(claims.IssuedAt, 0),
		ExpiresAt:   time.Unix(claims.Expires, 0),
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "Cannot create user: %v", err)
		return protodaemon.CreateUserResponse{
			Error: err,
		}
	}

	return protodaemon.CreateUserResponse{
		JWT: userJWT,
		Key: string(userKey),
	}
}

func (h *NATSHandler) getUser(ctx context.Context, params map[string]string, msg message.Msg) any {
	req, ok := msg.Data().(protodaemon.GetUserRequest)
	if !ok {
		return protodaemon.CreateUserResponse{
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

	return protodaemon.GetUserResponse{
		User: protodaemon.User{
			ID:          user.ID,
			Name:        user.Name,
			Description: user.Description,
			Kind:        user.Kind,
			IssuedAt:    user.IssuedAt.Unix(),
			ExpiresAt:   user.ExpiresAt.Unix(),
		},
	}
}

func (h *NATSHandler) listUsers(ctx context.Context, params map[string]string, msg message.Msg) any {
	return nil
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
