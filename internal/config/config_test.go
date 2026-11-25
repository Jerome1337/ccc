package config

import (
	"os"
	"slices"
	"testing"
)

func TestParseShippedConfig(t *testing.T) {
	data, err := os.ReadFile("../../ccc.yml")
	if err != nil {
		t.Fatal(err)
	}
	current, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(current.MCP) != 2 || len(current.Marketplaces) != 5 || len(current.Plugins) != 8 {
		t.Fatalf("got %d servers, %d marketplaces, %d plugins, want 2, 5, 8",
			len(current.MCP), len(current.Marketplaces), len(current.Plugins))
	}
	if !slices.Equal(current.MCP["elevenlabs"].Env, []string{"ELEVENLABS_API_KEY"}) {
		t.Errorf("elevenlabs env = %v", current.MCP["elevenlabs"].Env)
	}
	if !slices.Contains(current.Plugins, "humanizer@humanizer") {
		t.Errorf("plugins = %v, want humanizer@humanizer", current.Plugins)
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
