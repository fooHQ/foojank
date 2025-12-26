package proto

import "github.com/foohq/foojank/proto/capnp"

// StartWorkerRequest is a request to start a new worker.
type StartWorkerRequest struct {
	Command string
	Args    []string
	Env     []string
}

func marshalStartWorkerRequest(data StartWorkerRequest) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	m, err := capnp.NewStartWorkerRequest(msg.Segment())
	if err != nil {
		return nil, err
	}

	err = m.SetCommand(data.Command)
	if err != nil {
		return nil, err
	}

	argsList, err := newTextList(msg.Segment(), data.Args)
	if err != nil {
		return nil, err
	}

	err = m.SetArgs(argsList)
	if err != nil {
		return nil, err
	}

	envList, err := newTextList(msg.Segment(), data.Env)
	if err != nil {
		return nil, err
	}

	err = m.SetEnv(envList)
	if err != nil {
		return nil, err
	}

	err = msg.Content().SetStartWorkerRequest(m)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func unmarshalStartWorkerRequest(message capnp.Message) (StartWorkerRequest, error) {
	v, err := message.Content().StartWorkerRequest()
	if err != nil {
		return StartWorkerRequest{}, err
	}

	command, err := v.Command()
	if err != nil {
		return StartWorkerRequest{}, err
	}

	vArgs, err := v.Args()
	if err != nil {
		return StartWorkerRequest{}, err
	}

	args, err := textListToStringSlice(vArgs)
	if err != nil {
		return StartWorkerRequest{}, err
	}

	vEnv, err := v.Env()
	if err != nil {
		return StartWorkerRequest{}, err
	}

	env, err := textListToStringSlice(vEnv)
	if err != nil {
		return StartWorkerRequest{}, err
	}

	return StartWorkerRequest{
		Command: command,
		Args:    args,
		Env:     env,
	}, nil
}
