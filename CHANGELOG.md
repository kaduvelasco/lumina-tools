# Changelog

Todas as mudanças notáveis deste projeto serão documentadas aqui.

O formato segue o padrão [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/) e o projeto adota [Versionamento Semântico](https://semver.org/lang/pt-BR/).

---

## [1.0.4] — em desenvolvimento

### Adicionado

#### Fontes — JetBrains Mono Nerd Font (`lumina system fonts`)
- JetBrains Mono NF adicionado ao catálogo de fontes
- Instalação via `https://github.com/ryanoasis/nerd-fonts/releases/latest/download/JetBrainsMono.tar.xz`
- Remoção via glob `JetBrainsMonoNerd*.ttf` em `~/.local/share/fonts/`
- Instalador genérico `installFromURL` suporta arquivos `.zip` e `.tar.xz` — substitui o instalador dedicado da JetBrains Mono

#### Go — instalação e atualização automatizada (`lumina dev go`)
- Nova opção "Gerenciar Go" adicionada ao menu DevStuff > Gerenciar Ferramentas (segundo item, após Instalar Pré-requisitos)
- Se Go já estiver instalado em `/usr/local/go`, exibe a versão atual e pergunta se deseja atualizar
- Consulta a versão estável mais recente via `https://go.dev/dl/?mode=json`
- Se já estiver na versão mais recente, informa e encerra sem alterações
- Download via `curl` de `https://dl.google.com/go/<versão>.linux-amd64.tar.gz` para diretório temporário
- Extrai o tarball para `/usr/local/go-staging` antes de remover a instalação anterior; substituição é atômica via `mv` dentro de `/usr/local`
- Garante `export PATH=$PATH:/usr/local/go/bin` em `~/.bashrc` se ainda não estiver presente
- Arquivos temporários removidos automaticamente ao final da instalação

#### Node.js via nvm — pré-requisito do DevStuff (`lumina dev pre`)
- Instalação do Node.js LTS via nvm adicionada à etapa de pré-requisitos do DevStuff
- Verifica se Node.js já está disponível (inclusive via nvm existente) antes de instalar
- Instala o nvm via script oficial se ainda não estiver presente; caso contrário apenas executa `nvm install --lts` + `nvm use --lts`
- Exibe aviso de reinício do terminal quando o nvm é instalado pela primeira vez

### Removido

#### Pós-instalação — fastfetch
- `fastfetch` removido da lista de pacotes de pós-instalação de todas as distribuições (Fedora, Ubuntu, Mint, ZorinOS)

#### GNOME — Temas GTK
- WhiteSur removido do catálogo de temas GTK

#### GNOME — Cursores
- Oreo removido do catálogo de cursores

### Corrigido

#### Servidores MCP — instalação e desinstalação sem sudo (`lumina dev mcp`)
- `npmInstallMCP` atualizado para verificar disponibilidade do npm antes de instalar e emitir mensagem clara quando Node.js não está presente
- `npmUninstallMCP` adicionado: script bash com source do `nvm.sh` sem sudo — corrige falha ao desinstalar servidores MCP em sistemas com Node.js via nvm
- `uninstall.go` e `select.go` reescritos para usar as funções `npmInstallMCP`/`npmUninstallMCP`; removidos `RequiresSudo: true`, `env PATH=...` e dependência de `os.Getenv`

#### CLIs LLM — desinstalação sem sudo (`lumina dev llm`)
- `npmUninstall` reescrito para usar script bash com source do `nvm.sh` e sem sudo — consistente com `npmInstall`; corrige falha ao desinstalar Codex e OpenCode em sistemas com Node.js via nvm
- `rm -f` de Claude Code e Antigravity CLI removido de `RequiresSudo: true` — ambos instalam em `~/.local/bin` (diretório do usuário), sudo desnecessário

#### Flatpak — ESC sequences na saída (`lumina system apps`, `lumina system gnome`)
- `TERM=dumb` adicionado às chamadas `flatpak install` em `system/apps/install.go` e `gnome/prereqs.go` — corrige sequências como `[24;52R` aparecendo durante a instalação de aplicativos e extensões GNOME
- `TERM=dumb` adicionado à chamada `flatpak uninstall` em `system/apps/uninstall.go` pelo mesmo motivo

#### Stack — Docker sem buildx (`lumina stack config docker`)
- `docker-buildx` adicionado à instalação via apt-get nas distribuições Debian/Ubuntu/Mint/Zorin — corrige o erro "Docker Compose is configured to build using buildx, but buildx isn't installed" na primeira inicialização da stack
- `docker-buildx-plugin` adicionado à instalação via dnf no Fedora pelo mesmo motivo

#### Go — janela destrutiva na instalação eliminada (`lumina dev go`)
- Extração do tarball agora ocorre para `/usr/local/go-staging` antes de remover a instalação anterior — se a extração falhar, o Go existente permanece intacto
- Substituição é atômica: `rm -rf /usr/local/go` só ocorre após extração bem-sucedida, seguido de `mv /usr/local/go-staging /usr/local/go` no mesmo filesystem

#### Contexto AI — referências quebradas em update mode com todos os modelos desmarcados (`lumina manager ai`)
- `generateSharedFiles` agora é sempre executada em modo atualização, mesmo quando todos os modelos são desmarcados — evita que `CLAUDE.md`, `GEMINI.md`, `AGENTS.md` e demais arquivos fiquem com `@-references` apontando para arquivos de instrução já removidos

#### CLIs LLM — falha silenciosa na desinstalação do Claude Code (`lumina dev llm`)
- Erro de `rm -f` na remoção do binário do Claude Code não é mais descartado com `_ =` — comportamento agora consistente com o caso Antigravity CLI na mesma função; falha de permissão é reportada ao usuário

#### Pré-requisitos — mensagem de reinício enganosa em falha de instalação do Node.js (`lumina dev pre`)
- Aviso "Reinicie o terminal para ativar o nvm" agora só é exibido quando a instalação do Node.js é bem-sucedida — evita instruir o usuário a reiniciar para uma ferramenta que nunca chegou a ser instalada

#### Fontes — remoção via glob e instalação com URL vazia (`lumina system fonts`)
- Glob de remoção (`RemoveGlob`) agora é passado via variável de ambiente `LUMINA_GLOB` em vez de interpolado com `%s` na string do script — elimina risco de injeção de shell para entradas futuras do catálogo com metacaracteres
- Guard adicionado em `install()`: retorna erro descritivo quando `AptPkg` e `URL` são ambos vazios, em vez de passar URL vazia para o `curl`

#### Go — requisição HTTP podia bloquear indefinidamente (`lumina dev go`)
- `http.DefaultClient` substituído por cliente local com `Timeout: 30 * time.Second` na consulta da versão estável — evita travamento da TUI em ambientes com conectividade instável

### Alterado

#### Template de instruções AI — melhorias de comportamento (`manager/ai/templates/BASIC.md`)
- Nova seção `Communication` adicionada: resposta concisa, sem preâmbulo antes de agir, sem resumo pós-ação, e tratamento diferenciado para perguntas exploratórias (recomendar + tradeoff antes de implementar)
- `Agent Behavior`: novo bullet — preferir editar arquivos existentes a criar novos; criar apenas quando a tarefa exigir explicitamente
- `Agent Behavior`: dois bullets redundantes sobre refatoração fundidos em um só
- `Code Quality`: regra de comentários adicionada — sem comentários por padrão; comentar apenas o WHY quando não é óbvio pelo código
- `General Principles`: seção removida — todo conteúdo era duplicata de `Agent Behavior`
- `Agent Behavior`: "Sempre pergunte em caso de dúvidas" adicionado como primeiro bullet, com linguagem direta e sem condicional

#### Criar Contexto AI — detecção de execução anterior (`lumina manager ai`)
- Ao executar em um diretório onde o contexto já foi gerado, exibe "Deseja atualizar?" antes do multiselect
- Em modo atualização: todos os arquivos selecionados são regerados sem pedir confirmação por arquivo; modelos desmarcados têm seus arquivos de instrução removidos
- Em modo fresh: comportamento anterior preservado — pergunta "Sobrescrever?" individualmente para cada arquivo existente
- `writeFile`, `writeInstruction` e `generateSharedFiles` recebem parâmetro `overwrite bool`; `removeInstruction` extraída como função dedicada
- Testes do pacote `manager/ai` atualizados para refletir as novas assinaturas

#### Node.js — separação de responsabilidades
- Instalação automática do Node.js removida do fluxo de gerenciamento de CLIs LLM (`lumina dev llm`)
- `npmInstall` falha com mensagem clara ("Execute DevStuff → Instalar Pré-requisitos primeiro") quando o Node.js não está disponível

#### npm global — helper compartilhado entre `llm` e `mcp`
- `RunNPMGlobal(ctx, exe, stdout, action, pkg)` extraído para `internal/dev/localbin` — fonte única para install/uninstall de pacotes npm globais via nvm
- Funções duplicadas `npmInstall`, `npmUninstall` (pacote `llm`) e `npmInstallMCP`, `npmUninstallMCP` (pacote `mcp`) removidas; todos os call sites em `llm`, `mcp/install.go`, `mcp/select.go`, `mcp/uninstall.go` e `mcp/update.go` atualizados

#### Go — confirmação de atualização via stdin injetado (`lumina dev go`)
- `Manage` migrado de `exec` para `execInteractive`: aceita `stdin io.Reader` e usa `prompt.ReadLineFrom(stdin)` em vez de `prompt.ReadLine()` — consistente com `GenerateContext` e demais funções interativas, e testável sem terminal real

#### Criar Contexto AI — fluxo de fresh/update unificado (`lumina manager ai`)
- Branches `if !update / else` em `GenerateContext` substituídos por fluxo único parametrizado com `toWrite []Model`; `update` passa diretamente como `overwrite` para os helpers — elimina duplicação de ~30 linhas e garante que novos passos precisem ser adicionados em um único lugar

---

## [1.0.3] — 2026-05-29

### Adicionado

#### Starship Prompt (`lumina dev term`)
- Starship adicionado ao catálogo de terminais como opção selecionável
- Instalação via script oficial para `~/.local/bin`; seleção interativa de preset (gruvbox-rainbow, tokyo-night, pastel-powerline, pure-preset) na primeira instalação
- Configuração automática em todos os shells presentes (bash, zsh, fish) com hook de inicialização
- Desinstalação remove binário, `~/.config/starship.toml` e os hooks de todos os shells configurados
- Atualização via reinvocação do script oficial (preset e config preservados)

#### Aplicar tema no Flatpak (`lumina system gnome flatpak`)
- Novo item "Aplicar tema no Flatpak" no submenu de Customização GNOME
- Detecta automaticamente os temas GTK instalados pelo Lumina e apresenta single-select
- Aplica `flatpak override --user --env=GTK_THEME=<tema>` para todos os aplicativos Flatpak do usuário

#### Templates — Markdown
- Formato Markdown (`.md`) adicionado ao catálogo de templates de arquivo
- Arquivos presentes na pasta `~/Templates` mas fora do catálogo Lumina (criados por outros programas, como os padrões do ZorinOS) agora aparecem na listagem de remoção com sufixo `[externo]`

#### WebApps sugeridos (`lumina system apps webapps`)
- Novo item "WebApps sugeridos" no submenu de Aplicativos
- Exibe lista de aplicativos web com URL para abrir no navegador e instalar como PWA
- Lista inicial: WhatsApp (`https://web.whatsapp.com`) e Vectorpea (`https://www.vectorpea.com/`)

### Corrigido

#### Pós-instalação ZorinOS
- `software-properties-common` adicionado aos pacotes e repositórios `universe`/`multiverse` habilitados via `add-apt-repository` antes do `apt-get update` — corrige "impossível encontrar o pacote fastfetch"

#### GNOME — Temas GTK
- Estratégia de instalação corrigida para Tokyonight, Everforest e Gruvbox: executa `install.sh` do subdiretório `themes/` (os arquivos pré-construídos não estão no repositório e precisam ser gerados pelo script)
- Estratégia corrigida para Rose Pine: copia de `gtk3/` (subdiretório com arquivos pré-construídos) em vez do diretório raiz

#### GNOME — Cursores
- Oreo: adicionada etapa de build (`bash build.sh`) antes da cópia — o diretório `dist/` não existe no repositório e precisa ser gerado; `inkscape` e `x11-apps` adicionados aos pré-requisitos GNOME
- Sweet: caminho corrigido de `cursors/` para `kde/cursors/` na branch `nova` (o caminho anterior não existe nessa branch)

#### ESC sequences em saída de apt e Flatpak
- Adicionados `-o Dpkg::Use-Pty=0`, `-o Dpkg::Progress-Fancy=0`, `-o APT::Color=0` e `DEBIAN_FRONTEND=noninteractive` às chamadas apt-get — previne sequências de escape como `^[[34;1R` ao executar pós-instalação ou atualização do sistema
- Adicionado `TERM=dumb` às chamadas flatpak pelo mesmo motivo

#### CLIs LLM — `~/.local/bin` ausente do PATH
- `~/.local/bin` agora é adicionado ao `~/.bashrc` automaticamente após instalar Claude Code ou Antigravity CLI quando o diretório não está no PATH — corrige aviso "~/.local/bin is not in your PATH"

#### Servidores MCP — instalação npm
- Removido `sudo` na instalação e atualização de servidores MCP; npm agora é invocado via script bash que carrega o `nvm` do usuário — corrige "npm: arquivo ou diretório inexistente" em sistemas com Node.js instalado via nvm

#### Help viewer — carregamento lento
- `glamour.WithAutoStyle()` substituído por `glamour.WithStandardStyle()` com estilo derivado do tema configurado pelo usuário — elimina a query bloqueante de detecção de cor do terminal que causava lentidão ao abrir a ajuda

### Alterado

#### Pré-requisitos de desenvolvimento
- `lumina dev pre` garante que `~/.local/bin` está no PATH (adiciona ao `~/.bashrc` se ausente) e instrui o usuário a reiniciar o terminal ao concluir

### Refatorado

- `internal/dev/localbin`: novo pacote compartilhado com `EnsureInPath()`, extraído de `llm/install.go` e reutilizado em `dev/depends` e `dev/llm`

---

## [1.0.2] — 2026-05-28

### Adicionado

#### TUI — Menus com descrições (`bubbles/list`)
- Todos os itens de menu exibem uma descrição da ação ao serem destacados, utilizando o componente `bubbles/list` com `NewDefaultDelegate()`
- Delegate estilizado com as cores do tema ativo: título selecionado, descrição, normal e dimmed
- Itens de submenu indicados visualmente com `›`

#### Desinstalar Ulauncher
- Nova opção no submenu de Aplicativos > Desinstalar (TUI) e via CLI (`lumina system ulauncher uninstall`)
- Remove o pacote via apt/dnf, desfaz o repositório PPA (Debian) e remove o diretório de temas do usuário

#### Ajuda interativa com rolagem (Glamour + viewport)
- Help (`Configurações Lumina › Ajuda`) reescrito como visualizador scrollável com Markdown renderizado via Glamour
- Borda arredondada na cor primária do tema; navegação com `↑↓/jk/PgUp/PgDn`, fechado com `q/esc`
- Conteúdo completo: atalhos TUI, todos os comandos CLI por seção e referência de configuração

#### Stack — Pré-requisitos unificados
- Nova função `SetupPrereqs` (`stack/config/prereqs.go`) que combina instalação de pacotes base e Docker Engine em uma única etapa na TUI
- Verificação se o Docker já está instalado antes de qualquer tentativa de reinstalação

### Alterado

#### Reorganização dos menus TUI
- **Gerenciamento Linux:** novo submenu "Aplicativos" com opções Instalar e Desinstalar (Flatpak + Ulauncher)
- **DevStuff:** agrupa "Criar Stack de Desenvolvimento" e "Gerenciar Ferramentas de Desenvolvimento"
  - Criar Stack: pré-requisitos unificados (pacotes base + Docker numa etapa) + Workspace + docker-compose
  - Gerenciar Ferramentas: CLIs LLM, IDEs, terminais, MCP e atualização em lote
- **DevManager:** submenu "Gerenciar Stack" (Iniciar, Finalizar, Logs, Status, Dados DB, Permissões)

#### Pós-instalação ZorinOS
- Removidos: `gnome-tweaks`, `gparted`, instalação manual do Flathub e VLC via Flatpak (ZorinOS já inclui Flatpak/Flathub nativamente)

#### GitHub Actions — Node.js 24
- `actions/checkout@v4` → `@v6`, `actions/setup-go@v5` → `@v6`, `softprops/action-gh-release@v2` → `@v3`
- `FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true` removido (não é mais necessário)

#### Stack — guard de Docker no Compose
- `lumina stack config stack` verifica se o Docker está instalado antes de gerar os arquivos; exibe instrução para executar "Instalar Pré-requisitos" caso o Docker esteja ausente

### Dependências
- Adicionado `github.com/charmbracelet/glamour v1.0.0` (renderização Markdown no terminal)

---

## [1.0.1] — 2026-05-28

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
