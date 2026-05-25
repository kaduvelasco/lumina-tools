#!/usr/bin/env bash
# =============================================================================
# Script Name     : pos-install-popos.sh
# Description     : Post-install routine for Pop!_OS 24.04 LTS COSMIC
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
setup_fastfetch_repository() {
    info "Configurando repositório do Fastfetch..."
    sudo apt-get install -y -- software-properties-common
    sudo add-apt-repository -y ppa:zhangsongcui3371/fastfetch
}

# --- funções de negócio ---
run_post_install() {
    local log_file
    local -a apt_packages=(
        ubuntu-restricted-extras
        libavcodec-extra
        ffmpeg
        build-essential
        gparted
        gdebi
        libfuse2t64
        ntfs-3g
        unrar
        unzip
        p7zip-full
        curl
        wget
        git
        htop
        make
        tree
        jq
        plocate
        net-tools
        python3-pip
        fastfetch
    )
    local -a flatpak_packages=(
        org.videolan.VLC
    )

    require_not_root
    require_sudo
    require_internet
    assert_distro "pop"

    log_file="$(start_log "pos-install-popos")"

    show_header
    info "Log salvo em: ${log_file}"

    setup_fastfetch_repository

    info "Atualizando sistema base..."
    sudo apt-get update -y
    sudo apt-get full-upgrade -y

    info "Instalando pacotes APT..."
    sudo apt-get install -y -- "${apt_packages[@]}"

    info "Configurando fontes Microsoft..."
    printf '%s\n' \
        "ttf-mscorefonts-installer msttcorefonts/accepted-mscorefonts-eula select true" \
        | sudo debconf-set-selections
    sudo apt-get install -y -- ttf-mscorefonts-installer

    ensure_flatpak_ready
    info "Instalando Flatpaks essenciais..."
    flatpak install -y flathub "${flatpak_packages[@]}"

    configure_swappiness
    configure_inotify
    apply_sysctl

    if is_installed_cmd updatedb; then
        sudo updatedb
    fi

    success "Pós-instalação do Pop!_OS concluída."
    warn "Verifique atualizações do COSMIC regularmente."
    pause_screen
}

# --- ponto de entrada ---
main() {
    run_post_install
}

main "$@"
