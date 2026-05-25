#!/usr/bin/env bash
# =============================================================================
# Script Name     : pos-install-zorin.sh
# Description     : Post-install routine for ZorinOS 18.1
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

# --- funções de negócio ---
run_post_install() {
    local log_file
    local -a apt_packages=(
        # codecs and media support
        libavcodec-extra
        ffmpeg
        gstreamer1.0-plugins-bad
        gstreamer1.0-plugins-ugly
        gstreamer1.0-libav
        # gnome tools
        gnome-tweaks
        gnome-shell-extension-manager
        # essential tools
        build-essential
        git
        curl
        wget
        htop
        fastfetch
        # utilities
        gparted
        gdebi
        libfuse2t64
        unrar
        unzip
        ntfs-3g
        p7zip-full
        tree
        jq
        plocate
        net-tools
    )
    local -a flatpak_packages=(
        org.videolan.VLC
    )

    require_not_root
    require_sudo
    require_internet
    assert_distro "zorin"

    log_file="$(start_log "pos-install-zorin")"

    show_header
    info "Log salvo em: ${log_file}"

    info "Atualizando sistema base..."
    sudo apt-get update -y
    sudo apt-get full-upgrade -y

    info "Instalando pacotes APT..."
    sudo apt-get install -y -- "${apt_packages[@]}"

    info "Configurando fontes Microsoft e mídias restritas..."
    printf '%s\n' \
        "ttf-mscorefonts-installer msttcorefonts/accepted-mscorefonts-eula select true" \
        | sudo debconf-set-selections
    sudo apt-get install -y -- zorin-os-restricted-extras

    configure_swappiness
    configure_inotify
    apply_sysctl

    ensure_flatpak_ready
    info "Instalando Flatpaks essenciais..."
    flatpak install -y flathub "${flatpak_packages[@]}"

    if is_installed_cmd updatedb; then
        sudo updatedb
    fi

    success "Pós-instalação do ZorinOS concluída."
    warn "Reinicie o sistema para aplicar todas as mudanças."
    pause_screen
}

# --- ponto de entrada ---
main() {
    run_post_install
}

main "$@"
