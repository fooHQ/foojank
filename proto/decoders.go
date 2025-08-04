package proto

import (
	"errors"

	capnplib "capnproto.org/go/capnp/v3"

	"github.com/foohq/foojank/proto/capnp"
)

var (
	ErrUnknownAction   = errors.New("unknown action")
	ErrUnknownResponse = errors.New("unknown response")
	ErrUnknownMessage  = errors.New("unknown message")
)

func Unmarshal(b []byte) (any, error) {
	message, err := parseMessage(b)
	if err != nil {
		return nil, err
	}

	content := message.Content()
	switch {
	case content.HasCreateJobRequest():
		return unmarshalCreateJobRequest(message)

	case content.HasCreateJobResponse():
		return unmarshalCreateJobResponse(message)

	case content.HasCancelJobRequest():
		return unmarshalCancelJobRequest(message)

	case content.HasCancelJobResponse():
		return unmarshalCancelJobResponse(message)

	case content.HasUpdateJob():
		return unmarshalUpdateJob(message)

	case content.HasUpdateStdioLine():
		return unmarshalUpdateStdioLine(message)
	}

	return nil, ErrUnknownMessage
}

func unmarshalCreateJobRequest(message capnp.Message) (CreateJobRequest, error) {
	v, err := message.Content().CreateJobRequest()
	if err != nil {
		return CreateJobRequest{}, err
	}

	command, err := v.Command()
	if err != nil {
		return CreateJobRequest{}, err
	}

	vArgs, err := v.Args()
	if err != nil {
		return CreateJobRequest{}, err
	}

	args, err := textListToStringSlice(vArgs)
	if err != nil {
		return CreateJobRequest{}, err
	}

	vEnv, err := v.Env()
	if err != nil {
		return CreateJobRequest{}, err
	}

	env, err := textListToStringSlice(vEnv)
	if err != nil {
		return CreateJobRequest{}, err
	}

	return CreateJobRequest{
		Command: command,
		Args:    args,
		Env:     env,
	}, nil
}

func unmarshalCreateJobResponse(message capnp.Message) (CreateJobResponse, error) {
	v, err := message.Content().CreateJobResponse()
	if err != nil {
		return CreateJobResponse{}, err
	}

	jobID, err := v.JobID()
	if err != nil {
		return CreateJobResponse{}, err
	}

	stdinSubject, err := v.StdinSubject()
	if err != nil {
		return CreateJobResponse{}, err
	}

	stdoutSubject, err := v.StdoutSubject()
	if err != nil {
		return CreateJobResponse{}, err
	}

	errMsg, err := v.Error()
	if err != nil {
		return CreateJobResponse{}, err
	}

	var respErr error
	if errMsg != "" {
		respErr = errors.New(errMsg)
	}

	return CreateJobResponse{
		JobID:         jobID,
		StdinSubject:  stdinSubject,
		StdoutSubject: stdoutSubject,
		Error:         respErr,
	}, nil
}

func unmarshalCancelJobRequest(message capnp.Message) (CancelJobRequest, error) {
	v, err := message.Content().CancelJobRequest()
	if err != nil {
		return CancelJobRequest{}, err
	}

	jobID, err := v.JobID()
	if err != nil {
		return CancelJobRequest{}, err
	}

	return CancelJobRequest{
		JobID: jobID,
	}, nil
}

func unmarshalCancelJobResponse(message capnp.Message) (CancelJobResponse, error) {
	v, err := message.Content().CancelJobResponse()
	if err != nil {
		return CancelJobResponse{}, err
	}

	errMsg, err := v.Error()
	if err != nil {
		return CancelJobResponse{}, err
	}

	var respErr error
	if errMsg != "" {
		respErr = errors.New(errMsg)
	}

	return CancelJobResponse{
		Error: respErr,
	}, nil
}

func unmarshalUpdateJob(message capnp.Message) (UpdateJob, error) {
	v, err := message.Content().UpdateJob()
	if err != nil {
		return UpdateJob{}, err
	}

	jobID, err := v.JobID()
	if err != nil {
		return UpdateJob{}, err
	}

	return UpdateJob{
		JobID:      jobID,
		ExitStatus: v.ExitStatus(),
	}, nil
}

func unmarshalUpdateStdioLine(message capnp.Message) (UpdateStdioLine, error) {
	v, err := message.Content().UpdateStdioLine()
	if err != nil {
		return UpdateStdioLine{}, err
	}

	text, err := v.Text()
	if err != nil {
		return UpdateStdioLine{}, err
	}

	return UpdateStdioLine{
		Text: text,
	}, nil
}

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

	case action.HasCreateJob():
		return parseCreateJobRequest(message)

	case action.HasCancelJob():
		return parseCancelJobRequest(message)

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

	case response.HasCreateJob():
		return parseCreateJobResponse(message)

	case response.HasCancelJob():
		return parseCancelJobResponse(message)

	case response.HasUpdateJob():
		return parseUpdateJob(message)

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
	Args     []string
	FilePath string
}

func parseExecuteRequest(message capnp.Message) (ExecuteRequest, error) {
	v, err := message.Action().Execute()
	if err != nil {
		return ExecuteRequest{}, err
	}

	vArgs, err := v.Args()
	if err != nil {
		return ExecuteRequest{}, err
	}

	args, err := textListToStringSlice(vArgs)
	if err != nil {
		return ExecuteRequest{}, err
	}

	filePath, err := v.FilePath()
	if err != nil {
		return ExecuteRequest{}, err
	}

	return ExecuteRequest{
		Args:     args,
		FilePath: filePath,
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

func parseDummyRequest(_ capnp.Message) (DummyRequest, error) {
	return DummyRequest{}, nil
}

type CreateJobRequest struct {
	Command string
	Args    []string
	Env     []string
}

func parseCreateJobRequest(message capnp.Message) (CreateJobRequest, error) {
	v, err := message.Action().CreateJob()
	if err != nil {
		return CreateJobRequest{}, err
	}

	command, err := v.Command()
	if err != nil {
		return CreateJobRequest{}, err
	}

	vArgs, err := v.Args()
	if err != nil {
		return CreateJobRequest{}, err
	}

	args, err := textListToStringSlice(vArgs)
	if err != nil {
		return CreateJobRequest{}, err
	}

	vEnv, err := v.Env()
	if err != nil {
		return CreateJobRequest{}, err
	}

	env, err := textListToStringSlice(vEnv)
	if err != nil {
		return CreateJobRequest{}, err
	}

	return CreateJobRequest{
		Command: command,
		Args:    args,
		Env:     env,
	}, nil
}

type CreateJobResponse struct {
	JobID         string
	StdinSubject  string
	StdoutSubject string
	Error         error
}

func parseCreateJobResponse(message capnp.Message) (CreateJobResponse, error) {
	v, err := message.Response().CreateJob()
	if err != nil {
		return CreateJobResponse{}, err
	}

	jobID, err := v.JobID()
	if err != nil {
		return CreateJobResponse{}, err
	}

	stdinSubject, err := v.StdinSubject()
	if err != nil {
		return CreateJobResponse{}, err
	}

	stdoutSubject, err := v.StdoutSubject()
	if err != nil {
		return CreateJobResponse{}, err
	}

	errMsg, err := v.Error()
	if err != nil {
		return CreateJobResponse{}, err
	}

	return CreateJobResponse{
		JobID:         jobID,
		StdinSubject:  stdinSubject,
		StdoutSubject: stdoutSubject,
		Error:         errors.New(errMsg),
	}, nil
}

type CancelJobRequest struct {
	JobID string
}

func parseCancelJobRequest(message capnp.Message) (CancelJobRequest, error) {
	v, err := message.Action().CancelJob()
	if err != nil {
		return CancelJobRequest{}, err
	}

	jobID, err := v.JobID()
	if err != nil {
		return CancelJobRequest{}, err
	}

	return CancelJobRequest{
		JobID: jobID,
	}, nil
}

type CancelJobResponse struct {
	Error error
}

func parseCancelJobResponse(message capnp.Message) (CancelJobResponse, error) {
	v, err := message.Response().CancelJob()
	if err != nil {
		return CancelJobResponse{}, err
	}

	errMsg, err := v.Error()
	if err != nil {
		return CancelJobResponse{}, err
	}

	return CancelJobResponse{
		Error: errors.New(errMsg),
	}, nil
}

type UpdateJob struct {
	JobID      string
	ExitStatus int64
}

func parseUpdateJob(message capnp.Message) (UpdateJob, error) {
	v, err := message.Response().UpdateJob()
	if err != nil {
		return UpdateJob{}, err
	}

	jobID, err := v.JobID()
	if err != nil {
		return UpdateJob{}, err
	}

	return UpdateJob{
		JobID:      jobID,
		ExitStatus: v.ExitStatus(),
	}, nil
}

type UpdateStdioLine struct {
	Text string
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
