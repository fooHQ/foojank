package daemon

import (
	"context"
	"errors"
	"fmt"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nuid"

	protodaemon "github.com/foohq/foojank/proto/daemon"
)

func IssueUserJWT(ctx context.Context, nc *nats.Conn) (string, error) {
	userID, err := connectedUserID(nc)
	if err != nil {
		return "", err
	}

	b, err := protodaemon.Marshal(protodaemon.IssueJWTRequest{})
	if err != nil {
		return "", err
	}

	resp, err := nc.RequestMsgWithContext(ctx, &nats.Msg{
		Subject: protodaemon.IssueJWTSubject(userID),
		Header:  nats.Header{nats.MsgIdHdr: []string{nuid.Next()}},
		Data:    b,
	})
	if err != nil {
		return "", translate(err)
	}

	data, err := protodaemon.Unmarshal(resp.Data)
	if err != nil {
		return "", err
	}

	v, ok := data.(protodaemon.IssueJWTResponse)
	if !ok {
		return "", fmt.Errorf("invalid response: %T", data)
	}
	if v.Error != nil {
		return "", v.Error
	}
	if v.JWT == "" {
		return "", errors.New("empty user JWT")
	}
	return v.JWT, nil
}

func connectedUserID(nc *nats.Conn) (string, error) {
	if nc == nil || nc.Opts.UserJWT == nil {
		return "", errors.New("user JWT is not available")
	}

	token, err := nc.Opts.UserJWT()
	if err != nil {
		return "", err
	}

	claims, err := jwt.DecodeUserClaims(token)
	if err != nil {
		return "", err
	}

	if claims.Subject == "" {
		return "", errors.New("user JWT has no subject")
	}

	return claims.Subject, nil
}
