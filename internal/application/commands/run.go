package commands

import (
	"bufio"
	"fmt"
	vesselcli "github.com/foojank/foojank/clients/vessel"
	"github.com/muesli/cancelreader"
	"github.com/urfave/cli/v2"
	"os"
	"sync"
)

func NewRunCommand(vessel *vesselcli.Client) *cli.Command {
	return &cli.Command{
		Name: "run",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "id",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "script",
				Required: true,
			},
		},
		Action: newRunCommandAction(vessel),
	}
}

func newRunCommandAction(vessel *vesselcli.Client) cli.ActionFunc {
	return func(c *cli.Context) error {
		id := c.String("id")
		script := c.String("script")

		// TODO: make configurable!
		var serviceName = "vessel"

		ctx := c.Context
		info, err := vessel.GetInfo(ctx, vesselcli.NewID(serviceName, id))
		if err != nil {
			return err
		}

		wid, err := vessel.CreateWorker(ctx, info)
		if err != nil {
			return err
		}

		workerID, err := vessel.GetWorker(ctx, info, wid)
		if err != nil {
			return err
		}

		worker, err := vessel.GetInfo(ctx, workerID)
		if err != nil {
			return err
		}

		stdinCh := make(chan []byte, 128)
		stdoutCh := make(chan []byte, 1024)
		exitCh := make(chan int64, 1)

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case line, ok := <-stdoutCh:
					if !ok {
						return
					}
					fmt.Print(string(line))
				}
			}
		}()

		r, err := cancelreader.NewReader(os.Stdin)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			code, err := vessel.Execute(ctx, worker, stdinCh, stdoutCh, []byte(script))
			if err != nil {
				// TODO: handle error!
				//  return error message + code (define which codes should be used!)
			}

			// Cancel stdin scanner to unblock the main loop.
			_ = r.Cancel()
			exitCh <- code
		}()

		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text() + "\n"
			select {
			case stdinCh <- []byte(line):
			case <-ctx.Done():
				return nil
			default:
			}
		}

		close(stdoutCh)
		wg.Wait()

		code := <-exitCh
		if code != 0 {
			return cli.Exit("", int(code))
		}

		return nil

		/*
			{
				arena := capnp.SingleSegment(nil)
				_, seg, err := capnp.NewMessage(arena)
				if err != nil {
					return err
				}

				msg, err := proto.NewRootMessage(seg)
				if err != nil {
					return err
				}

				msgCreateWorker, err := proto.NewCreateWorkerRequest(seg)
				if err != nil {
					return err
				}

				err = msg.Action().SetCreateWorker(msgCreateWorker)
				if err != nil {
					return err
				}

				b, err := msg.Message().Marshal()
				if err != nil {
					return err
				}

				reqMsg := &nats.Msg{
					Subject: "$SRV.RPC.vessel." + id,
					Reply:   nats.NewInbox(),
					Data:    b,
				}
				respMsg, err := nc.RequestMsgWithContext(ctx, reqMsg)
				if err != nil {
					return err
				}

				capMsg, err := capnp.Unmarshal(respMsg.Data)
				if err != nil {
					return err
				}

				message, err := proto.ReadRootMessage(capMsg)
				if err != nil {
					return err
				}

				v, err := message.Response().CreateWorker()
				if err != nil {
					return err
				}

				workerID = v.Id()
			}

			var serviceID string

			{
				arena := capnp.SingleSegment(nil)
				_, seg, err := capnp.NewMessage(arena)
				if err != nil {
					return err
				}

				msg, err := proto.NewRootMessage(seg)
				if err != nil {
					return err
				}

				msgGetWorker, err := proto.NewGetWorkerRequest(seg)
				if err != nil {
					return err
				}

				msgGetWorker.SetId(workerID)

				err = msg.Action().SetGetWorker(msgGetWorker)
				if err != nil {
					return err
				}

				b, err := msg.Message().Marshal()
				if err != nil {
					return err
				}

				reqMsg := &nats.Msg{
					Subject: "$SRV.RPC.vessel." + id,
					Reply:   nats.NewInbox(),
					Data:    b,
				}
				respMsg, err := nc.RequestMsgWithContext(ctx, reqMsg)
				if err != nil {
					return err
				}

				errHeader := respMsg.Header.Get(micro.ErrorHeader)
				if errHeader != "" {
					println("err", errHeader)
					return nil
				}

				capMsg, err := capnp.Unmarshal(respMsg.Data)
				if err != nil {
					return err
				}

				message, err := proto.ReadRootMessage(capMsg)
				if err != nil {
					return err
				}

				v, err := message.Response().GetWorker()
				if err != nil {
					return err
				}

				serviceID, err = v.ServiceId()
				if err != nil {
					return err
				}
			}

			var stdinSubject string
			var stdoutSubject string
			var data string

			{
				inbox := nats.NewInbox()
				msgCh := make(chan *nats.Msg, 512)
				sub, err := nc.ChanSubscribe(inbox, msgCh)
				if err != nil {
					return err
				}
				defer sub.Drain()

				// TODO: service name is configurable in the agent!
				subject, err := micro.ControlSubject(micro.InfoVerb, "vessel-worker", serviceID)
				if err != nil {
					return err
				}

				reqMsg := &nats.Msg{
					Subject: subject,
					Reply:   inbox,
				}
				resp, err := nc.RequestMsgWithContext(ctx, reqMsg)
				if err != nil {
					return err
				}

				var info micro.Info
				err = json.Unmarshal(resp.Data, &info)
				if err != nil {
					return err
				}

				stdoutSubject = info.Metadata["stdout"]
				for _, endpoint := range info.Endpoints {
					if endpoint.Name == "stdin" {
						stdinSubject = endpoint.Subject
						continue
					}
					if endpoint.Name == "data" {
						data = endpoint.Subject
						continue
					}
				}
			}

			var script = c.String("script")

			{
				msgCh := make(chan *nats.Msg, 512)
				sub, err := nc.ChanSubscribe(stdoutSubject, msgCh)
				if err != nil {
					return err
				}
				defer sub.Drain()

				// TODO: not compatible with windows!
				stdin := os.NewFile(uintptr(syscall.Stdin), "stdin")
				fd := int(stdin.Fd())

				origState, err := xterm.GetState(fd)
				if err != nil {
					panic(err)
				}
				defer xterm.Restore(fd, origState)

				term := xterm.NewTerminal(stdin, "")

				// Start console write (stdout)!
				go func() {
					for {
						select {
						case msg := <-msgCh:
							_, _ = term.Write(msg.Data)
						case <-ctx.Done():
							return
						}
					}
				}()

				go func() {
					arena := capnp.SingleSegment(nil)
					_, seg, err := capnp.NewMessage(arena)
					if err != nil {
						// TODO: log error!
						return
					}

					msg, err := proto.NewRootMessage(seg)
					if err != nil {
						// TODO: log error!
						return
					}

					msgExecute, err := proto.NewExecuteRequest(seg)
					if err != nil {
						// TODO: log error!
						return
					}

					_ = msgExecute.SetData([]byte(script))

					err = msg.Action().SetExecute(msgExecute)
					if err != nil {
						// TODO: log error!
					}

					b, err := msg.Message().Marshal()
					if err != nil {
						// TODO: log error!
						return
					}

					reqMsg := &nats.Msg{
						Subject: data,
						Reply:   nats.NewInbox(),
						Data:    b,
					}
					_, err = nc.RequestMsgWithContext(ctx, reqMsg)
					if err != nil {
						// TODO: log error!
						return
					}

					_, _ = term.Write([]byte("done!"))
				}()

				for {
					// TODO: check if terminal!
					oldState, err := xterm.MakeRaw(fd)
					if err != nil {
						return err
					}

					err = stdin.SetDeadline(time.Now().Add(3 * time.Second))
					if err != nil {
						return err
					}

					isTimeout := false
					line, err := term.ReadLine()
					if err != nil {
						switch {
						case errors.Is(err, io.EOF):
							return err
						case errors.Is(err, os.ErrDeadlineExceeded):
							isTimeout = true
						default:
							return err
						}
					}

					err = xterm.Restore(fd, oldState)
					if err != nil {
						return err
					}

					if isTimeout {
						continue
					}

					// Adds back the newline eaten by ReadLine.
					line += "\n"

					reqMsg := &nats.Msg{
						Subject: stdinSubject,
						Reply:   nats.NewInbox(),
						Data:    []byte(line),
					}
					err = nc.PublishMsg(reqMsg)
					if err != nil {
						return err
					}
				}
			}*/
	}
}
