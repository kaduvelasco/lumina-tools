#!/usr/bin/env bash

# =============================================================================
# Nome do Script : install.sh
# Descrição      : Ponto de entrada do instalador interativo do LuminaStack
# Versão         : 3.0.0
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR

# --- carregamento de dependências ---
for _lib in utils.sh versions.sh menu.sh system.sh workspace.sh docker.sh; do
    if [[ ! -f "$SCRIPT_DIR/lib/$_lib" ]]; then
        printf '\033[0;31m❌ Erro fatal: lib/%s não encontrado.\033[0m\n' "$_lib" >&2
        exit 1
    fi
    # shellcheck source=/dev/null
    source "$SCRIPT_DIR/lib/$_lib"
done
unset _lib

# --- funções de interface ---
concluir_acao() {
    local label="$1"
    printf '\n'
    success "$label concluída com sucesso!"
    printf '%b\n' "${C4}──────────────────────────────────${RESET}"
    printf '%s' "   Pressione Enter para voltar ao menu..."
    read -r _ || true
}

# --- ponto de entrada ---
main() {
    detect_pkg_manager

    while true; do
        show_lumina_header

        show_menu

        printf '%s' "Selecione uma opção: "
        read -r opt

        case "$opt" in
            1)
                if install_prereqs; then
                    concluir_acao "Instalação de pré-requisitos"
                fi
                ;;
            2)
                if install_docker; then
                    concluir_acao "Instalação do Docker"
                fi
                ;;
            3)
                if create_workspace; then
                    concluir_acao "Criação do workspace"
                fi
                ;;
            4)
                if generate_docker_stack; then
                    concluir_acao "Geração da stack Docker"
                fi
                ;;
            0)
                printf '\n%b\n\n' "${C2}Saindo do LuminaStack. Até logo!${RESET}"
                exit 0
                ;;
            *)
                printf '\n%b\n' "${C1}❌ Opção inválida! Digite um número de 0 a 4.${RESET}"
                ;;
        esac
    done
}

main "$@"
