#!/usr/bin/env bash
# =============================================================================
# Script Name     : apps-uninstall.sh
# Description     : Uninstall installed Flatpak applications from a numbered menu
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

# --- globals compartilhados entre funções ---
declare -a INSTALLED_APP_NAMES=()
declare -a INSTALLED_APP_IDS=()
declare -a APPS_TO_UNINSTALL=()

# --- funções auxiliares ---
load_installed_flatpaks() {
    local entry
    local app_name
    local app_id

    INSTALLED_APP_NAMES=()
    INSTALLED_APP_IDS=()

    while IFS=$'\t' read -r app_name app_id; do
        [[ -z "${app_id}" ]] && continue

        entry="${app_name:-${app_id}}"
        INSTALLED_APP_NAMES+=("${entry}")
        INSTALLED_APP_IDS+=("${app_id}")
    done < <(flatpak list --app --columns=name,application 2>/dev/null)
}

show_installed_apps() {
    local index=1
    local app_name

    for app_name in "${INSTALLED_APP_NAMES[@]}"; do
        printf '%b\n' "  ${C2}${index}.${RESET} ${app_name}"
        index=$((index + 1))
    done
}

process_uninstall_selections() {
    local -a raw_choices=("$@")
    local choice
    local selected_name

    APPS_TO_UNINSTALL=()

    for choice in "${raw_choices[@]}"; do
        if [[ "$choice" == '0' ]]; then
            return 10
        fi

        if ! [[ "$choice" =~ ^[0-9]+$ ]]; then
            warn "Entrada ignorada: ${choice}"
            continue
        fi

        if ((choice < 1 || choice > ${#INSTALLED_APP_NAMES[@]})); then
            warn "Opção fora do intervalo: ${choice}"
            continue
        fi

        selected_name="${INSTALLED_APP_NAMES[$((choice - 1))]}"
        info "Selecionado para desinstalação: ${selected_name}"
        APPS_TO_UNINSTALL+=("${INSTALLED_APP_IDS[$((choice - 1))]}")
    done
}

uninstall_selected_apps() {
    if ((${#APPS_TO_UNINSTALL[@]} == 0)); then
        warn "Nenhum aplicativo foi selecionado."
        return 0
    fi

    local -a failed=()
    local app_id
    local exit_code

    info "Desinstalando ${#APPS_TO_UNINSTALL[@]} aplicativo(s)..."

    for app_id in "${APPS_TO_UNINSTALL[@]}"; do
        exit_code=0
        flatpak uninstall -y "${app_id}" || exit_code=$?
        if ((exit_code != 0)); then
            warn "Falha ao desinstalar: ${app_id}"
            failed+=("${app_id}")
        fi
    done

    if ((${#failed[@]} == 0)); then
        success "Todos os aplicativos foram desinstalados com sucesso."
    else
        warn "${#failed[@]} aplicativo(s) não foram desinstalados:"
        local f
        for f in "${failed[@]}"; do
            printf '%b\n' "  ${C1}  - ${f}${RESET}"
        done
        if ((${#APPS_TO_UNINSTALL[@]} > ${#failed[@]})); then
            success "$((${#APPS_TO_UNINSTALL[@]} - ${#failed[@]})) aplicativo(s) desinstalado(s) com sucesso."
        fi
    fi
}

# --- funções de interface ---
show_header() {
    show_lumina_header
}

show_menu() {
    local -a selections
    local process_status=0

    show_header

    if ! is_installed_cmd flatpak; then
        warn "Flatpak não está instalado neste sistema."
        pause_screen
        return 0
    fi

    load_installed_flatpaks

    if ((${#INSTALLED_APP_IDS[@]} == 0)); then
        warn "Nenhum Flatpak instalado foi encontrado."
        pause_screen
        return 0
    fi

    printf '%b\n' "Selecione os aplicativos instalados que deseja desinstalar."
    printf '%b\n' ""
    show_installed_apps
    printf '%b\n' ""
    printf '%b\n' "  ${C1}0.${RESET} Voltar"
    printf '%b\n' ""
    printf '%s' "Digite os números desejados: "
    read -r -a selections

    if process_uninstall_selections "${selections[@]}"; then
        process_status=0
    else
        process_status=$?
    fi

    if [[ ${process_status} -eq 10 ]]; then
        return 10
    fi
    if [[ ${process_status} -ne 0 ]]; then
        return 0
    fi

    uninstall_selected_apps
}

# --- ponto de entrada ---
main() {
    show_menu
}

main "$@"
