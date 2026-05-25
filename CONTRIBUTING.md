# Contribuindo com o Lumina Tools

Obrigado pelo interesse em contribuir! Este documento explica como configurar o ambiente de desenvolvimento, as convenções do projeto e como enviar alterações.

---

## Ambiente de Desenvolvimento

**Requisitos:**
- Go 1.26 ou superior
- Git

**Configuração:**

```bash
git clone https://github.com/kaduvelasco/lumina-tools.git
cd lumina-tools
go mod download
```

**Comandos comuns:**

```bash
# Compilar (modo desenvolvimento)
go build ./cmd/lumina

# Compilar com versão injetada
go build -ldflags "-X github.com/kaduvelasco/lumina-tools/internal/version.Version=v1.0.0" -o lumina ./cmd/lumina

# Executar diretamente
go run ./cmd/lumina [args]

# Testes
go test ./...
go test -race ./...

# Análise estática
go vet ./...
golangci-lint run
```

---

## Estrutura do Projeto

```
lumina-tools/
├── cmd/lumina/         # Ponto de entrada (main.go)
├── internal/
│   ├── app/            # Dispatch CLI — app.Run() e sub-dispatchers
│   ├── tui/            # TUI Bubble Tea (model, menus, estilos, temas)
│   ├── ui/             # Primitivos de terminal (PrintHeader, Info, Err, Success…)
│   ├── executor/       # Único ponto de escalada sudo
│   ├── config/         # ~/.lumina/config.yaml
│   ├── distro/         # Detecção de família de distro
│   ├── prompt/         # Helpers de leitura stdin
│   ├── sets/           # Estruturas de conjunto
│   ├── version/        # String de versão (injetada via -ldflags)
│   ├── selfupdate/     # Auto-atualização via GitHub Releases
│   ├── system/         # Gerenciamento Linux (update, fonts, apps, templates…)
│   ├── stack/          # Ciclo de vida do DevStack Docker
│   ├── dev/            # DevStuff (LLMs, IDEs, terminais, MCP)
│   └── manager/        # DevManager (contexto AI, .gitignore, banco de dados, repositórios)
│       ├── ai/         # Geração de contexto AI (CLAUDE.md, GEMINI.md, AGENTS.md…)
│       ├── gitignore/  # Geração de .gitignore com base na stack (.instructions/)
│       ├── db/         # Operações MariaDB (backup, restore, optimize…)
│       └── repo/       # Identidade Git (global, init, clone, ident)
├── assets/             # Templates e catálogos de referência
├── completions/        # Completions de shell (bash, zsh)
└── install.sh          # Instalador one-line
```

---

## Convenções de Código

- **Idioma:** todo o código, comentários e documentação em inglês; todas as strings visíveis ao usuário em português do Brasil.
- **Tratamento de erros:** sempre encapsule com `fmt.Errorf("contexto: %w", err)`.
- **Sem sudo direto:** todos os comandos privilegiados passam pelo `executor.Executor` com `RequiresSudo: true`.
- **Assinatura de funções de domínio:**
  ```go
  // Sem stdin (funcCmd):
  func DoSomething(ctx context.Context, exe *executor.Executor, stdout io.Writer) error

  // Com stdin (interactiveFuncCmd — multi-select):
  func DoSomething(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error
  ```
- **Padrão de UI:** toda ação de terminal segue a sequência `PrintHeader → Info/Warning → Err ou Success → WaitEnter`.
- **Sem abstrações prematuras:** implemente apenas o necessário.
- **Detecção de distro:** use sempre `distro.Detect()` — nunca crie funções locais de detecção em pacotes de domínio.

---

## Adicionando uma Nova Ação ao Menu

1. Adicione uma constante `actXxx` em `internal/tui/menus.go`.
2. Adicione `{label: "...", action: actXxx}` ao submenu correspondente em `itemsFor()`.
3. Implemente a função de domínio no pacote apropriado.
4. Adicione um `case actXxx:` em `runAction()` em `internal/tui/model.go`.
5. Adicione o subcomando CLI correspondente em `internal/app/app.go`.

---

## Adicionando um Novo Aplicativo Flatpak

Edite `internal/system/apps/catalogue.go` e adicione uma entrada ao slice `Catalogue`:

```go
{Name: "Nome do App", FlatID: "com.exemplo.AppID"},
```

---

## Adicionando um Novo Servidor MCP

Edite `internal/dev/mcp/servers.yaml` e adicione uma entrada:

```yaml
- name: "Nome do Servidor"
  package: "pacote-npm"
  cmd: "binario"
  description: "Descrição curta"
```

Recompile o binário para embutir o catálogo atualizado.

---

## Adicionando um Novo LLM, IDE ou Terminal

Edite o arquivo `catalogue.go` do pacote correspondente em `internal/dev/llm/`, `internal/dev/ide/` ou `internal/dev/terminal/`, e adicione uma entrada ao slice `Catalogue`. Em seguida, adicione o caso de instalação e desinstalação nos respectivos `install.go` e `uninstall.go`.

---

## Enviando Alterações

1. Crie um fork do repositório e uma branch para sua funcionalidade.
2. Implemente as alterações seguindo as convenções acima.
3. Execute `go test ./...` e `go vet ./...` — ambos devem passar sem erros.
4. Abra um Pull Request com uma descrição clara do que foi alterado e por quê.

---

Made with ❤️ and AI by [Kadu Velasco](https://github.com/kaduvelasco)
