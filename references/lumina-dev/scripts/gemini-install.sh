#!/usr/bin/env bash
# =============================================================================
# Nome do Script : gemini-install.sh
# Descrição      : Instalação e configuração do Gemini Code Assist CLI
# Versão         : 2.0.0
# =============================================================================

# --- opções do shell ---
set -euo pipefail

# --- constantes ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
readonly GEMINI_PKG="@google/gemini-cli"
readonly GEMINI_CMD="gemini"

# --- carregamento de dependências ---
# shellcheck source=scripts/utils.sh
source "$SCRIPT_DIR/utils.sh"

# --- funções de negócio ---

# =============================================================================
# Instala o Gemini CLI via npm.
# =============================================================================
install_gemini() {
    if is_installed_cmd "$GEMINI_CMD"; then
        local current_version
        current_version=$(gemini --version 2>/dev/null || echo "versão desconhecida")
        printf '%b\n' "${C2}✅ Gemini CLI já está instalado (${current_version}).${RESET}"
        echo -ne "   Deseja reinstalar / atualizar? (${C3}s${RESET}/N): "
        read -r confirm
        [[ ! "$confirm" =~ ^[sS]$ ]] && return 0
    fi

    if ! check_node_version; then
        echo -e "${C3}⚠️  O Gemini CLI requer o Node.js v${NODE_MIN_VERSION}+.${RESET}"
        echo -ne "   Deseja instalar agora? (${C3}S${RESET}/n): "
        read -r install_node_confirm
        [[ "$install_node_confirm" =~ ^[Nn]$ ]] && die "Instalação abortada."
        install_node
    fi

    require_internet

    printf '%b\n' "${C6}⚙️  Instalando ${GEMINI_PKG} globalmente...${RESET}"

    if sudo env PATH="$PATH" npm install -g "$GEMINI_PKG"; then
        success "Gemini CLI instalado com sucesso!"
    else
        die "Falha ao instalar via npm. Verifique as permissões e tente novamente."
    fi
}

# =============================================================================
# Configura a API Key do Gemini.
# =============================================================================
configure_gemini() {
    printf '\n%b\n' "${C5}🔑 Configuração do Gemini Code Assist${RESET}"
    printf '%b\n' "${C5}──────────────────────────────────${RESET}"
    echo -e "   📌 Obtenha sua chave em: ${C4}https://aistudio.google.com/apikey${RESET}"
    echo ""
    echo -ne "   Cole sua GOOGLE_API_KEY aqui (Enter para pular): "
    read -r api_key

    if [[ -z "$api_key" ]]; then
        warn "Configuração da API Key ignorada."
        return 0
    fi

    export GOOGLE_API_KEY="$api_key"

    local bashrc_local="$HOME/.bashrc.local"

    if [[ ! -f "$bashrc_local" ]]; then
        (umask 077; touch "$bashrc_local")
    fi

    grep -vF "GOOGLE_API_KEY" "$bashrc_local" 2>/dev/null > "${bashrc_local}.tmp" || true
    mv -- "${bashrc_local}.tmp" "$bashrc_local"
    chmod 600 "$bashrc_local"

    printf 'export GOOGLE_API_KEY=%q\n' "$api_key" >> "$bashrc_local"
    success "GOOGLE_API_KEY configurada em ${bashrc_local}."

    local source_line="[[ -f \"\$HOME/.bashrc.local\" ]] && source \"\$HOME/.bashrc.local\""
    if ! grep -qF ".bashrc.local" "$HOME/.bashrc" 2>/dev/null; then
        printf '\n# LuminaDev — configurações locais\n%s\n' "$source_line" >> "$HOME/.bashrc"
    fi

    printf '%b\n' "${C6}🔍 Verificando instalação do Gemini CLI...${RESET}"

    local version_output
    version_output=$(gemini --version 2>&1 || true)

    if [[ -z "$version_output" ]] || printf '%s' "$version_output" | grep -qF "ENOENT"; then
        warn "Aplicando correção de diretório de configuração..."
        mkdir -p "$HOME/.gemini"
        if [[ ! -f "$HOME/.gemini/projects.json" ]]; then
            echo '{"projects":[]}' > "$HOME/.gemini/projects.json"
        fi
        version_output=$(gemini --version 2>&1 || echo "erro")
    fi

    printf '%b\n' "${C2}✅ Gemini CLI operacional. Versão: ${C4}${version_output}${RESET}"
}

# --- ponto de entrada ---

# =============================================================================
# Ponto de entrada principal do script.
# =============================================================================
main() {
    detect_pkg_manager
    install_gemini

    printf '\n%b\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    printf '%b\n' "   ${C4}Próximos passos (instalação):${RESET}"
    printf '%b\n' "   1. Acesse ${C4}https://aistudio.google.com/apikey${RESET}"
    printf '%b\n' "   2. Configure agora ou adicione ao ${C4}~/.bashrc${RESET}"
    printf '%b\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"

    configure_gemini
}

main "$@"
