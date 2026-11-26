package config

import (
	"slices"
	"testing"
)

// readmeExample is the ccc.yml shown in the README.
const readmeExample = `mcp:
  heroui-react:
    command: npx -y @heroui/react-mcp@latest   # stdio server, arguments split on spaces
  elevenlabs:
    command: uvx elevenlabs-mcp
    env: [ELEVENLABS_API_KEY]                  # read from your environment
  sentry:
    url: https://mcp.sentry.dev/mcp            # http server

marketplaces:                                  # passed to claude plugin marketplace add
  - anthropics/claude-plugins-official

plugins:                                       # passed to claude plugin install
  - context7@claude-plugins-official
`

func TestParseReadmeExample(t *testing.T) {
	current, err := Parse([]byte(readmeExample))
	if err != nil {
		t.Fatal(err)
	}
	if len(current.MCP) != 3 || len(current.Marketplaces) != 1 || len(current.Plugins) != 1 {
		t.Fatalf("got %d servers, %d marketplaces, %d plugins, want 3, 1, 1",
			len(current.MCP), len(current.Marketplaces), len(current.Plugins))
	}
	if !slices.Equal(current.MCP["elevenlabs"].Env, []string{"ELEVENLABS_API_KEY"}) {
		t.Errorf("elevenlabs env = %v", current.MCP["elevenlabs"].Env)
	}
	if current.MCP["sentry"].URL != "https://mcp.sentry.dev/mcp" {
		t.Errorf("sentry url = %q", current.MCP["sentry"].URL)
	}
}

func TestParseLenientEntries(t *testing.T) {
	current, err := Parse([]byte(`
mcp:
  single-env:
    command: uvx tool
    env: API_KEY
  remote:
    url: https://example.com/mcp
  numeric:
    command: 42
  scalar: just-a-string
  empty:
marketplaces: owner/repo
plugins:
  - a@b
  - 12
  - [nested]
  - ""
  - null
`))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(current.MCP["single-env"].Env, []string{"API_KEY"}) {
		t.Errorf("env as a single string = %v", current.MCP["single-env"].Env)
	}
	if current.MCP["remote"].URL != "https://example.com/mcp" {
		t.Errorf("url = %q", current.MCP["remote"].URL)
	}
	for _, name := range []string{"numeric", "scalar", "empty"} {
		server := current.MCP[name]
		if server.Command != "" || server.URL != "" {
			t.Errorf("%s = %+v, want an empty server", name, server)
		}
	}
	if !slices.Equal(current.Marketplaces, []string{"owner/repo"}) {
		t.Errorf("marketplaces = %v", current.Marketplaces)
	}
	if !slices.Equal(current.Plugins, []string{"a@b"}) {
		t.Errorf("plugins = %v, want only the text entry", current.Plugins)
	}
}

func TestParseInvalidYAML(t *testing.T) {
	if _, err := Parse([]byte("mcp: [unclosed")); err == nil {
		t.Fatal("Parse() should fail on invalid YAML")
	}
}

func TestParseEmptyFile(t *testing.T) {
	current, err := Parse(nil)
	if err != nil || len(current.Plugins) != 0 {
		t.Fatalf("Parse(nil) = %+v, %v", current, err)
	}
}
