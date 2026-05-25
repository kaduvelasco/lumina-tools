#!/usr/bin/env bash
# =============================================================================
# Nome do Script : windsurf-install.sh
# Descrição      : Instalação do Windsurf Editor
# Versão         : 2.1.0
# =============================================================================

# --- opções do shell ---
set -euo pipefail

# --- constantes ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
readonly WINDSURF_CMD="windsurf"
readonly WINDSURF_GPG_URL="https://windsurf-stable.codeiumdata.com/wVxQEIWkwPUEAGf3/windsurf.gpg"
readonly WINDSURF_APT_KEYRING="/etc/apt/keyrings/windsurf-stable.gpg"
readonly WINDSURF_APT_SOURCES="/etc/apt/sources.list.d/windsurf.list"
readonly WINDSURF_RPM_KEY="https://windsurf-stable.codeiumdata.com/wVxQEIWkwPUEAGf3/yum/RPM-GPG-KEY-windsurf"
readonly WINDSURF_RPM_REPO="/etc/yum.repos.d/windsurf.repo"

# --- carregamento de dependências ---
if [[ ! -f "$SCRIPT_DIR/../scripts/utils.sh" ]]; then
    printf '\033[0;31m❌ Erro fatal: scripts/utils.sh não encontrado.\033[0m\n' >&2
    exit 1
fi
# shellcheck source=scripts/utils.sh
source "$SCRIPT_DIR/../scripts/utils.sh"

# --- funções de interface ---

# =============================================================================
# Exibe o cabeçalho padrão com identificação do módulo.
# =============================================================================
show_header() {
    show_lumina_header
    printf '%b\n' "   ${C5}MÓDULO : ${C1}Windsurf Editor — Instalador${RESET}"
    printf '%b\n\n' "   ${C5}Distro : ${C4}${PKG_MANAGER}${RESET}"
}

# --- funções de negócio ---

# =============================================================================
# Instala o Windsurf via repositório apt (Ubuntu/Debian e derivados).
# =============================================================================
install_windsurf_apt() {
    ensure_pkg "wget"
    ensure_pkg "gpg"

    if [[ ! -f "$WINDSURF_APT_KEYRING" ]]; then
        printf '%b\n' "${C6}🔑 Adicionando chave GPG do repositório...${RESET}"
        wget -qO- "$WINDSURF_GPG_URL" | gpg --dearmor \
            | sudo install -D -o root -g root -m 644 /dev/stdin "$WINDSURF_APT_KEYRING"
    fi

    if [[ ! -f "$WINDSURF_APT_SOURCES" ]]; then
        printf '%b\n' "${C6}📋 Adicionando repositório apt...${RESET}"
        echo "deb [arch=amd64 signed-by=${WINDSURF_APT_KEYRING}] https://windsurf-stable.codeiumdata.com/wVxQEIWkwPUEAGf3/apt stable main" \
            | sudo tee "$WINDSURF_APT_SOURCES" > /dev/null
    fi

    sudo apt-get update -qq
    ensure_pkg "$WINDSURF_CMD"
}

# =============================================================================
# Instala o Windsurf via repositório dnf (Fedora e derivados).
# =============================================================================
install_windsurf_dnf() {
    if [[ ! -f "$WINDSURF_RPM_REPO" ]]; then
        printf '%b\n' "${C6}🔑 Importando chave GPG do repositório...${RESET}"
        sudo rpm --import "$WINDSURF_RPM_KEY"

        printf '%b\n' "${C6}📋 Adicionando repositório dnf...${RESET}"
        printf '[windsurf]\nname=Windsurf Repository\nbaseurl=https://windsurf-stable.codeiumdata.com/wVxQEIWkwPUEAGf3/yum/repo/\nenabled=1\nautorefresh=1\ngpgcheck=1\ngpgkey=%s\n' \
            "$WINDSURF_RPM_KEY" | sudo tee "$WINDSURF_RPM_REPO" > /dev/null
    fi

    sudo dnf check-update -q || true
    ensure_pkg "$WINDSURF_CMD"
}

# =============================================================================
# Instala o Windsurf via AUR (Arch Linux e derivados).
# =============================================================================
install_windsurf_arch() {
    if is_installed_cmd "yay"; then
        yay -S --noconfirm windsurf
    elif is_installed_cmd "paru"; then
        paru -S --noconfirm windsurf
    else
        die "AUR helper não encontrado (yay ou paru)."
    fi
}

# =============================================================================
# Ponto de entrada principal do script.
# =============================================================================
main() {
    detect_pkg_manager
    show_header

    if is_installed_cmd "$WINDSURF_CMD"; then
        printf '%b\n' "${C2}✅ Windsurf já está instalado.${RESET}"
        echo -ne "   Deseja reinstalar / atualizar? (s/${C3}N${RESET}): "
        read -r confirm
        [[ ! "$confirm" =~ ^[sS]$ ]] && return 0
    fi

    printf '%b\n' "${C6}⚙️  Instalando Windsurf...${RESET}"

    case "$PKG_MANAGER" in
        apt)    install_windsurf_apt ;;
        dnf)    install_windsurf_dnf ;;
        pacman) install_windsurf_arch ;;
        *)      die "Gerenciador de pacotes '${PKG_MANAGER}' não suportado." ;;
    esac

    success "Windsurf instalado com sucesso."

    printf '\n%b\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    printf '%b\n' "   ${C4}Próximos passos:${RESET}"
    printf '%b\n' "   1. Inicie o Windsurf: ${C3}windsurf${RESET}"
    printf '%b\n\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
}

main "$@"
