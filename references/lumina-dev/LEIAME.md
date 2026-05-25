# 💡 LuminaDev — Configuração de Workstation

> Automação de workstation Linux para desenvolvedores PHP/Moodle com ergonomia JetBrains.

📄 English version: see [README.md](README.md)

![Licença](https://img.shields.io/badge/license-GPL--3.0-blue)
![Shell](https://img.shields.io/badge/shell-bash-green)
![Distros](https://img.shields.io/badge/distros-Ubuntu%20%7C%20Debian%20%7C%20Fedora%20%7C%20Arch-orange)
![CI](https://img.shields.io/github/actions/workflow/status/kaduvelasco/lumina-dev/lint.yml?label=lint%20%26%20smoke%20test)

---

## 📋 Índice

- [Sobre o projeto](#-sobre-o-projeto)
- [Diferenciais](#-diferenciais)
- [Estrutura do projeto](#-estrutura-do-projeto)
- [Pré-requisitos](#-pré-requisitos)
- [Instalação](#-instalação)
- [Desinstalação](#-desinstalação)
- [Scripts e módulos](#-scripts-e-módulos)
- [Distros suportadas](#-distros-suportadas)
- [CI e qualidade de código](#-ci-e-qualidade-de-código)
- [Contribuindo](#-contribuindo)
- [Licença](#-licença)

---

## 📖 Sobre o projeto

O **LuminaDev** é uma suíte de automação em Shell Script que transforma uma instalação base do Linux em uma estação de trabalho de alta performance para desenvolvedores PHP/Moodle.

Cada ferramenta é instalada de forma **idempotente** — o script verifica o que já existe antes de agir, nunca sobrescrevendo configurações sem confirmação. O suporte a múltiplas distribuições é gerenciado por um módulo central (`utils.sh`) que detecta automaticamente o gerenciador de pacotes disponível.

---

## 🚀 Diferenciais

### 1. Experiência Visual "Storm"

Todas as IDEs são configuradas para replicar a ergonomia do ecossistema JetBrains: fonte JetBrains Mono, tema One Dark / JetBrains New UI e atalhos de teclado no padrão IntelliJ IDEA.

### 2. Múltiplos Editores

Suporte a cinco editores prontos para PHP/Moodle — VS Code, VSCodium, Zed, Windsurf e PHPStorm — todos configurados com a mesma identidade visual JetBrains.

### 3. Servidores MCP

Instalação e configuração automatizada de servidores MCP (Model Context Protocol) para ampliar as capacidades dos CLIs de IA com contexto de Moodle e base de conhecimento persistente.

### 4. Idempotência Total

Todos os scripts verificam o estado atual do sistema antes de agir. Reinstalações e execuções repetidas são seguras: o que já está instalado é pulado, e configurações existentes são preservadas com backup antes de qualquer sobrescrita.

### 5. Módulo Utilitário Centralizado

O `utils.sh` provê detecção de distro, funções de verificação e instalação de pacotes compartilhadas por todos os scripts, eliminando duplicação de código e garantindo consistência.

---

## 📂 Estrutura do projeto

```text
lumina-dev/
├── lumina-dev.sh               # Menu principal (instalação + acesso ao desinstalador)
├── opencode-models.md          # Referência de modelos suportados pelo OpenCode
├── .aiexclude                  # Modelo de bloqueio para ferramentas de IA
├── .gitignore                  # Padrão para projetos PHP/Moodle
│
├── scripts/                    # Utilitários e instaladores de CLI
│   ├── utils.sh                # Módulo central: cores, distro, idempotência
│   ├── uninstall.sh            # Menu de remoção seletiva
│   ├── fonts-install.sh        # Instalação da JetBrains Mono
│   ├── claude-install.sh       # Instalação do Claude Code CLI
│   ├── gemini-install.sh       # Instalação do Gemini Code Assist CLI
│   ├── opencode-install.sh     # Instalação do OpenCode (CLI e Desktop)
│   ├── codex-install.sh        # Instalação do OpenAI Codex CLI
│   ├── mcp-install.sh          # Instalação de servidores MCP
│   └── kitty-install.sh        # Instalação do Kitty Terminal + Starship
│
├── ides/                       # Instaladores de IDEs e editores
│   ├── zed-install.sh          # Zed Editor — PHP/Moodle Edition
│   ├── vscodium-install.sh     # VSCodium — PHP/Moodle Edition
│   ├── vscode-install.sh       # VS Code — PHP/Moodle Edition
│   ├── phpstorm-install.sh     # Auxiliar de instalação do PHPStorm (.tar.gz)
│   └── windsurf-install.sh     # Windsurf Editor (via repositório oficial)
│
└── .github/
    └── workflows/
        └── lint.yml            # CI: ShellCheck + Smoke Test
```

---

## ✅ Pré-requisitos

- Linux com `bash` 4.0+
- `sudo` configurado para seu usuário
- Conexão com a internet
- Distro suportada (veja [Distros suportadas](#-distros-suportadas))

> **Arch Linux:** a instalação de VS Code, VSCodium e Windsurf requer um AUR helper (`yay` ou `paru`) previamente instalado.

Para o instalador do **PHPStorm**, é necessário baixar o pacote `.tar.gz` manualmente em [jetbrains.com/phpstorm/download](https://www.jetbrains.com/phpstorm/download) antes de executar o script.

Para os **Servidores MCP**, é necessário ter o Node.js v18+ instalado (o script oferece instalação automática caso ausente).

---

## 🛠️ Instalação

**1. Clone o repositório:**

```bash
git clone https://github.com/kaduvelasco/lumina-dev.git
cd lumina-dev
```

**2. Dê permissão ao instalador:**

```bash
chmod +x lumina-dev.sh
```

**3. Execute:**

```bash
./lumina-dev.sh
```

O menu principal será exibido. As permissões dos demais scripts são sincronizadas automaticamente ao iniciar.

### Opções do menu

| Opção | Descrição                                     |
| ----- | --------------------------------------------- |
| `1`   | Instalar fontes JetBrains Mono                |
| `2`   | Instalar Git e compilar libsecret             |
| `3`   | Instalar LLMs (submenu)                       |
| `3>1` | Instalar Claude Code CLI                      |
| `3>2` | Instalar Gemini Code Assist CLI               |
| `3>3` | Instalar OpenCode (CLI)                       |
| `3>4` | Instalar OpenCode (Desktop)                   |
| `3>5` | Instalar OpenAI Codex CLI                     |
| `4`   | Instalar IDEs (submenu)                       |
| `4>1` | Instalar Zed Editor                           |
| `4>2` | Instalar VSCodium                             |
| `4>3` | Instalar VS Code                              |
| `4>4` | Auxiliar de instalação do PHPStorm            |
| `4>5` | Instalar Windsurf                             |
| `5`   | Instalar Servidores MCP (submenu)             |
| `5>1` | Instalar Moodle Dev MCP                       |
| `5>2` | Instalar Lumina AI Vault                      |
| `5>3` | Instalar Code Review Graph                    |
| `6`   | Instalar Kitty Terminal                       |
| `7`   | Iniciar Desinstalador                         |
| `0`   | Sair                                          |

### Ordem recomendada de instalação

```
1 → Fontes
2 → Git e libsecret
3 → LLMs (Claude, Gemini, OpenCode CLI e Desktop, Codex)
4 → IDEs (Zed, VSCodium, VS Code, PHPStorm, Windsurf)
5 → Servidores MCP (Moodle Dev MCP, Lumina AI Vault, Code Review Graph)
6 → Terminal Kitty (opcional)
```

---

## 🧹 Desinstalação

O desinstalador está acessível diretamente pelo menu principal (opção `7`) ou pode ser executado de forma independente:

```bash
bash scripts/uninstall.sh
```

O menu de remoção permite escolher individualmente o que desinstalar. Cada opção exige confirmação explícita antes de agir.

| Opção  | Descrição                              |
| ------ | -------------------------------------- |
| `1`    | Remover fontes JetBrains Mono          |
| `2`    | Remover IDEs (submenu)                 |
| `2>1`  | Remover Zed Editor                     |
| `2>2`  | Remover VSCodium                       |
| `2>3`  | Remover VS Code                        |
| `2>4`  | Remover PHPStorm                       |
| `2>5`  | Remover Windsurf                       |
| `3`    | Remover LLMs (submenu)                 |
| `3>1`  | Remover Claude Code CLI                |
| `3>2`  | Remover Gemini Code Assist CLI         |
| `3>3`  | Remover OpenCode CLI                   |
| `3>4`  | Remover OpenCode Desktop               |
| `3>5`  | Remover OpenAI Codex CLI               |
| `4`    | Remover Servidores MCP                 |
| `5`    | Remover Kitty Terminal                 |
| `0`    | Voltar                                 |

---

## 📦 Scripts e módulos

### `scripts/utils.sh`

Módulo central carregado por todos os outros scripts via `source`. Não deve ser executado diretamente.

| Função                     | Descrição                                                  |
| -------------------------- | ---------------------------------------------------------- |
| `detect_pkg_manager`       | Detecta e exporta `PKG_MANAGER` (apt / dnf / pacman)       |
| `is_installed_cmd`         | Verifica se um comando existe no PATH                      |
| `is_installed_pkg`         | Verifica se um pacote está instalado no sistema            |
| `ensure_pkg`               | Instala um pacote apenas se ausente                        |
| `ensure_local_bin_in_path` | Garante que `~/.local/bin` está no PATH (bashrc e zshrc)   |
| `check_node_version`       | Verifica se Node.js ≥ v18 está instalado                   |
| `install_node`             | Instala Node.js LTS via NodeSource (apt/dnf/pacman)        |
| `require_not_root`         | Aborta se executado como root                              |
| `require_sudo`             | Valida que sudo está disponível e funcional                |
| `require_internet`         | Verifica conexão antes de operações de download            |
| `print_version`            | Exibe a versão instalada de um comando                     |
| `show_lumina_header`       | Exibe o cabeçalho ASCII padrão LuminaDev                   |
| `die`                      | Encerra com mensagem de erro e código de saída             |
| `warn`                     | Exibe aviso em stderr                                      |
| `info`                     | Exibe mensagem informativa                                 |
| `success`                  | Exibe mensagem de sucesso                                  |

---

### `scripts/fonts-install.sh`

Instala a **JetBrains Mono v2.304** em `~/.local/share/fonts` e atualiza o cache de fontes do sistema. Verifica se a fonte já está presente antes de baixar.

---

### `scripts/kitty-install.sh`

Instala o **Kitty Terminal** via script oficial e o **Starship prompt**. Aplica configurações completas:

- Fonte **JetBrains Mono Nerd Font v3.3.0** — instalada automaticamente pelo script
- Scrollback de 10.000 linhas, suporte a layouts e múltiplas janelas/tabs
- Configuração de tema interativa via `ctrl+shift+F2`
- Starship prompt com seleção de preset interativa
- Integração com o sistema: entrada `.desktop` e ícone registrados
- Opção para definir como terminal padrão

---

### `scripts/claude-install.sh`

Instala o **Claude Code CLI** via script oficial da Anthropic. Verifica e instala o Node.js LTS (v18+) se necessário.

**Pós-instalação:** execute `claude` no terminal para autenticar com sua conta Anthropic.

---

### `scripts/gemini-install.sh`

Instala o **Gemini Code Assist CLI** via npm (`@google/gemini-cli`). Verifica e instala o Node.js LTS (v18+) se necessário.

**Pós-instalação:** a configuração da `GOOGLE_API_KEY` é solicitada ao final da instalação.

---

### `scripts/opencode-install.sh`

Instala o **OpenCode** com menu interativo para escolher o que instalar:

- **CLI** via npm (`opencode-ai@latest`) — requer Node.js v18+
- **Desktop** via GitHub Releases — detecta automaticamente a versão mais recente e a arquitetura do sistema

Pode ser chamado diretamente com argumento para pular o menu:

```bash
bash scripts/opencode-install.sh cli
bash scripts/opencode-install.sh desktop
```

---

### `scripts/codex-install.sh`

Instala o **OpenAI Codex CLI** via npm (`codex-cli`). Verifica e instala o Node.js LTS (v18+) se necessário.

**Pós-instalação:** configure a variável `OPENAI_API_KEY` com sua chave obtida em [platform.openai.com/api-keys](https://platform.openai.com/api-keys) e execute `codex` no terminal.

---

### `scripts/mcp-install.sh`

Instala e configura **servidores MCP** (Model Context Protocol) para integração com Claude Code, Gemini e outros clientes compatíveis. Requer Node.js v18+.

Pode ser chamado diretamente com argumento para pular o menu:

```bash
bash scripts/mcp-install.sh moodle-dev-mcp
bash scripts/mcp-install.sh lumina-ai-vault
bash scripts/mcp-install.sh code-review-graph
```

#### Moodle Dev MCP (`5>1`)

Instala o pacote `moodle-dev-mcp` globalmente via npm e registra o servidor no Claude Code. Requer a variável `MOODLE_PATH` apontando para a instalação local do Moodle.

- **Requisitos:** Moodle 4.1+ | Node.js 18+
- **Docs:** [github.com/kaduvelasco/moodle-dev-mcp](https://github.com/kaduvelasco/moodle-dev-mcp)

#### Lumina AI Vault (`5>2`)

Instala o pacote `lumina-ai-vault` globalmente via npm e registra o servidor no Claude Code. Mantém uma base de conhecimento persistente acessível pelos CLIs de IA. O vault path padrão é `~/.lumina-aivault/knowledge`.

- **Requisitos:** Node.js 18+
- **Docs:** [github.com/kaduvelasco/lumina-ai-vault](https://github.com/kaduvelasco/lumina-ai-vault)

#### Code Review Graph (`5>3`)

Instala o `code-review-graph` via **pipx** (Python), usando UV como backend. Constrói um grafo de conhecimento incremental do codebase e provê análise de impacto e revisão de código baseada em contexto estrutural. Após a instalação, executa `code-review-graph install` para configurar automaticamente as plataformas suportadas.

- **Requisitos:** UV + pipx (instalados automaticamente pelo script)

---

### `ides/zed-install.sh`

Instala o **Zed Editor** via script oficial e aplica configurações com JetBrains Mono, tema One Dark, assistant Claude pré-configurado e suporte a PHP/Moodle via `phpactor`.

> **Atenção:** o `phpactor` deve ser instalado separadamente via Extensions (`ctrl+shift+x`).

---

### `ides/vscodium-install.sh`

Instala o **VSCodium** via repositório oficial e configura o ambiente completo:

- Extensões PHP/Moodle: Intelephense, PHP CS Fixer, PHP Namespace Resolver, Moodle Snippets, Mustache
- Extensões Docker: vscode-docker, remote-containers
- Interface: JetBrains Mono, tema JetBrains New UI, keybindings IntelliJ

> **Arch Linux:** requer `yay` ou `paru` instalado previamente.

---

### `ides/vscode-install.sh`

Instala o **VS Code** via repositório oficial da Microsoft e configura o ambiente completo:

- Extensões PHP/Moodle: Intelephense, PHP CS Fixer, PHP Namespace Resolver, Moodle Snippets, Mustache
- Extensões Docker: vscode-docker, remote-containers
- Interface: JetBrains Mono, tema JetBrains New UI, keybindings IntelliJ

> **Arch Linux:** requer `yay` ou `paru` instalado previamente.

---

### `ides/phpstorm-install.sh`

Instalador auxiliar para o **PHPStorm** a partir de um pacote `.tar.gz` baixado manualmente. Extrai para `/opt/phpstorm`, cria link simbólico em `/usr/local/bin/phpstorm` e gera entrada `.desktop` no menu do sistema.

> **Atenção:** o PHPStorm requer licença ativa. Baixe em [jetbrains.com/phpstorm](https://www.jetbrains.com/phpstorm/download).

---

### `ides/windsurf-install.sh`

Instala o **Windsurf Editor** via repositório oficial — sem necessidade de download manual.

- **apt (Ubuntu/Debian):** adiciona chave GPG e repositório, instala via `apt-get`
- **dnf (Fedora):** importa chave RPM e cria repositório, instala via `dnf`
- **pacman (Arch/CachyOS):** instala via AUR (`windsurf`) com `yay` ou `paru`

---

## 🐧 Distros suportadas

| Distro        | Gerenciador | Status         | Observações                                        |
| ------------- | ----------- | -------------- | -------------------------------------------------- |
| Ubuntu 22.04+ | apt         | ✅ Suportado   | —                                                  |
| Linux Mint    | apt         | ✅ Suportado   | Baseado em Ubuntu                                  |
| Pop!\_OS      | apt         | ✅ Suportado   | Baseado em Ubuntu                                  |
| Fedora 39+    | dnf         | ✅ Suportado   | —                                                  |
| Arch Linux    | pacman      | ✅ Suportado   | VS Code, VSCodium e Windsurf requerem `yay`/`paru` |
| CachyOS       | pacman      | ✅ Suportado   | Baseado em Arch; `yay` geralmente pré-instalado    |
| Manjaro       | pacman      | ✅ Suportado   | Baseado em Arch; requer `yay`/`paru`               |
| Outras        | —           | ⚠️ Não testado | —                                                  |

---

## ⚙️ CI e qualidade de código

O workflow `.github/workflows/lint.yml` executa automaticamente a cada push ou pull request na branch `main`.

**ShellCheck** — lint estático em todos os arquivos `.sh` com `--severity=warning`.

**Smoke Test** — valida a sintaxe de cada script com `bash -n`, carrega o `utils.sh` e verifica que todas as funções essenciais estão presentes.

Para testar localmente:

```bash
find . -name "*.sh" | xargs shellcheck --severity=warning --shell=bash --exclude=SC1091
```

---

## 🤝 Contribuindo

Contribuições são bem-vindas! Consulte o [CONTRIBUINDO.md](CONTRIBUINDO.md) para detalhes sobre o processo e os padrões de código.

---

## ⚖️ Licença

Este projeto está licenciado sob a [GPL-3.0 License](LICENSE).

---

Made with ❤️ and AI by [Kadu Velasco](https://github.com/kaduvelasco)
