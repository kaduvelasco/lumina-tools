package repo

import (
	"strings"
	"testing"
)

func TestConductEN_RequiredSections(t *testing.T) {
	for _, section := range []string{
		"# Contributor Covenant",
		"## Our Pledge",
		"## Our Standards",
		"## Enforcement Responsibilities",
		"## Scope",
		"## Enforcement",
		"## Enforcement Guidelines",
		"## Attribution",
		"kadu.velasco@gmail.com",
	} {
		if !strings.Contains(conductEN, section) {
			t.Errorf("conductEN missing section: %q", section)
		}
	}
}

func TestConductPT_RequiredSections(t *testing.T) {
	for _, section := range []string{
		"# Contributor Covenant",
		"## Nosso Compromisso",
		"## Nossos Padrões",
		"## Responsabilidades de Aplicação",
		"## Escopo",
		"## Aplicação",
		"## Diretrizes de Aplicação",
		"## Atribuição",
		"kadu.velasco@gmail.com",
	} {
		if !strings.Contains(conductPT, section) {
			t.Errorf("conductPT missing section: %q", section)
		}
	}
}

func TestConductFileNames(t *testing.T) {
	if fileEN != "CODE_OF_CONDUCT.md" {
		t.Errorf("fileEN = %q, want CODE_OF_CONDUCT.md", fileEN)
	}
	if filePT != "CODIGO_DE_CONDUTA.md" {
		t.Errorf("filePT = %q, want CODIGO_DE_CONDUTA.md", filePT)
	}
}
