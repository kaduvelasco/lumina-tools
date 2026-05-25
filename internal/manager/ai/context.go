package ai

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

//go:embed all:templates
var templateFS embed.FS

// Model represents a project type with its own instruction file.
type Model struct {
	Name        string
	Instruction string // path inside templateFS
}

var models = []Model{
	{Name: "Go", Instruction: "templates/instructions/GOLANG.md"},
	{Name: "Linux Bash", Instruction: "templates/instructions/BASH.md"},
	{Name: "MCP Server", Instruction: "templates/instructions/MCP.md"},
	{Name: "PHP", Instruction: "templates/instructions/PHP.md"},
	{Name: "Moodle", Instruction: "templates/instructions/MOODLE.md"},
}

// GenerateContext shows a multiselect of all project models.
// Models whose instruction file already exists in .instructions/ start selected.
// Selecting = generate/add; deselecting = remove. After confirming, shared
// context files are regenerated to reference all currently active models.
func GenerateContext(ctx context.Context, _ *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Criar Contexto AI")
	d, _ := os.Getwd()
	ui.Info(stdout, "Diretório atual: "+d)

	present := detectActiveModels()

	items := make([]ui.SelectItem, len(models))
	for i, m := range models {
		items[i] = ui.SelectItem{Label: m.Name, ID: m.Name, Selected: present[m.Name]}
	}

	finalItems, confirmed, err := ui.RunMultiSelect(ctx, stdin, stdout, items)
	if err != nil {
		return err
	}
	if !confirmed {
		ui.Warning(stdout, "Operação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}

	// Compute diff between previous state and new selection.
	var toAdd, toRemove []Model
	for i, item := range finalItems {
		m := models[i]
		switch {
		case item.Selected && !present[m.Name]:
			toAdd = append(toAdd, m)
		case !item.Selected && present[m.Name]:
			toRemove = append(toRemove, m)
		}
	}

	if len(toAdd) == 0 && len(toRemove) == 0 {
		ui.Info(stdout, "Nenhuma alteração necessária.")
		ui.WaitEnter(stdout)
		return nil
	}

	// Determine which models remain active after the operation.
	// finalItems[i] corresponds to models[i] — same order as the items slice built above.
	var active []Model
	for i, item := range finalItems {
		if item.Selected {
			active = append(active, models[i])
		}
	}

	ui.PrintHeader(stdout, "Criar Contexto AI")

	// Remove instruction files for deselected models.
	for _, m := range toRemove {
		dest := filepath.Join(".instructions", filepath.Base(m.Instruction))
		if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
			ui.Warning(stdout, "Falha ao remover "+dest+": "+err.Error())
		} else {
			fmt.Fprintf(stdout, "  - %s removido.\n", dest)
		}
	}

	// Generate instruction files for newly selected models.
	for _, m := range toAdd {
		if err := writeInstruction(m, stdin, stdout); err != nil {
			ui.Err(stdout, "Falha ao gerar instrução para "+m.Name+": "+err.Error())
			ui.WaitEnter(stdout)
			return err
		}
	}

	// Regenerate shared context files based on the full active set.
	if len(active) > 0 {
		ui.Info(stdout, "Atualizando arquivos de contexto para: "+modelNames(active))
		if err := generateSharedFiles(active, stdin, stdout); err != nil {
			ui.Err(stdout, "Falha ao gerar arquivos: "+err.Error())
			ui.WaitEnter(stdout)
			return err
		}
	}

	ui.Success(stdout, "Contexto AI atualizado com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}

// detectActiveModels checks the .instructions/ directory in the current working
// directory and returns which models already have their instruction file present.
func detectActiveModels() map[string]bool {
	present := make(map[string]bool, len(models))
	for _, m := range models {
		dest := filepath.Join(".instructions", filepath.Base(m.Instruction))
		if _, err := os.Stat(dest); err == nil {
			present[m.Name] = true
		}
	}
	return present
}

// generateSharedFiles writes CLAUDE.md, GEMINI.md, AGENTS.md, .windsurfrules
// and .cursorrules referencing all active models. When multiple models are
// active their @-references are stacked; for inline files the content is concatenated.
func generateSharedFiles(active []Model, stdin io.Reader, stdout io.Writer) error {
	rawBasic, err := readTpl("templates/BASIC.md")
	if err != nil {
		return err
	}
	onlyClaude, err := readTpl("templates/ONLY-CLAUDE.md")
	if err != nil {
		return err
	}
	onlyGemini, err := readTpl("templates/ONLY-GEMINI.md")
	if err != nil {
		return err
	}

	// Build the @-reference block (one line per active model).
	var refBlock strings.Builder
	refBlock.WriteString("\n\n## Language-Specific Standards\n")
	for _, m := range active {
		refBlock.WriteString("\n@.instructions/")
		refBlock.WriteString(filepath.Base(m.Instruction))
	}
	instructionRef := refBlock.String()

	// Build the inline block (concatenated instruction content).
	var inlineBlock strings.Builder
	inlineBlock.WriteString("\n\n## Language-Specific Standards\n\n")
	for _, m := range active {
		content, err := readTpl(m.Instruction)
		if err != nil {
			return err
		}
		inlineBlock.WriteString(content)
		inlineBlock.WriteString("\n\n")
	}

	buildContent := func(filename, extra string) string {
		base := strings.ReplaceAll(rawBasic, "{{AGENT_FILE}}", filename)
		if extra == "" {
			return base + instructionRef
		}
		return base + "\n\n" + extra + instructionRef
	}

	buildContentInline := func(filename string) string {
		base := strings.ReplaceAll(rawBasic, "{{AGENT_FILE}}", filename)
		return base + inlineBlock.String()
	}

	type entry struct {
		filename string
		content  string
	}
	files := []entry{
		{"CLAUDE.md", buildContent("CLAUDE.md", onlyClaude)},
		{"GEMINI.md", buildContent("GEMINI.md", onlyGemini)},
		{"AGENTS.md", buildContent("AGENTS.md", "")},
		{".windsurfrules", buildContentInline(".windsurfrules")},
		{".cursorrules", buildContentInline(".cursorrules")},
	}
	for _, f := range files {
		if err := writeFile(f.filename, f.content, stdin, stdout); err != nil {
			return err
		}
	}

	// PHP references when any active model is PHP.
	for _, m := range active {
		if m.Name == "PHP" {
			if err := copyPHPReferences(stdout); err != nil {
				return err
			}
			break
		}
	}

	// Ignore files (shared, always regenerated).
	aiexclude, err := readTpl("templates/.aiexclude")
	if err != nil {
		return err
	}
	for _, name := range []string{".aiexclude", ".claudeignore", ".geminiignore"} {
		if err := writeFile(name, aiexclude, stdin, stdout); err != nil {
			return err
		}
	}

	return nil
}

func modelNames(ms []Model) string {
	names := make([]string, len(ms))
	for i, m := range ms {
		names[i] = m.Name
	}
	return strings.Join(names, ", ")
}

func writeFile(name, content string, stdin io.Reader, stdout io.Writer) error {
	if _, err := os.Stat(name); err == nil {
		fmt.Fprintf(stdout, "  %s ja existe. Sobrescrever? (s/N): ", name)
		line, _ := prompt.ReadLineFrom(stdin)
		confirm := strings.TrimSpace(line)
		if confirm != "s" && confirm != "S" {
			fmt.Fprintf(stdout, "  %s mantido.\n", name)
			return nil
		}
	}
	if err := os.WriteFile(name, []byte(content), 0o644); err != nil {
		return fmt.Errorf("escrever %s: %w", name, err)
	}
	fmt.Fprintf(stdout, "  + %s criado.\n", name)
	return nil
}

func writeInstruction(model Model, stdin io.Reader, stdout io.Writer) error {
	dir := ".instructions"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	dest := filepath.Join(dir, filepath.Base(model.Instruction))
	content, err := readTpl(model.Instruction)
	if err != nil {
		return err
	}
	return writeFile(dest, content, stdin, stdout)
}

func copyPHPReferences(stdout io.Writer) error {
	dir := ".instructions/php-references"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	entries, err := fs.ReadDir(templateFS, "templates/instructions/php-references")
	if err != nil {
		return nil // not fatal if missing
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := templateFS.ReadFile("templates/instructions/php-references/" + e.Name())
		if err != nil {
			fmt.Fprintf(stdout, "  aviso: ler %s: %v\n", e.Name(), err)
			continue
		}
		dest := filepath.Join(dir, e.Name())
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			fmt.Fprintf(stdout, "  aviso: %s: %v\n", dest, err)
		}
	}
	return nil
}

func readTpl(path string) (string, error) {
	data, err := templateFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("template %s nao encontrado: %w", path, err)
	}
	return string(data), nil
}

