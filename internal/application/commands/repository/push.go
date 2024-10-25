package repository

import (
	"github.com/foojank/foojank/clients/repository"
	"github.com/urfave/cli/v2"
)

func NewCopyCommand(repo *repository.Client) *cli.Command {
	return &cli.Command{
		Name:        "copy",
		Description: "Copy file from/to a repository",
		Action:      newCopyCommandAction(repo),
	}
}

func newCopyCommandAction(repo *repository.Client) cli.ActionFunc {
	return func(c *cli.Context) error {
		// Possible cases:
		// [local -> repository]
		// ./path/to/file repository:/                      => repository:/path/to/file
		// ./path/to/file repository:/test                  => repository:/test
		// ./path/to/file repository:/test/                 => repository:/test/path/to/file
		// ./path/to/file ./path/to/file2 repository:/test  => repository:/test/path/to/file
		//                                                  => repository:/test/path/to/file2
		// ./path/to/file ./path/to/file2 repository:/test/ => repository:/test/path/to/file
		//                                                  => repository:/test/path/to/file2
		// [repository -> local]
		// repository:/path/to/file ./ => file
		// repository:/path/to/file repository:/path/to/file ./ => file (!!! SHOW WARNING THAT THIS WILL OVERWRITE THE FIRST FILE !!!)
		/*if c.Args().Len() < 2 {
			return fmt.Errorf("command expects at least two arguments")
		}

		args := c.Args().Slice()
		dest := args[len(args)-1]
		destPath, err := path.Parse(dest)
		if err != nil {
			return err
		}

		if destPath.Repository == "" {
			return fmt.Errorf("destination path must be in format repository:/file/path")
		}

		for item := range c.Args().Slice() {

		}*/

		return nil
	}
}
