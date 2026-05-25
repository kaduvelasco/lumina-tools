#!/usr/bin/env bash
# =============================================================================
# Script Name     : utils.sh
# Description     : Shared output and package helpers
# Version         : 1.0.0
# =============================================================================

set -euo pipefail

[[ -n "${LUMINA_UTILS_LOADED:-}" ]] && return 0
readonly LUMINA_UTILS_LOADED=1

readonly C1='\033[0;31m'
readonly C2='\033[0;32m'
readonly C3='\033[1;33m'
readonly C4='\033[0;34m'
readonly C5='\033[0;35m'
readonly C6='\033[0;36m'
readonly H1='\033[1;32m'
readonly H2='\033[0;32m'
readonly TS=''
readonly RESET='\033[0m'

readonly SIM_OK='✅'
readonly SIM_WARN='⚠️'
readonly SIM_INFO='ℹ️'
readonly SIM_FAIL='❌'

export C1 C2 C3 C4 C5 C6 H1 H2 TS RESET

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

show_lumina_header() {
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
    printf '%b\n' "${H2}LUMINA LINUX MANAGEMENT${RESET}"
    printf '%b\n' ""
}

pause_screen() {
    printf '%s' "Pressione Enter para continuar..."
    read -r _
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
        export PKG_INSTALL='sudo apt-get install -y --'
        export PKG_UPDATE='sudo apt-get update -y'
    elif is_installed_cmd pacman; then
        export PKG_MANAGER='pacman'
        export PKG_INSTALL='sudo pacman -S --needed --noconfirm --'
        export PKG_UPDATE='sudo pacman -Sy --noconfirm'
    elif is_installed_cmd dnf; then
        export PKG_MANAGER='dnf'
        export PKG_INSTALL='sudo dnf install -y --'
        export PKG_UPDATE='sudo dnf makecache -y'
    else
        die "No supported package manager was detected."
    fi

    readonly PKG_MANAGER PKG_INSTALL PKG_UPDATE
}

ensure_pkg() {
    local package_name="$1"

    detect_pkg_manager

    if [[ "${PKG_MANAGER}" == 'apt' ]]; then
        dpkg -s "${package_name}" >/dev/null 2>&1 && return 0
        sudo apt-get update -y
        sudo apt-get install -y -- "${package_name}"
        return 0
    fi

    if [[ "${PKG_MANAGER}" == 'pacman' ]]; then
        pacman -Q "${package_name}" >/dev/null 2>&1 && return 0
        sudo pacman -S --needed --noconfirm -- "${package_name}"
        return 0
    fi

    rpm -q "${package_name}" >/dev/null 2>&1 && return 0
    sudo dnf install -y -- "${package_name}"
}

ensure_flatpak_ready() {
    if ! is_installed_cmd flatpak; then
        detect_pkg_manager
        info "Flatpak not found. Installing dependency..."
        if [[ "${PKG_MANAGER}" == 'apt' ]]; then
            sudo apt-get update -y
            sudo apt-get install -y -- flatpak
        elif [[ "${PKG_MANAGER}" == 'pacman' ]]; then
            sudo pacman -S --needed --noconfirm -- flatpak
        else
            sudo dnf install -y -- flatpak
        fi
    fi

    if ! flatpak remotes --columns=name 2>/dev/null | grep -qxF 'flathub'; then
        info "Adding Flathub repository..."
        flatpak remote-add --if-not-exists flathub \
            https://flathub.org/repo/flathub.flatpakrepo \
            || die "Falha ao adicionar o repositório Flathub. Verifique a conexão com a internet."
    fi
}
