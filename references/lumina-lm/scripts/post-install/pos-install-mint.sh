#!/usr/bin/env bash
# =============================================================================
# Script Name     : pos-install-mint.sh
# Description     : Post-install routine for Linux Mint 22.3
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
enable_base_repositories() {
    local repo

    info "Habilitando repositórios base necessários..."
    sudo apt-get install -y -- software-properties-common

    for repo in universe multiverse; do
        sudo add-apt-repository -y "${repo}"
    done
}

# --- funções de negócio ---
run_post_install() {
    local log_file
    local -a apt_packages=(
        mint-meta-codecs
        ubuntu-drivers-common
        libavcodec-extra
        ffmpeg
        build-essential
        gparted
        gdebi
        libfuse2t64
        unrar
        unzip
        ntfs-3g
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
        net.codelogistics.webapps
    )

    require_not_root
    require_sudo
    require_internet
    assert_distro "linuxmint"

    log_file="$(start_log "pos-install-mint")"

    show_header
    info "Log salvo em: ${log_file}"
    enable_base_repositories

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

    if is_installed_cmd ubuntu-drivers; then
        sudo ubuntu-drivers autoinstall || warn "Nenhum driver adicional foi necessário."
    fi

    if is_installed_cmd updatedb; then
        sudo updatedb
    fi

    success "Pós-instalação do Linux Mint concluída."
    warn "Reinicie o sistema para aplicar todas as mudanças."
    pause_screen
}

# --- ponto de entrada ---
main() {
    run_post_install
}

main "$@"
