# Changelog

Todas as mudanças notáveis deste projeto serão documentadas aqui.

O formato segue o padrão [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/) e o projeto adota [Versionamento Semântico](https://semver.org/lang/pt-BR/).

---

## [2.2.2] — 2026-06-18

### Corrigido

#### Stack PHP — Dockerfile: download do PHPUnit retornando 404 ao iniciar a stack
- Mesma causa raiz do fix de PHPCS da v2.2.1: `phar.phpunit.de` parou de publicar os sidecars `phpunit-${VER}.phar.sha256` (apenas os `.phar` continuam disponíveis) — o `curl` do hash retornava 404, interrompendo o build da imagem PHP (`docker compose up` falhava com `exit code: 22` na etapa de instalação do PHPUnit)
- URLs "rolling" por major version (`phpunit-10.phar`, `phpunit-11.phar`, `phpunit-12.phar`) substituídas por versões de patch fixadas: PHP 8.1 → PHPUnit `10.5.55`, PHP 8.2/8.3 → `11.5.39`, PHP 8.4+ → `12.5.30`. Versões "rolling" foram evitadas porque o conteúdo do arquivo muda silenciosamente a cada patch release, o que invalidaria qualquer hash fixado no futuro
- Verificação de integridade mantida via hash SHA256 fixado inline (`echo "<hash>  phpunit.phar" | sha256sum -c -`), mesmo padrão adotado para o PHPCS
- **Stacks existentes precisam reconstruir as imagens** para receber a correção: `docker compose build --no-cache && docker compose up -d --force-recreate`

---

## [2.2.1] — 2026-06-18

### Corrigido

#### Stack PHP — Dockerfile: download do PHP_CodeSniffer retornando 404 ao iniciar a stack
- `squizlabs.github.io/PHP_CodeSniffer` parou de publicar os sidecars `phpcs.phar.sha256`/`phpcbf.phar.sha256` — o `curl` desses arquivos retornava 404, interrompendo o build da imagem PHP (`docker compose up` falhava com `exit code: 22` na etapa de instalação do PHPCS)
- Origem do download trocada para os GitHub Releases de `PHPCSStandards/PHP_CodeSniffer` (fork ativo da ferramenta), com versão fixada em `3.13.5` (série 3.x, compatível com `moodlehq/moodle-cs` e a maioria dos standards de terceiros — PHPCS 4.x ainda não é amplamente suportado por sniffs externos)
- Verificação de integridade mantida via hash SHA256 fixado inline (`echo "<hash>  arquivo" | sha256sum -c -`), já que não há mais sidecar `.sha256` oficial disponível para os PHARs do PHPCS
- **Stacks existentes precisam reconstruir as imagens** para receber a correção: `docker compose build --no-cache && docker compose up -d --force-recreate`

---

## [2.2.0] — 2026-06-17

### Adicionado

#### Stack PHP — ferramentas de qualidade e wrappers de linha de comando (`lumina dev create-stack-php`)
- **PHP_CodeSniffer** (`phpcs` + `phpcbf`) instalado no container via PHAR oficial durante o build da imagem Docker — disponível em todas as versões PHP selecionadas
- **PHPUnit** instalado no container com versão mapeada à versão do PHP: `8.1` → v10, `8.2`/`8.3` → v11, `8.4+` → v12
- **Wrapper scripts** gerados automaticamente em `~/.local/bin/` ao criar a stack:
  - Genéricos (apontam para o primeiro container criado): `php`, `phpcs`, `phpcbf`, `phpunit`, `composer`
  - Por versão: `php81`, `php82`, ..., `phpcs82`, `phpcbf83`, `phpunit84`, `composer81` etc.
  - Tradução automática de paths: argumentos sob `workspace/www/html/` são convertidos para `/var/www/html/` dentro do container
  - Detecção de TTY: usa `docker exec -it` quando conectado a terminal, `docker exec -i` em pipe — transparente para o chamador
- `~/.local/bin/` adicionado automaticamente ao PATH via `localbin.EnsureInPath()` ao final da geração

#### Personalizar Linux — catálogos de temas GTK expandidos (`lumina theme gnome` / `lumina theme cinnamon`)
- **Catppuccin** expandido de 1 para 4 variantes: Mocha, Latte, Frappé, Macchiato — pergunta borda e botões macOS na instalação
- **Everforest** expandido de 1 para 3 variantes explícitas: Hard, Medium, Soft — pergunta borda e botões macOS
- **Material GTK** adicionado ao catálogo (GNOME e Cinnamon) com 4 variantes: Lighter, Oceanic, Palenight, Darker — pergunta borda e botões macOS
- **Nightfox** adicionado ao catálogo (GNOME e Cinnamon) com 5 variantes: Nightfox, Duskfox, Nordfox, Terafox, Carbonfox — pergunta borda e botões macOS
- **Graphite** expandido: instala Normal e Nord em uma única operação (`tweak_variants: [[], ["nord"]]`); pergunta borda rimless/com borda na instalação; no Cinnamon, `--tweaks normal` aplicado automaticamente via `fixed_tweaks` (sidebar Nautilus com ícones coloridos)
- Catálogo GNOME passa de 11 para 26 variantes; catálogo Cinnamon passa de 5 para 18 variantes

#### Personalizar Linux — seleção de tema Flatpak ampliada (`lumina theme flatpak`)
- Seletor de tema Flatpak passa a listar todos os diretórios presentes em `~/.themes/` (via `os.ReadDir`), cobrindo todas as variantes instaladas; antes usava o campo estático `flatpak_name` por entrada YAML, o que excluía variantes extras (ex.: Graphite-Dark-Nord, Catppuccin-Light)

#### Gerenciar Contextos IA — Docker nos templates PHP e Moodle (`lumina ai context`)
- **`instructions/MOODLE.md`**: nova seção "Docker Development Environment" com tabela de path mapping (`workspace/www/html` ↔ `/var/www/html`), lista completa dos wrappers disponíveis e exemplos de comandos Moodle via Docker (`php82 admin/cli/purge_caches.php`, `phpunit82`, `php82 vendor/bin/phpcs --standard=moodle local/myplugin/`); seção Quality atualizada com comandos Docker
- **`instructions/PHP.md`**: seção Quality expandida com subseção "Docker commands" cobrindo `phpcs82`, `phpcbf82`, `phpunit82`, `php82 -l` e composer scripts via `composer82`

### Alterado

#### Stack PHP — Dockerfile da imagem PHP (`lumina dev create-stack-php`)
- `curl` adicionado ao `apt-get install` (necessário para download dos PHARs no build)
- **Stacks existentes precisam reconstruir as imagens** para receber phpcs, phpcbf e phpunit: `docker compose build --no-cache && docker compose up -d --force-recreate`

### Corrigido

#### Stack PHP — wrappers gerados em `~/.local/bin/`
- **TTY check incorreto**: verificação usava apenas `[ -t 1 ]` (stdout) — pipes de stdin (ex.: `echo '<?php ...' | php82`) causavam falha `"input device is not a TTY"` no `docker exec -it`; corrigido para `[ -t 0 ] && [ -t 1 ]` (stdin E stdout)
- **Colisão de prefixo de path**: substituição de caminho usava `${WS_HOST}*`, fazendo `/workspace/www/html_extra` ser incorretamente reescrito como subpath de `/workspace/www/html`; corrigido para `"${WS_HOST}/"*` (exige separador de diretório) com fallback de match exato `"${WS_HOST}"`
- **Injeção de aspas simples**: caminhos de workspace contendo `'` (ex.: `/home/d'artagnan/workspace`) quebravam a sintaxe bash das variáveis `CONTAINER` e `WS_HOST` nos scripts gerados; corrigido com `bashSingleQuote()` — converte `'` → `'\''`
- **Guard de versões vazia**: `writeToolWrappers` acessava `versions[0]` sem verificar slice vazio — adicionado retorno antecipado quando `versions` é nil ou vazio

#### Stack PHP — Dockerfile: PHARs sem verificação de integridade
- `phpcs.phar`, `phpcbf.phar` e `phpunit-*.phar` eram instalados diretamente sem checagem de integridade; adicionada verificação SHA256 via arquivos `.sha256` oficiais (`sha256sum -c`) antes de mover os PHARs para `/usr/local/bin`; arquivos temporários de hash removidos após verificação

#### Gerenciar Contextos IA — `MOODLE.md`: caminho de workspace fixo
- Template `MOODLE.md` gerado com `~/workspace/www/html/` hardcoded — incorreto para usuários com workspace personalizado; substituído pelo placeholder `{{WORKSPACE_PATH}}` resolvido em runtime a partir de `~/.lumina/config.yaml` via `config.Load()`

### Refatorado

#### Stack PHP — wrappers: DRY e helpers extraídos
- `var phpTools []string` extraído como variável de pacote — lista de ferramentas (`php`, `phpcs`, `phpcbf`, `phpunit`, `composer`) antes duplicada em dois pontos do código; `func phpSuffix(version string) string` elimina 7 ocorrências de `strings.ReplaceAll(v, ".", "")` espalhadas por `Compose()`, `buildCompose()` e `writeToolWrappers()`
- Pré-alocação do slice de wrappers com capacidade calculada: `make([]wrapDef, 0, len(phpTools)+len(versions)*len(phpTools))`

### Testes

#### Stack PHP — `internal/stack/config/compose_test.go`
- 8 testes unitários para `bashSingleQuote` e `buildWrapperScript`: shebang correto, container/tool presentes em ambos os branches `exec`, TTY check em stdin E stdout, colisão de prefixo de path (trailing slash), match exato de `WS_HOST`, aspas simples no caminho do workspace e guard de versões vazia (nil e slice vazio)

### Removido

#### Personalizar Linux — temas GTK Cinnamon
- **WhiteSur** removido do catálogo Cinnamon

---

## [2.1.0] — 2026-06-16

### Adicionado

#### Personalizar Linux — catálogos YAML embutidos (`lumina theme`)
- Temas GTK (GNOME e Cinnamon), cursores e pacotes de ícones agora definidos em arquivos YAML externos embutidos no binário (`//go:embed`) em vez de slices Go hardcoded
- Novos arquivos: `gnome_themes.yaml` (11 temas), `cinnamon_themes.yaml` (5 temas), `cursors.yaml` (4 cursores), `icons.yaml` (7 ícones)
- Cada catálogo é parseado uma única vez com `sync.Once` — mesmo padrão do catálogo MCP (`internal/dev/mcp/servers.yaml`)
- Permite atualizar temas, cursores e ícones sem recompilar o Go: basta editar o YAML antes do build

### Corrigido

#### Aplicativos Flatpak — override falhava para apps instalados no escopo system (`lumina apps install`)
- `flatpak override <id>` chamado sem flag de escopo tentava o escopo system, exigia root e falhava com `exit status 1` — corrigido passando o mesmo `scope` (`--system` ou `--user`) retornado por `config.FlatpakFlag()` e ajustando `RequiresSudo: scope == "--system"` dinamicamente
- Afetava principalmente o DBeaver Community (`io.dbeaver.DBeaverCommunity`), que usa `FlatpakOverride: ["--share=network"]`

### Alterado

#### Stack PHP — seleção de versões PHP via multi-select (`lumina stack config stack`)
- Substitui prompt de texto livre para as versões PHP por seleção interativa (`ui.RunMultiSelect`)
- Apenas versões suportadas (`8.1`, `8.2`, `8.3`, `8.4`) são exibidas como opções; ao menos uma deve ser selecionada
- `Compose` passa a aceitar `stdin io.Reader` — mesmo padrão das demais funções interativas

---

## [2.0.2] — 2026-06-12

### Adicionado

#### Aplicativos Flatpak — novos apps (`lumina apps install`)
- **Google Chrome** (`com.google.Chrome`) adicionado ao catálogo
- **AppImage Pool** (`io.github.prateekmedia.appimagepool`) adicionado ao catálogo
- **Bazaar** (`io.github.kolunmi.Bazaar`) adicionado ao catálogo
- **DBeaver Community** (`io.dbeaver.DBeaverCommunity`) adicionado ao catálogo; executa `flatpak override --share=network` automaticamente após instalação
- **Beekeeper Studio** (`io.beekeeperstudio.Studio`) adicionado ao catálogo
- Catálogo passa de 31 para 36 aplicativos (Krita, FileZilla, LibreOffice e Extension Manager já constavam)
- Suporte técnico: `App` ganhou campo `FlatpakOverride []string` — args passados a `flatpak override <id>` após instalação bem-sucedida

### Alterado

#### Aplicativos Flatpak — correções de ESC sequences (`lumina apps install/uninstall`)
- `--noninteractive` adicionado a todas as chamadas `flatpak install`, `flatpak uninstall` e `flatpak update` no codebase (`apps/install.go`, `apps/uninstall.go`, `gnome/prereqs.go`, `terminal/install.go`, `terminal/uninstall.go`, `update/update.go`, `postinstall/common.go`)
- Corrige escape sequences (`^[[24;63R`) que apareciam na saída do Flatpak — `TERM=dumb` suprime o nome do terminal mas não o `isatty()` check que o Flatpak usa para ativar barras de progresso com cursor position queries; `--noninteractive` desativa isso na camada do Flatpak

#### Aplicativos Flatpak — desinstalação (`lumina apps uninstall`)
- `RequiresSudo` agora é definido dinamicamente por escopo: `scope == "--system"` → `RequiresSudo: true`
- `SelectUninstall` reduzido de 4 para 2 chamadas a `flatpak list` — `scopeMap` construído uma única vez e reutilizado na desinstalação
- Iteração dupla sobre `Catalogue` em `SelectUninstall` eliminada — `inCatalogue` e `items` agora montados em um único loop
- Apps não instalados em `Uninstall()` geram aviso descritivo em vez de tentar desinstalar com escopo `--system` por padrão

#### IDEs — DBeaver (`lumina dev ide`)
- DBeaver CE removido do catálogo de IDEs gerenciadas (instalação via apt/rpm descontinuada)
- Migrado para o catálogo Flatpak como **DBeaver Community** (`io.dbeaver.DBeaverCommunity`)

#### Pós-instalação — Google Chrome
- Instalação automática do Google Chrome removida de todas as pós-instalações (Ubuntu, Mint, ZorinOS, Fedora)
- Chrome disponível exclusivamente via catálogo Flatpak (`lumina apps install`)

#### GNOME — Pré-requisitos de customização (`lumina gnome prereqs`)
- `inkscape` e `x11-apps` removidos dos pré-requisitos — eram dependências dos cursores Oreo (removidos em versão anterior); não têm mais uso no fluxo de customização

#### GNOME — Aplicar tema no Flatpak (`lumina gnome flatpak`)
- Lógica de `flatpak override` extraída para função privada `applyFlatpakTheme` — elimina duplicação entre `offerFlatpak` e `ApplyFlatpakTheme`
- Erro no `flatpak override --filesystem` deixou de ser silenciado; falha exibe warning e interrompe o fluxo antes do segundo override

#### GNOME — Temas GTK — Yaru (Ubuntu 26.04)
- Script de instalação migrado de URLs hardcoded com versão (`yaru-theme-gtk_26.04.5.1ubuntu_all.deb`) para descoberta dinâmica via índice HTML do pool Ubuntu — resistente a atualizações de pacote

#### Pós-instalação — Ubuntu 26.04 (`lumina system pos ubuntu`)
- Remoção seletiva de snaps (`removeSnaps`) agora recebe `stdin` injetado em vez de ler `os.Stdin` diretamente
- `Ubuntu()` migrado para assinatura interativa (`stdin io.Reader`) — TUI passou de `exec` para `execInteractive`

---

## [2.0.1] — 2026-06-10

### Adicionado

#### GNOME — Temas GTK (`lumina gnome themes`)
- **Yaru (Ubuntu 24.04)** adicionado ao catálogo — instala `fonts-ubuntu`, `yaru-theme-gtk`, `yaru-theme-icon`, `yaru-theme-sound` e `yaru-theme-gnome-shell` via `apt-get install`; remoção via `apt-get purge`
- **Yaru (Ubuntu 26.04)** adicionado ao catálogo — baixa os pacotes `.deb` diretamente do archive Ubuntu 26.04 (`yaru-theme-gtk`, `yaru-theme-icon`, `yaru-theme-sound`) e instala via `apt-get install`; remoção via `apt-get purge`
- Catálogo de temas GTK passa de 10 para 12 entradas
- Ambas as variantes instalam em `/usr/share/themes/` (sistema) em vez de `~/.themes/` (usuário)
- Suporte técnico: `themeEntry` ganhou dois novos campos — `CustomScript string` (bash executado como root via `RequiresSudo`) e `PurgePackages []string` (remoção via `apt-get purge`); `isThemeInstalled` atualizado para aceitar `DirPattern` como caminho absoluto

---

## [2.0.0] — 2026-06-10

### Adicionado

#### Pós-instalação — Google Chrome (`lumina system pos`)
- Google Chrome instalado automaticamente ao final da pós-instalação em todas as distribuições suportadas
- Debian (Mint, ZorinOS, Ubuntu): baixa `google-chrome-stable_current_amd64.deb` direto do CDN do Google e instala via `apt-get install`
- Fedora: instala `google-chrome-stable_current_x86_64.rpm` via `dnf install <URL>`
- Pulado automaticamente se `google-chrome` já estiver presente no PATH

#### Gerenciamento Linux — Instalar Linux Toys (`lumina system toys`)
- Novo item na seção "Gerenciamento Linux" da TUI e subcomando CLI `lumina system toys`
- Executa o instalador oficial: `curl -fsSL https://linux.toys/install.sh | bash`
- Implementado em `internal/system/linuxtoys/install.go`

#### Gerenciamento Linux — Instalar MegaSync (`lumina system megasync`)
- Novo item na seção "Gerenciamento Linux" da TUI e subcomando CLI `lumina system megasync`
- Detecta automaticamente a distribuição via `ID` + `VERSION_ID` em `/etc/os-release` e baixa o pacote correto:
  - Linux Mint 22.x → `megasync-xUbuntu_24.04_amd64.deb`
  - Ubuntu 24.04 → `megasync-xUbuntu_24.04_amd64.deb`
  - Ubuntu 26.04 → `megasync-xUbuntu_26.04_amd64.deb`
  - ZorinOS 18.x → `megasync-xUbuntu_24.04_amd64.deb`
  - Fedora 44 → `megasync-Fedora_44.x86_64.rpm`
- Pulado automaticamente se `megasync` já estiver presente no PATH
- Implementado em `internal/system/megasync/install.go`

#### IDEs — DBeaver CE (`lumina dev ide`)
- DBeaver CE adicionado ao catálogo de IDEs gerenciadas
- Debian: repositório APT oficial com GPG keyring em `/usr/share/keyrings/dbeaver.gpg.key`
- Fedora: instala RPM direto via URL (`dbeaver-ce-latest-linux-x86_64.rpm` via `dnf install`)
- Arch/Unknown: mensagem orientando instalação manual pelo AUR (`yay -S dbeaver`)

#### Aplicativos Flatpak — novos apps (`lumina apps install`)
- Vivaldi (`com.vivaldi.Vivaldi`) adicionado ao catálogo
- Android Studio (`com.google.AndroidStudio`) adicionado ao catálogo
- Tomatillo — Pomodoro (`io.github.diegopvlk.Tomatillo`) adicionado ao catálogo
- Catálogo passa de 28 para 31 aplicativos

#### Pós-instalação — Remoção seletiva de Snaps (`lumina system pos ubuntu`)
- Novo passo interativo no início da pós-instalação do Ubuntu 26.04
- Executa `snap list`, exibe multi-select com todos os snaps instalados e remove os selecionados com `snap remove --purge`
- Múltiplas passadas garantem a ordem correta de remoção respeitando dependências entre snaps (ex.: `snap-store` antes de `gnome-42-2204`)
- Passo ignorado silenciosamente se o `snap` não estiver disponível

#### Pós-instalação — Swapfile 4 GB (`lumina system pos`)
- Criação automática de swapfile de 4 GB nas pós-instalações do Ubuntu 26.04, Linux Mint 22.3 e ZorinOS 18.1
- Cria `/swapfile` com `fallocate -l 4G`, define permissão `600`, formata via `mkswap` e ativa com `swapon`
- Persiste automaticamente em `/etc/fstab` (entrada `none swap sw 0 0`); guard `grep -qF` evita duplicatas em reexecuções
- Passo ignorado silenciosamente se `/swapfile` já existir (idempotente)
- Limpeza automática de `/swapfile` se `chmod` ou `mkswap` falharem
- Implementado em `postinstall/common.go` como `setupSwapfile`; compartilhado entre Ubuntu, Mint e ZorinOS

#### Pós-instalação — Ubuntu 26.04 — `gnome-software-plugin-flatpak`
- `gnome-software-plugin-flatpak` adicionado à lista de pacotes essenciais do Ubuntu 26.04
- Habilita a gestão de aplicativos Flatpak diretamente pelo GNOME Software

#### Repositórios — Criar Código de Conduta (`lumina repo conduct`)
- Novo subcomando `lumina repo conduct` e item "Criar Código de Conduta" na seção "Gerenciar Repositórios" da TUI
- Cria `CODE_OF_CONDUCT.md` (inglês) e `CODIGO_DE_CONDUTA.md` (português) no diretório atual, baseados no Contributor Covenant v2.1
- Se algum arquivo já existir, exibe aviso e solicita confirmação antes de sobrescrever
- Implementado em `internal/manager/repo/conduct.go`

#### Configurar Lumina (`lumina self-config`)
- Nova ação interativa para editar workspace, diretório Docker Compose, tema e escopo Flatpak
- Disponível via TUI (Home → Configurar Lumina) e CLI (`lumina self-config`)
- Implementado em `internal/selfupdate/configure.go`

#### Stack — reiniciar containers (`lumina stack restart`)
- `stack.Restart` em `internal/stack/lifecycle.go`: `docker compose down` + `up -d --remove-orphans` em uma única ação
- Disponível via TUI e CLI

#### Novos comandos top-level: `lumina apps` e `lumina gnome`
- `lumina apps <install|uninstall|web>` — atalho direto para aplicativos Flatpak (era `lumina system apps`)
- `lumina gnome <pre|ext|themes|icons|cursor|flatpak>` — atalho direto para personalização GNOME (era `lumina system gnome`)
- `lumina system apps` e `lumina system gnome` mantidos por compatibilidade

#### `lumina ai context` e `lumina ai clear`
- `lumina ai` reestruturado com subcomandos explícitos: `context` (gerar) e `clear` (remover)
- `ClearContext` em `internal/manager/ai/context.go`: remove CLAUDE.md, GEMINI.md, AGENTS.md, arquivos de regras, ignore files, instruções de modelo e diretório php-references; reporta contagem de itens removidos

#### `lumina repo gitignore`
- `gitignore` adicionado como subcomando de `lumina repo`
- Anterior: `lumina gitignore` era top-level

#### `lumina dev create-workspace` e `lumina dev create-stack-php`
- Atalhos diretos no CLI para criação de workspace e docker-compose da stack PHP
- Equivalentes a `lumina stack config workspace` e `lumina stack config stack`

#### Terminais — integração com gerenciadores de arquivos (`lumina dev term`)
- Ao instalar um terminal, entradas "Abrir no [Terminal]" são criadas automaticamente nos menus de contexto do Nautilus, Nemo e Dolphin
- Ao desinstalar, as entradas são removidas automaticamente
- Suporte: Kitty (`--directory`), Alacritty (`--working-directory`), Black Box (`flatpak run com.raggesilver.BlackBox --working-directory`), GNOME Console (`--working-directory`)
- Starship não recebe entradas de menu (sem conceito de diretório de abertura)
- Implementado em `internal/dev/terminal/contextmenu.go`

### Alterado

#### Templates de IA — padronização de seções entre `instructions/*.md`
- `MOODLE.md`: adicionadas seções `## Language` e `## Quality` (moodle-cs) que estavam ausentes
- `MCP.md`: adicionada seção `## Language`; `## Pre-commit Checklist` renomeada para `## Quality` com comandos de tooling (`tsc --noEmit`, `eslint src/`, `npm test`) antes da checklist
- `BASH.md`: adicionada seção `## Anti-Patterns` com 8 padrões a evitar (echo colorizado, variáveis sem aspas, parsing de `ls`, `cd` sem guard, `eval`, `which`, `local` ausente, declaração + substituição de comando em passo único)
- `PHP.md`: adicionada seção `## Anti-Patterns` com 8 padrões a evitar (catch broad, SQL hardcoded, `$_POST` direto, `new` inline, `echo` em domain classes, magic numbers, `@` supressor, retornar null em vez de exceção)

#### Templates de IA — `ONLY-CLAUDE.md`
- Modelo Opus atualizado de `claude-opus-4-7` (Opus 4.7) para `claude-opus-4-8` (Opus 4.8)
- Nota da tabela de modelos traduzida de português para inglês: `(use no orquestrador, não em subagents)` → `(use in the orchestrator, not in subagents)`
- Parágrafo de Escalation traduzido para inglês — arquivo misturava idiomas violando a regra do próprio `BASIC.md`

#### Criar Contexto AI — saída agrupada em painel único (`lumina ai context`)
- Todas as mensagens `+ arquivo criado.` e `- arquivo removido.` são coletadas durante a execução e exibidas em um único `PrintBox` ao final, em vez de serem impressas inline uma a uma
- A mensagem de info "Atualizando arquivos de contexto para: ..." é exibida antes das operações de arquivo, garantindo a ordem correta: info → box com lista → success → WaitEnter

#### Limpar Contexto AI — seleção antes da remoção (`lumina ai clear`)
- Antes de remover qualquer arquivo, exibe multi-select com todos os arquivos de contexto presentes no diretório atual — todos pré-selecionados (remover tudo por padrão)
- O usuário pode desmarcar arquivos que deseja preservar; apenas os itens marcados são removidos
- Função promovida de não-interativa para interativa: `ClearContext` agora aceita `stdin io.Reader`, despacho na TUI atualizado de `exec` para `execInteractive`

### Corrigido

#### Testes — `manager/ai`: argumentos faltando em chamadas de funções internas
- `context_test.go`: 4 chamadas a `generateSharedFiles` e `writeInstruction` estavam sem o argumento `log *strings.Builder` adicionado recentemente às assinaturas — causava falha de compilação no CI (`not enough arguments in call`)

#### Documentação — `README.md` revisado para refletir o estado atual da aplicação
- Ulauncher removido das features e da referência de comandos (não acessível via TUI)
- Teclas numéricas `1`–`4` removidas da tabela de atalhos (não implementadas no layout v2)
- Seções renomeadas para corresponder à TUI: "Aplicativos" → "Aplicativos Linux", "Personalização GNOME" → "Personalizar Linux", "DevStack" → "Gerenciar Stack PHP", "DevStuff" → "Ambiente de Desenvolvimento", "Configurações Lumina" → "Home"
- "DevManager" dividido em três seções independentes: "Gerenciar banco de Dados", "Gerenciar Repositórios", "Gerenciar Contextos IA"
- Referência CLI reorganizada e alinhada com os subcomandos reais de cada dispatcher

#### Pós-instalação — Chrome e MegaSync: arquivo temporário não removido em falha de download
- `installChromeDeb` em `postinstall/common.go`: arquivo `.deb` temporário agora é removido quando o `wget` falha, eliminando arquivo órfão em `/tmp`
- `megasync/install.go`: mesma correção aplicada; o arquivo temporário é removido em todos os caminhos de erro

#### Pós-instalação — Chrome e MegaSync: caminhos previsíveis em `/tmp` (TOCTOU)
- Migração de caminhos fixos como `/tmp/google-chrome.deb` para `os.CreateTemp("", "google-chrome-*.deb")` em `installChromeDeb` e `megasync/install.go`
- Elimina condição de corrida TOCTOU: um arquivo temporário de nome previsível pode ser substituído por um link simbólico entre a checagem e a abertura

#### IDEs — DBeaver CE: keyring sobrescrito a cada reinstalação (`lumina dev ide`)
- Guard `[ -f "$KEYRING" ] ||` adicionado ao script de instalação APT — o keyring e o arquivo de fontes só são criados se ainda não existirem

#### IDEs — DBeaver CE: prompts interativos e ESC sequences durante instalação (`lumina dev ide`)
- `DEBIAN_FRONTEND=noninteractive` e flags `-o Dpkg::Use-Pty=0 -o Dpkg::Progress-Fancy=0 -o APT::Color=0` adicionados à instalação APT do DBeaver, alinhando com o padrão dos demais instaladores

#### MegaSync — falha silenciosa com `/etc/os-release` incompleto
- Guard adicionado em `resolvePackage`: retorna erro descritivo se o campo `ID` estiver vazio no `/etc/os-release`

#### Chrome, MegaSync, DBeaver — instalação silenciosa falha em arquiteturas não-amd64
- Guard `runtime.GOARCH != "amd64"` adicionado em `installChromeDeb`, `installChromeFedora`, `megasync.Install` e `installDBeaver` (caso Fedora)
- Em vez de falhar no meio da instalação com erro de pacote incompatível, exibe mensagem clara e orienta instalação manual

### Refatorado

#### Detecção de distribuição — `distro.RawID()` e `distro.VersionID()`
- `internal/distro` expõe `RawID()` (campo `ID=` bruto, ex.: `"linuxmint"`) e `VersionID()` (campo `VERSION_ID=`, ex.: `"24.04"`), ambos com cache `sync.Once`
- `megasync/install.go` migrado para usar as novas funções; funções locais `parseOSRelease` e `osReleaseTrim` removidas (eliminada duplicação de lógica de parsing)
- `resolvePackage` passa a aceitar `(id, ver string)` — testável sem I/O de sistema de arquivos

#### MegaSync — helper de download encapsulado
- `downloadToTemp(ctx, exe, url, pattern)` extraído em `megasync/install.go`: encapsula `os.CreateTemp` + `wget`, remove o arquivo temporário automaticamente se o download falhar

#### Testes unitários — `megasync` e `conduct`
- `megasync/install_test.go`: 10 casos de tabela para `resolvePackage` cobrindo Mint 22.x, Ubuntu 24.04/26.04, ZorinOS 18.x, Fedora 44 e casos não suportados
- `manager/repo/conduct_test.go`: verifica presença de todas as seções obrigatórias em `conductEN` e `conductPT` e os nomes dos arquivos de saída

### Removido

#### Pós-instalação — Fedora KDE 44 e Kubuntu 26.04
- Scripts `internal/system/postinstall/fedora_kde.go` e `kubuntu.go` removidos
- Menu "Personalizar Linux": "Pré-requisitos KDE" e "Temas KDE" removidos

#### `lumina gitignore` (top-level)
- Removido — use `lumina repo gitignore`

#### Limpeza de planejamento
- `new-lumina-tolls/` (scratch Go de planejamento) removido
- `lumina-tools-2-planner.md` removido

### Corrigido

#### Stack — `$CFG->dataroot is not writable` no Moodle (`lumina dev create-stack-php`)
- Dockerfile do PHP gerado por `Compose()` agora recebe o UID do host como build arg (`ARG UID=1000` com fallback; `RUN usermod -u ${UID} www-data`) em vez de fixar `usermod -u 1000 www-data`
- `buildCompose` passa a receber `uid int` (detectado via `os.Getuid()` em `Compose()`) e injeta `UID: %d` em `build.args` de cada serviço PHP no `docker-compose.yml`
- Causa: em máquinas onde o usuário principal não tem UID 1000 (ex.: contas adicionais, ambientes corporativos), o `www-data` do container ficava com dono diferente do diretório montado como `$CFG->dataroot`, mesmo após `FixPerms` — resultando em "dataroot is not writable"
- **Stacks já existentes precisam reconstruir as imagens PHP** para a correção ter efeito: `docker compose build --no-cache && docker compose up -d --force-recreate`, seguido de "Ajustar Permissões"

#### Flatpak — desinstalação usava escopo configurado em vez do escopo real do app (`lumina apps uninstall`)
- `Uninstall` em `apps/uninstall.go` agora detecta via `InstalledScopeMap` o escopo real (`--system` ou `--user`) de cada app antes de desinstalar
- Anterior: usava sempre `config.FlatpakFlag()` — apps instalados em escopo diferente do configurado falhavam na desinstalação

#### TUI — entrada direta para "Criar Stack PHP" navegava para seção errada
- `RunAtStackConfig` apontava para seção "Aplicativos Linux" (cursor em "Desinstalar") em vez de "Ambiente de Desenvolvimento" → "Criar Stack PHP"
- Introduzido durante a migração para o layout v2; corrigido com constantes nomeadas para todos os índices de seção em `internal/tui/content.go`

#### Pós-instalação — execuções duplicadas de `add-apt-repository`
- Mint, Ubuntu e ZorinOS verificam se o componente apt já está habilitado antes de chamar `add-apt-repository universe/multiverse`
- `aptComponentEnabled` em `postinstall/common.go` suporta formato `.list` (legado) e DEB822 `.sources`

#### Pós-instalação — Flathub duplicado em ZorinOS e Fedora
- `zorin.go` e `fedora.go`: chamadas a `ensureFlatpakReady` removidas — Flathub já vem pré-configurado nessas distros
- `fedora.go`: `rpmFusionEnabled` guard adicionado — RPM Fusion não é reinstalado se já estiver presente

### Reformulado

#### TUI — layout e navegação completamente redesenhados

O modelo de navegação em pilha (menu → submenu → ação) foi substituído por um layout de três níveis fixos:

- **Cabeçalho persistente**: marca `◈ lumina.tools · <versão>` e indicador de conectividade à internet (DNS lookup periódico de `github.com` a cada 45s)
- **Painel duplo**: submenu à esquerda + painel de conteúdo à direita
  - `Tab` / `Shift+Tab` alternam foco entre os painéis
  - No submenu: `↑↓` navegam o cursor; seleção atualiza o conteúdo em tempo real (master-detail)
  - No conteúdo: título e comando CLI do item selecionado; executável com `Enter`
- **Rodapé contextual**: hints mudam conforme o painel em foco ou overlay ativo

#### TUI — overlays

- **Confirmação de saída**: `q` abre overlay "Deseja encerrar o Lumina Tools?"; `Enter`/`y`/`s` confirmam; `Esc`/`q`/`n` cancelam — `Ctrl+C` sai instantaneamente sem overlay
- **Seletor de tema**: `t` abre overlay com lista de temas; preview ao vivo ao mover o cursor; `Enter` confirma e salva; `Esc` cancela e restaura o tema anterior

#### TUI — remoção do código v1

- `internal/tui/model.go`, `menus.go`, `header.go`, `footer.go`, `styles.go` removidos — substituídos por `model_v2.go`, `content.go` e `chrome.go`

## [1.0.5] — 2026-06-04

### Adicionado

#### DevStuff — Pré-Requisitos unificados (`lumina dev pre`)
- `internal/dev/prereqs/` — novo pacote com catálogo de 5 grupos instaláveis via multi-select, seguindo o mesmo padrão de "Gerenciar CLIs LLM" (itens instalados pré-selecionados; selecionar instala, desmarcar desinstala)
- **Pacotes base**: `curl`, `git`, `openssl`, `lsof`
- **Ferramentas DevStuff**: `libsecret`, `gnome-keyring` (nomes de pacotes por distro)
- **GitHub CLI**: `gh` — repositório oficial por distro (keyring + apt source no Debian; `dnf config-manager` no Fedora; `pacman` no Arch)
- **Docker Engine**: docker + buildx, habilita serviço systemd e adiciona o usuário ao grupo docker
- **Node.js**: Node.js LTS via nvm (migrado de `dev/depends`)
- `lumina dev pre` agora abre o multi-select unificado — anterior: fluxo linear sem seleção de componentes
- `lumina stack config pre` removido — funcionalidade absorvida pelo fluxo unificado
- TUI: "Instalar Pré-Requisitos" adicionado como primeiro item em DevStuff; itens duplicados removidos de "Criar Stack" e "Gerenciar Ferramentas"

#### Terminais — GNOME Console (`lumina dev term`)
- GNOME Console (`kgx`) adicionado ao catálogo de terminais entre Black Box e Starship
- Instalação via gerenciador de pacotes: `apt-get install gnome-console` (Debian/Ubuntu) ou `dnf install gnome-console` (Fedora)

#### GNOME — Temas GTK (`lumina system gnome themes`)
- **Graphite** adicionado ao catálogo — instala via `./install.sh -t all` (variantes: default, purple, pink, red, orange, yellow, green, teal, blue)
- **Zorin** adicionado ao catálogo — copia os diretórios pré-construídos da raiz do repositório (`ZorinBlue-Dark/`, `ZorinBlue-Light/`, `ZorinGreen-Dark/`, etc.) para `~/.themes/`

### Corrigido

#### `lumina dev go` ausente na CLI
- `case "go"` adicionado em `app.go:dispatchDev` — o comando CLI não roteava para `devgolang.Manage`; apenas a TUI funcionava corretamente

#### Self-update — sem retorno claro quando já na versão mais recente
- Ao detectar que a versão instalada é a mais recente, exibe painel de sucesso "Você já possui a versão mais recente." e aguarda Enter antes de retornar — anterior: retornava silenciosamente sem feedback

#### Pós-instalação — falha no `sysctl` (`lumina system pos`)
- `configureSysctl` em `postinstall/common.go`: `printf '%s'` → `printf '%b'` — corrige "Argumento inválido" reportado pelo `sysctl -p` em todas as distros
- Causa: `%q` do Go formata `\n` como escape literal (barra-n); `printf '%s'` escreve o arquivo sem quebras de linha reais; `sysctl` recebia uma única linha longa e falhava ao definir `vm.swappiness`

#### Stack — config não era relido após ações que o modificam (`lumina stack`)
- `actionDoneMsg` no modelo Bubble Tea agora recarrega `~/.lumina/config.yaml` do disco após cada ação concluída
- Corrige o bug onde "Iniciar Stack" exibia "Stack não configurada" mesmo após executar "Criar Workspace" ou "Criar Stack" na mesma sessão da TUI

#### Flatpak — ESC sequences em chamadas faltantes (`lumina dev term`, pós-instalação)
- `TERM=dumb` adicionado às chamadas `flatpak install` e `flatpak uninstall` do Black Box em `dev/terminal/install.go` e `dev/terminal/uninstall.go`
- `TERM=dumb` adicionado às chamadas `flatpak remote-add` em `system/postinstall/common.go` e `system/apps/install.go`

### Refatorado

#### Pré-Requisitos — remoção de arquivos obsoletos
- `internal/dev/depends/depends.go` removido — lógica migrada para `internal/dev/prereqs`
- `internal/stack/config/prereqs.go` removido — `SetupPrereqs` migrado para `internal/dev/prereqs`
- `internal/stack/config/depends.go` removido — `Depends` migrado para `internal/dev/prereqs`

#### Fluxos Select — helper compartilhado (`ide`, `llm`, `terminal`)
- `ui.RunManagedSelect` adicionado a `internal/ui`: encapsula o fluxo multiselect → diff → execute comum aos domínios de ferramentas gerenciadas
- `ide/select.go`, `llm/select.go` e `terminal/select.go` simplificados de ~75 para ~30 linhas cada

#### Instalação de pacotes — helper compartilhado (`distro.InstallPkgs`)
- `distro.InstallPkgs(ctx, exe, stdout, family, pkgs...)` adicionado a `internal/distro`: encapsula o bloco `switch family { apt-get / dnf / pacman }` duplicado

#### Help — fonte única de verdade (`selfupdate.HelpMarkdown`)
- `helpMarkdown()` exportada como `HelpMarkdown()` em `internal/selfupdate/help.go`
- `lumina dev go` adicionado à tabela de DevStuff (estava ausente em ambos os arquivos de help)
- `app.go:printHelp` delega a `selfupdate.HelpMarkdown()` — uma única fonte para `lumina help` e o viewer interativo da TUI

---

## [1.0.4] — 2026-06-03

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
