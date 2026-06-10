package repo

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

const conductEN = `# Contributor Covenant — Code of Conduct

## Our Pledge

We as members, contributors, and leaders pledge to make participation in our community a harassment-free experience for everyone, regardless of age, body size, visible or invisible disability, ethnicity, sex characteristics, gender identity and expression, level of experience, education, socio-economic status, nationality, personal appearance, race, religion, or sexual identity and orientation.

We pledge to act and interact in ways that contribute to an open, welcoming, diverse, inclusive, and healthy community.

## Our Standards

Examples of behavior that contributes to a positive environment for our community include:

- Demonstrating empathy and kindness toward other people
- Being respectful of differing opinions, viewpoints, and experiences
- Giving and gracefully accepting constructive feedback
- Accepting responsibility and apologizing to those affected by our mistakes, and learning from the experience
- Focusing on what is best not just for us as individuals, but for the overall community

Examples of unacceptable behavior include:

- The use of sexualized language or imagery, and sexual attention or advances of any kind
- Trolling, insulting or derogatory comments, and personal or political attacks
- Public or private harassment
- Publishing others' private information, such as a physical or email address, without their explicit permission
- Other conduct which could reasonably be considered inappropriate in a professional setting

## Enforcement Responsibilities

Community leaders are responsible for clarifying and enforcing our standards of acceptable behavior and will take appropriate and fair corrective action in response to any behavior that they deem inappropriate, threatening, offensive, or harmful.

Community leaders have the right and responsibility to remove, edit, or reject comments, commits, code, wiki edits, issues, and other contributions that are not aligned to this Code of Conduct, and will communicate reasons for moderation decisions when appropriate.

## Scope

This Code of Conduct applies within all community spaces, and also applies when an individual is officially representing the community in public spaces. Examples of representing our community include using an official e-mail address, posting via an official social media account, or acting as an appointed representative at an online or offline event.

## Enforcement

Instances of abusive, harassing, or otherwise unacceptable behavior may be reported to the community leaders responsible for enforcement at **kadu.velasco@gmail.com**. All complaints will be reviewed and investigated promptly and fairly.

All community leaders are obligated to respect the privacy and security of the reporter of any incident.

## Enforcement Guidelines

Community leaders will follow these Community Impact Guidelines in determining the consequences for any action they deem in violation of this Code of Conduct:

### 1. Correction

**Community Impact:** Use of inappropriate language or other behavior deemed unprofessional or unwelcome in the community.
**Consequence:** A private, written warning from community leaders, providing clarity around the nature of the violation and an explanation of why the behavior was inappropriate. A public apology may be requested.

### 2. Warning

**Community Impact:** A violation through a single incident or series of actions.
**Consequence:** A warning with consequences for continued behavior. No interaction with the people involved, including unsolicited interaction with those enforcing the Code of Conduct, for a specified period of time. This includes avoiding interaction in community spaces as well as external channels like social media. Violating these terms may lead to a temporary or permanent ban.

### 3. Temporary Ban

**Community Impact:** A serious violation of community standards, including sustained inappropriate behavior.
**Consequence:** A temporary ban from any sort of interaction or public communication with the community for a specified period of time. No public or private interaction with the people involved, including unsolicited interaction with those enforcing the Code of Conduct, is allowed during this period. Violating these terms may lead to a permanent ban.

### 4. Permanent Ban

**Community Impact:** Demonstrating a pattern of violation of community standards, including sustained inappropriate behavior, harassment of an individual, or aggression toward or disparagement of classes of individuals.
**Consequence:** A permanent ban from any sort of public interaction within the community.

## Attribution

This Code of Conduct is adapted from the [Contributor Covenant](https://www.contributor-covenant.org/version/2/1/code_of_conduct.html), version 2.1.
`

const conductPT = `# Contributor Covenant — Código de Conduta

## Nosso Compromisso

Nós, como membros, colaboradores e líderes, nos comprometemos a tornar a participação em nossa comunidade uma experiência livre de assédio para todos, independentemente de idade, porte físico, deficiência visível ou invisível, etnia, características sexuais, identidade e expressão de gênero, nível de experiência, educação, status socioeconômico, nacionalidade, aparência pessoal, raça, religião ou identidade e orientação sexual.

Nos comprometemos a agir e interagir de formas que contribuam para uma comunidade aberta, acolhedora, diversa, inclusiva e saudável.

## Nossos Padrões

Exemplos de comportamento que contribuem para um ambiente positivo em nossa comunidade incluem:

- Demonstrar empatia e gentileza com outras pessoas
- Respeitar opiniões, pontos de vista e experiências diferentes
- Dar e receber feedback construtivo com elegância
- Aceitar responsabilidade e pedir desculpas às pessoas afetadas por nossos erros, aprendendo com a experiência
- Focar no que é melhor não apenas para nós como indivíduos, mas para a comunidade em geral

Exemplos de comportamento inaceitável incluem:

- O uso de linguagem ou imagens sexualizadas, e atenção ou avanços sexuais de qualquer tipo
- Trollagem, comentários insultuosos ou depreciativos, e ataques pessoais ou políticos
- Assédio público ou privado
- Publicar informações privadas de outras pessoas, como endereço físico ou de e-mail, sem permissão explícita
- Outros comportamentos que possam ser razoavelmente considerados inadequados em um ambiente profissional

## Responsabilidades de Aplicação

Os líderes da comunidade são responsáveis por esclarecer e fazer cumprir nossos padrões de comportamento aceitável e tomarão as medidas corretivas adequadas e justas em resposta a qualquer comportamento que considerem inapropriado, ameaçador, ofensivo ou prejudicial.

Os líderes da comunidade têm o direito e a responsabilidade de remover, editar ou rejeitar comentários, commits, código, edições de wiki, issues e outras contribuições que não estejam alinhadas a este Código de Conduta, e comunicarão as razões das decisões de moderação quando apropriado.

## Escopo

Este Código de Conduta se aplica a todos os espaços da comunidade e também quando um indivíduo está representando oficialmente a comunidade em espaços públicos. Exemplos de representação da nossa comunidade incluem o uso de um endereço de e-mail oficial, publicações em conta oficial de mídia social, ou atuação como representante designado em um evento online ou presencial.

## Aplicação

Casos de comportamento abusivo, de assédio ou de outra forma inaceitável podem ser relatados aos líderes da comunidade responsáveis pela aplicação em **kadu.velasco@gmail.com**. Todas as reclamações serão revisadas e investigadas de forma imediata e justa.

Todos os líderes da comunidade são obrigados a respeitar a privacidade e a segurança do denunciante de qualquer incidente.

## Diretrizes de Aplicação

Os líderes da comunidade seguirão estas Diretrizes de Impacto na Comunidade para determinar as consequências de qualquer ação que considerem uma violação deste Código de Conduta:

### 1. Correção

**Impacto na Comunidade:** Uso de linguagem inadequada ou outro comportamento considerado não profissional ou indesejado na comunidade.
**Consequência:** Um aviso privado e escrito dos líderes da comunidade, esclarecendo a natureza da violação e explicando por que o comportamento foi inadequado. Um pedido público de desculpas pode ser solicitado.

### 2. Aviso

**Impacto na Comunidade:** Uma violação por meio de um único incidente ou série de ações.
**Consequência:** Um aviso com consequências para o comportamento continuado. Sem interação com as pessoas envolvidas, incluindo interação não solicitada com aquelas que aplicam o Código de Conduta, por um período de tempo especificado. Isso inclui evitar interações em espaços da comunidade, bem como em canais externos como mídias sociais. A violação desses termos pode levar a um banimento temporário ou permanente.

### 3. Banimento Temporário

**Impacto na Comunidade:** Uma violação grave dos padrões da comunidade, incluindo comportamento inadequado sustentado.
**Consequência:** Um banimento temporário de qualquer tipo de interação ou comunicação pública com a comunidade por um período de tempo especificado. Nenhuma interação pública ou privada com as pessoas envolvidas, incluindo interação não solicitada com aquelas que aplicam o Código de Conduta, é permitida durante este período. A violação desses termos pode levar a um banimento permanente.

### 4. Banimento Permanente

**Impacto na Comunidade:** Demonstrar um padrão de violação dos padrões da comunidade, incluindo comportamento inadequado sustentado, assédio a um indivíduo, ou agressão ou menosprezo a classes de indivíduos.
**Consequência:** Um banimento permanente de qualquer tipo de interação pública dentro da comunidade.

## Atribuição

Este Código de Conduta é adaptado do [Contributor Covenant](https://www.contributor-covenant.org/version/2/1/code_of_conduct.html), versão 2.1.
`

const (
	fileEN = "CODE_OF_CONDUCT.md"
	filePT = "CODIGO_DE_CONDUTA.md"
)

// CreateConduct writes CODE_OF_CONDUCT.md and CODIGO_DE_CONDUTA.md to the
// current directory. If either file already exists the user is prompted for
// confirmation before overwriting.
func CreateConduct(_ context.Context, _ *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Repositórios :: Criar Código de Conduta")
	ui.Info(stdout, "Diretório: "+cwd())

	existing := make([]string, 0, 2)
	for _, f := range []string{fileEN, filePT} {
		if _, err := os.Stat(f); err == nil {
			existing = append(existing, f)
		}
	}

	if len(existing) > 0 {
		ui.Warning(stdout, "Arquivo(s) já existente(s): "+strings.Join(existing, ", "))
		fmt.Fprint(stdout, "Sobrescrever? (s/N): ")
		if confirm := strings.TrimSpace(prompt.ReadLine()); confirm != "s" && confirm != "S" {
			ui.Info(stdout, "Operação cancelada.")
			ui.WaitEnter(stdout)
			return nil
		}
	}

	for _, f := range []struct {
		name    string
		content string
	}{
		{fileEN, conductEN},
		{filePT, conductPT},
	} {
		if err := os.WriteFile(f.name, []byte(f.content), 0644); err != nil {
			ui.Err(stdout, "Falha ao criar "+f.name+": "+err.Error())
			ui.WaitEnter(stdout)
			return fmt.Errorf("criar %s: %w", f.name, err)
		}
		ui.Info(stdout, "Criado: "+f.name)
	}

	ui.Success(stdout, "Código de conduta criado com sucesso.")
	ui.WaitEnter(stdout)
	return nil
}
