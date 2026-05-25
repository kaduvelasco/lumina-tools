package mcp

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"gopkg.in/yaml.v3"
)

// Server describes an MCP server managed by lumina.
type Server struct {
	Name        string `yaml:"name"`
	Package     string `yaml:"package"`
	Cmd         string `yaml:"cmd"`
	Description string `yaml:"description"`
}

type catalogue struct {
	Servers []Server `yaml:"servers"`
}

//go:embed servers.yaml
var embeddedYAML []byte

var (
	catalogueOnce  sync.Once
	catalogueCache []Server
	catalogueErr   error
)

// Catalogue returns all MCP servers from the embedded YAML.
// The YAML is parsed only once; subsequent calls return the cached result.
func Catalogue() ([]Server, error) {
	catalogueOnce.Do(func() {
		var c catalogue
		if err := yaml.Unmarshal(embeddedYAML, &c); err != nil {
			catalogueErr = fmt.Errorf("parse mcp catalogue: %w", err)
			return
		}
		catalogueCache = c.Servers
	})
	return catalogueCache, catalogueErr
}

// InstalledMap returns which servers are currently installed (by Name).
func InstalledMap(ctx context.Context, exe *executor.Executor, servers []Server) map[string]bool {
	result := make(map[string]bool, len(servers))
	for _, s := range servers {
		if _, err := exe.Output(ctx, executor.Options{}, "which", s.Cmd); err == nil {
			result[s.Name] = true
		}
	}
	return result
}
