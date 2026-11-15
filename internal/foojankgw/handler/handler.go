package handler

import (
	"context"
	"strings"

	"github.com/nutsdb/nutsdb"

	"github.com/foohq/foojank/clients/gateway"
	protogw "github.com/foohq/foojank/proto/gateway"

	"github.com/foohq/foojank/internal/log"
	"github.com/foohq/foojank/internal/message"
)

type Config struct {
	Gateway *gateway.Client
	DB      *nutsdb.DB
}

type Handler struct {
	conf   Config
	logger *log.Logger
}

func New(logger *log.Logger, conf Config) *Handler {
	return &Handler{
		conf:   conf,
		logger: logger,
	}
}

func (h *Handler) Match(msg message.Msg) (func(context.Context) message.Msg, bool) {
	routes := map[string]adapterFn{
		protogw.RegisterAgentSubject("<gateway>"):   adapter(h.RegisterAgent),
		protogw.UnregisterAgentSubject("<gateway>"): adapter(h.UnregisterAgent),
		protogw.CreateJobSubject("<gateway>"):       adapter(h.CreateJob),
		protogw.CancelJobSubject("<gateway>"):       adapter(h.CancelJob),
		protogw.JobInputSubject("<gateway>"):        adapter(h.WriteJobInput),
		protogw.JobOutputSubject("<gateway>"):       adapter(h.ReadJobOutput),
	}
	for route, handler := range routes {
		params, ok := match(route, msg.Subject())
		if !ok {
			continue
		}
		return func(ctx context.Context) message.Msg {
			v := handler(ctx, params, msg.Data())
			return Message{
				msg:     msg,
				subject: msg.ReplySubject(),
				data:    v,
			}
		}, true
	}
	h.logger.WarnContext(context.Background(), "Cannot match message with subject %q to route", msg.Subject())
	return nil, false
}

func (h *Handler) RegisterAgent(ctx context.Context, params map[string]string, req protogw.RegisterAgentRequest) protogw.RegisterAgentResponse {
	return protogw.RegisterAgentResponse{}
}

func (h *Handler) UnregisterAgent(ctx context.Context, params map[string]string, req protogw.UnregisterAgentRequest) protogw.UnregisterAgentResponse {
	return protogw.UnregisterAgentResponse{}
}

func (h *Handler) CreateJob(ctx context.Context, params map[string]string, req protogw.CreateJobRequest) protogw.CreateJobResponse {
	// TODO: validate request!
	agentID := req.AgentID

	err := h.conf.DB.Update(func(tx *nutsdb.Tx) error {
		// TODO: change the bucket name!
		// TODO: assign the value!
		err := tx.RPush("test", []byte(agentID), []byte(agentID), []byte("start_job_request_data"))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		h.logger.WarnContext(ctx, "Cannot create job: %v", err)
		return protogw.CreateJobResponse{
			Error: err,
		}
	}

	return protogw.CreateJobResponse{}
}

func (h *Handler) CancelJob(ctx context.Context, params map[string]string, req protogw.CancelJobRequest) protogw.CancelJobResponse {
	return protogw.CancelJobResponse{}
}

func (h *Handler) WriteJobInput(ctx context.Context, params map[string]string, req protogw.JobInputRequest) protogw.JobInputResponse {
	return protogw.JobInputResponse{}
}

func (h *Handler) ReadJobOutput(ctx context.Context, params map[string]string, req protogw.JobOutputRequest) protogw.JobOutputResponse {
	return protogw.JobOutputResponse{}
}

type Message struct {
	msg     message.Msg
	subject string
	data    any
}

func (m Message) ID() string {
	return m.msg.ID()
}

func (m Message) Subject() string {
	return m.subject
}

func (m Message) ReplySubject() string {
	return ""
}

func (m Message) Data() any {
	return m.data
}

func (m Message) Ack() error {
	return m.msg.Ack()
}

type (
	adapterFn         func(context.Context, map[string]string, any) any
	fn[Req, Resp any] func(context.Context, map[string]string, Req) Resp
)

func adapter[Req, Resp any](fn fn[Req, Resp]) adapterFn {
	return func(ctx context.Context, params map[string]string, req any) any {
		return fn(ctx, params, req.(Req))
	}
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
