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
| Atualizar Sistema | Atualiza pacotes (apt/dnf/pacman), Snap e Flatpak em uma única etapa |
| Ulauncher | Instalar ou desinstalar o Ulauncher com temas libadwaita |
| Linux Toys | Instalar Linux Toys via instalador oficial |
| MegaSync | Instalar o cliente MegaSync — pacote detectado automaticamente pela distro/versão |

### Aplicativos (`lumina apps`)

| Funcionalidade | Descrição |
|---|---|
| Aplicativos Flatpak | Instalar/remover a partir de um catálogo curado de 31 aplicativos |
| WebApps sugeridos | Lista de aplicativos web com URL para abrir no navegador como PWA |

#### O que a pós-instalação configura

Cada script é adaptado à distro alvo e realiza as seguintes etapas:

| Etapa | Mint 22.3 | ZorinOS 18.1 | Ubuntu 26.04 | Fedora 44 |
|---|:---:|:---:|:---:|:---:|
| Remoção seletiva de Snaps (multi-select) | — | — | ✅ | — |
| Atualização completa do sistema | ✅ | ✅ | ✅ | ✅ |
| Codecs multimídia | ✅ | ✅ | ✅ | ✅ |
| Ferramentas essenciais (build, compactação, utilitários) | ✅ | ✅ | ✅ | ✅ |
| Compiladores C/C++ (`gcc`, `gcc-c++`) | — | — | — | ✅ |
| Fontes Microsoft | ✅ | ✅ | ✅ | — |
| Flatpak + Flathub | ✅ | — ¹ | ✅ | — ¹ |
| Swapfile 4 GB (criação e ativação automática) | ✅ | ✅ | ✅ | — |
| Otimização de kernel (sysctl: swappiness, inotify) | ✅ | ✅ | ✅ | ✅ |
| TRIM para SSDs (`fstrim.timer`) | ✅ | ✅ | ✅ | ✅ |
| Aceleração de vídeo por hardware (VA-API — Intel ou AMD) | ✅ | ✅ | ✅ | ✅ |
| Timeshift (snapshots do sistema) | ✅ | — | ✅ | — |
| RPM Fusion (free + non-free) | — | — | — | ✅ |
| Detecção de drivers proprietários | ✅ | — | ✅ | — |
| Google Chrome | ✅ | ✅ | ✅ | ✅ |

> ¹ ZorinOS e Fedora já incluem Flatpak e Flathub pré-configurados — o script não os reinstala.

### Personalização GNOME (`lumina gnome`)

Requer GNOME como desktop ativo. Todas as operações verificam o ambiente antes de executar.

| Funcionalidade | Descrição |
|---|---|
| Pré-requisitos | gnome-tweaks, murrine-engine (por distro), sassc, git e extensões Flatpak |
| Extensões | Lista de extensões recomendadas com links de instalação |
| Temas GTK | 10 temas: Orchis, Nordic, Colloid, Fluent, Tokyonight, Everforest, Rose Pine, Gruvbox, Graphite, Zorin |
| Ícones | 5 pacotes: Gruvbox Plus, Kora, Candy Icons, Flatery, Newaita |
| Cursores | 4 temas: Layan, Sweet, Colloid, Future |
| Flatpak | Aplicar tema GTK a todos os apps Flatpak via `flatpak override --user` |

### DevStack (`lumina stack`)

Ambiente de desenvolvimento PHP com Docker (multi-versão PHP + Nginx + MariaDB).

| Funcionalidade | Descrição |
|---|---|
| Configurar | Docker Engine, workspace e docker-compose |
| Ciclo de vida | Iniciar, finalizar, reiniciar, visualizar logs, monitorar recursos em tempo real |
| Banco de Dados | Exibir credenciais de conexão MariaDB |
| Permissões | Corrigir propriedade e permissões do workspace |

### DevStuff (`lumina dev`)

| Funcionalidade | Descrição |
|---|---|
| Pré-requisitos | Selecionar e instalar: pacotes base, ferramentas DevStuff, GitHub CLI, Docker Engine e Node.js via nvm (multi-seleção) |
| Go | Instalar ou atualizar Go via tarball oficial (`go.dev/dl`) |
| LLMs | Instalar/remover Claude Code, Antigravity CLI, Codex CLI, OpenCode CLI |
| IDEs | Instalar/remover Zed, Windsurf, VS Code, VSCodium, DBeaver CE |
| Terminais | Instalar/remover Kitty, Alacritty, Black Box, GNOME Console, Starship Prompt — instalação integra entradas "Abrir aqui" ao Nautilus, Nemo e Dolphin |
| Servidores MCP | Instalar/remover servidores a partir de catálogo YAML embutido |
| Atualizar Ferramentas | Atualizar todos os CLIs, IDEs e terminais instalados |

### DevManager

| Funcionalidade | Descrição |
|---|---|
| Contexto AI | Gerar/atualizar `CLAUDE.md`, `GEMINI.md`, `AGENTS.md` e arquivos de regras para o projeto — saída agrupada exibida em painel único ao final |
| Limpar Contexto AI | Remover contextos AI — exibe multi-select com todos os arquivos presentes para o usuário escolher o que remover |
| .gitignore | Criar/atualizar `.gitignore` com base na stack detectada em `.instructions/` |
| Banco de Dados | Backup, restaurar, remover, otimizar tabelas e ajustar MariaDB para Moodle |
| Repositórios | Configurar identidade Git global/local, init, clone, aplicar credenciais e criar código de conduta |

### Configurações Lumina

| Funcionalidade | Descrição |
|---|---|
| Configurar | Editar workspace, diretório Docker Compose, tema e escopo Flatpak interativamente |
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
go build -ldflags "-X github.com/kaduvelasco/lumina-tools/internal/version.Version=v2.0.0" -o lumina ./cmd/lumina
sudo mv lumina /usr/local/bin/lumina
```

---

## Uso

### TUI interativa

```bash
lumina
```

A interface usa um layout em dois painéis: submenu à esquerda e painel de conteúdo à direita.

| Tecla | Ação |
|---|---|
| `1` `2` `3` `4` | Trocar seção |
| `↑` `↓` | Navegar no submenu (ou ciclar botões no painel de conteúdo) |
| `Tab` / `Shift+Tab` | Alternar foco entre submenu e conteúdo |
| `Enter` | Executar botão selecionado |
| `Esc` | Voltar o foco ao submenu |
| `t` | Abrir seletor de temas (preview ao vivo) |
| `q` | Abrir confirmação de saída |
| `Ctrl+C` | Sair imediatamente |

---

### Referência de Comandos CLI

#### Comandos principais

```
lumina                   Abre a interface TUI interativa
lumina self-update       Verifica e instala atualizações
lumina self-uninstall    Remove o binário e as configurações
lumina self-config       Configura o Lumina interativamente
lumina version           Exibe a versão instalada
lumina help              Exibe esta referência
```

#### Gerenciamento Linux

```
lumina system pos [mint|zorin|ubuntu|fedora]   Pós-instalação (sem arg abre menu)
lumina system fonts                            Gerenciar fontes (multi-seleção)
lumina system templates                        Gerenciar templates de arquivos
lumina system update                           Atualizar o sistema completo
lumina system ulauncher                        Instalar Ulauncher e temas
lumina system ulauncher uninstall              Desinstalar Ulauncher e remover dados
lumina system toys                             Instalar Linux Toys
lumina system megasync                         Instalar MegaSync (pacote detectado pela distro)
```

#### Aplicativos

```
lumina apps install     Instalar aplicativos Flatpak (multi-seleção)
lumina apps uninstall   Desinstalar aplicativos Flatpak (multi-seleção)
lumina apps web         Lista de WebApps sugeridos
```

#### Personalização GNOME

```
lumina gnome pre        Instalar pré-requisitos GNOME
lumina gnome ext        Exibir extensões recomendadas
lumina gnome themes     Gerenciar temas GTK (multi-seleção)
lumina gnome icons      Gerenciar pacotes de ícones (multi-seleção)
lumina gnome cursor     Gerenciar temas de cursor (multi-seleção)
lumina gnome flatpak    Aplicar tema GTK em apps Flatpak
```

#### DevStack

```
lumina stack config [docker|workspace|stack]   Configurar stack (sem arg abre menu)
lumina stack start                             Iniciar stack de containers
lumina stack end                               Finalizar stack de containers
lumina stack restart                           Reiniciar stack de containers
lumina stack log                               Visualizar logs em tempo real
lumina stack status                            Status e uso de recursos
lumina stack db                                Exibir dados de conexão do banco
lumina stack fix-perm                          Corrigir permissões do workspace
```

#### DevStuff

```
lumina dev pre                Selecionar e instalar pré-requisitos (pacotes base, DevStuff, GitHub CLI, Docker, Node.js)
lumina dev go                 Instalar ou atualizar Go via tarball oficial
lumina dev llm                Gerenciar CLIs LLM (multi-seleção)
lumina dev ide                Gerenciar IDEs (multi-seleção)
lumina dev term               Gerenciar terminais — Kitty, Alacritty, Black Box, GNOME Console, Starship (multi-seleção)
lumina dev mcp                Gerenciar servidores MCP (multi-seleção)
lumina dev update             Atualizar todas as ferramentas de desenvolvimento
lumina dev create-workspace   Criar estrutura de workspace
lumina dev create-stack-php   Criar/atualizar docker-compose da stack PHP
```

#### DevManager

```
lumina ai context                                      Gerar/atualizar contexto AI (multi-seleção)
lumina ai clear                                        Remover todos os contextos AI do diretório atual
lumina db [backup|restore|remove|optimize|moodle]      Gerenciar banco de dados MariaDB
lumina repo [global|init|clone|ident|gitignore|conduct] Gerenciar identidade Git e arquivos de projeto
```

#### Configuração via CLI

```
lumina set workspace <caminho>                                    Define o caminho do workspace
lumina set docker <caminho>                                       Define o diretório do docker-compose
lumina set theme [lumina|light|dracula|nord|tokyo|gruvbox]        Define o tema da TUI
lumina set flatpak [user|system]                                  Define o escopo de instalação Flatpak
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
