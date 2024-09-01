package reader

import (
	"context"
	xterm "golang.org/x/term"
	"os"
	"strings"
)

type Arguments struct {
	OutputCh chan<- Message
}

type Service struct {
	args Arguments
}

func New(args Arguments) *Service {
	return &Service{
		args: args,
	}
}

func (s *Service) Start(ctx context.Context) error {
	responseCh := make(chan MessageResponse, 1)

	// TODO: check if terminal!
	term := xterm.NewTerminal(os.Stdin, "(foojank)=> ")

	for {
		oldState, err := xterm.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}

		line, err := term.ReadLine()
		if err != nil {
			return err
		}

		xterm.Restore(int(os.Stdin.Fd()), oldState)

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		msg := Message{
			ctx:        ctx,
			data:       "foojank" + " " + line,
			responseCh: responseCh,
		}
		select {
		case s.args.OutputCh <- msg:
		case <-ctx.Done():
			return nil
		}

		select {
		case <-responseCh:
		case <-ctx.Done():
			return nil
		}
	}
}
