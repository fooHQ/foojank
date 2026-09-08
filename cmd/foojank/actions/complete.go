package actions

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/foohq/foojank/cmd/foojank/flags"
	"github.com/foohq/foojank/internal/auth"
	"github.com/foohq/foojank/internal/clients/daemon"
	"github.com/foohq/foojank/internal/clients/server"
	"github.com/foohq/foojank/internal/formatter"
	"github.com/foohq/foojank/internal/profile"
)

const generateShellCompletionFlag = "--generate-shell-completion"

// IsShellCompletion reports whether the current process was invoked by a
// shell completion script.
func IsShellCompletion() bool {
	n := len(os.Args)
	return n > 0 && os.Args[n-1] == generateShellCompletionFlag
}

// CompleteAgentName is a ShellComplete func for commands that take a single
// agent name. Flags are still completed after a positional argument.
func CompleteAgentName(ctx context.Context, c *cli.Command) {
	ctx, cancel := completionContext(ctx)
	defer cancel()
	completeName(ctx, c, listAgentNames)
}

// CompleteGatewayName is a ShellComplete func for commands that take a single
// gateway name. Flags are still completed after a positional argument.
func CompleteGatewayName(ctx context.Context, c *cli.Command) {
	ctx, cancel := completionContext(ctx)
	defer cancel()
	completeName(ctx, c, listGatewayNames)
}

// CompleteProfileName is a ShellComplete func for commands that take a single
// profile name. Flags are still completed after a positional argument.
func CompleteProfileName(ctx context.Context, c *cli.Command) {
	ctx, cancel := completionContext(ctx)
	defer cancel()
	completeName(ctx, c, listProfileNames)
}

func completeName(ctx context.Context, c *cli.Command, list func(context.Context) ([]string, error)) {
	if completeFlagToken(ctx, c) {
		return
	}
	if c.Args().Len() > 0 {
		return
	}

	lines, err := list(ctx)
	if err != nil {
		return
	}
	printCompletionLines(c, "", lines)
}

// CompleteFlags is a ShellComplete func for commands that have no positional
// name argument. It emits flag suggestions and values for --agent, --gateway,
// --profile, and --format.
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
	case flags.Agent:
		list = listAgentNames
	case flags.Gateway:
		list = listGatewayNames
	case flags.Profile:
		list = listProfileNames
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

func completionLine(name, description string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	description = strings.Join(strings.Fields(description), " ")
	if description == "" {
		return name
	}
	return name + ":" + description
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

func listAgentNames(ctx context.Context) ([]string, error) {
	client, err := daemonClient(ctx)
	if err != nil {
		return nil, err
	}

	agents, err := client.ListAgents(ctx)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(agents))
	for _, agent := range agents {
		line := completionLine(agent.Name, agent.Description)
		if line == "" {
			continue
		}
		names = append(names, line)
	}
	sort.Strings(names)
	return names, nil
}

func listGatewayNames(ctx context.Context) ([]string, error) {
	client, err := daemonClient(ctx)
	if err != nil {
		return nil, err
	}

	gateways, err := client.ListGateways(ctx)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(gateways))
	for _, gateway := range gateways {
		line := completionLine(gateway.Name, gateway.Description)
		if line == "" {
			continue
		}
		names = append(names, line)
	}
	sort.Strings(names)
	return names, nil
}

func listFormatNames(_ context.Context) ([]string, error) {
	return []string{formatter.FormatASCII, formatter.FormatJSON}, nil
}

func listProfileNames(ctx context.Context) ([]string, error) {
	v := ctx.Value(profilesKey)
	if v == nil {
		return nil, errors.New("profiles not loaded")
	}
	profs, ok := v.(*profile.Profiles)
	if !ok {
		return nil, errors.New("profiles not loaded")
	}

	names := profs.List()
	sort.Strings(names)

	lines := make([]string, 0, len(names))
	for _, name := range names {
		line := completionLine(name, "")
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines, nil
}

func daemonClient(ctx context.Context) (*daemon.Client, error) {
	conf := GetConfigFromContext(ctx)

	serverURL, _ := conf.String(flags.ServerURL)
	if serverURL == "" {
		return nil, errors.New("server URL not configured")
	}
	serverCert, _ := conf.String(flags.ServerCertificate)
	accountName, _ := conf.String(flags.Account)

	userJWT, userSeed, err := auth.ReadUser(accountName)
	if err != nil {
		return nil, err
	}

	srv, err := server.New([]string{serverURL}, userJWT, string(userSeed), serverCert)
	if err != nil {
		return nil, err
	}

	return daemon.New(srv), nil
}
