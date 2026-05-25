#!/usr/bin/env bash
# =============================================================================
# Nome do Script : claude-install.sh
# Descrição      : Instalação do Claude Code CLI via instalador oficial
# Versão         : 2.0.0
# =============================================================================

# --- opções do shell ---
set -euo pipefail

# --- constantes ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
readonly CLAUDE_CMD="claude"

# --- carregamento de dependências ---
# shellcheck source=scripts/utils.sh
source "$SCRIPT_DIR/utils.sh"

# --- funções de negócio ---

# =============================================================================
# Instala o Claude Code CLI.
# =============================================================================
install_claude() {
    if is_installed_cmd "$CLAUDE_CMD"; then
        local current_version
        current_version=$(claude --version 2>/dev/null || echo "versão desconhecida")
        printf '%b\n' "${C2}✅ Claude Code já está instalado (${current_version}).${RESET}"
        echo -ne "   Deseja reinstalar / atualizar? (${C3}s${RESET}/N): "
        read -r confirm
        [[ ! "$confirm" =~ ^[sS]$ ]] && return 0
    fi

    if ! check_node_version; then
        echo -e "${C3}⚠️  O Claude Code requer o Node.js v${NODE_MIN_VERSION}+.${RESET}"
        echo -ne "   Deseja instalar agora? (${C3}S${RESET}/n): "
        read -r install_node_confirm
        [[ "$install_node_confirm" =~ ^[Nn]$ ]] && die "Instalação abortada."
        install_node
    fi

    require_internet

    printf '%b\n' "${C6}⚙️  Baixando instalador oficial da Anthropic...${RESET}"

    local installer
    installer=$(mktemp)
    trap 'rm -f "$installer"' EXIT
    if ! curl -fsSL https://claude.ai/install.sh -o "$installer"; then
        die "Falha ao baixar o instalador do Claude Code."
    fi

    printf '%b\n' "${C6}⚙️  Executando instalador...${RESET}"
    if bash "$installer"; then
        rm -f "$installer"
        trap - EXIT
        success "Claude Code instalado com sucesso!"
        printf '\n%b\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
        printf '%b\n' "   ${C4}Próximos passos:${RESET}"
        printf '%b\n' "   1. Execute ${C3}claude${RESET} no terminal para iniciar o login."
        printf '%b\n' "   2. Autentique com sua conta em ${C4}https://claude.ai${RESET}"
        printf '%b\n' "   3. Navegue até um projeto e execute ${C3}claude${RESET} para começar."
        printf '%b\n\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    else
        die "Erro durante a execução do instalador oficial."
    fi
}

# --- ponto de entrada ---

# =============================================================================
# Ponto de entrada principal do script.
# =============================================================================
main() {
    detect_pkg_manager
    install_claude
}

main "$@"
