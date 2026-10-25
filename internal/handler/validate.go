package handler

import (
	"errors"

	"github.com/nats-io/nkeys"

	"github.com/foohq/foojank/internal/directory"
	protodaemon "github.com/foohq/foojank/proto/daemon"
)

func validateName(name string) error {
	if name == "" {
		return errors.New("name is required")
	}
	return directory.ValidateKey(name)
}

func validateRef(ref string) error {
	if ref == "" {
		return errors.New("ref is required")
	}
	return directory.ValidateKey(ref)
}

func ValidateCreateUserRequest(req protodaemon.CreateUserRequest) error {
	if !nkeys.IsValidPublicUserKey(req.ID) {
		return errors.New("invalid user id")
	}
	return validateName(req.Name)
}

func ValidateGetUserRequest(req protodaemon.GetUserRequest) error {
	return validateRef(req.UserRef)
}

func ValidateUpdateUserRequest(req protodaemon.UpdateUserRequest) error {
	return validateRef(req.UserRef)
}

func ValidateDeleteUserRequest(req protodaemon.DeleteUserRequest) error {
	return validateRef(req.UserRef)
}

func ValidateCreateAgentRequest(req protodaemon.CreateAgentRequest) error {
	err := validateName(req.Name)
	if err != nil {
		return err
	}
	err = validateRef(req.GatewayRef)
	if err != nil {
		return err
	}
	if req.Config.OS == "" {
		return errors.New("os is required")
	}
	if req.Config.Arch == "" {
		return errors.New("arch is required")
	}
	return nil
}

func ValidateGetAgentRequest(req protodaemon.GetAgentRequest) error {
	return validateRef(req.AgentRef)
}

func ValidateDeleteAgentRequest(req protodaemon.DeleteAgentRequest) error {
	return validateRef(req.AgentRef)
}

func ValidateGetCommandRequest(req protodaemon.GetCommandRequest) error {
	err := validateRef(req.AgentRef)
	if err != nil {
		return err
	}
	return validateName(req.Name)
}

func ValidateListCommandsRequest(req protodaemon.ListCommandsRequest) error {
	return validateRef(req.AgentRef)
}

func ValidateCreateGatewayRequest(req protodaemon.CreateGatewayRequest) error {
	return validateName(req.Name)
}

func ValidateGetGatewayRequest(req protodaemon.GetGatewayRequest) error {
	return validateRef(req.GatewayRef)
}

func ValidateDeleteGatewayRequest(req protodaemon.DeleteGatewayRequest) error {
	return validateRef(req.GatewayRef)
}

func ValidateCreateRoleRequest(req protodaemon.CreateRoleRequest) error {
	return validateName(req.Name)
}

func ValidateGetRoleRequest(req protodaemon.GetRoleRequest) error {
	return validateRef(req.RoleRef)
}

func ValidateDeleteRoleRequest(req protodaemon.DeleteRoleRequest) error {
	return validateRef(req.RoleRef)
}

func ValidateCreateJobRequest(req protodaemon.CreateJobRequest) error {
	err := validateRef(req.AgentRef)
	if err != nil {
		return err
	}
	if req.Command == "" {
		return errors.New("command is required")
	}
	return nil
}

func ValidateGetJobRequest(req protodaemon.GetJobRequest) error {
	return validateRef(req.JobRef)
}

func ValidateCancelJobRequest(req protodaemon.CancelJobRequest) error {
	return validateRef(req.JobRef)
}

func ValidateGetJobOutputRequest(req protodaemon.GetJobOutputRequest) error {
	return validateRef(req.JobRef)
}

func ValidateListJobsRequest(req protodaemon.ListJobsRequest) error {
	if req.AgentRef == "" {
		return nil
	}
	return validateRef(req.AgentRef)
}
