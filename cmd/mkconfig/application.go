package main

import (
	"context"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank"
)

func NewApplication() *cli.Command {
	return &cli.Command{
		Name:      "mkconfig",
		ArgsUsage: "<input-file> <output-file>",
		Action: func(ctx context.Context, c *cli.Command) error {
			args := c.Args()
			if args.Len() != 2 {
				err := fmt.Errorf("command expects the following arguments: %s", c.ArgsUsage)
				return err
			}

			b, err := os.ReadFile(args.Get(0))
			if err != nil {
				return err
			}

			var data Data
			err = yaml.Unmarshal(b, &data)
			if err != nil {
				return err
			}

			if data.Service.Version == "" {
				data.Service.Version = foojank.Version()
			}

			template := NewTemplate()
			output, err := template.Render(data)
			if err != nil {
				return err
			}

			f, err := os.Create(args.Get(1))
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = f.Write([]byte(output))
			if err != nil {
				return err
			}

			return nil
		},
	}
}
