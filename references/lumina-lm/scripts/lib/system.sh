#!/usr/bin/env bash
# =============================================================================
# Script Name     : system.sh
# Description     : Shared system validation helpers
# Version         : 1.0.0
# =============================================================================

set -euo pipefail

[[ -n "${LUMINA_SYSTEM_LOADED:-}" ]] && return 0
readonly LUMINA_SYSTEM_LOADED=1

SYSTEM_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SYSTEM_LIB_DIR

if [[ ! -f "$SYSTEM_LIB_DIR/utils.sh" ]]; then
    printf '\033[0;31m❌ Fatal error: lib/utils.sh not found.\033[0m\n' >&2
    exit 1
fi
# shellcheck source=/dev/null
source "$SYSTEM_LIB_DIR/utils.sh"

require_not_root() {
    if [[ $EUID -eq 0 ]]; then
        die "Este script não deve ser executado como root."
    fi
}

require_sudo() {
    info "Será solicitada a senha de administrador para continuar."
    sudo -v || die "Este script requer privilégios de sudo."
}

require_internet() {
    curl -fsSL https://checkip.amazonaws.com >/dev/null 2>&1 || \
        die "Este script requer conexão com a internet."
}

get_distro_id() {
    local distro_id='unknown'

    if [[ -f /etc/os-release ]]; then
        # shellcheck disable=SC1091
        source /etc/os-release
        distro_id="${ID:-unknown}"
    fi

    printf '%s' "${distro_id}"
}

assert_distro() {
    local expected_distro="$1"
    local current_distro

    current_distro="$(get_distro_id)"
    [[ "${current_distro}" == "${expected_distro}" ]] || \
        die "Este script é destinado a '${expected_distro}', mas a distribuição atual é '${current_distro}'."
}

start_log() {
    local prefix="$1"
    local log_file

    log_file="${HOME}/${prefix}-$(date +%Y%m%d_%H%M%S).log"
    exec > >(tee -i "${log_file}") 2>&1
    printf '%s' "${log_file}"
}

set_sysctl_value() {
    local config_file="$1"
    local key="$2"
    local value="$3"
    local temp_file

    temp_file=$(mktemp)
    trap 'rm -f -- "$temp_file"; trap - RETURN' RETURN

    sudo touch "${config_file}"
    grep -vF "${key}=" "${config_file}" > "${temp_file}" 2>/dev/null || true
    printf '%s=%s\n' "${key}" "${value}" >> "${temp_file}"
    sudo install -m 644 "${temp_file}" "${config_file}"
}

configure_swappiness() {
    info "Configurando swappiness..."
    set_sysctl_value "/etc/sysctl.d/99-lumina-swappiness.conf" "vm.swappiness" "10"
}

configure_inotify() {
    info "Configurando inotify..."
    set_sysctl_value "/etc/sysctl.d/99-lumina-inotify.conf" "fs.inotify.max_user_watches" "524288"
}

apply_sysctl() {
    info "Aplicando parâmetros do kernel..."
    sudo sysctl --system >/dev/null
}
