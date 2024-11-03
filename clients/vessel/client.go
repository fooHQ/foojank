package vessel

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/foohq/foojank/proto"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"sync"
)

type Client struct {
	nc *nats.Conn
}

func New(nc *nats.Conn) *Client {
	return &Client{
		nc: nc,
	}
}

type ID struct {
	serviceName string
	serviceID   string
}

func (i ID) String() string {
	return i.serviceName + "." + i.serviceID
}

func (i ID) infoSubject() string {
	subject, _ := micro.ControlSubject(micro.InfoVerb, i.serviceName, i.serviceID)
	return subject
}

func NewID(serviceName, id string) ID {
	return ID{
		serviceName: serviceName,
		serviceID:   id,
	}
}

type Endpoint struct {
	name    string
	subject string
}

type Service struct {
	ID        ID
	Metadata  map[string]string
	Endpoints map[string]Endpoint
}

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message + " " + "(" + e.Code + ")"
}

func (c *Client) Discover(ctx context.Context, serviceName string, outputCh chan<- Service) error {
	subject, err := micro.ControlSubject(micro.InfoVerb, serviceName, "")
	if err != nil {
		return err
	}

	inbox := nats.NewInbox()
	msgCh := make(chan *nats.Msg, 1024)
	sub, err := c.nc.ChanSubscribe(inbox, msgCh)
	if err != nil {
		return err
	}
	defer func() {
		_ = sub.Drain()
	}()

	reqMsg := &nats.Msg{
		Subject: subject,
		Reply:   inbox,
	}
	err = c.nc.PublishMsg(reqMsg)
	if err != nil {
		return err
	}

	for {
		select {
		case msg := <-msgCh:
			res, err := c.parseInfo(msg.Data)
			if err != nil {
				continue
			}

			select {
			case outputCh <- res:
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (c *Client) GetInfo(ctx context.Context, id ID) (Service, error) {
	subject := id.infoSubject()
	msg := &nats.Msg{
		Subject: subject,
		Reply:   nats.NewInbox(),
	}
	resp, err := c.request(ctx, msg)
	if err != nil {
		return Service{}, err
	}

	return c.parseInfo(resp.Data)
}

func (c *Client) CreateWorker(ctx context.Context, s Service) (uint64, error) {
	b, err := proto.NewCreateWorkerRequest()
	if err != nil {
		return 0, err
	}

	rpcEndpoint, ok := s.Endpoints["rpc"]
	if !ok {
		return 0, errors.New("rpc endpoint not found")
	}

	msg := &nats.Msg{
		Subject: rpcEndpoint.subject,
		Reply:   nats.NewInbox(),
		Data:    b,
	}
	resp, err := c.request(ctx, msg)
	if err != nil {
		return 0, err
	}

	parsed, err := proto.ParseResponse(resp.Data)
	if err != nil {
		return 0, err
	}

	v, ok := parsed.(proto.CreateWorkerResponse)
	if !ok {
		return 0, errors.New("invalid response type")
	}

	return v.ID, nil
}

func (c *Client) GetWorker(ctx context.Context, s Service, workerID uint64) (ID, error) {
	b, err := proto.NewGetWorkerRequest(workerID)
	if err != nil {
		return ID{}, err
	}

	rpcEndpoint, ok := s.Endpoints["rpc"]
	if !ok {
		return ID{}, errors.New("rpc endpoint not found")
	}

	msg := &nats.Msg{
		Subject: rpcEndpoint.subject,
		Reply:   nats.NewInbox(),
		Data:    b,
	}
	resp, err := c.request(ctx, msg)
	if err != nil {
		return ID{}, err
	}

	parsed, err := proto.ParseResponse(resp.Data)
	if err != nil {
		return ID{}, err
	}

	v, ok := parsed.(proto.GetWorkerResponse)
	if !ok {
		return ID{}, errors.New("invalid response type")
	}

	return ID{
		serviceName: v.ServiceName,
		serviceID:   v.ServiceID,
	}, nil
}

func (c *Client) DestroyWorker(ctx context.Context, s Service, workerID uint64) error {
	b, err := proto.NewDestroyWorkerRequest(workerID)
	if err != nil {
		return err
	}

	rpcEndpoint, ok := s.Endpoints["rpc"]
	if !ok {
		return errors.New("rpc endpoint not found")
	}

	msg := &nats.Msg{
		Subject: rpcEndpoint.subject,
		Reply:   nats.NewInbox(),
		Data:    b,
	}
	resp, err := c.request(ctx, msg)
	if err != nil {
		return err
	}

	parsed, err := proto.ParseResponse(resp.Data)
	if err != nil {
		return err
	}

	_, ok = parsed.(proto.DestroyWorkerResponse)
	if !ok {
		return errors.New("invalid response type")
	}

	return nil
}

func (c *Client) Execute(ctx context.Context, s Service, repository, filePath string, stdin <-chan []byte, stdout chan<- []byte) (int64, error) {
	b, err := proto.NewExecuteRequest(repository, filePath)
	if err != nil {
		return 0, err
	}

	dataEndpoint, ok := s.Endpoints["data"]
	if !ok {
		return 0, errors.New("data endpoint not found")
	}

	stdinEndpoint, ok := s.Endpoints["stdin"]
	if !ok {
		return 0, errors.New("stdin endpoint not found")
	}

	stdoutSubject, ok := s.Metadata["stdout"]
	if !ok {
		return 0, errors.New("stdout subject not found")
	}

	msgCh := make(chan *nats.Msg, 1024)
	sub, err := c.nc.ChanSubscribe(stdoutSubject, msgCh)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = sub.Drain()
	}()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start stdout messages passthrough
	wg.Add(1)
	go func() {
		defer wg.Done()
		for loop := true; loop; {
			select {
			case msg := <-msgCh:
				stdout <- msg.Data

			case <-ctx.Done():
				loop = false
				continue
			}
		}

		// Drain the message channel
		close(msgCh)
		for msg := range msgCh {
			stdout <- msg.Data
		}
	}()

	// Start stdin messages passthrough
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case line := <-stdin:
				msg := &nats.Msg{
					Subject: stdinEndpoint.subject,
					Reply:   nats.NewInbox(),
					Data:    line,
				}
				err := c.nc.PublishMsg(msg)
				if err != nil {
					// TODO: handle error!
					return
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	msg := &nats.Msg{
		Subject: dataEndpoint.subject,
		Reply:   nats.NewInbox(),
		Data:    b,
	}
	resp, respErr := c.request(ctx, msg)

	// From this point we know the worker has already responded to our request therefore we can
	// drain the channel and proceed to the shutdown.
	_ = sub.Drain()
	cancel()
	wg.Wait()

	// Delayed error handling
	if respErr != nil {
		return 0, respErr
	}

	parsed, err := proto.ParseResponse(resp.Data)
	if err != nil {
		return 0, err
	}

	v, ok := parsed.(proto.ExecuteResponse)
	if !ok {
		return 0, err
	}

	return v.Code, nil
}

func (c *Client) request(ctx context.Context, msg *nats.Msg) (*nats.Msg, error) {
	resp, err := c.nc.RequestMsgWithContext(ctx, msg)
	if err != nil {
		return nil, err
	}

	code := resp.Header.Get(micro.ErrorCodeHeader)
	errMsg := resp.Header.Get(micro.ErrorHeader)
	if code != "" || errMsg != "" {
		err := &Error{
			Code:    code,
			Message: errMsg,
		}
		return nil, err
	}

	return resp, nil
}

func (c *Client) parseInfo(b []byte) (Service, error) {
	var data micro.Info
	err := json.Unmarshal(b, &data)
	if err != nil {
		return Service{}, err
	}

	metadata := make(map[string]string, len(data.Metadata))
	for k, v := range data.Metadata {
		metadata[k] = v
	}

	endpoints := make(map[string]Endpoint, len(data.Endpoints))
	for _, endpoint := range data.Endpoints {
		endpoints[endpoint.Name] = Endpoint{
			name:    endpoint.Name,
			subject: endpoint.Subject,
		}
	}

	return Service{
		ID: ID{
			serviceName: data.Name,
			serviceID:   data.ID,
		},
		Metadata:  metadata,
		Endpoints: endpoints,
	}, nil
}
