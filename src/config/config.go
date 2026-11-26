// Package config reads ccc.yml, the list of MCP servers, marketplaces and plugins to install.
package config

import (
	"errors"
	"strings"

	"go.yaml.in/yaml/v3"
)

const textTag = "!!str"

// Config is lenient per entry, like the Node version: a wrong type drops that entry, not the whole file.
type Config struct {
	MCP          map[string]Server `yaml:"mcp"`
	Marketplaces textList          `yaml:"marketplaces"`
	Plugins      textList          `yaml:"plugins"`
}

type Server struct {
	Command text     `yaml:"command"`
	URL     text     `yaml:"url"`
	Env     textList `yaml:"env"`
}

// text keeps only YAML strings, so `command: 42` or `url: [a]` reads as empty instead of failing.
type text string

func (value *text) UnmarshalYAML(node *yaml.Node) error {
	if isText(node) {
		*value = text(strings.TrimSpace(node.Value))
	}

	return nil
}

// textList accepts a single string or a list, and drops entries that aren't non-empty strings.
type textList []string

func (list *textList) UnmarshalYAML(node *yaml.Node) error {
	nodes := node.Content
	if node.Kind == yaml.ScalarNode {
		nodes = []*yaml.Node{node}
	}
	for _, item := range nodes {
		if isText(item) && strings.TrimSpace(item.Value) != "" {
			*list = append(*list, strings.TrimSpace(item.Value))
		}
	}

	return nil
}

// UnmarshalYAML ignores a server that isn't a mapping; it's reported later as missing a url or command.
func (server *Server) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return nil
	}
	type plain Server

	return node.Decode((*plain)(server))
}

func isText(node *yaml.Node) bool {
	return node.Kind == yaml.ScalarNode && node.Tag == textTag
}

func Parse(data []byte) (Config, error) {
	var current Config
	if err := yaml.Unmarshal(data, &current); err != nil {
		// yaml.v3 errors can span lines; the first one names the problem and its line.
		return Config{}, errors.New(strings.SplitN(err.Error(), "\n", 2)[0])
	}

	return current, nil
}
