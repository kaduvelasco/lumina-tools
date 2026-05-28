# Changelog

Todas as mudanças notáveis deste projeto serão documentadas aqui.

O formato segue o padrão [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/) e o projeto adota [Versionamento Semântico](https://semver.org/lang/pt-BR/).

---

## [1.0.1] — não publicado

### Adicionado

#### Customização GNOME (`lumina system gnome`)

- Verificação automática do ambiente GNOME antes de executar qualquer operação (verifica `XDG_CURRENT_DESKTOP`, `DESKTOP_SESSION`, `GDMSESSION`)
- Instalação de pré-requisitos: `gnome-tweaks`, `gnome-themes-extra`, `murrine-engine` (por família de distro: Debian/Ubuntu/Mint/Zorin vs Fedora vs Arch), `sassc`, `git` e Flatpak extensions `org.gnome.Extensions` e `com.mattjakeman.ExtensionManager`
- Lista de extensões recomendadas com URLs de instalação: Tiling Shell, User Themes, ArcMenu, Dash to Panel
- Gerenciamento de 9 temas GTK via seleção múltipla (itens instalados pré-selecionados): Orchis, WhiteSur, Nordic, Colloid, Fluent, Tokyonight, Everforest, Rose Pine, Gruvbox
  - Temas com script (Orchis, WhiteSur, Colloid, Fluent): clone em diretório temporário + `./install.sh -t all`; variações completas instaladas automaticamente
  - Temas diretos (Nordic, Tokyonight, Everforest, Rose Pine, Gruvbox): `git clone --depth=1` direto em `~/.themes/`
  - WhiteSur: seleção interativa do ícone da barra de título (14 opções via single-select)
  - Após instalar temas: opção de aplicar `GTK_THEME` a todos os apps Flatpak via `flatpak override --user`
- Gerenciamento de 5 pacotes de ícones via seleção múltipla (itens instalados pré-selecionados): Gruvbox Plus, Kora, Candy Icons, Flatery, Newaita — instalados em `~/.local/share/icons/` sem sudo
- Gerenciamento de 5 temas de cursor via seleção múltipla (itens instalados pré-selecionados): Layan, Oreo, Sweet, Colloid, Future — instalados em `~/.local/share/icons/` sem sudo
- CLI: `lumina system gnome <pre|ext|themes|icons|cursors>`

### Corrigido

- `dev/llm`: nome do pacote npm do Codex CLI corrigido de `codex-cli` para `@openai/codex` (instalação e desinstalação)
- `dev/llm`: desinstalação do Claude Code simplificada — removida tentativa de `npm uninstall` que era no-op (Claude Code é instalado via script, não npm); mantido apenas `which claude` + `rm -f`
- `manager/db/remove`: credencial MariaDB agora passa via arquivo temporário com `chmod 0600` e flag `--env-file`, consistente com `backup`, `restore` e `optimize`; eliminada exposição da senha em `/proc/<pid>/environ`
- `system/fonts`: caminhos do diretório de fontes entre aspas duplas no script bash — previne falha quando `$HOME` contém espaços
- `app`: subcomando `lumina set flatpak` ausente das mensagens de erro de uso do `dispatchSet` — adicionado em todos os pontos

### Alterado

- `executor`: adicionada `executor.CurrentUser()` como função pública centralizada (resolve via `SUDO_USER → USER → LOGNAME`)
- `stack/perms`, `stack/config/docker`, `stack/config/workspace`: eliminada duplicação de `currentUser()` — todos usam `executor.CurrentUser()`
- Formatação: `gofmt -w .` aplicado

### Templates de IA (`assets/ai/templates`)

- `BASIC.md`: regras de Code Quality para testes; documentação de placement expandida
- `ONLY-CLAUDE.md`: coluna de API string na tabela de modelos; nota sobre Opus para orquestrador; protocolo de escalação para subagentes
- `ONLY-GEMINI.md`: reescrito com critérios de spawn/não-spawn, tabela de modelos e seção de Escalation Rule
- `instructions/MCP.md`: seção de bootstrap do servidor adicionada (SDK v1.29.0, API `setRequestHandler`); versão explicitada nas dependências
- `instructions/MCP-Migration.md`: removido — SDK v2 não existe no npm (versão estável é 1.29.0)
- `instructions/BASH.md`: dois boilerplates distintos (Lumina Ecosystem e Standalone); exemplos de uso das cores; `store_secret()` corrigido (ShellCheck SC2168 — `local` fora de função); seção de argument parsing adicionada
- `instructions/GOLANG.md`: seção de context/timeout; verificação de status HTTP em `Fetch`; `sync.Pool` corrigido para `buf.String()` com explicação do risco de `buf.Bytes()`; seção de testes com table-driven e `t.Helper()`
- `instructions/MOODLE.md`: seção Hook API vs lib.php com suporte multi-versão (<4.3 e ≥4.3); Privacy API completa (null e full provider); External API; Events; AMD JavaScript sem jQuery
- `instructions/PHP.md`: estrutura de projeto DDD; regras de tipagem (`declare(strict_types=1)` não é padrão); features PHP 8.x; error handling com domain exceptions; segurança; testes com `@dataProvider`; quality tools (`phpcs`, `phpstan`, `phpunit`)

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
