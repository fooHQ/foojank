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
	return validateName(req.Name)
}

func ValidateCreateAgentRequest(req protodaemon.CreateAgentRequest) error {
	err := validateName(req.Name)
	if err != nil {
		return err
	}
	if req.Gateway == "" {
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
	return validateName(req.Name)
}
