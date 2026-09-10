package actions

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojankd/flags"
	"github.com/foohq/foojank/internal/formatter"
)

const generateShellCompletionFlag = "--generate-shell-completion"

// IsShellCompletion reports whether the current process was invoked by a
// shell completion script.
func IsShellCompletion() bool {
	n := len(os.Args)
	return n > 0 && os.Args[n-1] == generateShellCompletionFlag
}

// CompleteFlags is a ShellComplete func for commands that have no positional
// name argument. It emits flag suggestions and values for --agent, --gateway,
// and --format.
func CompleteFlags(ctx context.Context, c *cli.Command) {
	ctx, cancel := completionContext(ctx)
	defer cancel()
	completeFlagToken(ctx, c)
}

func completionContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 4*time.Second)
}

func completeFlagToken(ctx context.Context, c *cli.Command) bool {
	token := completeToken()
	if list, prefix, ok := flagValueCompleter(c, token); ok {
		lines, err := list(ctx)
		if err == nil {
			printCompletionLines(c, prefix, lines)
			return true
		}
		// Value lookup failed (server down, missing config). Fall through to
		// flag-name completion so TAB still does something useful.
	}
	if strings.HasPrefix(token, "-") {
		printFlagCompletions(c, token)
		return true
	}
	return false
}

func completeToken() string {
	args := os.Args
	if n := len(args); n > 0 && args[n-1] == generateShellCompletionFlag {
		args = args[:n-1]
	}
	if len(args) == 0 {
		return ""
	}
	return args[len(args)-1]
}

func flagValueCompleter(c *cli.Command, token string) (list func(context.Context) ([]string, error), prefix string, ok bool) {
	if !strings.HasPrefix(token, "--") {
		return nil, "", false
	}
	body := strings.TrimPrefix(token, "--")
	name, _, hasEq := strings.Cut(body, "=")
	if !commandHasFlag(c, name) {
		return nil, "", false
	}
	switch name {
	case flags.Format:
		list = listFormatNames
	default:
		return nil, "", false
	}
	if hasEq {
		return list, "--" + name + "=", true
	}
	return list, "", true
}

func commandHasFlag(c *cli.Command, name string) bool {
	for _, flag := range c.VisibleFlags() {
		if slices.Contains(flag.Names(), name) {
			return true
		}
	}
	return false
}

func printCompletionLines(c *cli.Command, prefix string, lines []string) {
	w := c.Root().Writer
	for _, line := range lines {
		if prefix == "" {
			_, _ = fmt.Fprintln(w, line)
			continue
		}
		name, desc, found := strings.Cut(line, ":")
		suggestion := prefix + name
		if found && desc != "" {
			suggestion += ":" + desc
		}
		_, _ = fmt.Fprintln(w, suggestion)
	}
}

func printFlagCompletions(c *cli.Command, lastArg string) {
	cur := strings.TrimLeft(lastArg, "-")
	w := c.Root().Writer
	for _, flag := range c.Flags {
		if vf, ok := flag.(cli.VisibleFlag); ok && !vf.IsVisible() {
			continue
		}
		names := flag.Names()
		if len(names) == 0 {
			continue
		}
		name := strings.TrimSpace(names[0])
		nDashes := 2
		if len(name) == 1 {
			nDashes = 1
		}
		if strings.HasPrefix(lastArg, "--") && nDashes == 1 {
			continue
		}
		if !strings.HasPrefix(name, cur) || cur == name {
			continue
		}
		suggestion := strings.Repeat("-", nDashes) + name
		if doc, ok := flag.(cli.DocGenerationFlag); ok {
			if usage := doc.GetUsage(); usage != "" {
				suggestion += ":" + usage
			}
		}
		_, _ = fmt.Fprintln(w, suggestion)
	}
}

func listFormatNames(_ context.Context) ([]string, error) {
	return []string{formatter.FormatASCII, formatter.FormatJSON}, nil
}
