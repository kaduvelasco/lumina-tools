#!/usr/bin/env bash

# =============================================================================
# Nome do Script : menu.sh
# Descrição      : Menu interativo do instalador
# Versão         : 3.0.0
# =============================================================================

[[ -n "${LUMINA_MENU_LOADED:-}" ]] && return 0
readonly LUMINA_MENU_LOADED=1

show_header() {
    show_lumina_header "LuminaStack — Linux + Nginx + PHP + MariaDb + Docker"
}

show_menu() {
    printf '%b\n' "O que você deseja instalar?"
    printf '%b\n' ""
    printf '%b\n' "  ${C2}1.${RESET} Instalar pré-requisitos (curl, git, openssl, lsof)"
    printf '%b\n' "  ${C2}2.${RESET} Instalar Docker (Engine + configuração)"
    printf '%b\n' "  ${C2}3.${RESET} Criar workspace (Cria estrutura /srv/workspace)"
    printf '%b\n' "  ${C2}4.${RESET} Gerar stack Docker (PHP, Nginx e MariaDB via docker-compose)"
    printf '%b\n' "  ${C1}0.${RESET} Sair"
    printf '%b\n' ""
}
