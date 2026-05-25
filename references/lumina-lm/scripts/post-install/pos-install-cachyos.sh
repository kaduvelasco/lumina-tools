#!/usr/bin/env bash
# =============================================================================
# Script Name     : pos-install-cachyos.sh
# Description     : Post-install routine for CachyOS
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
detect_desktop() {
    local desktop_session="${XDG_CURRENT_DESKTOP:-Unknown}"
    info "Ambiente detectado: ${desktop_session}"
}

validate_desktop_support() {
    local desktop_session="${XDG_CURRENT_DESKTOP:-}"

    [[ -z "${desktop_session}" ]] && return 0

    case "${desktop_session,,}" in
        *plasma*|*kde*|*gnome*|*cosmic*|*niri*)
            return 0
            ;;
        *)
            warn "Ambiente não listado como alvo principal: ${desktop_session}"
            ;;
    esac
}

# --- funções de negócio ---
run_post_install() {
    local log_file
    local -a pacman_packages=(
        base-devel
        git
        wget
        curl
        htop
        fastfetch
        tree
        jq
        plocate
        python-pip
        ffmpeg
        vlc
        unrar
        p7zip
        unzip
        ttf-dejavu
        ttf-liberation
        flatpak
    )

    require_not_root
    require_sudo
    require_internet
    assert_distro "cachyos"

    log_file="$(start_log "pos-install-cachyos")"

    show_header
    info "Log salvo em: ${log_file}"

    info "Atualizando sistema base..."
    sudo pacman -Syu --noconfirm

    info "Instalando pacotes Pacman..."
    sudo pacman -S --needed --noconfirm -- "${pacman_packages[@]}"

    ensure_flatpak_ready
    detect_desktop
    validate_desktop_support
    configure_swappiness
    configure_inotify
    apply_sysctl

    if is_installed_cmd updatedb; then
        sudo updatedb
    fi

    success "Pós-instalação do CachyOS concluída."
    pause_screen
}

# --- ponto de entrada ---
main() {
    run_post_install
}

main "$@"
