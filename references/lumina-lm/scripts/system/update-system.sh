#!/usr/bin/env bash
# =============================================================================
# Script Name     : update-system.sh
# Description     : Run system update and cleanup for supported distros
# Version         : 1.0.0
# =============================================================================

set -euo pipefail

readonly C1='\033[0;31m'
readonly C2='\033[0;32m'
readonly C3='\033[1;33m'
readonly C4='\033[0;34m'
readonly H2='\033[0;32m'
readonly RESET='\033[0m'

readonly SIM_OK='✅'
readonly SIM_WARN='⚠️'
readonly SIM_INFO='ℹ️'
readonly SIM_FAIL='❌'

# --- funções de interface ---
show_header() {
    clear
    printf '%b\n' ""
    printf '%b\n' "░██                            ░██                      "
    printf '%b\n' "░██                                                     "
    printf '%b\n' "░██ ░██    ░██ ░█████████████  ░██░████████   ░██████   "
    printf '%b\n' "░██ ░██    ░██ ░██   ░██   ░██ ░██░██    ░██       ░██  "
    printf '%b\n' "░██ ░██    ░██ ░██   ░██   ░██ ░██░██    ░██  ░███████  "
    printf '%b\n' "░██ ░██   ░███ ░██   ░██   ░██ ░██░██    ░██ ░██   ░██  "
    printf '%b\n' "░██  ░█████░██ ░██   ░██   ░██ ░██░██    ░██  ░█████░██ "
    printf '%b\n' ""
    printf '%b\n' "${H2}LUMINA LINUX UPDATE${RESET}"
    printf '%b\n' ""
}

# --- funções auxiliares ---
die() {
    local message="$1"
    local exit_code="${2:-1}"
    printf '%b\n' "${C1}${SIM_FAIL} ${message}${RESET}" >&2
    exit "${exit_code}"
}

warn() {
    local message="$1"
    printf '%b\n' "${C3}${SIM_WARN} ${message}${RESET}" >&2
}

info() {
    local message="$1"
    printf '%b\n' "${C4}${SIM_INFO} ${message}${RESET}"
}

success() {
    local message="$1"
    printf '%b\n' "${C2}${SIM_OK} ${message}${RESET}"
}

is_installed_cmd() {
    type -P "$1" >/dev/null 2>&1
}

detect_pkg_manager() {
    if [[ -n "${PKG_MANAGER:-}" ]]; then
        return 0
    fi

    if is_installed_cmd apt-get; then
        export PKG_MANAGER='apt'
    elif is_installed_cmd dnf; then
        export PKG_MANAGER='dnf'
    elif is_installed_cmd pacman; then
        export PKG_MANAGER='pacman'
    else
        die "Nenhum gerenciador de pacotes suportado foi detectado."
    fi

    readonly PKG_MANAGER
}

require_not_root() {
    if [[ $EUID -eq 0 ]]; then
        die "Este script não deve ser executado como root."
    fi
}

require_sudo() {
    info "Será solicitada a senha de administrador para continuar."
    sudo -v || die "Este script requer privilégios de sudo."
}

cleanup_thumbnails() {
    if [[ -d "${HOME}/.cache/thumbnails" ]]; then
        rm -rf "${HOME}/.cache/thumbnails/"*
    fi
}

log_path() {
    printf '%s' "${HOME}/manutencao_sistema.log"
}

run_apt_update() {
    sudo apt-get update -y
    sudo apt-get upgrade -y
    sudo apt-get full-upgrade -y
    sudo apt-get autoremove -y
    sudo apt-get autoclean -y
    sudo apt-get clean
}

run_dnf_update() {
    sudo dnf upgrade --refresh -y
    sudo dnf autoremove -y
    sudo dnf clean all
}

run_pacman_update() {
    sudo pacman -Syu --noconfirm
    sudo pacman -Sc --noconfirm
}

run_snap_update() {
    sudo snap refresh
}

run_flatpak_update() {
    flatpak update -y
    flatpak uninstall --unused -y
}

# --- funções de negócio ---
update_system() {
    local log_file

    require_not_root
    require_sudo
    detect_pkg_manager

    log_file="$(log_path)"
    exec > >(tee -i "${log_file}") 2>&1

    show_header
    info "Log salvo em: ${log_file}"

    if [[ "${PKG_MANAGER}" == 'apt' ]]; then
        info "Atualizando pacotes APT..."
        run_apt_update
    elif [[ "${PKG_MANAGER}" == 'dnf' ]]; then
        info "Atualizando pacotes DNF..."
        run_dnf_update
    else
        info "Atualizando pacotes Pacman..."
        run_pacman_update
    fi

    if is_installed_cmd snap; then
        info "Atualizando Snaps..."
        run_snap_update
    else
        info "Snap não está instalado. Etapa ignorada."
    fi

    if is_installed_cmd flatpak; then
        info "Atualizando Flatpaks..."
        run_flatpak_update
    else
        info "Flatpak não está instalado. Etapa ignorada."
    fi

    if is_installed_cmd journalctl; then
        info "Removendo logs antigos..."
        sudo journalctl --vacuum-time=7d
    fi

    cleanup_thumbnails

    if is_installed_cmd updatedb; then
        info "Atualizando índice de busca..."
        sudo updatedb
    fi

    success "Atualização do sistema concluída."
}

# --- ponto de entrada ---
main() {
    update_system
}

main "$@"
