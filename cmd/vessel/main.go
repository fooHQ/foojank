package main

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"os"
	"os/signal"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/foohq/foojank/internal/vessel"
	"github.com/foohq/foojank/internal/vessel/dialer"
	"github.com/foohq/foojank/internal/vessel/log"
	"github.com/foohq/foojank/internal/vessel/subjects"
)

var (
	ID                           = ""
	Server                       = ""
	UserJWT                      = ""
	UserKey                      = ""
	CACertificate                = ""
	Stream                       = ""
	Consumer                     = ""
	InboxPrefix                  = ""
	ObjectStoreName              = ""
	SubjectApiWorkerStartT       = ""
	SubjectApiWorkerStopT        = ""
	SubjectApiWorkerWriteStdinT  = ""
	SubjectApiWorkerWriteStdoutT = ""
	SubjectApiWorkerStatusT      = ""
	SubjectApiConnInfoT          = ""
	SubjectApiReplyT             = ""
	ReconnectInterval            = "" // time.Duration
	ReconnectJitter              = "" // time.Duration
	AwaitMessagesDuration        = "" // time.Duration
)

func main() {
	log.Debug("Vessel started")
	defer log.Debug("Vessel stopped")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	connDialer := dialer.New(mustGetAwaitMessagesDuration())
	defer func() {
		_ = connDialer.Close()
	}()

	conn, err := connect(ctx, Server, UserJWT, UserKey, CACertificate, connDialer)
	if err != nil {
		log.Debug("Cannot connect to the server", "server", Server, "error", err)
		return
	}
	defer conn.Conn().Close()

	stream, err := getStream(ctx, conn, Stream)
	if err != nil {
		log.Debug("Cannot obtain stream", "error", err)
		return
	}

	consumer, err := getConsumer(ctx, conn, Stream, Consumer)
	if err != nil {
		log.Debug("Cannot obtain durable consumer", "error", err)
		return
	}

	store, err := getObjectStore(ctx, conn, ObjectStoreName)
	if err != nil {
		log.Debug("Cannot obtain object store", "error", err)
		return
	}

	err = vessel.New(vessel.Arguments{
		ID:          ID,
		Connection:  conn,
		Stream:      stream,
		Consumer:    consumer,
		ObjectStore: store,
		Templates: map[int]string{
			subjects.StartWorker:       SubjectApiWorkerStartT,
			subjects.StopWorker:        SubjectApiWorkerStopT,
			subjects.WorkerWriteStdin:  SubjectApiWorkerWriteStdinT,
			subjects.WorkerWriteStdout: SubjectApiWorkerWriteStdoutT,
			subjects.WorkerStatus:      SubjectApiWorkerStatusT,
			subjects.ConnInfo:          SubjectApiConnInfoT,
			subjects.Reply:             SubjectApiReplyT,
		},
	}).Start(ctx)
	if err != nil {
		log.Debug("Cannot start the agent", "error", err)
		return
	}
}

func connect(
	ctx context.Context,
	server string,
	userJWT,
	userKey,
	caCertificate string,
	dialer nats.CustomDialer,
) (jetstream.JetStream, error) {
	opts := []nats.Option{
		nats.CustomInboxPrefix(InboxPrefix),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(mustGetReconnectInterval()),
		nats.ReconnectJitter(mustGetReconnectJitter(), mustGetReconnectJitter()),
		nats.ConnectHandler(connected),
		nats.ReconnectHandler(connected),
		nats.DisconnectErrHandler(disconnected),
		nats.ErrorHandler(failed),
		nats.SetCustomDialer(dialer),
	}

	if userJWT != "" && userKey != "" {
		opts = append(opts, nats.UserJWTAndSeed(userJWT, userKey))
	}

	if caCertificate != "" {
		opts = append(opts, nats.ClientTLSConfig(nil, decodeCertificateHandler(caCertificate)))
	}

	nc, err := nats.Connect(server, opts...)
	if err != nil {
		return nil, err
	}

	for !nc.IsConnected() {
		select {
		case <-time.After(3 * time.Second):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	jetStream, err := jetstream.New(
		nc,
		jetstream.WithDefaultTimeout(10*time.Second),
		jetstream.WithPublishAsyncTimeout(mustGetReconnectInterval()+mustGetReconnectJitter()+(15*time.Second)),
		jetstream.WithPublishAsyncMaxPending(120),
	)
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

func getStream(ctx context.Context, conn jetstream.JetStream, stream string) (jetstream.Stream, error) {
	s, err := conn.Stream(ctx, stream)
	if err != nil {
		return nil, err
	}

	// IMPORTANT: Info is called so that StreamInfo is cached and retrievable with CachedInfo().
	_, err = s.Info(ctx)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func getObjectStore(ctx context.Context, conn jetstream.JetStream, store string) (jetstream.ObjectStore, error) {
	return conn.ObjectStore(ctx, store)
}

func getConsumer(ctx context.Context, conn jetstream.JetStream, stream, consumer string) (jetstream.Consumer, error) {
	c, err := conn.Consumer(ctx, stream, consumer)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func mustGetReconnectInterval() time.Duration {
	d, err := time.ParseDuration(ReconnectInterval)
	if err != nil {
		panic(err)
	}
	return d
}

func mustGetReconnectJitter() time.Duration {
	d, err := time.ParseDuration(ReconnectJitter)
	if err != nil {
		panic(err)
	}
	return d
}

func mustGetAwaitMessagesDuration() time.Duration {
	d, err := time.ParseDuration(AwaitMessagesDuration)
	if err != nil {
		panic(err)
	}
	return d
}

func connected(_ *nats.Conn) {
	log.Debug("Connected to the server")
}

func disconnected(_ *nats.Conn, err error) {
	if err != nil {
		log.Debug("Disconnected from the server", "error", err)
	} else {
		log.Debug("Disconnected from the server")
	}
}

func failed(_ *nats.Conn, _ *nats.Subscription, err error) {
	log.Debug("Connection error", "error", err)
}
