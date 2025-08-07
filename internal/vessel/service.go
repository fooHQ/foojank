package vessel

import (
	"context"

	"github.com/nats-io/nats.go"
)

type Arguments struct {
	Name       string
	Version    string
	Metadata   map[string]string
	Connection *nats.Conn
}

type Service struct {
	args Arguments
}

func New(args Arguments) *Service {
	return &Service{
		args: args,
	}
}

func (s *Service) Start(ctx context.Context) error {
	panic("temporarily disabled")
	/*
		//----
		conn, err := connect(
			ctx,
			s.args.Servers,
			s.args.UserJWT,
			s.args.UserKey,
			s.args.CACertificate,
		)
		if err != nil {
			return err
		}
		//----
		connectorOutCh := make(chan connector.Message)
		decoderOutCh := make(chan decoder.Message)

		group, groupCtx := errgroup.WithContext(ctx)
		group.Go(func() error {
			return consumer.New(connector.Arguments{
				Name:       s.args.Name,
				Version:    s.args.Version,
				Metadata:   s.args.Metadata,
				RPCSubject: rpcSubject,
				Connection: s.args.Connection,
				OutputCh:   connectorOutCh,
			}).Start(groupCtx)
		})
		group.Go(func() error {
			return decoder.New(decoder.Arguments{
				InputCh:  connectorOutCh,
				OutputCh: decoderOutCh,
			}).Start(groupCtx)
		})
		group.Go(func() error {
			return scheduler.New(scheduler.Arguments{
				Connection: s.args.Connection,
				InputCh:    decoderOutCh,
			}).Start(groupCtx)
		})

		return group.Wait()*/
}

/*func connect(
	ctx context.Context,
	servers []string,
	userJWT,
	userKey,
	caCertificate string,
) (jetstream.JetStream, error) {
	opts := []nats.Option{
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ConnectHandler(func(_ *nats.Conn) {
			log.Debug("connected to the server")
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			log.Debug("reconnected to the server")
		}),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			if err != nil {
				log.Debug("disconnected from the server", "error", err.Error())
			} else {
				log.Debug("disconnected from the server")
			}
		}),
		nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
			log.Debug("server error", "error", err.Error())
		}),
	}

	if userJWT != "" && userKey != "" {
		opts = append(opts, nats.UserJWTAndSeed(userJWT, userKey))
	}

	if caCertificate != "" {
		opts = append(opts, nats.ClientTLSConfig(nil, decodeCertificateHandler(caCertificate)))
	}

	nc, err := nats.Connect(strings.Join(servers, ","), opts...)
	if err != nil {
		return nil, err
	}

	for !nc.IsConnected() {
		select {
		case <-time.After(3 * time.Second):
		case <-ctx.Done():
			return nil, nil
		}
	}

	jetStream, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	return jetStream, nil
}

func decodeCertificateHandler(s string) func() (*x509.CertPool, error) {
	return func() (*x509.CertPool, error) {
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, err
		}

		cert, err := x509.ParseCertificate(b)
		if err != nil {
			return nil, err
		}

		pool := x509.NewCertPool()
		pool.AddCert(cert)
		return pool, nil
	}
}
*/
