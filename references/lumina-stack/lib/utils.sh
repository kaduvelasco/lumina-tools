#!/usr/bin/env bash

# =============================================================================
# Nome do Script : utils.sh
# Descrição      : Paleta de cores, funções de saída e utilitários compartilhados
# Versão         : 3.0.0
# =============================================================================

[[ -n "${LUMINA_UTILS_LOADED:-}" ]] && return 0
readonly LUMINA_UTILS_LOADED=1

# --- cores ---
export C1=$'\033[0;31m'    # vermelho  — erros, ações destrutivas
export C2=$'\033[0;32m'    # verde     — sucesso, operações normais
export C3=$'\033[1;33m'    # amarelo   — avisos, manutenção
export C4=$'\033[0;34m'    # azul      — informativas, consulta
export C5=$'\033[0;35m'    # magenta   — rótulos, bordas de menu
export C6=$'\033[0;36m'    # ciano     — dicas, decorativos
export H1=$'\033[1;32m'    # verde bold — logo, linha principal
export H2=$'\033[0;32m'    # verde     — logo, subtítulo
export TS=''
export RESET=$'\033[0m'

# --- saída padronizada ---
die() {
    printf '%b\n' "${C1}❌ ${1}${RESET}" >&2
    exit "${2:-1}"
}

warn() {
    printf '%b\n' "${C3}⚠️  ${1}${RESET}" >&2
}

info() {
    printf '%b\n' "${C4}ℹ  ${1}${RESET}"
}

success() {
    printf '%b\n' "${C2}✅ ${1}${RESET}"
}

# --- cabeçalho ---
# =============================================================================
# Exibe o cabeçalho ASCII padrão Lumina. $1 = subtítulo (opcional).
# =============================================================================
show_lumina_header() {
    local subtitle="${1:-LUMINA STACK}"
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
    printf '%b\n' "${H2}${subtitle}${RESET} "
    printf '%b\n' ""
}

# --- verificação de comandos ---
is_installed_cmd() {
    type -P "$1" >/dev/null 2>&1
}

# --- gerenciador de pacotes ---
detect_pkg_manager() {
    [[ -n "${PKG_MANAGER:-}" ]] && return 0
    if type -P apt-get >/dev/null 2>&1; then
        PKG_MANAGER="apt"
    elif type -P dnf >/dev/null 2>&1; then
        PKG_MANAGER="dnf"
    elif type -P pacman >/dev/null 2>&1; then
        PKG_MANAGER="pacman"
    else
        die "Nenhum gerenciador de pacotes suportado encontrado."
    fi
    export PKG_MANAGER
    readonly PKG_MANAGER
}

is_installed_pkg() {
    local pkg="$1"
    case "${PKG_MANAGER:-}" in
        apt)    dpkg -l "$pkg" 2>/dev/null | grep -q '^ii' ;;
        dnf)    rpm -q "$pkg" >/dev/null 2>&1 ;;
        pacman) pacman -Q "$pkg" >/dev/null 2>&1 ;;
        *)      type -P "$pkg" >/dev/null 2>&1 ;;
    esac
}

ensure_pkg() {
    local pkg="$1"
    is_installed_pkg "$pkg" && return 0
    case "${PKG_MANAGER:-}" in
        apt)    sudo apt-get install -y -- "$pkg" ;;
        dnf)    sudo dnf install -y -- "$pkg" ;;
        pacman) sudo pacman -S --noconfirm -- "$pkg" ;;
        *)      die "PKG_MANAGER não definido. Execute detect_pkg_manager primeiro." ;;
    esac
}
