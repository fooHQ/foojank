package proto

import "github.com/foohq/foojank/proto/capnp"

type UpdateClientInfo struct {
	Username string
	Hostname string
	System   string
	Address  string
}

func marshalUpdateClientInfo(data UpdateClientInfo) ([]byte, error) {
	msg, err := newMessage()
	if err != nil {
		return nil, err
	}

	m, err := capnp.NewUpdateClientInfo(msg.Segment())
	if err != nil {
		return nil, err
	}

	err = m.SetUsername(data.Username)
	if err != nil {
		return nil, err
	}

	err = m.SetHostname(data.Hostname)
	if err != nil {
		return nil, err
	}

	err = m.SetSystem(data.System)
	if err != nil {
		return nil, err
	}

	err = m.SetAddress(data.Address)
	if err != nil {
		return nil, err
	}

	err = msg.Content().SetUpdateClientInfo(m)
	if err != nil {
		return nil, err
	}

	return msg.Message().Marshal()
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
