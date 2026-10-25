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

func ValidateCreateUserRequest(req protodaemon.CreateUserRequest) error {
	if !nkeys.IsValidPublicUserKey(req.ID) {
		return errors.New("invalid user id")
	}
	return validateName(req.Name)
}

func ValidateGetUserRequest(req protodaemon.GetUserRequest) error {
	return validateName(req.UserRef)
}

func ValidateUpdateUserRequest(req protodaemon.UpdateUserRequest) error {
	return validateName(req.UserRef)
}

func ValidateDeleteUserRequest(req protodaemon.DeleteUserRequest) error {
	return validateName(req.UserRef)
}

func ValidateCreateAgentRequest(req protodaemon.CreateAgentRequest) error {
	err := validateName(req.Name)
	if err != nil {
		return err
	}
	if req.GatewayRef == "" {
		return errors.New("gateway is required")
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
	return validateName(req.AgentRef)
}

func ValidateDeleteAgentRequest(req protodaemon.DeleteAgentRequest) error {
	return validateName(req.AgentRef)
}

func ValidateGetCommandRequest(req protodaemon.GetCommandRequest) error {
	if !nkeys.IsValidPublicUserKey(req.AgentRef) {
		return errors.New("invalid agent id")
	}
	return validateName(req.Name)
}

func ValidateListCommandsRequest(req protodaemon.ListCommandsRequest) error {
	if !nkeys.IsValidPublicUserKey(req.AgentRef) {
		return errors.New("invalid agent id")
	}
	return nil
}

func ValidateCreateGatewayRequest(req protodaemon.CreateGatewayRequest) error {
	return validateName(req.Name)
}

func ValidateGetGatewayRequest(req protodaemon.GetGatewayRequest) error {
	return validateName(req.GatewayRef)
}

func ValidateDeleteGatewayRequest(req protodaemon.DeleteGatewayRequest) error {
	return validateName(req.GatewayRef)
}

func ValidateCreateRoleRequest(req protodaemon.CreateRoleRequest) error {
	return validateName(req.Name)
}

func ValidateGetRoleRequest(req protodaemon.GetRoleRequest) error {
	return validateName(req.RoleRef)
}

func ValidateDeleteRoleRequest(req protodaemon.DeleteRoleRequest) error {
	return validateName(req.RoleRef)
}

func ValidateCreateJobRequest(req protodaemon.CreateJobRequest) error {
	if !nkeys.IsValidPublicUserKey(req.AgentRef) {
		return errors.New("invalid agent id")
	}
	if req.Command == "" {
		return errors.New("command is required")
	}
	return nil
}

func ValidateGetJobRequest(req protodaemon.GetJobRequest) error {
	if req.JobRef == "" {
		return errors.New("job id is required")
	}
	return directory.ValidateKey(req.JobRef)
}

func ValidateCancelJobRequest(req protodaemon.CancelJobRequest) error {
	if req.JobRef == "" {
		return errors.New("job id is required")
	}
	return directory.ValidateKey(req.JobRef)
}

func ValidateGetJobOutputRequest(req protodaemon.GetJobOutputRequest) error {
	if req.JobRef == "" {
		return errors.New("job id is required")
	}
	return directory.ValidateKey(req.JobRef)
}

func ValidateListJobsRequest(req protodaemon.ListJobsRequest) error {
	if req.AgentRef == "" {
		return nil
	}
	if !nkeys.IsValidPublicUserKey(req.AgentRef) {
		return errors.New("invalid agent id")
	}
	return nil
}
