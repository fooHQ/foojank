package commands

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/urfave/cli/v2"
	"time"
)

func NewListCommand(nc *nats.Conn) *cli.Command {
	return &cli.Command{
		Name:   "list",
		Action: newListCommandAction(nc),
	}
}

func newListCommandAction(nc *nats.Conn) cli.ActionFunc {
	return func(c *cli.Context) error {
		ctx := c.Context
		inbox := nats.NewInbox()
		msgCh := make(chan *nats.Msg, 1024)
		sub, err := nc.ChanSubscribe(inbox, msgCh)
		if err != nil {
			return err
		}
		defer sub.Drain()

		// TODO: configurable service name!
		subject, err := micro.ControlSubject(micro.InfoVerb, "vessel", "")
		if err != nil {
			return err
		}

		reqMsg := &nats.Msg{
			Subject: subject,
			Reply:   inbox,
		}
		err = nc.PublishMsg(reqMsg)
		if err != nil {
			return err
		}

		for {
			select {
			case msg := <-msgCh:
				if len(msg.Data) == 0 {
					continue
				}

				var data micro.Info
				err := json.Unmarshal(msg.Data, &data)
				if err != nil {
					return err
				}

				username := data.Metadata["user"]
				hostname := data.Metadata["hostname"]
				osName := data.Metadata["os"]
				fmt.Printf("ID: %s WHO: %s@%s OS: %s\n", data.ID, username, hostname, osName)

			case <-time.After(1 * time.Second): // TODO: configurable timeout!
				return nil

			case <-ctx.Done():
				return nil
			}
		}
	}
}
