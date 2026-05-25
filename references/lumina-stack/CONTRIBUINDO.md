# Contribuindo com o LuminaStack

📄 English version: see [CONTRIBUTING.md](CONTRIBUTING.md)

Obrigado pelo interesse em contribuir com o LuminaStack! Este documento descreve o processo para reportar bugs, sugerir melhorias e enviar pull requests.

---

## Reportando Bugs

Antes de abrir uma issue, verifique se:

- O bug é reproduzível na versão mais recente
- Não existe uma issue já existente cobrindo o mesmo problema

Ao abrir uma issue, inclua:

- Sistema operacional e versão (ex: Ubuntu 24.04)
- Versão do Bash (`bash --version`)
- Versão do Docker e Docker Compose (`docker --version`, `docker compose version`)
- Passos para reproduzir o problema
- Comportamento esperado vs. comportamento atual
- Saída ou mensagens de erro relevantes

---

## Sugerindo Melhorias

Abra uma issue descrevendo:

- O problema que você está tentando resolver
- A solução que você propõe
- Por que isso seria útil para outros usuários

---

## Enviando Pull Requests

### 1. Fork e branch

```bash
git clone https://github.com/kaduvelasco/lumina-stack.git
cd lumina-stack
git checkout -b feature/minha-melhoria
```

### 2. Siga o estilo de código do projeto

Todos os scripts devem seguir as convenções documentadas no projeto:

- Shebang `#!/usr/bin/env bash` na primeira linha
- Bloco de cabeçalho com `Nome do Script`, `Descrição`, `Versão`
- `set -euo pipefail` em scripts de entrada
- Guard de carregamento (`[[ -n "${LIB_LOADED:-}" ]] && return 0`) em todos os arquivos de biblioteca
- Todas as variáveis de função declaradas com `local`
- Use `printf` em vez de `echo -e`
- Use as funções de saída de `lib/utils.sh` (`die`, `warn`, `info`, `success`)
- Menus usam loops `while true` — nunca recursão
- `mktemp` sempre acompanhado de `trap ... EXIT`
- Sem chamadas diretas a `apt-get`, `dnf` ou `pacman` — use `ensure_pkg` de `lib/utils.sh`

### 3. Execute o ShellCheck antes de enviar

```bash
shellcheck --severity=warning --shell=bash --exclude=SC1091 install.sh
shellcheck --severity=warning --shell=bash --exclude=SC1091 clean-docker.sh
shellcheck --severity=warning --shell=bash --exclude=SC1091 lib/utils.sh
shellcheck --severity=warning --shell=bash --exclude=SC1091 lib/versions.sh
shellcheck --severity=warning --shell=bash --exclude=SC1091 lib/menu.sh
shellcheck --severity=warning --shell=bash --exclude=SC1091 lib/system.sh
shellcheck --severity=warning --shell=bash --exclude=SC1091 lib/workspace.sh
shellcheck --severity=warning --shell=bash --exclude=SC1091 lib/docker.sh
```

Todos os arquivos devem passar com zero warnings antes de abrir o PR.

### 4. Atualize `lib/versions.sh` para mudanças de versão

Se sua mudança introduz ou atualiza uma versão de componente (PHP, Nginx, MariaDB), atualize as constantes em `lib/versions.sh`. Não coloque versões hardcoded diretamente em scripts ou templates.

### 5. Abra o Pull Request

Descreva no PR:

- Qual problema ele resolve
- O que foi alterado e por quê
- Como testar a mudança

---

## Código de Conduta

Seja respeitoso e construtivo. Contribuições de todos os níveis de experiência são bem-vindas.

---

Made with ❤️ and AI by [Kadu Velasco](https://github.com/kaduvelasco)
