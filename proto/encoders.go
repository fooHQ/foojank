package proto

import (
	capnplib "capnproto.org/go/capnp/v3"

	"github.com/foohq/foojank/proto/capnp"
)

func NewCreateWorkerRequest() ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	msgCreateWorker, err := capnp.NewCreateWorkerRequest(msg.Segment())
	if err != nil {
		return nil, err
	}

	err = msg.Action().SetCreateWorker(msgCreateWorker)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func NewCreateWorkerResponse(id uint64) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	msgCreateWorker, err := capnp.NewCreateWorkerResponse(msg.Segment())
	if err != nil {
		return nil, err
	}

	msgCreateWorker.SetId(id)

	err = msg.Response().SetCreateWorker(msgCreateWorker)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func NewDestroyWorkerRequest(id uint64) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	msgDestroyWorker, err := capnp.NewDestroyWorkerRequest(msg.Segment())
	if err != nil {
		return nil, err
	}

	msgDestroyWorker.SetId(id)

	err = msg.Action().SetDestroyWorker(msgDestroyWorker)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func NewDestroyWorkerResponse() ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	msgDestroyWorker, err := capnp.NewDestroyWorkerResponse(msg.Segment())
	if err != nil {
		return nil, err
	}

	err = msg.Response().SetDestroyWorker(msgDestroyWorker)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func NewGetWorkerRequest(id uint64) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	msgGetWorker, err := capnp.NewGetWorkerRequest(msg.Segment())
	if err != nil {
		return nil, err
	}

	msgGetWorker.SetId(id)

	err = msg.Action().SetGetWorker(msgGetWorker)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func NewGetWorkerResponse(serviceName, serviceID string) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	msgGetWorker, err := capnp.NewGetWorkerResponse(msg.Segment())
	if err != nil {
		return nil, err
	}

	err = msgGetWorker.SetServiceName(serviceName)
	if err != nil {
		return nil, err
	}

	err = msgGetWorker.SetServiceId(serviceID)
	if err != nil {
		return nil, err
	}

	err = msg.Response().SetGetWorker(msgGetWorker)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func NewExecuteRequest(filePath string, args []string) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	msgExecute, err := capnp.NewExecuteRequest(msg.Segment())
	if err != nil {
		return nil, err
	}

	argsList, err := newTextList(msg.Segment(), args)
	if err != nil {
		return nil, err
	}

	err = msgExecute.SetArgs(argsList)
	if err != nil {
		return nil, err
	}

	err = msgExecute.SetFilePath(filePath)
	if err != nil {
		return nil, err
	}

	err = msg.Action().SetExecute(msgExecute)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func NewExecuteResponse(code int64) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	msgExecute, err := capnp.NewExecuteResponse(msg.Segment())
	if err != nil {
		return nil, err
	}

	msgExecute.SetCode(code)

	err = msg.Response().SetExecute(msgExecute)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func NewDummyRequest() ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	msgDummy, err := capnp.NewDummyRequest(msg.Segment())
	if err != nil {
		return nil, err
	}

	err = msg.Action().SetDummyRequest(msgDummy)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func NewCreateJobRequest(command string, args, env []string) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	msgCreateJob, err := capnp.NewCreateJobRequest(msg.Segment())
	if err != nil {
		return nil, err
	}

	err = msgCreateJob.SetCommand(command)
	if err != nil {
		return nil, err
	}

	argsList, err := newTextList(msg.Segment(), args)
	if err != nil {
		return nil, err
	}

	err = msgCreateJob.SetArgs(argsList)
	if err != nil {
		return nil, err
	}

	envList, err := newTextList(msg.Segment(), env)
	if err != nil {
		return nil, err
	}

	err = msgCreateJob.SetEnv(envList)
	if err != nil {
		return nil, err
	}

	err = msg.Action().SetCreateJob(msgCreateJob)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func NewCreateJobResponse(jobID, stdinSubject, stdoutSubject string, respErr error) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	msgCreateJob, err := capnp.NewCreateJobResponse(msg.Segment())
	if err != nil {
		return nil, err
	}

	err = msgCreateJob.SetJobID(jobID)
	if err != nil {
		return nil, err
	}

	err = msgCreateJob.SetStdinSubject(stdinSubject)
	if err != nil {
		return nil, err
	}

	err = msgCreateJob.SetStdoutSubject(stdoutSubject)
	if err != nil {
		return nil, err
	}

	if respErr != nil {
		err = msgCreateJob.SetError(respErr.Error())
		if err != nil {
			return nil, err
		}
	}

	err = msg.Response().SetCreateJob(msgCreateJob)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func NewCancelJobRequest(jobID string) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	msgCancelJob, err := capnp.NewCancelJobRequest(msg.Segment())
	if err != nil {
		return nil, err
	}

	err = msgCancelJob.SetJobID(jobID)
	if err != nil {
		return nil, err
	}

	err = msg.Action().SetCancelJob(msgCancelJob)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func NewCancelJobResponse(respErr error) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	msgCancelJob, err := capnp.NewCancelJobResponse(msg.Segment())
	if err != nil {
		return nil, err
	}

	if respErr != nil {
		err = msgCancelJob.SetError(respErr.Error())
		if err != nil {
			return nil, err
		}
	}

	err = msg.Response().SetCancelJob(msgCancelJob)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func NewUpdateJob(jobID string, exitStatus int64) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	msgUpdateJob, err := capnp.NewUpdateJob(msg.Segment())
	if err != nil {
		return nil, err
	}

	err = msgUpdateJob.SetJobID(jobID)
	if err != nil {
		return nil, err
	}

	msgUpdateJob.SetExitStatus(exitStatus)

	err = msg.Response().SetUpdateJob(msgUpdateJob)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
}

func newMessage() (capnp.Message, error) {
	arena := capnplib.SingleSegment(nil)
	_, seg, err := capnplib.NewMessage(arena)
	if err != nil {
		return capnp.Message{}, err
	}

	msg, err := capnp.NewRootMessage(seg)
	if err != nil {
		return capnp.Message{}, err
	}

	return msg, nil
}

func newTextList(segment *capnplib.Segment, ss []string) (capnplib.TextList, error) {
	tl, err := capnplib.NewTextList(segment, int32(len(ss)))
	if err != nil {
		return capnplib.TextList{}, err
	}

	for i, s := range ss {
		err := tl.Set(i, s)
		if err != nil {
			return capnplib.TextList{}, err
		}
	}

	return tl, nil
}
