package proto

import (
	capnplib "capnproto.org/go/capnp/v3"
	"errors"
	"github.com/foojank/foojank/proto/capnp"
)

var (
	ErrUnknownAction   = errors.New("unknown action")
	ErrUnknownResponse = errors.New("unknown response")
)

func ParseAction(b []byte) (any, error) {
	message, err := parseMessage(b)
	if err != nil {
		return nil, err
	}

	action := message.Action()
	switch {
	case action.HasCreateWorker():
		return parseCreateWorkerRequest(message)

	case action.HasDestroyWorker():
		return parseDestroyWorkerRequest(message)

	case action.HasGetWorker():
		return parseGetWorkerRequest(message)

	case action.HasExecute():
		return parseExecuteRequest(message)

	case action.HasDummyRequest():
		return parseDummyRequest(message)

	default:
		return nil, ErrUnknownAction
	}
}

func ParseResponse(b []byte) (any, error) {
	message, err := parseMessage(b)
	if err != nil {
		return nil, err
	}

	response := message.Response()
	switch {
	case response.HasCreateWorker():
		return parseCreateWorkerResponse(message)

	case response.HasDestroyWorker():
		return parseDestroyWorkerResponse(message)

	case response.HasGetWorker():
		return parseGetWorkerResponse(message)

	case response.HasExecute():
		return parseExecuteResponse(message)

	default:
		return nil, ErrUnknownResponse
	}
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

type CreateWorkerRequest struct{}

func parseCreateWorkerRequest(message capnp.Message) (CreateWorkerRequest, error) {
	_, err := message.Action().CreateWorker()
	if err != nil {
		return CreateWorkerRequest{}, err
	}

	return CreateWorkerRequest{}, nil
}

type CreateWorkerResponse struct {
	ID uint64
}

func parseCreateWorkerResponse(message capnp.Message) (CreateWorkerResponse, error) {
	v, err := message.Response().CreateWorker()
	if err != nil {
		return CreateWorkerResponse{}, err
	}

	return CreateWorkerResponse{
		ID: v.Id(),
	}, nil
}

type DestroyWorkerRequest struct {
	ID uint64
}

func parseDestroyWorkerRequest(message capnp.Message) (DestroyWorkerRequest, error) {
	v, err := message.Action().DestroyWorker()
	if err != nil {
		return DestroyWorkerRequest{}, err
	}

	return DestroyWorkerRequest{
		ID: v.Id(),
	}, nil
}

type DestroyWorkerResponse struct{}

func parseDestroyWorkerResponse(message capnp.Message) (DestroyWorkerResponse, error) {
	_, err := message.Response().DestroyWorker()
	if err != nil {
		return DestroyWorkerResponse{}, err
	}

	return DestroyWorkerResponse{}, nil
}

type GetWorkerRequest struct {
	ID uint64
}

func parseGetWorkerRequest(message capnp.Message) (GetWorkerRequest, error) {
	v, err := message.Action().GetWorker()
	if err != nil {
		return GetWorkerRequest{}, err
	}

	return GetWorkerRequest{
		ID: v.Id(),
	}, nil
}

type GetWorkerResponse struct {
	ServiceName string
	ServiceID   string
}

func parseGetWorkerResponse(message capnp.Message) (GetWorkerResponse, error) {
	v, err := message.Response().GetWorker()
	if err != nil {
		return GetWorkerResponse{}, err
	}

	serviceName, err := v.ServiceName()
	if err != nil {
		return GetWorkerResponse{}, err
	}

	serviceID, err := v.ServiceId()
	if err != nil {
		return GetWorkerResponse{}, err
	}

	return GetWorkerResponse{
		ServiceName: serviceName,
		ServiceID:   serviceID,
	}, nil
}

type ExecuteRequest struct {
	Repository string
	FilePath   string
}

func parseExecuteRequest(message capnp.Message) (ExecuteRequest, error) {
	v, err := message.Action().Execute()
	if err != nil {
		return ExecuteRequest{}, err
	}

	repository, err := v.Repository()
	if err != nil {
		return ExecuteRequest{}, err
	}

	filePath, err := v.FilePath()
	if err != nil {
		return ExecuteRequest{}, err
	}

	return ExecuteRequest{
		Repository: repository,
		FilePath:   filePath,
	}, nil
}

type ExecuteResponse struct {
	Code int64
}

func parseExecuteResponse(message capnp.Message) (ExecuteResponse, error) {
	v, err := message.Response().Execute()
	if err != nil {
		return ExecuteResponse{}, err
	}

	return ExecuteResponse{
		Code: v.Code(),
	}, nil
}

type DummyRequest struct{}

func parseDummyRequest(message capnp.Message) (DummyRequest, error) {
	return DummyRequest{}, nil
}
