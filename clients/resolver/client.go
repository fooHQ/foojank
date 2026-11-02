package resolver

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/foohq/foojank/clients/server"
)

const updateAccountClaimsSubject = "$SYS.REQ.CLAIMS.UPDATE"

// Client publishes account JWTs to a NATS account resolver.
type Client struct {
	srv *server.Client
}

func New(srv *server.Client) *Client {
	return &Client{
		srv: srv,
	}
}

// UpdateAccountClaims uploads accountJWT to the NATS account resolver.
// The payload is the account JWT. The call returns when the first resolver replies or ctx ends.
func (c *Client) UpdateAccountClaims(ctx context.Context, accountJWT string) error {
	msg, err := c.srv.Conn().RequestWithContext(ctx, updateAccountClaimsSubject, []byte(accountJWT))
	if err != nil {
		return err
	}

	var resp updateAccountClaimsResponse
	err = json.Unmarshal(msg.Data, &resp)
	if err != nil {
		return err
	}

	data := resp.Data
	if data.Code != 200 {
		return errors.New(data.Message)
	}

	return nil
}

type updateAccountClaimsResponse struct {
	Data updateAccountClaimsData `json:"data"`
}

type updateAccountClaimsData struct {
	Code      int    `json:"code"`
	ErrorCode int    `json:"err_code"`
	Message   string `json:"message"`
}
