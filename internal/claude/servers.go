package claude

import (
	"maps"
	"slices"
	"strings"

	"github.com/Jerome1337/ccc/internal/config"
	"github.com/Jerome1337/ccc/internal/term"
)

const existsMarker = "already exists"

var mcpAdd = []string{"mcp", "add", "-s", "user"}

// installServers uses `claude mcp add`. `mcp add-json -` would keep env values off the command line,
// but Claude Code refuses to read it from stdin when ccc runs inside a Claude Code session.
func (installer *Installer) installServers(servers map[string]config.Server) {
	for _, name := range slices.Sorted(maps.Keys(servers)) {
		args, skipReason := installer.addArgs(name, servers[name])
		if skipReason != "" {
			installer.print(term.Yellow("skip  "), name, skipReason)
			installer.summary.Skipped++
			continue
		}

		output, err := installer.Runner.Capture(args...)
		message := strings.TrimSpace(string(output))
		switch {
		case err != nil && strings.Contains(message, existsMarker):
			installer.summary.Present++
		case err != nil:
			installer.print(term.Red("fail  "), name, message)
			installer.summary.Failed++
		default:
			installer.print(term.Green("add   "), name, "")
			installer.summary.Installed++
		}
	}
}

func (installer *Installer) addArgs(name string, server config.Server) ([]string, string) {
	if server.URL != "" {
		return append(slices.Clone(mcpAdd), "--transport", "http", name, string(server.URL)), ""
	}

	fields := strings.Fields(string(server.Command))
	if len(fields) == 0 {
		return nil, "needs a url or a command"
	}

	args := append(slices.Clone(mcpAdd), name)
	var missing []string
	for _, key := range server.Env {
		value, isSet := installer.LookupEnv(key)
		if !isSet || value == "" {
			missing = append(missing, key)
			continue
		}
		args = append(args, "-e", key+"="+value)
	}
	if len(missing) > 0 {
		return nil, "set " + strings.Join(missing, ", ") + " and run again"
	}

	return append(append(args, "--"), fields...), ""
}
