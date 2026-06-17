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

	"github.com/kaduvelasco/lumina-tools/internal/config"
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
// If the context was previously generated, asks whether to update before
// proceeding. In update mode all files are overwritten without confirmation.
// Selecting = generate/add; deselecting = remove.
func GenerateContext(ctx context.Context, _ *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Criar Contexto AI")
	d, err := os.Getwd()
	if err != nil {
		d = "desconhecido"
	}
	ui.Info(stdout, "Diretório atual: "+d)

	present := detectActiveModels()
	update := false

	if len(present) > 0 {
		fmt.Fprint(stdout, "\nContexto AI já existe neste diretório. Deseja atualizar? (s/N): ")
		line, _ := prompt.ReadLineFrom(stdin)
		if c := strings.TrimSpace(line); c != "s" && c != "S" {
			ui.Info(stdout, "Operação cancelada.")
			ui.WaitEnter(stdout)
			return nil
		}
		update = true
	}

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

	// Determine active set and models to remove.
	var active, toRemove []Model
	for i, item := range finalItems {
		m := models[i]
		if item.Selected {
			active = append(active, m)
		} else if present[m.Name] {
			toRemove = append(toRemove, m)
		}
	}

	// In fresh mode only newly selected models are written; in update mode all
	// selected models are rewritten. The overwrite flag mirrors the update flag.
	var toWrite []Model
	if !update {
		for i, item := range finalItems {
			if item.Selected && !present[models[i].Name] {
				toWrite = append(toWrite, models[i])
			}
		}
		if len(toWrite) == 0 && len(toRemove) == 0 {
			ui.Info(stdout, "Nenhuma alteração necessária.")
			ui.WaitEnter(stdout)
			return nil
		}
	} else {
		toWrite = active
	}

	// Resolve workspace path for template substitution (e.g. {{WORKSPACE_PATH}} in MOODLE.md).
	workspacePath := filepath.Join(os.Getenv("HOME"), "workspace")
	if cfg, cfgErr := config.Load(); cfgErr == nil && cfg.WorkspacePath != "" {
		workspacePath = cfg.WorkspacePath
	} else if home, homeErr := os.UserHomeDir(); homeErr == nil {
		workspacePath = filepath.Join(home, "workspace")
	}

	ui.PrintHeader(stdout, "Criar Contexto AI")

	// Show the info panel before any file operations so it appears above the results box.
	if update || len(active) > 0 {
		switch {
		case update && len(active) == 0:
			ui.Info(stdout, "Removendo referências dos arquivos de contexto compartilhados...")
		case update:
			ui.Info(stdout, "Regenerando arquivos de contexto para: "+modelNames(active))
		default:
			ui.Info(stdout, "Atualizando arquivos de contexto para: "+modelNames(active))
		}
	}

	// Collect all file operations into a log; display them together after all writes.
	var log strings.Builder

	for _, m := range toRemove {
		removeInstruction(m, stdout, &log)
	}
	for _, m := range toWrite {
		var vars map[string]string
		if m.Name == "Moodle" {
			vars = map[string]string{"WORKSPACE_PATH": workspacePath}
		}
		if err := writeInstruction(m, update, stdin, stdout, &log, vars); err != nil {
			ui.Err(stdout, "Falha ao gerar instrução para "+m.Name+": "+err.Error())
			ui.WaitEnter(stdout)
			return err
		}
	}

	// Regenerate shared files. In update mode this always runs (even when active
	// is empty) to clear dangling @-references left by removed instructions.
	// In fresh mode it only runs when at least one model remains active.
	if update || len(active) > 0 {
		if err := generateSharedFiles(active, update, stdin, stdout, &log); err != nil {
			ui.Err(stdout, "Falha ao gerar arquivos: "+err.Error())
			ui.WaitEnter(stdout)
			return err
		}
	}

	if log.Len() > 0 {
		ui.PrintBox(stdout, strings.TrimRight(log.String(), "\n"))
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
// When overwrite is true, existing files are replaced without prompting.
func generateSharedFiles(active []Model, overwrite bool, stdin io.Reader, stdout io.Writer, log *strings.Builder) error {
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
		if err := writeFile(f.filename, f.content, overwrite, stdin, stdout, log); err != nil {
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
		if err := writeFile(name, aiexclude, overwrite, stdin, stdout, log); err != nil {
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

func writeFile(name, content string, overwrite bool, stdin io.Reader, stdout io.Writer, log *strings.Builder) error {
	if _, err := os.Stat(name); err == nil && !overwrite {
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
	fmt.Fprintf(log, "  + %s criado.\n", name)
	return nil
}

func writeInstruction(model Model, overwrite bool, stdin io.Reader, stdout io.Writer, log *strings.Builder, vars map[string]string) error {
	dir := ".instructions"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	dest := filepath.Join(dir, filepath.Base(model.Instruction))
	content, err := readTpl(model.Instruction)
	if err != nil {
		return err
	}
	for k, v := range vars {
		content = strings.ReplaceAll(content, "{{"+k+"}}", v)
	}
	return writeFile(dest, content, overwrite, stdin, stdout, log)
}

// ClearContext lists all AI context files present in the current directory,
// shows a multi-select so the user can choose which to remove, then deletes
// only the selected items.
func ClearContext(ctx context.Context, _ *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Remover Contextos AI")
	d, err := os.Getwd()
	if err != nil {
		d = "desconhecido"
	}
	ui.Info(stdout, "Diretório atual: "+d)

	type candidate struct {
		path  string
		label string
		isDir bool
	}

	sharedFiles := []string{
		"CLAUDE.md", "GEMINI.md", "AGENTS.md",
		".windsurfrules", ".cursorrules",
		".aiexclude", ".claudeignore", ".geminiignore",
	}

	var found []candidate
	for _, f := range sharedFiles {
		if _, err := os.Stat(f); err == nil {
			found = append(found, candidate{path: f, label: f})
		}
	}
	for _, m := range models {
		dest := filepath.Join(".instructions", filepath.Base(m.Instruction))
		if _, err := os.Stat(dest); err == nil {
			found = append(found, candidate{path: dest, label: dest})
		}
	}
	phpRefDir := filepath.Join(".instructions", "php-references")
	if _, err := os.Stat(phpRefDir); err == nil {
		found = append(found, candidate{path: phpRefDir, label: phpRefDir + "/", isDir: true})
	}

	if len(found) == 0 {
		ui.Info(stdout, "Nenhum contexto AI encontrado neste diretório.")
		ui.WaitEnter(stdout)
		return nil
	}

	// All items pre-selected — user deselects what they want to keep.
	items := make([]ui.SelectItem, len(found))
	for i, c := range found {
		items[i] = ui.SelectItem{ID: c.path, Label: c.label, Selected: true}
	}

	finalItems, confirmed, err := ui.RunMultiSelect(ctx, stdin, stdout, items)
	if err != nil {
		return err
	}
	if !confirmed {
		ui.Info(stdout, "Operação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}

	var sb strings.Builder
	removed := 0
	for i, item := range finalItems {
		if !item.Selected {
			continue
		}
		c := found[i]
		var removeErr error
		if c.isDir {
			removeErr = os.RemoveAll(c.path)
		} else {
			removeErr = os.Remove(c.path)
		}
		if removeErr != nil && !os.IsNotExist(removeErr) {
			ui.Warning(stdout, "Falha ao remover "+c.path+": "+removeErr.Error())
		} else if removeErr == nil {
			fmt.Fprintf(&sb, "  - %s\n", c.label)
			removed++
		}
	}

	// Remove .instructions/ if it is now empty.
	if entries, err := os.ReadDir(".instructions"); err == nil && len(entries) == 0 {
		if err := os.Remove(".instructions"); err == nil {
			fmt.Fprintf(&sb, "  - .instructions/\n")
			removed++
		}
	}

	if removed > 0 {
		ui.PrintBox(stdout, "Arquivos removidos:\n"+sb.String())
	}
	ui.Success(stdout, fmt.Sprintf("%d item(ns) removido(s).", removed))
	ui.WaitEnter(stdout)
	return nil
}

func removeInstruction(model Model, stdout io.Writer, log *strings.Builder) {
	dest := filepath.Join(".instructions", filepath.Base(model.Instruction))
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		ui.Warning(stdout, "Falha ao remover "+dest+": "+err.Error())
	} else if err == nil {
		fmt.Fprintf(log, "  - %s removido.\n", dest)
	}
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
