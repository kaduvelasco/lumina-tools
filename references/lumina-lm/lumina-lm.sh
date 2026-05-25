#!/usr/bin/env bash
# =============================================================================
# Script Name     : lumina-lm.sh
# Description     : Main menu for Lumina Linux Management
# Version         : 2.0.0
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR

if [[ ! -f "$SCRIPT_DIR/scripts/lib/utils.sh" ]]; then
    printf '\033[0;31m❌ Erro fatal: scripts/lib/utils.sh não encontrado.\033[0m\n' >&2
    exit 1
fi
# shellcheck source=/dev/null
source "$SCRIPT_DIR/scripts/lib/utils.sh"

# --- funções de interface ---
show_header() {
    show_lumina_header
}

confirm_return_to_main_menu() {
    success "Operação concluída."
    printf '%s' "Pressione ENTER para voltar ao menu: "
    read -r _
}

run_menu_script() {
    local script_path="$1"
    local exit_code=0

    if bash "$script_path"; then
        exit_code=0
    else
        exit_code=$?
    fi

    if [[ ${exit_code} -eq 10 ]]; then
        return 0
    fi
    if [[ ${exit_code} -eq 0 ]]; then
        confirm_return_to_main_menu
        return 0
    fi

    warn "A execução falhou: $(basename "$script_path")"
    pause_screen
    return 1
}

show_main_menu() {
    while true; do
        show_header
        printf '%b\n' "O que você deseja fazer?"
        printf '%b\n' ""
        printf '%b\n' "  ${C2}1.${RESET} Executar Pós-instalação"
        printf '%b\n' "  ${C2}2.${RESET} Criar modelos de arquivos"
        printf '%b\n' "  ${C2}3.${RESET} Instalar aplicativos"
        printf '%b\n' "  ${C2}4.${RESET} Desinstalar aplicativos"
        printf '%b\n' "  ${C2}5.${RESET} Instalar comando update-system"
        printf '%b\n' "  ${C1}0.${RESET} Sair"
        printf '%b\n' ""
        printf '%s' "Selecione uma opção: "

        local choice
        read -r choice

        case "$choice" in
            1) run_menu_script "$SCRIPT_DIR/scripts/menus/post-install-menu.sh" ;;
            2) run_menu_script "$SCRIPT_DIR/scripts/templates/file-models.sh" ;;
            3) run_menu_script "$SCRIPT_DIR/scripts/apps/apps-install.sh" ;;
            4) run_menu_script "$SCRIPT_DIR/scripts/apps/apps-uninstall.sh" ;;
            5) run_menu_script "$SCRIPT_DIR/scripts/installers/install-update-system.sh" ;;
            0) return 0 ;;
            *) warn "Opção inválida." ; pause_screen ;;
        esac
    done
}

# --- ponto de entrada ---
main() {
    show_main_menu
}

main "$@"
