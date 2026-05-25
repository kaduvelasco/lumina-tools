package templates

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/sets"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Template describes a file template managed by lumina.
type Template struct {
	Label    string
	Filename string
}

// Catalogue lists all templates managed by lumina.
var Catalogue = []Template{
	{Label: "Documento Word (.docx)", Filename: "Documento.docx"},
	{Label: "Planilha Excel (.xlsx)", Filename: "Planilha.xlsx"},
	{Label: "Apresentação PowerPoint (.pptx)", Filename: "Apresentacao.pptx"},
	{Label: "Documento LibreOffice (.odt)", Filename: "Documento.odt"},
	{Label: "Planilha LibreOffice (.ods)", Filename: "Planilha.ods"},
	{Label: "Texto (.txt)", Filename: "Texto.txt"},
	{Label: "HTML (.html)", Filename: "HTML.html"},
	{Label: "CSS (.css)", Filename: "Estilo.css"},
	{Label: "JavaScript (.js)", Filename: "Script.js"},
	{Label: "Python (.py)", Filename: "Script.py"},
	{Label: "PHP (.php)", Filename: "PHP.php"},
	{Label: "Shell (.sh)", Filename: "Shell.sh"},
}

// Dir returns the user templates directory, creating it when needed.
func Dir() (string, error) {
	out, err := exec.Command("xdg-user-dir", "TEMPLATES").Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", herr
		}
		return filepath.Join(home, "Templates"), nil
	}
	return strings.TrimSpace(string(out)), nil
}

// PresentNames returns a set of template filenames that already exist in dir.
func PresentNames(dir string) map[string]bool {
	result := make(map[string]bool, len(Catalogue))
	for _, t := range Catalogue {
		if _, err := os.Stat(filepath.Join(dir, t.Filename)); err == nil {
			result[t.Filename] = true
		}
	}
	return result
}

// Select shows an interactive multi-select for templates and applies the diff.
func Select(ctx context.Context, _ *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Templates de Arquivos")

	dir, err := Dir()
	if err != nil {
		ui.Err(stdout, "Erro ao obter diretorio de templates: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	present := PresentNames(dir)
	items := make([]ui.SelectItem, len(Catalogue))
	for i, t := range Catalogue {
		items[i] = ui.SelectItem{Label: t.Label, ID: t.Filename, Selected: present[t.Filename]}
	}

	finalItems, confirmed, err := ui.RunMultiSelect(ctx, stdin, stdout, items)
	if err != nil {
		return err
	}
	if !confirmed {
		ui.Warning(stdout, "Operacao cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}

	var toCreate, toRemove []string
	for _, item := range finalItems {
		switch {
		case item.Selected && !present[item.ID]:
			toCreate = append(toCreate, item.ID)
		case !item.Selected && present[item.ID]:
			toRemove = append(toRemove, item.ID)
		}
	}

	if len(toCreate) == 0 && len(toRemove) == 0 {
		ui.Info(stdout, "Nenhuma alteracao necessaria.")
		ui.WaitEnter(stdout)
		return nil
	}

	ui.PrintHeader(stdout, "Templates de Arquivos")
	if err := Apply(stdout, dir, toCreate, toRemove); err != nil {
		ui.Err(stdout, "Erro ao aplicar alteracoes: "+err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	ui.Success(stdout, "Templates atualizados com sucesso!")
	ui.WaitEnter(stdout)
	return nil
}

// Apply creates templates in toCreate and removes those in toRemove.
func Apply(stdout io.Writer, dir string, toCreate, toRemove []string) error {
	createSet := sets.Of(toCreate)
	removeSet := sets.Of(toRemove)

	for _, t := range Catalogue {
		dest := filepath.Join(dir, t.Filename)
		switch {
		case createSet[t.Filename]:
			if err := create(dest, t); err != nil {
				ui.Warning(stdout, fmt.Sprintf("Falha ao criar %s: %v", t.Filename, err))
			} else {
				ui.Info(stdout, "Criado: "+t.Filename)
			}
		case removeSet[t.Filename]:
			if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
				ui.Warning(stdout, fmt.Sprintf("Falha ao remover %s: %v", t.Filename, err))
			} else {
				ui.Info(stdout, "Removido: "+t.Filename)
			}
		}
	}
	return nil
}

func create(dest string, t Template) error {
	ext := filepath.Ext(t.Filename)
	switch ext {
	case ".docx":
		return createDocx(dest)
	case ".xlsx":
		return createXlsx(dest)
	case ".pptx":
		return createPptx(dest)
	case ".odt":
		return createOdt(dest)
	case ".ods":
		return createOds(dest)
	default:
		return createText(dest, textContent(ext))
	}
}

func createZip(dest string, entries map[string]string) error {
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	w := zip.NewWriter(f)
	for name, content := range entries {
		fw, err := w.Create(name)
		if err != nil {
			_ = w.Close()
			_ = f.Close()
			_ = os.Remove(dest)
			return err
		}
		if _, err := io.WriteString(fw, content); err != nil {
			_ = w.Close()
			_ = f.Close()
			_ = os.Remove(dest)
			return err
		}
	}
	if err := w.Close(); err != nil {
		_ = f.Close()
		_ = os.Remove(dest)
		return err
	}
	return f.Close()
}

func createDocx(dest string) error {
	return createZip(dest, map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`,
		"_rels/.rels":         `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`,
		"word/document.xml":   `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p/></w:body></w:document>`,
	})
}

func createXlsx(dest string) error {
	return createZip(dest, map[string]string{
		"[Content_Types].xml":          `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`,
		"_rels/.rels":                  `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`,
		"xl/workbook.xml":              `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Sheet1" sheetId="1" r:id="rId1"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels":   `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`,
		"xl/worksheets/sheet1.xml":     `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData/></worksheet>`,
	})
}

func createPptx(dest string) error {
	return createZip(dest, map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/></Types>`,
		"_rels/.rels":         `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/></Relationships>`,
		"ppt/presentation.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><p:sldMasterIdLst/><p:sldSz cx="9144000" cy="6858000"/><p:notesSz cx="6858000" cy="9144000"/></p:presentation>`,
	})
}

func createOdt(dest string) error {
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	w := zip.NewWriter(f)

	// mimetype must be first and uncompressed
	mw, err := w.CreateHeader(&zip.FileHeader{Name: "mimetype", Method: zip.Store})
	if err != nil {
		_ = w.Close()
		_ = f.Close()
		_ = os.Remove(dest)
		return err
	}
	if _, err := io.WriteString(mw, "application/vnd.oasis.opendocument.text"); err != nil {
		_ = w.Close()
		_ = f.Close()
		_ = os.Remove(dest)
		return err
	}

	entries := map[string]string{
		"META-INF/manifest.xml": `<?xml version="1.0" encoding="UTF-8"?><manifest:manifest xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0" manifest:version="1.2"><manifest:file-entry manifest:full-path="/" manifest:media-type="application/vnd.oasis.opendocument.text"/><manifest:file-entry manifest:full-path="content.xml" manifest:media-type="text/xml"/></manifest:manifest>`,
		"content.xml":           `<?xml version="1.0" encoding="UTF-8"?><office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0" office:version="1.2"><office:body><office:text><text:p/></office:text></office:body></office:document-content>`,
	}
	for name, content := range entries {
		fw, err := w.Create(name)
		if err != nil {
			_ = w.Close()
			_ = f.Close()
			_ = os.Remove(dest)
			return err
		}
		if _, err := io.WriteString(fw, content); err != nil {
			_ = w.Close()
			_ = f.Close()
			_ = os.Remove(dest)
			return err
		}
	}
	if err := w.Close(); err != nil {
		_ = f.Close()
		_ = os.Remove(dest)
		return err
	}
	return f.Close()
}

func createOds(dest string) error {
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	w := zip.NewWriter(f)

	mw, err := w.CreateHeader(&zip.FileHeader{Name: "mimetype", Method: zip.Store})
	if err != nil {
		_ = w.Close()
		_ = f.Close()
		_ = os.Remove(dest)
		return err
	}
	if _, err := io.WriteString(mw, "application/vnd.oasis.opendocument.spreadsheet"); err != nil {
		_ = w.Close()
		_ = f.Close()
		_ = os.Remove(dest)
		return err
	}

	entries := map[string]string{
		"META-INF/manifest.xml": `<?xml version="1.0" encoding="UTF-8"?><manifest:manifest xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0" manifest:version="1.2"><manifest:file-entry manifest:full-path="/" manifest:media-type="application/vnd.oasis.opendocument.spreadsheet"/><manifest:file-entry manifest:full-path="content.xml" manifest:media-type="text/xml"/></manifest:manifest>`,
		"content.xml":           `<?xml version="1.0" encoding="UTF-8"?><office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0" office:version="1.2"><office:body><office:spreadsheet><table:table table:name="Planilha1"><table:table-row><table:table-cell/></table:table-row></table:table></office:spreadsheet></office:body></office:document-content>`,
	}
	for name, content := range entries {
		fw, err := w.Create(name)
		if err != nil {
			_ = w.Close()
			_ = f.Close()
			_ = os.Remove(dest)
			return err
		}
		if _, err := io.WriteString(fw, content); err != nil {
			_ = w.Close()
			_ = f.Close()
			_ = os.Remove(dest)
			return err
		}
	}
	if err := w.Close(); err != nil {
		_ = f.Close()
		_ = os.Remove(dest)
		return err
	}
	return f.Close()
}

func createText(dest, content string) error {
	perm := os.FileMode(0o644)
	ext := filepath.Ext(dest)
	if ext == ".sh" || ext == ".py" {
		perm = 0o755
	}
	return os.WriteFile(dest, []byte(content), perm)
}

func textContent(ext string) string {
	switch ext {
	case ".txt":
		return ""
	case ".html":
		return "<!DOCTYPE html>\n<html lang=\"pt-BR\">\n<head>\n    <meta charset=\"UTF-8\">\n    <title></title>\n</head>\n<body>\n</body>\n</html>\n"
	case ".css":
		return "/* ============================================================\n   Stylesheet\n   ============================================================ */\n\n*, *::before, *::after {\n    box-sizing: border-box;\n    margin: 0;\n    padding: 0;\n}\n"
	case ".js":
		return "'use strict';\n"
	case ".py":
		return "#!/usr/bin/env python3\n\n\ndef main():\n    pass\n\n\nif __name__ == '__main__':\n    main()\n"
	case ".php":
		return "<?php\n\ndeclare(strict_types=1);\n"
	case ".sh":
		return "#!/usr/bin/env bash\nset -euo pipefail\n"
	}
	return ""
}

