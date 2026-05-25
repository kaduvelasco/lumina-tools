# Changelog

Todas as mudanças notáveis deste projeto serão documentadas aqui.

O formato segue o padrão [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/) e o projeto adota [Versionamento Semântico](https://semver.org/lang/pt-BR/).

---

## [1.0.0] — 2026-05-25

### Adicionado

#### Gerenciamento Linux (`lumina system`)
- Pós-instalação automatizada para Linux Mint 22.3, ZorinOS 18.1, Ubuntu 26.04 e Fedora 44
- Gerenciamento de fontes: JetBrains Mono, Carlito, Caladea, Noto e mais (seleção múltipla)
- Criação de templates de arquivos em branco (Office, LibreOffice, código) em `~/Templates`
- Catálogo Flatpak com 28 aplicativos: instalação e desinstalação via seleção múltipla
- Atualização completa do sistema: apt/dnf/pacman, Snap e Flatpak em uma única etapa
- Instalação do Ulauncher com temas libadwaita (famílias Debian e Fedora)

#### DevStack (`lumina stack`)
- Configuração completa de stack Docker: pré-requisitos, Docker Engine, workspace e docker-compose
- Gerenciamento do ciclo de vida: iniciar, finalizar, logs em tempo real e monitoramento de recursos
- Ambiente PHP multi-versão + Nginx + MariaDB
- Exibição de credenciais de conexão MariaDB
- Correção de permissões do workspace

#### DevStuff (`lumina dev`)
- Gerenciamento de CLIs LLM: Claude Code, Gemini CLI, Codex CLI, OpenCode (seleção múltipla)
- Gerenciamento de IDEs: Zed, Windsurf, VS Code, VSCodium (seleção múltipla)
- Gerenciamento de terminais: Kitty, Alacritty, Black Box (seleção múltipla)
- Catálogo de servidores MCP embutido via YAML: instalação e desinstalação
- Atualização em lote de todas as ferramentas de desenvolvimento instaladas
- Instalação de pré-requisitos: git, libsecret, gnome-keyring

#### DevManager
- Geração de contexto AI: `CLAUDE.md`, `GEMINI.md`, `AGENTS.md`, `.windsurfrules`, `.cursorrules` e `.aiexclude`
- Geração de `.gitignore` com detecção automática de stack via pasta `.instructions/` (Go, Shell, PHP, Node.js, Python, Ruby, Rust, Java); fallback genérico quando a pasta não existe (`lumina gitignore`)
- Operações de banco de dados MariaDB via `docker exec`: backup, restauração, remoção de banco, otimização de tabelas e ajuste de configuração para Moodle
- Gerenciamento de identidade Git: configuração global, init, clone e aplicação de credenciais

#### TUI Interativa
- Interface Bubble Tea navegável com teclado (`↑↓`, `jk`, `Enter`, `Espaço`, `Esc`, `q`)
- Seis temas visuais com preview ao vivo: Lumina, Claro, Dracula, Nord, Tokyo Night, Gruvbox
- Alternância de tema a qualquer momento com a tecla `t`

#### CLI Completa
- Todos os recursos da TUI disponíveis via linha de comando (`lumina <subcomando>`)
- Auto-atualização via GitHub Releases (`lumina self-update`)
- Auto-desinstalação (`lumina self-uninstall`)
- Configuração persistente em `~/.lumina/config.yaml` com suporte a `workspace_path`, `docker_compose_dir`, `theme` e `flatpak_scope`
- Escopo Flatpak configurável: `system` (padrão) ou `user`, evitando ambiguidade quando o Flathub existe em múltiplas instalações (`lumina set flatpak user|system`)
- Completions de shell para Bash e Zsh com suporte a todos os subcomandos e argumentos

---

Made with ❤️ and AI by [Kadu Velasco](https://github.com/kaduvelasco)
