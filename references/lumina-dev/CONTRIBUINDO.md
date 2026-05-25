# Contribuindo com o LuminaDev

📄 English version: see [CONTRIBUTING.md](CONTRIBUTING.md)

Obrigado pelo interesse em contribuir com o LuminaDev! Este documento descreve os padrões e o processo para contribuição.

---

## Como Contribuir

1. Faça um fork do repositório
2. Crie uma branch: `git checkout -b feature/minha-melhoria`
3. Siga os padrões de código descritos abaixo
4. Certifique-se de que o ShellCheck passa sem avisos
5. Abra um Pull Request descrevendo o que foi alterado e por quê

---

## Ambiente de Desenvolvimento

**Requisitos:**

- Bash 4.0+
- ShellCheck (para lint local)

**Instalando o ShellCheck:**

```bash
sudo apt install shellcheck   # Ubuntu/Debian
sudo dnf install ShellCheck   # Fedora
sudo pacman -S shellcheck     # Arch
```

**Executando o linter localmente:**

```bash
find . -name "*.sh" | xargs shellcheck --severity=warning --shell=bash --exclude=SC1091
```

**Validando sintaxe:**

```bash
bash -n lumina-dev.sh
bash -n scripts/utils.sh
```

---

## Padrões de Código

Todos os scripts devem seguir as convenções do guia de estilo do projeto. Regras principais:

### Estrutura do Script

Todo script segue exatamente esta ordem de seções:

1. Shebang: `#!/usr/bin/env bash`
2. Cabeçalho com nome, descrição, versão
3. `set -euo pipefail`
4. `readonly SCRIPT_DIR=...`
5. Guarda de existência e `source` do `utils.sh`
6. Funções de interface (`show_header`)
7. Funções auxiliares
8. Funções de negócio
9. Ponto de entrada `main()`
10. `main "$@"`

### Opções do Shell

Obrigatório em todos os scripts, imediatamente após o cabeçalho:

```bash
set -euo pipefail
```

### Variáveis

- Constantes de script usam `readonly`
- Variáveis dentro de funções usam `local`
- Nunca use variáveis globais implícitas dentro de funções

### Saída Padronizada

Use exclusivamente as funções de `utils.sh` para mensagens de status:

```bash
die "mensagem de erro"    # stderr + exit 1
warn "aviso"              # stderr, sem sair
info "informação"         # stdout, informativo
success "confirmação"     # stdout, sucesso
```

Para mensagens de progresso com ícones customizados, use `printf '%b\n'` em vez de `echo -e`:

```bash
printf '%b\n' "${C6}⚙️  Instalando...${RESET}"
```

### Menus

Menus interativos usam `while true` com `break` ou `return 0` para sair. Nunca use recursão.

### Arquivos Temporários

Todo `mktemp` deve ser acompanhado de `trap 'rm ...' EXIT`. Sempre faça limpeza explícita e `trap - EXIT` ao final da função:

```bash
local tmp
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
# ... trabalho ...
rm -f "$tmp"
trap - EXIT
```

### Instalação de Pacotes

Nunca chame `apt-get`, `dnf` ou `pacman` diretamente nos scripts de módulo. Use as abstrações do `utils.sh`:

```bash
detect_pkg_manager   # chamar no início de main()
ensure_pkg "nome"    # instala se não instalado
```

### Idempotência

Toda função de instalação deve verificar se a ferramenta já está instalada e oferecer pular ou reinstalar:

```bash
if is_installed_cmd "ferramenta"; then
    printf '%b\n' "${C2}✅ ferramenta já está instalada.${RESET}"
    echo -ne "   Reinstalar / Atualizar? (${C3}s${RESET}/N): "
    read -r confirm
    [[ ! "$confirm" =~ ^[sS]$ ]] && return 0
fi
```

---

## Adicionando um Novo Script

Ao adicionar um novo script instalador:

1. Coloque em `scripts/` (ferramentas CLI) ou `ides/` (editores)
2. Faça `source` do `utils.sh` pelo caminho relativo correto
3. Chame `detect_pkg_manager` no início de `main()`
4. Implemente `require_not_root` e `require_sudo` se o script usa `sudo`
5. Adicione o script à lista de verificação do CI em `.github/workflows/lint.yml`
6. Adicione uma entrada no menu principal em `lumina-dev.sh`
7. Adicione a função de remoção correspondente em `scripts/uninstall.sh`
8. Atualize o `README.md` e o `LEIAME.md` para documentar o novo módulo

---

## Processo de Pull Request

- Mantenha PRs focados em uma única alteração ou funcionalidade
- Descreva o que foi alterado e por quê no corpo do PR
- Garanta que o ShellCheck passa: `find . -name "*.sh" | xargs shellcheck --severity=warning --shell=bash --exclude=SC1091`
- Teste em pelo menos uma distro suportada antes de enviar

---

Made with ❤️ and AI by [Kadu Velasco](https://github.com/kaduvelasco)
