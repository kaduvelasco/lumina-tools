<div align="center">
  <img src="docs/img/logo.png" alt="Lumina Tools" />
</div>

# Lumina Tools

> Binário Go unificado para Linux com TUI interativa e CLI completa — 100% em português do Brasil.

![Version](https://img.shields.io/github/v/release/kaduvelasco/lumina-tools?label=vers%C3%A3o&color=brightgreen)
![Go](https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go)
![Platform](https://img.shields.io/badge/plataforma-Linux-FCC624?logo=linux)
![License](https://img.shields.io/github/license/kaduvelasco/lumina-tools?label=licen%C3%A7a)

---

## Recursos

### Gerenciamento Linux (`lumina system`)

| Funcionalidade | Descrição |
|---|---|
| Pós-instalação | Scripts automatizados para Mint 22.3, ZorinOS 18.1, Ubuntu 26.04 e Fedora 44 |
| Fontes | Instalar/remover JetBrains Mono, Noto, Carlito, Caladea e outros |
| Templates de Arquivos | Criar templates em branco (Office, LibreOffice, código) em `~/Templates` |
| Aplicativos Flatpak | Instalar/remover a partir de um catálogo curado de 28 aplicativos |
| WebApps sugeridos | Lista de aplicativos web com URL para abrir no navegador como PWA |
| Atualizar Sistema | Atualiza pacotes (apt/dnf/pacman), Snap e Flatpak em uma única etapa |
| Ulauncher | Instalar ou desinstalar o Ulauncher com temas libadwaita |
| Customizar GNOME | Temas GTK, ícones, cursores, extensões recomendadas e pré-requisitos |

#### O que a pós-instalação configura

Cada script é adaptado à distro alvo e realiza as seguintes etapas:

| Etapa | Mint 22.3 | ZorinOS 18.1 | Ubuntu 26.04 | Fedora 44 |
|---|:---:|:---:|:---:|:---:|
| Atualização completa do sistema | ✅ | ✅ | ✅ | ✅ |
| Codecs multimídia | ✅ | ✅ | ✅ | ✅ |
| Ferramentas essenciais (build, compactação, utilitários) | ✅ | ✅ | ✅ | ✅ |
| Compiladores C/C++ (`gcc`, `gcc-c++`) | — | — | — | ✅ |
| Fontes Microsoft | ✅ | ✅ | ✅ | — |
| Flatpak + Flathub | ✅ | ✅ | ✅ | ✅ |
| Otimização de kernel (sysctl: swappiness, inotify) | ✅ | ✅ | ✅ | ✅ |
| TRIM para SSDs (`fstrim.timer`) | ✅ | ✅ | ✅ | ✅ |
| Aceleração de vídeo por hardware (VA-API — Intel ou AMD) | ✅ | ✅ | ✅ | ✅ |
| Timeshift (snapshots do sistema) | ✅ | — | ✅ | — |
| RPM Fusion (free + non-free) | — | — | — | ✅ |
| Detecção de drivers proprietários | ✅ | — | ✅ | — |

#### Customização GNOME (`lumina system gnome`)

Requer GNOME como desktop ativo. Todas as operações verificam o ambiente antes de executar.

| Funcionalidade | Descrição |
|---|---|
| Pré-requisitos | gnome-tweaks, murrine-engine (por distro), sassc, git e extensões Flatpak |
| Extensões | Lista de extensões recomendadas com links de instalação |
| Temas GTK | 9 temas: Orchis, WhiteSur, Nordic, Colloid, Fluent, Tokyonight, Everforest, Rose Pine, Gruvbox |
| Ícones | 5 pacotes: Gruvbox Plus, Kora, Candy Icons, Flatery, Newaita |
| Cursores | 5 temas: Layan, Oreo, Sweet, Colloid, Future |
| Flatpak | Aplicar tema GTK a todos os apps Flatpak via `flatpak override --user` |

### DevStack (`lumina stack`)

Ambiente de desenvolvimento PHP com Docker (multi-versão PHP + Nginx + MariaDB).

| Funcionalidade | Descrição |
|---|---|
| Configurar | Pré-requisitos, Docker Engine, workspace e docker-compose |
| Ciclo de vida | Iniciar, finalizar, visualizar logs, monitorar recursos em tempo real |
| Banco de Dados | Exibir credenciais de conexão MariaDB |
| Permissões | Corrigir propriedade e permissões do workspace |

### DevStuff (`lumina dev`)

| Funcionalidade | Descrição |
|---|---|
| LLMs | Instalar/remover Claude Code, Antigravity CLI, Codex CLI, OpenCode CLI |
| IDEs | Instalar/remover Zed, Windsurf, VS Code, VSCodium |
| Terminais | Instalar/remover Kitty, Alacritty, Black Box, Starship Prompt |
| Servidores MCP | Instalar/remover servidores a partir de catálogo YAML embutido |
| Atualizar Ferramentas | Atualizar todos os CLIs, IDEs e terminais instalados |

### DevManager

| Funcionalidade | Descrição |
|---|---|
| Contexto AI | Gerar `CLAUDE.md`, `GEMINI.md`, `AGENTS.md` e arquivos de regras para o projeto |
| .gitignore | Criar/atualizar `.gitignore` com base na stack detectada em `.instructions/` |
| Banco de Dados | Backup, restaurar, remover, otimizar tabelas e ajustar MariaDB para Moodle |
| Repositórios | Configurar identidade Git global/local, init, clone e aplicar credenciais |

### Configurações Lumina

| Funcionalidade | Descrição |
|---|---|
| Atualizar | Verificar e instalar a versão mais recente |
| Desinstalar | Remover o binário e as configurações do sistema |
| Ajuda | Referência completa de comandos com rolagem (Markdown renderizado via Glamour) |

---

## Requisitos

- Linux: Ubuntu 26.04, Linux Mint 22+, ZorinOS 18.1, Pop!_OS 24.04+, Fedora 44+
- Terminal com suporte a 256 cores

---

## Instalação

### Instalador automático (recomendado)

```bash
curl -fsSL https://raw.githubusercontent.com/kaduvelasco/lumina-tools/main/install.sh | bash
```

### Manual

1. Baixe o binário para sua arquitetura na página de [Releases](https://github.com/kaduvelasco/lumina-tools/releases).
2. Torne-o executável e mova para o `$PATH`:

```bash
chmod +x lumina-linux-amd64
sudo mv lumina-linux-amd64 /usr/local/bin/lumina
```

### Compilar do código-fonte

```bash
git clone https://github.com/kaduvelasco/lumina-tools.git
cd lumina-tools
go build -ldflags "-X github.com/kaduvelasco/lumina-tools/internal/version.Version=v1.0.3" -o lumina ./cmd/lumina
sudo mv lumina /usr/local/bin/lumina
```

---

## Uso

### TUI interativa

```bash
lumina
```

Navegue com `↑ ↓` ou `j k`, confirme com `Enter` ou `Espaço`, volte com `Esc`, saia com `q`.
Pressione `t` a qualquer momento para abrir o seletor de temas com preview ao vivo.
Cada item do menu exibe uma descrição da ação ao ser destacado.

---

### Referência de Comandos CLI

#### Comandos principais

```
lumina                   Abre a interface TUI interativa
lumina self-update       Verifica e instala atualizações
lumina self-uninstall    Remove o binário e as configurações
lumina version           Exibe a versão instalada
lumina help              Exibe esta referência
```

#### Gerenciamento Linux

```
lumina system pos [mint|zorin|ubuntu|fedora]   Pós-instalação (sem arg abre menu)
lumina system fonts                            Gerenciar fontes (multi-seleção)
lumina system templates                        Gerenciar templates de arquivos
lumina system apps install                     Instalar aplicativos Flatpak
lumina system apps uninstall                   Desinstalar aplicativos Flatpak
lumina system apps webapps                     Listar WebApps sugeridos
lumina system update                           Atualizar o sistema completo
lumina system ulauncher                        Instalar Ulauncher e temas
lumina system ulauncher uninstall              Desinstalar Ulauncher e remover dados
lumina system gnome pre                        Instalar pré-requisitos GNOME
lumina system gnome ext                        Exibir extensões recomendadas
lumina system gnome themes                     Gerenciar temas GTK (multi-seleção)
lumina system gnome icons                      Gerenciar pacotes de ícones (multi-seleção)
lumina system gnome cursors                    Gerenciar temas de cursor (multi-seleção)
lumina system gnome flatpak                    Aplicar tema GTK em apps Flatpak
```

#### DevStack

```
lumina stack config [pre|docker|workspace|stack]   Configurar stack (sem arg abre menu)
lumina stack start                                 Iniciar stack de containers
lumina stack end                                   Finalizar stack de containers
lumina stack log                                   Visualizar logs em tempo real
lumina stack status                                Status e uso de recursos
lumina stack db                                    Exibir dados de conexão do banco
lumina stack fix-perm                              Corrigir permissões do workspace
```

#### DevStuff

```
lumina dev pre      Instalar pré-requisitos (git, libsecret, gnome-keyring)
lumina dev llm      Gerenciar CLIs LLM (multi-seleção)
lumina dev ide      Gerenciar IDEs (multi-seleção)
lumina dev term     Gerenciar terminais (multi-seleção)
lumina dev mcp      Gerenciar servidores MCP (multi-seleção)
lumina dev update   Atualizar todas as ferramentas de desenvolvimento
```

#### DevManager

```
lumina ai                                  Gerar contexto AI (multi-seleção)
lumina gitignore                           Criar/atualizar .gitignore
lumina db [backup|restore|remove|optimize|moodle]  Gerenciar banco de dados MariaDB
lumina repo [global|init|clone|ident]              Gerenciar identidade Git
```

#### Configuração via CLI

```
lumina set workspace <caminho>                       Define o caminho do workspace
lumina set docker <caminho>                          Define o diretório do docker-compose
lumina set theme [lumina|light|dracula|nord|tokyo|gruvbox]   Define o tema da TUI
lumina set flatpak [user|system]                     Define o escopo de instalação Flatpak
```

---

## Configuração

As configurações são salvas em `~/.lumina/config.yaml`:

```yaml
workspace_path: ~/workspace
docker_compose_dir: ~/workspace/docker
theme: Lumina
flatpak_scope: system
stack:
  php_versions: "8.1 8.2"
  db_user: admin
  db_pass: ""
  db_root_pass: ""
```

| Campo | Descrição |
|---|---|
| `workspace_path` | Raiz do workspace de desenvolvimento |
| `docker_compose_dir` | Diretório onde o `docker-compose.yml` está localizado |
| `theme` | Tema da TUI: `Lumina`, `Claro`, `Dracula`, `Nord`, `Tokyo Night`, `Gruvbox` |
| `flatpak_scope` | Escopo de instalação Flatpak: `system` (padrão) ou `user` |

---

## Catálogo MCP

A lista de servidores MCP fica em `internal/dev/mcp/servers.yaml` (embutida no binário). Para adicionar um servidor, edite o arquivo antes de compilar:

```yaml
servers:
  - name: "Nome do Servidor"
    package: "pacote-npm"
    cmd: "binario"
    description: "Descrição curta"
```

---

## Completions de Shell

**Bash** — adicione ao `~/.bashrc`:
```bash
source /path/to/completions/lumina.bash
```

**Zsh** — copie para um diretório no `$fpath`:
```bash
cp completions/_lumina /usr/local/share/zsh/site-functions/_lumina
```

---

## Contribuindo

Veja [CONTRIBUTING.md](CONTRIBUTING.md) para diretrizes de contribuição.

---

## Licença

Este projeto está licenciado sob a [Licença MIT](LICENSE).

---

Made with ❤️ and AI by [Kadu Velasco](https://github.com/kaduvelasco)
