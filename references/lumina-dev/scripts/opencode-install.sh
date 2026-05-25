#!/usr/bin/env bash
# =============================================================================
# Nome do Script : opencode-install.sh
# Descrição      : Instalação do OpenCode CLI e Desktop
# Versão         : 2.0.0
# =============================================================================

# --- opções do shell ---
set -euo pipefail

# --- constantes ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
readonly OPENCODE_REPO="anomalyco/opencode"
readonly GITHUB_API="https://api.github.com/repos/${OPENCODE_REPO}/releases/latest"

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
    printf '%b\n' "   ${C5}INSTALADOR OPENCODE${RESET}"
    printf '%b\n\n' "   ${C5}Distro : ${C4}${PKG_MANAGER}${RESET}"
}

# =============================================================================
# Exibe o menu de instalação do OpenCode.
# =============================================================================
show_opencode_menu() {
    while true; do
        show_header
        printf '%b\n' "OPENCODE - CLI & DESKTOP"
        printf '%b\n' ""
        printf '%b\n' "  ${C2}1.${RESET} Instalar CLI + Desktop"
        printf '%b\n' "  ${C2}2.${RESET} Apenas OpenCode CLI"
        printf '%b\n' "  ${C2}3.${RESET} Apenas OpenCode Desktop"
        printf '%b\n' "  ${C1}0.${RESET} Voltar"
        printf '%b\n' ""
        echo -ne "Selecione uma opção: "
        read -r choice

        case "$choice" in
            1) install_opencode_cli; install_opencode_desktop ;;
            2) install_opencode_cli ;;
            3) install_opencode_desktop ;;
            0) return 0 ;;
            *) printf '%b\n' "${C1}❌ Opção inválida.${RESET}" ;;
        esac
    done
}

# --- funções auxiliares ---

# =============================================================================
# Obtém a versão mais recente do OpenCode via GitHub API.
# =============================================================================
get_opencode_latest_version() {
    local version
    if is_installed_cmd "jq"; then
        version=$(curl -fsSL "$GITHUB_API" 2>/dev/null | jq -r '.tag_name // empty' | sed 's/^v//')
    else
        version=$(curl -fsSL "$GITHUB_API" 2>/dev/null | grep '"tag_name"' | sed 's/.*"tag_name":[[:space:]]*"v\([^"]*\)".*/\1/')
    fi

    if [[ -z "$version" ]]; then
        die "Não foi possível obter a versão mais recente do OpenCode. Verifique sua conexão."
    fi

    echo "$version"
}

# --- funções de negócio ---

# =============================================================================
# Instala o OpenCode CLI via npm.
# =============================================================================
install_opencode_cli() {
    printf '\n%b\n' "${C6}⚙️  Instalando OpenCode CLI...${RESET}"

    if is_installed_cmd "opencode"; then
        printf '%b\n' "${C2}✅ OpenCode CLI já está instalado.${RESET}"
        echo -ne "   Reinstalar / Atualizar? (${C3}s${RESET}/N): "
        read -r confirm
        [[ ! "$confirm" =~ ^[sS]$ ]] && return 0
    fi

    if ! check_node_version; then
        printf '%b\n' "${C3}⚠️  O OpenCode CLI requer o Node.js v${NODE_MIN_VERSION}+.${RESET}"
        echo -ne "   Instalar Node.js agora? (${C3}S${RESET}/n): "
        read -r confirm
        [[ "$confirm" =~ ^[Nn]$ ]] && return 1
        install_node
    fi

    printf '%b\n' "${C6}⚙️  Instalando opencode-ai via npm...${RESET}"
    local npm_prefix
    npm_prefix=$(npm config get prefix 2>/dev/null || echo "/usr/local")
    sudo rm -rf "${npm_prefix}/lib/node_modules/opencode-ai" 2>/dev/null || true
    if sudo env PATH="$PATH" npm install -g opencode-ai@latest; then
        success "OpenCode CLI instalado com sucesso!"
    else
        die "Falha ao instalar OpenCode CLI via npm."
    fi
}

# =============================================================================
# Instala o OpenCode Desktop via GitHub Releases.
# =============================================================================
install_opencode_desktop() {
    printf '\n%b\n' "${C6}🖥️  Instalando OpenCode Desktop...${RESET}"
    ensure_pkg "curl"

    local arch version base_url pkg_file tmp_file
    arch=$(uname -m)
    version=$(get_opencode_latest_version)
    base_url="https://github.com/${OPENCODE_REPO}/releases/download/v${version}"

    case "$PKG_MANAGER" in
        apt)
            case "$arch" in
                x86_64)        pkg_file="opencode-desktop-linux-amd64.deb" ;;
                aarch64|arm64) pkg_file="opencode-desktop-linux-arm64.deb" ;;
                *)             die "Arquitetura '${arch}' não suportada para .deb." ;;
            esac
            tmp_file=$(mktemp --suffix=".deb")
            trap 'rm -f "$tmp_file"' EXIT
            printf '%b\n' "${C6}📥 Baixando ${pkg_file}...${RESET}"
            curl -fSL "${base_url}/${pkg_file}" -o "$tmp_file"
            sudo apt-get install -y "$tmp_file"
            rm -f -- "$tmp_file"
            trap - EXIT
            ;;
        dnf)
            case "$arch" in
                x86_64)        pkg_file="opencode-desktop-linux-x86_64.rpm" ;;
                aarch64|arm64) pkg_file="opencode-desktop-linux-aarch64.rpm" ;;
                *)             die "Arquitetura '${arch}' não suportada para .rpm." ;;
            esac
            tmp_file=$(mktemp --suffix=".rpm")
            trap 'rm -f "$tmp_file"' EXIT
            printf '%b\n' "${C6}📥 Baixando ${pkg_file}...${RESET}"
            curl -fSL "${base_url}/${pkg_file}" -o "$tmp_file"
            sudo dnf install -y "$tmp_file"
            rm -f -- "$tmp_file"
            trap - EXIT
            ;;
        pacman)
            ensure_pkg "fuse2"
            local appimage_file
            case "$arch" in
                x86_64)        appimage_file="opencode-electron-linux-x86_64.AppImage" ;;
                aarch64|arm64) appimage_file="opencode-electron-linux-aarch64.AppImage" ;;
                *)             die "Arquitetura '${arch}' não suportada para AppImage." ;;
            esac
            local install_path="$HOME/.local/bin/opencode-desktop"
            printf '%b\n' "${C6}📥 Baixando AppImage para Arch Linux...${RESET}"
            curl -fSL "${base_url}/${appimage_file}" -o "$install_path"
            chmod +x "$install_path"
            ;;
    esac

    success "OpenCode Desktop instalado com sucesso."
}

# --- ponto de entrada ---

# =============================================================================
# Ponto de entrada principal do script.
# =============================================================================

# Chamado com argumento por lumina-dev.sh; sem argumento exibe menu standalone.
main() {
    detect_pkg_manager
    require_not_root
    require_sudo
    require_internet

    case "${1:-menu}" in
        cli)     install_opencode_cli ;;
        desktop) install_opencode_desktop ;;
        *)       show_opencode_menu ;;
    esac

    printf '\n%b\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    printf '%b\n' "   ${C4}Próximos passos:${RESET}"
    printf '%b\n' "   1. Execute ${C3}opencode${RESET} para iniciar o CLI"
    printf '%b\n\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
}

main "$@"
