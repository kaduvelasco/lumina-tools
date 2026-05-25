#!/usr/bin/env bash
# =============================================================================
# Script Name     : pos-install-fedora.sh
# Description     : Post-install routine for Fedora 44
# Version         : 1.0.0
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR

for _lib in utils.sh system.sh; do
    if [[ ! -f "$SCRIPT_DIR/../lib/${_lib}" ]]; then
        printf '\033[0;31m❌ Erro fatal: ../lib/%s não encontrado.\033[0m\n' "$_lib" >&2
        exit 1
    fi
    # shellcheck source=/dev/null
    source "$SCRIPT_DIR/../lib/${_lib}"
done
unset _lib

# --- funções de interface ---
show_header() {
    show_lumina_header
}

# --- funções auxiliares ---
enable_rpmfusion() {
    info "Habilitando repositórios RPM Fusion..."
    local fedora_version
    fedora_version="$(rpm -E %fedora)"
    sudo dnf install -y \
        "https://download1.rpmfusion.org/free/fedora/rpmfusion-free-release-${fedora_version}.noarch.rpm" \
        "https://download1.rpmfusion.org/nonfree/fedora/rpmfusion-nonfree-release-${fedora_version}.noarch.rpm"
}

install_multimedia_codecs() {
    info "Instalando codecs multimídia..."
    sudo dnf group install -y multimedia
    sudo dnf group install -y sound-and-video
}

install_ms_fonts() {
    info "Instalando fontes Microsoft..."
    sudo dnf install -y -- curl cabextract xorg-x11-font-utils fontconfig
    sudo dnf install -y \
        "https://downloads.sourceforge.net/corefonts/the%20fonts/fetchmsttfonts-11.0-17.noarch.rpm" \
        || warn "Falha ao instalar fontes Microsoft. Continue manualmente se necessário."
}

# --- funções de negócio ---
run_post_install() {
    local log_file
    local -a dnf_packages=(
        git
        curl
        wget
        htop
        fastfetch
        make
        tree
        jq
        p7zip
        p7zip-plugins
        unrar
        net-tools
        ntfs-3g
        plocate
        python3-pip
    )

    require_not_root
    require_sudo
    require_internet
    assert_distro "fedora"

    log_file="$(start_log "pos-install-fedora")"

    show_header
    info "Log salvo em: ${log_file}"

    info "Atualizando sistema base..."
    sudo dnf upgrade --refresh -y

    enable_rpmfusion
    install_multimedia_codecs

    info "Instalando pacotes DNF..."
    sudo dnf install -y -- "${dnf_packages[@]}"

    install_ms_fonts

    ensure_flatpak_ready

    configure_swappiness
    configure_inotify
    apply_sysctl

    info "Ativando TRIM para SSDs..."
    sudo systemctl enable --now fstrim.timer

    info "Limpando pacotes desnecessários..."
    sudo dnf autoremove -y
    sudo dnf clean all

    if is_installed_cmd updatedb; then
        sudo updatedb
    fi

    success "Pós-instalação do Fedora 44 concluída."
    warn "Reinicie o sistema para aplicar todas as mudanças."
    pause_screen
}

# --- ponto de entrada ---
main() {
    run_post_install
}

main "$@"
