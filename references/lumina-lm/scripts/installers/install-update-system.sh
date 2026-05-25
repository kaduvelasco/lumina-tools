#!/usr/bin/env bash
# =============================================================================
# Script Name     : install-update-system.sh
# Description     : Install the update-system command system-wide
# Version         : 1.0.0
# =============================================================================

set -euo pipefail

trap 'die "Falha inesperada na instalação do comando update-system (linha ${LINENO})."' ERR

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR

if [[ ! -f "$SCRIPT_DIR/../lib/utils.sh" ]]; then
    printf '\033[0;31m❌ Erro fatal: ../lib/utils.sh não encontrado.\033[0m\n' >&2
    exit 1
fi
# shellcheck source=/dev/null
source "$SCRIPT_DIR/../lib/utils.sh"

if [[ ! -f "$SCRIPT_DIR/../lib/system.sh" ]]; then
    printf '\033[0;31m❌ Erro fatal: ../lib/system.sh não encontrado.\033[0m\n' >&2
    exit 1
fi
# shellcheck source=/dev/null
source "$SCRIPT_DIR/../lib/system.sh"

# --- funções de negócio ---
install_command() {
    local source_script="${SCRIPT_DIR}/../system/update-system.sh"
    local target_file="/usr/local/bin/update-system"

    require_not_root
    require_sudo

    if [[ ! -f "${source_script}" ]]; then
        die "update-system.sh não encontrado em: ${source_script}"
    fi

    info "Copiando ${source_script} para ${target_file}..."
    if ! sudo cp "${source_script}" "${target_file}"; then
        die "Falha ao copiar o comando para ${target_file}."
    fi

    info "Aplicando permissão de execução..."
    if ! sudo chmod +x "${target_file}"; then
        die "Falha ao aplicar permissão de execução em ${target_file}."
    fi

    if ! sudo test -f "${target_file}"; then
        die "O arquivo ${target_file} não foi encontrado após a instalação."
    fi

    success "Comando instalado em ${target_file}"
    info "Execute 'update-system' em um novo terminal ou rode 'hash -r'."
}

# --- funções de interface ---
show_header() {
    show_lumina_header
}

# --- ponto de entrada ---
main() {
    show_header
    install_command
}

main "$@"
