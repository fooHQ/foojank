package repository

import (
	"fmt"
	"github.com/foojank/foojank/clients/repository"
	"github.com/foojank/foojank/internal/application/path"
	"github.com/urfave/cli/v2"
	"log/slog"
	"os"
	"path/filepath"
)

type CopyArguments struct {
	Logger     *slog.Logger
	Repository *repository.Client
}

func NewCopyCommand(args CopyArguments) *cli.Command {
	return &cli.Command{
		Name:        "copy",
		Args:        true,
		ArgsUsage:   "<file-path> <destination-path>",
		Description: "Copy file from/to a repository",
		Action:      newCopyCommandAction(args),
	}
}

func newCopyCommandAction(args CopyArguments) cli.ActionFunc {
	// Possible use cases:
	// [Destination is a repository]
	// ./path/to/file repository:/                      => repository:/path/to/file
	// ./path/to/file repository:/test                  => repository:/test
	// ./path/to/file repository:/test/                 => repository:/test/path/to/file
	// ./path/to/file ./path/to/file2 repository:/test  => repository:/test/path/to/file
	//                                                  => repository:/test/path/to/file2
	// ./path/to/file ./path/to/file2 repository:/test/ => repository:/test/path/to/file
	//                                                  => repository:/test/path/to/file2
	// [Destination is a local directory]
	// repository:/path/to/file ./ => file
	// repository:/path/to/file repository:/path/to/file ./ => file (!!! SHOW WARNING THAT THIS WILL OVERWRITE THE FIRST FILE !!!)
	return func(c *cli.Context) error {
		cnt := c.Args().Len()
		if cnt != 2 {
			err := fmt.Errorf("command '%s' expects the following arguments: %s", c.Command.Name, c.Command.ArgsUsage)
			args.Logger.Error(err.Error())
			return err
		}

		files := c.Args().Slice()
		src := files[0]
		srcPath, err := path.Parse(src)
		if err != nil {
			args.Logger.Error("invalid file path", "error", err)
			return err
		}

		dst := files[len(files)-1]
		dstPath, err := path.Parse(dst)
		if err != nil {
			args.Logger.Error("invalid destination path", "error", err)
			return err
		}

		if srcPath.IsDir() {
			err := fmt.Errorf("file '%s' is a directory, copying directories is currently not supported", srcPath)
			args.Logger.Error(err.Error())
			return err
		}

		if srcPath.IsLocal() && dstPath.IsLocal() {
			err := fmt.Errorf("both paths are local paths, this operation is currently not supported")
			args.Logger.Error(err.Error())
			return err
		}

		if !srcPath.IsLocal() && !dstPath.IsLocal() {
			err := fmt.Errorf("both paths are repository paths, this operation is currently not supported")
			args.Logger.Error(err.Error())
			return err
		}

		ctx := c.Context

		// Copy local file to a remote repository
		if srcPath.IsLocal() {
			f, err := os.Open(srcPath.FilePath)
			if err != nil {
				err := fmt.Errorf("cannot open local file: %v", err)
				args.Logger.Error(err.Error())
				return err
			}
			defer func() {
				_ = f.Close()
			}()

			var filename string
			if dstPath.IsDir() {
				filename = filepath.Join("/", dstPath.FilePath, srcPath.Base())
			} else {
				filename = filepath.Join("/", dstPath.FilePath)
			}

			args.Logger.Debug("put local file to a remote repository", "src", srcPath, "repository", dstPath.Repository, "dst", filename)

			err = args.Repository.PutFile(ctx, dstPath.Repository, filename, f)
			if err != nil {
				err := fmt.Errorf("cannot put local file '%s' to a remote repository '%s' as '%s': %v", srcPath, dstPath.Repository, filename, err)
				args.Logger.Error(err.Error())
				return err
			}

			return nil
		}

		// TODO
		// Copy file from a remote repository to a local directory
		//if !srcPath.IsLocal() {
		//}

		return nil
	}
}
