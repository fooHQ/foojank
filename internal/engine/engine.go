package engine

import (
	"context"
	"errors"
	"io"

	"github.com/risor-io/risor"
	"github.com/risor-io/risor/builtins"
	"github.com/risor-io/risor/compiler"
	modbase64 "github.com/risor-io/risor/modules/base64"
	modbytes "github.com/risor-io/risor/modules/bytes"
	modcli "github.com/risor-io/risor/modules/cli"
	moderrors "github.com/risor-io/risor/modules/errors"
	modexec "github.com/risor-io/risor/modules/exec"
	modfilepath "github.com/risor-io/risor/modules/filepath"
	modfmt "github.com/risor-io/risor/modules/fmt"
	modhttp "github.com/risor-io/risor/modules/http"
	modjson "github.com/risor-io/risor/modules/json"
	modmath "github.com/risor-io/risor/modules/math"
	modnet "github.com/risor-io/risor/modules/net"
	modos "github.com/risor-io/risor/modules/os"
	modrand "github.com/risor-io/risor/modules/rand"
	modregexp "github.com/risor-io/risor/modules/regexp"
	modstrconv "github.com/risor-io/risor/modules/strconv"
	modstrings "github.com/risor-io/risor/modules/strings"
	modtime "github.com/risor-io/risor/modules/time"
	"github.com/risor-io/risor/parser"
	"github.com/risor-io/risor/vm"

	"github.com/foohq/foojank/internal/engine/importers"
)

func defaultModules() map[string]any {
	return map[string]any{
		"base64":   modbase64.Module(),
		"bytes":    modbytes.Module(),
		"errors":   moderrors.Module(),
		"exec":     modexec.Module(),
		"filepath": modfilepath.Module(),
		"fmt":      modfmt.Module(),
		"http":     modhttp.Module(),
		"json":     modjson.Module(),
		"math":     modmath.Module(),
		"net":      modnet.Module(),
		"os":       modos.Module(),
		"cli":      modcli.Module(),
		"rand":     modrand.Module(),
		"regexp":   modregexp.Module(),
		"strconv":  modstrconv.Module(),
		"strings":  modstrings.Module(),
		"time":     modtime.Module(),
	}
}

func defaultBuiltins() map[string]any {
	result := make(map[string]any)
	for name, obj := range builtins.Builtins() {
		result[name] = obj
	}
	for name, obj := range modfmt.Builtins() {
		result[name] = obj
	}
	for name, obj := range modos.Builtins() {
		result[name] = obj
	}
	return result
}

func CompilePackage(ctx context.Context, reader io.ReaderAt, size int64) (*Code, error) {
	opts := append([]risor.Option{
		risor.WithoutDefaultGlobals(),
		risor.WithGlobals(defaultModules()),
		risor.WithGlobals(defaultBuiltins()),
	})
	conf := risor.NewConfig(opts...)
	prog, err := parser.Parse(ctx, "import main")
	if err != nil {
		return nil, err
	}

	code, err := compiler.Compile(prog, conf.CompilerOpts()...)
	if err != nil {
		return nil, err
	}

	importer, err := importers.NewZipImporter(reader, size, conf.CompilerOpts()...)
	if err != nil {
		return nil, err
	}

	// Recreate risor config but include the custom importer this time.
	opts = append(opts, risor.WithImporter(importer))
	return &Code{
		code: code,
		opts: opts,
	}, nil
}

type Code struct {
	code *compiler.Code
	opts []risor.Option
}

func (c *Code) Run(ctx context.Context) error {
	opts := risor.NewConfig(c.opts...)
	_, err := vm.Run(ctx, c.code, opts.VMOpts()...)
	if err != nil {
		return &Error{err}
	}

	return nil
}

type Error struct {
	err error
}

func (e *Error) Error() string {
	var parserErr parser.ParserError
	if errors.As(e.err, &parserErr) {
		return parserErr.FriendlyErrorMessage()
	}
	return e.err.Error()
}
