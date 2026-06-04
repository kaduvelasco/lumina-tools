#!/usr/bin/env bash
# =============================================================================
# Nome do Script : codex-install.sh
# Descrição      : Instalação e atualização do OpenAI Codex CLI via npm
# Versão         : 1.0.0
# =============================================================================

# --- opções do shell ---
set -euo pipefail

# --- constantes ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
readonly CODEX_PKG="codex-cli"
readonly CODEX_CMD="codex"

# --- carregamento de dependências ---
# shellcheck source=scripts/utils.sh
source "$SCRIPT_DIR/utils.sh"

# --- funções de negócio ---

# =============================================================================
# Instala ou atualiza o OpenAI Codex CLI via npm.
# =============================================================================
install_codex() {
    if is_installed_cmd "$CODEX_CMD"; then
        local current_version
        current_version=$(codex --version 2>/dev/null || echo "versão desconhecida")
        printf '%b\n' "${C2}✅ OpenAI Codex CLI já está instalado (${current_version}).${RESET}"
        echo -ne "   Deseja atualizar? (${C3}s${RESET}/N): "
        read -r confirm
        if [[ ! "$confirm" =~ ^[sS]$ ]]; then
            return 0
        fi

        printf '%b\n' "${C6}⚙️  Atualizando ${CODEX_PKG}...${RESET}"
        if sudo env PATH="$PATH" npm update -g "$CODEX_PKG"; then
            local new_version
            new_version=$(codex --version 2>/dev/null || echo "versão desconhecida")
            success "OpenAI Codex CLI atualizado para ${new_version}."
        else
            die "Falha ao atualizar via npm. Verifique as permissões e tente novamente."
        fi
        return 0
    fi

    if ! check_node_version; then
        printf '%b\n' "${C3}⚠️  O OpenAI Codex CLI requer o Node.js v${NODE_MIN_VERSION}+.${RESET}"
        echo -ne "   Deseja instalar agora? (${C3}S${RESET}/n): "
        read -r install_node_confirm
        [[ "$install_node_confirm" =~ ^[Nn]$ ]] && die "Instalação abortada."
        install_node
    fi

    require_internet

    printf '%b\n' "${C6}⚙️  Instalando ${CODEX_PKG} globalmente...${RESET}"

    if sudo env PATH="$PATH" npm install -g "$CODEX_PKG"; then
        success "OpenAI Codex CLI instalado com sucesso!"
        printf '\n%b\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
        printf '%b\n' "   ${C4}Próximos passos:${RESET}"
        printf '%b\n' "   1. Obtenha sua chave em: ${C4}https://platform.openai.com/api-keys${RESET}"
        printf '%b\n' "   2. Exporte a variável: ${C3}export OPENAI_API_KEY=<sua-chave>${RESET}"
        printf '%b\n' "   3. Execute ${C3}codex${RESET} no terminal para começar."
        printf '%b\n\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    else
        die "Falha ao instalar via npm. Verifique as permissões e tente novamente."
    fi
}

# --- ponto de entrada ---

# =============================================================================
# Ponto de entrada principal do script.
# =============================================================================
main() {
    detect_pkg_manager
    install_codex
}

main "$@"
