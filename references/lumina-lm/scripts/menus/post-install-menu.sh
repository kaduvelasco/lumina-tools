#!/usr/bin/env bash
# =============================================================================
# Script Name     : post-install-menu.sh
# Description     : Post-install submenu dispatcher
# Version         : 1.0.0
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR

if [[ ! -f "$SCRIPT_DIR/../lib/utils.sh" ]]; then
    printf '\033[0;31m❌ Erro fatal: ../lib/utils.sh não encontrado.\033[0m\n' >&2
    exit 1
fi
# shellcheck source=/dev/null
source "$SCRIPT_DIR/../lib/utils.sh"

# --- funções de interface ---
show_header() {
    show_lumina_header
}

show_post_install_menu() {
    while true; do
        show_header
        printf '%b\n' "Escolha a distribuição para pós-instalação:"
        printf '%b\n' ""
        printf '%b\n' "  ${C2}1.${RESET} Linux Mint 22.3"
        printf '%b\n' "  ${C2}2.${RESET} Pop!_OS 24.04 LTS (COSMIC)"
        printf '%b\n' "  ${C2}3.${RESET} CachyOS"
        printf '%b\n' "  ${C2}4.${RESET} ZorinOS 18.1 (Core)"
        printf '%b\n' "  ${C2}5.${RESET} ZorinOS 18.1 (Lite / XFCE)"
        printf '%b\n' "  ${C2}6.${RESET} Fedora 44"
        printf '%b\n' "  ${C1}0.${RESET} Voltar"
        printf '%b\n' ""
        printf '%s' "Selecione uma opção: "

        local choice
        read -r choice

        case "$choice" in
            1) bash "$SCRIPT_DIR/../post-install/pos-install-mint.sh" ;;
            2) bash "$SCRIPT_DIR/../post-install/pos-install-popos.sh" ;;
            3) bash "$SCRIPT_DIR/../post-install/pos-install-cachyos.sh" ;;
            4) bash "$SCRIPT_DIR/../post-install/pos-install-zorin.sh" ;;
            5) bash "$SCRIPT_DIR/../post-install/pos-install-zorin-lite.sh" ;;
            6) bash "$SCRIPT_DIR/../post-install/pos-install-fedora.sh" ;;
            0) return 10 ;;
            *) warn "Opção inválida." ; pause_screen ;;
        esac
    done
}

# --- ponto de entrada ---
main() {
    show_post_install_menu
}

main "$@"
