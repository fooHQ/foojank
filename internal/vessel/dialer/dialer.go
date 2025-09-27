package dialer

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

var _ nats.CustomDialer = &Dialer{}

type Dialer struct {
	connMux  sync.Mutex
	cancel   context.CancelFunc
	duration time.Duration
}

func New(duration time.Duration) *Dialer {
	return &Dialer{
		duration: duration,
	}
}

func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	d.connMux.Lock()
	defer d.connMux.Unlock()

	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}

	if d.duration > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), d.duration)
		d.cancel = cancel
		go func() {
			<-ctx.Done()
			_ = conn.Close()
			cancel()
		}()
	}

	return conn, nil
}

func (d *Dialer) Close() error {
	d.connMux.Lock()
	defer d.connMux.Unlock()
	if d.cancel != nil {
		d.cancel()
	}
	return nil
}
