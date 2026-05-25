package gitignore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

const instructionsDir = ".instructions"

// Generate creates or replaces .gitignore in the current directory.
// Stack detection is based on filenames inside .instructions/.
// Falls back to a generic template when the folder is absent.
func Generate(_ context.Context, _ *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "DevManager :: Criar/Atualizar .gitignore")

	dir, _ := os.Getwd()

	stacks, err := detectStacks()
	if err != nil {
		ui.Err(stdout, "Falha ao ler .instructions: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("detectar stack: %w", err)
	}

	var content string
	if len(stacks) == 0 {
		ui.Info(stdout, "Pasta .instructions não encontrada — gerando .gitignore genérico.")
		content = genericContent()
	} else {
		ui.Info(stdout, "Stack detectada: "+strings.Join(stacks, ", "))
		content = buildContent(stacks)
	}

	dest := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
		ui.Err(stdout, "Falha ao escrever .gitignore: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("escrever .gitignore: %w", err)
	}

	ui.Success(stdout, ".gitignore criado/atualizado em "+dest)
	ui.WaitEnter(stdout)
	return nil
}

// detectStacks reads .instructions/ and returns recognized stack names.
func detectStacks() ([]string, error) {
	entries, err := os.ReadDir(instructionsDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	var stacks []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		key := strings.ToUpper(strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())))
		if name, ok := knownStacks[key]; ok && !seen[name] {
			seen[name] = true
			stacks = append(stacks, name)
		}
	}
	return stacks, nil
}

var knownStacks = map[string]string{
	"GOLANG":     "Go",
	"GO":         "Go",
	"BASH":       "Shell",
	"SHELL":      "Shell",
	"PHP":        "PHP",
	"NODE":       "Node.js",
	"JAVASCRIPT": "Node.js",
	"TYPESCRIPT": "Node.js",
	"PYTHON":     "Python",
	"RUBY":       "Ruby",
	"RUST":       "Rust",
	"JAVA":       "Java",
}

// ── Content sections ──────────────────────────────────────────────────────────

const sectionEditors = `# ─── Editors ────────────────────────────────────────────────────────────────
.vscode/
.idea/
*.swp
*.swo
*~
`

const sectionOS = `# ─── OS ─────────────────────────────────────────────────────────────────────
.DS_Store
Thumbs.db
`

const sectionGo = `# ─── Go ─────────────────────────────────────────────────────────────────────
*.exe
*.test
*.out
dist/
vendor/
`

const sectionShell = `# ─── Shell ───────────────────────────────────────────────────────────────────
.env
.env.*
!.env.example
`

const sectionPHP = `# ─── PHP ─────────────────────────────────────────────────────────────────────
vendor/
.env
.env.*
!.env.example
*.phar
`

const sectionNode = `# ─── Node.js ─────────────────────────────────────────────────────────────────
node_modules/
dist/
build/
.env
.env.*
!.env.example
npm-debug.log*
yarn-debug.log*
.pnp
.pnp.js
`

const sectionPython = `# ─── Python ──────────────────────────────────────────────────────────────────
__pycache__/
*.py[cod]
*.egg-info/
.venv/
venv/
dist/
build/
.pytest_cache/
`

const sectionRuby = `# ─── Ruby ────────────────────────────────────────────────────────────────────
.bundle/
vendor/bundle/
*.gem
.env
.env.*
!.env.example
`

const sectionRust = `# ─── Rust ────────────────────────────────────────────────────────────────────
target/
Cargo.lock
`

const sectionJava = `# ─── Java ────────────────────────────────────────────────────────────────────
*.class
*.jar
*.war
target/
.gradle/
build/
`

const sectionGenericExtra = `# ─── Environment ─────────────────────────────────────────────────────────────
.env
.env.*
!.env.example

# ─── Logs / Temp ─────────────────────────────────────────────────────────────
*.log
*.tmp
*.bak
logs/

# ─── Build / Dependencies ─────────────────────────────────────────────────────
vendor/
node_modules/
dist/
build/
`

var stackSections = map[string]string{
	"Go":     sectionGo,
	"Shell":  sectionShell,
	"PHP":    sectionPHP,
	"Node.js": sectionNode,
	"Python": sectionPython,
	"Ruby":   sectionRuby,
	"Rust":   sectionRust,
	"Java":   sectionJava,
}

func buildContent(stacks []string) string {
	var sb strings.Builder
	sb.WriteString(sectionEditors)
	sb.WriteString("\n")
	sb.WriteString(sectionOS)
	for _, stack := range stacks {
		if section, ok := stackSections[stack]; ok {
			sb.WriteString("\n")
			sb.WriteString(section)
		}
	}
	return sb.String()
}

func genericContent() string {
	return sectionEditors + "\n" + sectionOS + "\n" + sectionGenericExtra
}
