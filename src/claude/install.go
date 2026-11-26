package claude

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/Jerome1337/ccc/src/config"
	"github.com/Jerome1337/ccc/src/term"
)

var sourcePrefixes = []string{"https://github.com/", "http://github.com/", "git@github.com:"}

// Summary counts what happened, for the closing line and the /mcp and /plugin reminder.
type Summary struct {
	Installed int
	Present   int
	Skipped   int
	Failed    int
}

type Installer struct {
	Runner    Runner
	LookupEnv func(string) (string, bool)
	Out       io.Writer
	summary   Summary
}

type installed struct {
	plugins      []string
	marketplaces []string
}

// Install adds what's missing. It fails closed: if the installed lists can't be read, it changes nothing.
func (installer *Installer) Install(current config.Config) (Summary, error) {
	state, err := installer.readInstalled()
	if err != nil {
		return Summary{}, err
	}

	installer.installServers(current.MCP)
	installer.addMarketplaces(current.Marketplaces, state.marketplaces)
	installer.installPlugins(current.Plugins, state.plugins)

	return installer.summary, nil
}

func (installer *Installer) readInstalled() (installed, error) {
	var plugins []struct {
		ID string `json:"id"`
	}
	if err := installer.readJSON(&plugins, "plugin", "list", "--json"); err != nil {
		return installed{}, err
	}

	var marketplaces []struct {
		Repo string `json:"repo"`
		URL  string `json:"url"`
	}
	if err := installer.readJSON(&marketplaces, "plugin", "marketplace", "list", "--json"); err != nil {
		return installed{}, err
	}

	state := installed{}
	for _, plugin := range plugins {
		state.plugins = append(state.plugins, plugin.ID)
	}
	for _, marketplace := range marketplaces {
		state.marketplaces = append(state.marketplaces, normalizeSource(marketplace.Repo), normalizeSource(marketplace.URL))
	}

	return state, nil
}

func (installer *Installer) readJSON(target any, args ...string) error {
	command := "claude " + strings.Join(args, " ")
	data, err := installer.Runner.Output(args...)
	if err != nil {
		return fmt.Errorf("%s failed: %w", command, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("%s returned unexpected output: %w", command, err)
	}

	return nil
}

// normalizeSource lets `owner/repo`, its GitHub URL and its .git URL match the same marketplace.
func normalizeSource(source string) string {
	normalized := strings.ToLower(strings.TrimSpace(source))
	for _, prefix := range sourcePrefixes {
		normalized = strings.TrimPrefix(normalized, prefix)
	}

	return strings.TrimSuffix(strings.TrimSuffix(normalized, "/"), ".git")
}

func (installer *Installer) addMarketplaces(sources, present []string) {
	for _, source := range sources {
		if slices.Contains(present, normalizeSource(source)) {
			installer.summary.Present++
			continue
		}
		installer.report(source, installer.Runner.Run("plugin", "marketplace", "add", source))
	}
}

func (installer *Installer) installPlugins(plugins, present []string) {
	for _, plugin := range plugins {
		if isPluginInstalled(plugin, present) {
			installer.summary.Present++
			continue
		}
		installer.report(plugin, installer.Runner.Run("plugin", "install", plugin))
	}
}

// isPluginInstalled matches `name@marketplace` exactly, and a bare `name` against any marketplace.
func isPluginInstalled(plugin string, present []string) bool {
	if strings.Contains(plugin, "@") {
		return slices.Contains(present, plugin)
	}

	return slices.ContainsFunc(present, func(id string) bool {
		return strings.HasPrefix(id, plugin+"@")
	})
}

func (installer *Installer) report(name string, err error) {
	if err != nil {
		installer.print(term.Red("fail  "), name, err.Error())
		installer.summary.Failed++

		return
	}
	installer.summary.Installed++
}

func (installer *Installer) print(label, name, detail string) {
	if detail == "" {
		fmt.Fprintf(installer.Out, "  %s  %s\n", label, name)

		return
	}
	fmt.Fprintf(installer.Out, "  %s  %s %s\n", label, name, term.Dim("("+detail+")"))
}
