#!/usr/bin/env bash
# =============================================================================
# Nome do Script : mcp-install.sh
# Descrição      : Instalação de Servidores MCP (Model Context Protocol)
# Versão         : 2.0.0
# =============================================================================

# --- opções do shell ---
set -euo pipefail

# --- constantes ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR

# --- carregamento de dependências ---
if [[ ! -f "$SCRIPT_DIR/utils.sh" ]]; then
    printf '\033[0;31m❌ Erro fatal: utils.sh não encontrado. Abortando.\033[0m\n' >&2
    exit 1
fi
# shellcheck source=scripts/utils.sh
source "$SCRIPT_DIR/utils.sh"

# --- funções de interface ---

# =============================================================================
# Exibe o cabeçalho padrão com identificação do módulo.
# =============================================================================
show_header() {
    show_lumina_header
    printf '%b\n' "   ${C5}INSTALADOR DE SERVIDORES MCP${RESET}"
    printf '%b\n\n' "   ${C5}Distro : ${C4}${PKG_MANAGER}${RESET}"
}

# =============================================================================
# Exibe o menu de servidores MCP.
# =============================================================================
show_mcp_menu() {
    while true; do
        show_header
        printf '%b\n' "SERVIDORES MCP"
        printf '%b\n' ""
        printf '%b\n' "  ${C2}1.${RESET} Moodle Dev MCP"
        printf '%b\n' "  ${C2}2.${RESET} Lumina AI Vault"
        printf '%b\n' "  ${C2}3.${RESET} Code Review Graph"
        printf '%b\n' "  ${C1}0.${RESET} Voltar"
        printf '%b\n' ""
        echo -ne "Selecione uma opção: "
        read -r choice

        case "$choice" in
            1) install_moodle_dev_mcp ;;
            2) install_lumina_ai_vault ;;
            3) install_code_review_graph ;;
            0) return 0 ;;
            *) printf '%b\n' "${C1}❌ Opção inválida.${RESET}" ;;
        esac
    done
}

# --- funções auxiliares ---

# =============================================================================
# Verifica se o Node.js está instalado e atende aos requisitos.
# =============================================================================
check_nodejs_mcp() {
    if ! check_node_version; then
        printf '%b\n' "${C3}⚠️  Servidores MCP requerem Node.js v${NODE_MIN_VERSION}+.${RESET}"
        echo -ne "   Instalar Node.js agora? (${C3}S${RESET}/n): "
        read -r confirm
        [[ "$confirm" =~ ^[Nn]$ ]] && return 1
        install_node
    fi
    return 0
}

# --- funções de negócio ---

# =============================================================================
# Instala o Moodle Dev MCP.
# =============================================================================
install_moodle_dev_mcp() {
    printf '\n%b\n' "${C6}⚙️  Configurando Moodle Dev MCP...${RESET}"
    check_nodejs_mcp || return 1

    if npm list -g moodle-dev-mcp &>/dev/null; then
        printf '%b\n' "${C2}✅ moodle-dev-mcp já está instalado.${RESET}"
        echo -ne "   Reinstalar / Atualizar? (${C3}s${RESET}/N): "
        read -r confirm
        [[ ! "$confirm" =~ ^[sS]$ ]] && return 0
    fi

    printf '%b\n' "${C6}⚙️  Instalando moodle-dev-mcp via npm...${RESET}"
    if ! sudo env PATH="$PATH" npm install -g moodle-dev-mcp@latest; then
        die "Falha ao instalar moodle-dev-mcp."
    fi

    echo -ne "\nCaminho da instalação do Moodle (ex: /var/www/moodle): "
    read -r moodle_path
    moodle_path="${moodle_path#"${moodle_path%%[![:space:]]*}"}"
    moodle_path="${moodle_path%"${moodle_path##*[![:space:]]}"}"

    if [[ -n "$moodle_path" ]]; then
        if [[ ! "$moodle_path" =~ ^[a-zA-Z0-9/_.-]+$ ]]; then
            warn "Caminho contém caracteres inválidos. Registro ignorado."
        elif is_installed_cmd "claude"; then
            printf '%b\n' "${C6}🔌 Registrando no Claude Code...${RESET}"
            claude mcp add moodle-dev-mcp -e "MOODLE_PATH=${moodle_path}" -- npx -y moodle-dev-mcp || \
                warn "Falha no registro automático. Registre manualmente."
        fi
    fi

    success "Moodle Dev MCP instalado com sucesso."
}

# =============================================================================
# Instala o Lumina AI Vault.
# =============================================================================
install_lumina_ai_vault() {
    printf '\n%b\n' "${C6}⚙️  Configurando Lumina AI Vault...${RESET}"
    check_nodejs_mcp || return 1

    if npm list -g lumina-ai-vault &>/dev/null; then
        printf '%b\n' "${C2}✅ lumina-ai-vault já está instalado.${RESET}"
        echo -ne "   Reinstalar / Atualizar? (${C3}s${RESET}/N): "
        read -r confirm
        [[ ! "$confirm" =~ ^[sS]$ ]] && return 0
    fi

    printf '%b\n' "${C6}⚙️  Instalando lumina-ai-vault via npm...${RESET}"
    if ! sudo env PATH="$PATH" npm install -g lumina-ai-vault@latest; then
        die "Falha ao instalar lumina-ai-vault."
    fi

    echo -ne "\nUsar caminho personalizado para o vault? (Vazio para padrão): "
    read -r vault_path
    vault_path="${vault_path#"${vault_path%%[![:space:]]*}"}"
    vault_path="${vault_path%"${vault_path##*[![:space:]]}"}"

    if is_installed_cmd "claude"; then
        printf '%b\n' "${C6}🔌 Registrando no Claude Code...${RESET}"
        if [[ -n "$vault_path" ]]; then
            if [[ ! "$vault_path" =~ ^[a-zA-Z0-9/_.-]+$ ]]; then
                warn "Caminho contém caracteres inválidos. Registrando sem caminho personalizado."
                claude mcp add lumina-aivault -- npx lumina-ai-vault || true
            else
                claude mcp add lumina-aivault -- npx lumina-ai-vault "$vault_path" || true
            fi
        else
            claude mcp add lumina-aivault -- npx lumina-ai-vault || true
        fi
    fi

    success "Lumina AI Vault instalado com sucesso."
}

# =============================================================================
# Instala o Code Review Graph MCP.
# =============================================================================
install_code_review_graph() {
    printf '\n%b\n' "${C6}⚙️  Configurando Code Review Graph MCP...${RESET}"

    # Garante ~/.local/bin no PATH desta sessão
    export PATH="$HOME/.local/bin:$PATH"

    # --- Passo 1: UV ---
    if ! is_installed_cmd "uv"; then
        printf '%b\n' "${C6}📥 Instalando UV...${RESET}"
        local uv_installer
        uv_installer=$(mktemp)
        trap 'rm -f "$uv_installer"' EXIT
        if ! curl -LsSf https://astral.sh/uv/install.sh -o "$uv_installer"; then
            die "Falha ao baixar instalador do UV."
        fi
        sh "$uv_installer"
        rm -f "$uv_installer"
        trap - EXIT
        if ! is_installed_cmd "uv"; then
            die "UV não encontrado após instalação. Reinicie o terminal e tente novamente."
        fi
        success "UV instalado com sucesso."
    else
        printf '%b\n' "${C2}✅ UV já está instalado.${RESET}"
    fi

    # --- Passo 2: pipx ---
    if ! is_installed_cmd "pipx"; then
        printf '%b\n' "${C6}📥 Instalando pipx...${RESET}"
        case "$PKG_MANAGER" in
            apt)
                sudo apt-get update -qq
                ensure_pkg "pipx"
                ;;
            dnf)
                ensure_pkg "pipx"
                ;;
            pacman)
                ensure_pkg "python-pipx"
                ;;
        esac
        pipx ensurepath
        success "pipx instalado com sucesso."
    else
        printf '%b\n' "${C2}✅ pipx já está instalado.${RESET}"
    fi

    # --- Passo 3: code-review-graph ---
    if pipx list 2>/dev/null | grep -qF "code-review-graph"; then
        printf '%b\n' "${C2}✅ code-review-graph já está instalado.${RESET}"
        printf '%s' "   Reinstalar / Atualizar? (${C3}s${RESET}/N): "
        read -r confirm
        [[ ! "$confirm" =~ ^[sS]$ ]] && return 0
        pipx upgrade code-review-graph || pipx install code-review-graph
    else
        printf '%b\n' "${C6}📥 Instalando code-review-graph via pipx...${RESET}"
        if ! pipx install code-review-graph; then
            die "Falha ao instalar code-review-graph."
        fi
    fi

    # --- Passo 4: configurar plataformas suportadas ---
    printf '%b\n' "${C6}⚙️  Configurando plataformas suportadas...${RESET}"
    if ! code-review-graph install; then
        warn "Falha na configuração automática. Execute manualmente: code-review-graph install"
    fi

    success "Code Review Graph MCP instalado e configurado com sucesso."
}

# --- ponto de entrada ---

# =============================================================================
# Ponto de entrada principal do script.
# Chamado com argumento por lumina-dev.sh; sem argumento exibe menu standalone.
# =============================================================================
main() {
    detect_pkg_manager
    require_not_root
    require_sudo
    require_internet

    case "${1:-menu}" in
        moodle-dev-mcp)    install_moodle_dev_mcp ;;
        lumina-ai-vault)   install_lumina_ai_vault ;;
        code-review-graph) install_code_review_graph ;;
        *)                 show_mcp_menu ;;
    esac
}

main "$@"
