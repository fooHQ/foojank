package modules

import (
	"github.com/risor-io/risor/object"

	modbase64 "github.com/foohq/foojank/internal/engine/modules/base64"
	modbuiltins "github.com/foohq/foojank/internal/engine/modules/builtins"
	modbytes "github.com/foohq/foojank/internal/engine/modules/bytes"
	modcli "github.com/foohq/foojank/internal/engine/modules/cli"
	moderrors "github.com/foohq/foojank/internal/engine/modules/errors"
	modexec "github.com/foohq/foojank/internal/engine/modules/exec"
	modfilepath "github.com/foohq/foojank/internal/engine/modules/filepath"
	modfmt "github.com/foohq/foojank/internal/engine/modules/fmt"
	modhttp "github.com/foohq/foojank/internal/engine/modules/http"
	modjson "github.com/foohq/foojank/internal/engine/modules/json"
	modmath "github.com/foohq/foojank/internal/engine/modules/math"
	modnet "github.com/foohq/foojank/internal/engine/modules/net"
	modos "github.com/foohq/foojank/internal/engine/modules/os"
	modrand "github.com/foohq/foojank/internal/engine/modules/rand"
	modregexp "github.com/foohq/foojank/internal/engine/modules/regexp"
	modshlex "github.com/foohq/foojank/internal/engine/modules/shlex"
	modstrconv "github.com/foohq/foojank/internal/engine/modules/strconv"
	modstrings "github.com/foohq/foojank/internal/engine/modules/strings"
	modtime "github.com/foohq/foojank/internal/engine/modules/time"
	modurlpath "github.com/foohq/foojank/internal/engine/modules/urlpath"
)

var (
	modules = []*object.Module{
		modbase64.Module(),
		modbuiltins.Module(),
		modbytes.Module(),
		modcli.Module(),
		moderrors.Module(),
		modexec.Module(),
		modfilepath.Module(),
		modfmt.Module(),
		modhttp.Module(),
		modjson.Module(),
		modmath.Module(),
		modnet.Module(),
		modos.Module(),
		modrand.Module(),
		modregexp.Module(),
		modshlex.Module(),
		modstrconv.Module(),
		modstrings.Module(),
		modtime.Module(),
		modurlpath.Module(),
	}
	builtins = []map[string]object.Object{
		modbase64.Builtins(),
		modbuiltins.Builtins(),
		modbytes.Builtins(),
		modcli.Builtins(),
		moderrors.Builtins(),
		modexec.Builtins(),
		modfilepath.Builtins(),
		modfmt.Builtins(),
		modhttp.Builtins(),
		modjson.Builtins(),
		modmath.Builtins(),
		modnet.Builtins(),
		modos.Builtins(),
		modrand.Builtins(),
		modregexp.Builtins(),
		modshlex.Builtins(),
		modstrconv.Builtins(),
		modstrings.Builtins(),
		modtime.Builtins(),
		modurlpath.Builtins(),
	}
)

func Globals() map[string]any {
	result := make(map[string]any, len(modules)+len(builtins))
	for _, module := range modules {
		if module == nil {
			continue
		}
		name := module.Name().String()
		result[name] = module
	}
	for _, builtin := range builtins {
		if builtin == nil {
			continue
		}
		for name, fn := range builtin {
			result[name] = fn
		}
	}
	return result
}

// StubBuildTag returns stub build tag for a module name. The function does not check existence of the module.
func StubBuildTag(name string) string {
	return "module_" + name + "_stub"
}
