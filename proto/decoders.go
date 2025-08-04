package proto

import (
	"errors"

	capnplib "capnproto.org/go/capnp/v3"

	"github.com/foohq/foojank/proto/capnp"
)

var (
	ErrUnknownMessage = errors.New("unknown message")
)

func Unmarshal(b []byte) (any, error) {
	message, err := parseMessage(b)
	if err != nil {
		return nil, err
	}

	content := message.Content()
	switch {
	case content.HasStartWorkerRequest():
		return unmarshalStartWorkerRequest(message)

	case content.HasStartWorkerResponse():
		return unmarshalStartWorkerResponse(message)

	case content.HasStopWorkerRequest():
		return unmarshalStopWorkerRequest(message)

	case content.HasStopWorkerResponse():
		return unmarshalStopWorkerResponse(message)

	case content.HasUpdateWorkerStatus():
		return unmarshalUpdateWorkerStatus(message)

	case content.HasUpdateWorkerStdio():
		return unmarshalUpdateWorkerStdio(message)

	case content.HasUpdateClientInfo():
		return unmarshalUpdateClientInfo(message)
	}

	return nil, ErrUnknownMessage
}

type StartWorkerRequest struct {
	File string
	Args []string
	Env  []string
}

func unmarshalStartWorkerRequest(message capnp.Message) (StartWorkerRequest, error) {
	v, err := message.Content().StartWorkerRequest()
	if err != nil {
		return StartWorkerRequest{}, err
	}

	file, err := v.File()
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
		File: file,
		Args: args,
		Env:  env,
	}, nil
}

type StartWorkerResponse struct {
	Error error
}

func unmarshalStartWorkerResponse(message capnp.Message) (StartWorkerResponse, error) {
	v, err := message.Content().StartWorkerResponse()
	if err != nil {
		return StartWorkerResponse{}, err
	}

	errMsg, err := v.Error()
	if err != nil {
		return StartWorkerResponse{}, err
	}

	var respErr error
	if errMsg != "" {
		respErr = errors.New(errMsg)
	}

	return StartWorkerResponse{
		Error: respErr,
	}, nil
}

type StopWorkerRequest struct{}

func unmarshalStopWorkerRequest(message capnp.Message) (StopWorkerRequest, error) {
	_, err := message.Content().StopWorkerRequest()
	if err != nil {
		return StopWorkerRequest{}, err
	}

	return StopWorkerRequest{}, nil
}

type StopWorkerResponse struct {
	Error error
}

func unmarshalStopWorkerResponse(message capnp.Message) (StopWorkerResponse, error) {
	v, err := message.Content().StopWorkerResponse()
	if err != nil {
		return StopWorkerResponse{}, err
	}

	errMsg, err := v.Error()
	if err != nil {
		return StopWorkerResponse{}, err
	}

	var respErr error
	if errMsg != "" {
		respErr = errors.New(errMsg)
	}

	return StopWorkerResponse{
		Error: respErr,
	}, nil
}

type UpdateWorkerStatus struct {
	Status int64
}

func unmarshalUpdateWorkerStatus(message capnp.Message) (UpdateWorkerStatus, error) {
	v, err := message.Content().UpdateWorkerStatus()
	if err != nil {
		return UpdateWorkerStatus{}, err
	}

	return UpdateWorkerStatus{
		Status: v.Status(),
	}, nil
}

type UpdateWorkerStdio struct {
	Data []byte
}

func unmarshalUpdateWorkerStdio(message capnp.Message) (UpdateWorkerStdio, error) {
	v, err := message.Content().UpdateWorkerStdio()
	if err != nil {
		return UpdateWorkerStdio{}, err
	}

	data, err := v.Data()
	if err != nil {
		return UpdateWorkerStdio{}, err
	}

	return UpdateWorkerStdio{
		Data: data,
	}, nil
}

type UpdateClientInfo struct {
	Username string
	Hostname string
	System   string
	Address  string
}

func unmarshalUpdateClientInfo(message capnp.Message) (UpdateClientInfo, error) {
	v, err := message.Content().UpdateClientInfo()
	if err != nil {
		return UpdateClientInfo{}, err
	}

	username, err := v.Username()
	if err != nil {
		return UpdateClientInfo{}, err
	}

	hostname, err := v.Hostname()
	if err != nil {
		return UpdateClientInfo{}, err
	}

	system, err := v.System()
	if err != nil {
		return UpdateClientInfo{}, err
	}

	address, err := v.Address()
	if err != nil {
		return UpdateClientInfo{}, err
	}

	return UpdateClientInfo{
		Username: username,
		Hostname: hostname,
		System:   system,
		Address:  address,
	}, nil
}

func parseMessage(b []byte) (capnp.Message, error) {
	capMsg, err := capnplib.Unmarshal(b)
	if err != nil {
		return capnp.Message{}, err
	}

	message, err := capnp.ReadRootMessage(capMsg)
	if err != nil {
		return capnp.Message{}, err
	}

	return message, nil
}

func textListToStringSlice(list capnplib.TextList) ([]string, error) {
	if list.Len() == 0 {
		return nil, nil
	}
	result := make([]string, 0, list.Len())
	for i := 0; i < list.Len(); i++ {
		v, err := list.At(i)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}
