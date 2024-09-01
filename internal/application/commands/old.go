package commands

/*{
	Name: "create",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "id",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		arena := capnp.SingleSegment(nil)
		_, seg, err := capnp.NewMessage(arena)
		if err != nil {
			panic(err)
		}

		msg, err := proto.NewRootMessage(seg)
		if err != nil {
			panic(err)
		}

		msgCreateWorker, err := proto.NewCreateWorkerRequest(seg)
		if err != nil {
			panic(err)
		}

		err = msg.Action().SetCreateWorker(msgCreateWorker)
		if err != nil {
			panic(err)
		}

		b, err := msg.Message().Marshal()
		if err != nil {
			panic(err)
		}

		id := c.String("id")
		reqMsg := &nats.Msg{
			Subject: "$SRV.RPC.vessel." + id,
			Reply:   nats.NewInbox(),
			Data:    b,
		}
		respMsg, err := nc.RequestMsgWithContext(ctx, reqMsg)
		if err != nil {
			panic(err)
		}

		capMsg, err := capnp.Unmarshal(respMsg.Data)
		if err != nil {
			panic(err)
		}

		message, err := proto.ReadRootMessage(capMsg)
		if err != nil {
			panic(err)
		}

		v, err := message.Response().CreateWorker()
		fmt.Printf("%v:%v", v.Id(), err)

		return nil
	},
},*/
/*{
	Name: "destroy",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "id",
			Required: true,
		},
		&cli.Uint64Flag{
			Name:     "worker-id",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		arena := capnp.SingleSegment(nil)
		_, seg, err := capnp.NewMessage(arena)
		if err != nil {
			panic(err)
		}

		msg, err := proto.NewRootMessage(seg)
		if err != nil {
			panic(err)
		}

		msgDestroyWorker, err := proto.NewDestroyWorkerRequest(seg)
		if err != nil {
			panic(err)
		}
		msgDestroyWorker.SetId(c.Uint64("worker-id"))

		err = msg.Action().SetDestroyWorker(msgDestroyWorker)
		if err != nil {
			panic(err)
		}

		b, err := msg.Message().Marshal()
		if err != nil {
			panic(err)
		}

		id := c.String("id")
		reqMsg := &nats.Msg{
			Subject: "$SRV.RPC.vessel." + id,
			Reply:   nats.NewInbox(),
			Data:    b,
		}
		respMsg, err := nc.RequestMsgWithContext(ctx, reqMsg)
		if err != nil {
			panic(err)
		}

		errHeader := respMsg.Header.Get(micro.ErrorHeader)
		if errHeader != "" {
			println(errHeader)
		}

		return nil
	},
},*/
/*{
	Name: "get",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "id",
			Required: true,
		},
		&cli.Uint64Flag{
			Name:     "worker-id",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		arena := capnp.SingleSegment(nil)
		_, seg, err := capnp.NewMessage(arena)
		if err != nil {
			panic(err)
		}

		msg, err := proto.NewRootMessage(seg)
		if err != nil {
			panic(err)
		}

		msgGetWorker, err := proto.NewGetWorkerRequest(seg)
		if err != nil {
			panic(err)
		}

		workerID := c.Uint64("worker-id")
		msgGetWorker.SetId(workerID)

		err = msg.Action().SetGetWorker(msgGetWorker)
		if err != nil {
			panic(err)
		}

		b, err := msg.Message().Marshal()
		if err != nil {
			panic(err)
		}

		id := c.String("id")
		reqMsg := &nats.Msg{
			Subject: "$SRV.RPC.vessel." + id,
			Reply:   nats.NewInbox(),
			Data:    b,
		}
		respMsg, err := nc.RequestMsgWithContext(ctx, reqMsg)
		if err != nil {
			panic(err)
		}

		errHeader := respMsg.Header.Get(micro.ErrorHeader)
		if errHeader != "" {
			println("err", errHeader)
			return nil
		}

		capMsg, err := capnp.Unmarshal(respMsg.Data)
		if err != nil {
			panic(err)
		}

		message, err := proto.ReadRootMessage(capMsg)
		if err != nil {
			panic(err)
		}

		v, err := message.Response().GetWorker()
		if err != nil {
			panic(err)
		}

		serviceID, err := v.ServiceId()
		if err != nil {
			panic(err)
		}

		fmt.Printf("WorkerID: %d ServiceID: %s\n", workerID, serviceID)

		return nil
	},
},*/
/*{
	Name: "list",
	Action: func(c *cli.Context) error {
		inbox := nats.NewInbox()
		msgCh := make(chan *nats.Msg, 512)
		sub, err := nc.ChanSubscribe(inbox, msgCh)
		if err != nil {
			panic(err)
		}
		defer sub.Drain()

		reqMsg := &nats.Msg{
			Subject: "$SRV.INFO.vessel", // TODO: configurable service name!
			Reply:   inbox,
		}
		err = nc.PublishMsg(reqMsg)
		if err != nil {
			panic(err)
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

			case <-time.After(1 * time.Second):
				return nil

			case <-ctx.Done():
				return nil
			}
		}

		return nil
	},
},*/
/*{
	Name: "attach",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "id",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		// TODO: check if terminal!
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)

		id := c.String("id")
		prompt := fmt.Sprintf("vessel=> ")
		term := term.NewTerminal(os.Stdin, prompt)

		for {
			line, err := term.ReadLine()
			if err != nil {
				if err != io.EOF {
					panic(err)
				}
				return nil
			}

			reqMsg := &nats.Msg{
				Subject: "$SRV.RPC.vessel." + id,
				Reply:   nats.NewInbox(),
				Data:    []byte(line),
			}
			respMsg, err := nc.RequestMsgWithContext(ctx, reqMsg)
			if err != nil {
				panic(err)
			}

			errHeader := respMsg.Header.Get(micro.ErrorHeader)
			if errHeader != "" {
				_, _ = term.Write([]byte(errHeader))
			} else {
				_, _ = term.Write(respMsg.Data)
			}

			//_, _ = term.Write([]byte("\n"))
		}

		return nil
	},
},*/
