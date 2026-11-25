package claude

import (
	"bytes"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/Jerome1337/ccc/internal/config"
)

// fakeRunner answers the two listings from testdata and records every other call.
type fakeRunner struct {
	t            *testing.T
	plugins      []byte
	marketplaces []byte
	addOutput    map[string]string
	calls        []string
}

func newFakeRunner(t *testing.T) *fakeRunner {
	plugins, err := os.ReadFile("testdata/plugins.json")
	if err != nil {
		t.Fatal(err)
	}
	marketplaces, err := os.ReadFile("testdata/marketplaces.json")
	if err != nil {
		t.Fatal(err)
	}

	return &fakeRunner{t: t, plugins: plugins, marketplaces: marketplaces, addOutput: map[string]string{}}
}

func (runner *fakeRunner) Output(args ...string) ([]byte, error) {
	switch strings.Join(args, " ") {
	case "plugin list --json":
		return runner.plugins, nil
	case "plugin marketplace list --json":
		return runner.marketplaces, nil
	}
	runner.t.Fatalf("unexpected Output(%v)", args)

	return nil, nil
}

func (runner *fakeRunner) Run(args ...string) error {
	runner.calls = append(runner.calls, strings.Join(args, " "))

	return nil
}

// Capture is only used by `mcp add`, where the server name follows the scope.
func (runner *fakeRunner) Capture(args ...string) ([]byte, error) {
	runner.calls = append(runner.calls, strings.Join(args, " "))
	name := args[4]
	if args[4] == "--transport" {
		name = args[6]
	}
	if output, hasOutput := runner.addOutput[name]; hasOutput {
		return []byte(output), errors.New("exit status 1")
	}

	return nil, nil
}

func noEnv(string) (string, bool) { return "", false }

func TestInstallOnlyAddsMissingEntries(t *testing.T) {
	runner := newFakeRunner(t)
	installer := Installer{Runner: runner, LookupEnv: noEnv, Out: &bytes.Buffer{}}
	summary, err := installer.Install(config.Config{
		Marketplaces: []string{"latent-spaces/brag", "https://github.com/AgriciDaniel/claude-seo", "https://github.com/DietrichGebert/ponytail.git", "blader/humanizer"},
		Plugins:      []string{"brag@brag", "claude-seo", "humanizer@humanizer", "ponytail@ponytail", "frontend-design@claude-plugins-official"},
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"plugin marketplace add blader/humanizer",
		"plugin install ponytail@ponytail",
		"plugin install frontend-design@claude-plugins-official",
	}
	if !slices.Equal(runner.calls, want) {
		t.Errorf("calls = %q\nwant %q", runner.calls, want)
	}
	if summary.Installed != 3 || summary.Present != 6 {
		t.Errorf("summary = %+v, want 3 installed and 6 present", summary)
	}
}

func TestInstallServers(t *testing.T) {
	runner := newFakeRunner(t)
	runner.addOutput["existing"] = "✘ MCP server existing already exists in user config"
	runner.addOutput["broken"] = "invalid command"
	env := map[string]string{"API_KEY": "secret-value"}
	lookupEnv := func(key string) (string, bool) {
		value, isSet := env[key]

		return value, isSet
	}
	out := &bytes.Buffer{}
	installer := Installer{Runner: runner, LookupEnv: lookupEnv, Out: out}

	summary, err := installer.Install(config.Config{MCP: map[string]config.Server{
		"stdio":    {Command: "npx -y tool@latest", Env: []string{"API_KEY"}},
		"remote":   {URL: "https://example.com/mcp"},
		"needs":    {Command: "uvx tool", Env: []string{"MISSING_KEY"}},
		"empty":    {},
		"existing": {Command: "tool"},
		"broken":   {Command: "tool"},
	}})
	if err != nil {
		t.Fatal(err)
	}

	if summary.Installed != 2 || summary.Present != 1 || summary.Skipped != 2 || summary.Failed != 1 {
		t.Errorf("summary = %+v", summary)
	}
	wantCalls := []string{
		"mcp add -s user broken -- tool",
		"mcp add -s user existing -- tool",
		"mcp add -s user --transport http remote https://example.com/mcp",
		"mcp add -s user stdio -e API_KEY=secret-value -- npx -y tool@latest",
	}
	if !slices.Equal(runner.calls, wantCalls) {
		t.Errorf("calls = %q\nwant %q", runner.calls, wantCalls)
	}
	if strings.Contains(out.String(), "secret-value") {
		t.Error("an env value was printed")
	}
	if !strings.Contains(out.String(), "set MISSING_KEY and run again") {
		t.Errorf("output = %s, want the missing env hint", out.String())
	}
}

func TestInstallFailsClosed(t *testing.T) {
	runner := newFakeRunner(t)
	runner.plugins = []byte("Warning: not JSON")
	installer := Installer{Runner: runner, LookupEnv: noEnv, Out: &bytes.Buffer{}}
	_, err := installer.Install(config.Config{Plugins: []string{"a@b"}, MCP: map[string]config.Server{"x": {Command: "x"}}})
	if err == nil {
		t.Fatal("Install() should fail when the plugin list can't be parsed")
	}
	if len(runner.calls) != 0 {
		t.Errorf("calls = %v, want nothing run", runner.calls)
	}
}

func TestNormalizeSource(t *testing.T) {
	for _, source := range []string{"Owner/Repo", "https://github.com/owner/repo", "https://github.com/owner/repo.git", "git@github.com:owner/repo.git", "owner/repo/"} {
		if got := normalizeSource(source); got != "owner/repo" {
			t.Errorf("normalizeSource(%q) = %q", source, got)
		}
	}
}
