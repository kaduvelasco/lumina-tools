#!/usr/bin/env bash
# =============================================================================
# Nome do Script : zed-install.sh
# Descrição      : Instalação e configuração do Zed Editor
# Versão         : 2.0.0
# =============================================================================

# --- opções do shell ---
set -euo pipefail

# --- constantes ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
readonly ZED_CMD="zed"
readonly ZED_CONFIG_DIR="$HOME/.config/zed"
readonly ZED_CONFIG_FILE="$ZED_CONFIG_DIR/settings.json"
readonly LOCAL_BIN="$HOME/.local/bin"

# --- carregamento de dependências ---
if [[ ! -f "$SCRIPT_DIR/../scripts/utils.sh" ]]; then
    printf '\033[0;31m❌ Erro fatal: scripts/utils.sh não encontrado.\033[0m\n' >&2
    exit 1
fi
# shellcheck source=scripts/utils.sh
source "$SCRIPT_DIR/../scripts/utils.sh"

# --- funções de interface ---

# =============================================================================
# Exibe o cabeçalho padrão com identificação do módulo.
# =============================================================================
show_header() {
    show_lumina_header
    printf '%b\n' "   ${C5}MÓDULO : ${C1}Zed Editor — Instalador${RESET}"
    printf '%b\n\n' "   ${C5}Distro : ${C4}${PKG_MANAGER}${RESET}"
}

# --- funções de negócio ---

# =============================================================================
# Instala o Zed Editor via script oficial.
# =============================================================================
install_zed() {
    if is_installed_cmd "$ZED_CMD"; then
        printf '%b\n' "${C2}✅ Zed já está instalado.${RESET}"
        printf '%s' "   Reinstalar / Atualizar? (s/${C3}N${RESET}): "
        read -r confirm
        [[ ! "$confirm" =~ ^[sS]$ ]] && return 0
    fi

    require_internet
    printf '%b\n' "${C6}⚙️  Baixando instalador oficial do Zed...${RESET}"

    local zed_installer
    zed_installer=$(mktemp)
    trap 'rm -f "$zed_installer"' EXIT
    if ! curl -fsSL https://zed.dev/install.sh -o "$zed_installer"; then
        die "Falha ao baixar instalador do Zed."
    fi
    if ! sh "$zed_installer"; then
        die "Falha ao instalar o Zed."
    fi
    rm -f "$zed_installer"
    trap - EXIT

    if [[ ":$PATH:" != *":$LOCAL_BIN:"* ]]; then
        ensure_local_bin_in_path
    fi

    success "Zed instalado com sucesso."
}

# =============================================================================
# Aplica configurações personalizadas do Zed.
# =============================================================================
apply_zed_settings() {
    printf '%b\n' "${C6}⚙️  Aplicando configurações...${RESET}"
    mkdir -p "$ZED_CONFIG_DIR"

    if [[ -f "$ZED_CONFIG_FILE" ]]; then
        printf '%s' "   Configuração existente encontrada. Sobrescrever? (s/${C3}N${RESET}): "
        read -r confirm
        if [[ ! "$confirm" =~ ^[sS]$ ]]; then
            printf '%b\n' "${C4}↩️  Configuração mantida.${RESET}"
            return 0
        fi
        cp "$ZED_CONFIG_FILE" "${ZED_CONFIG_FILE}.bak.$(date +%Y%m%d%H%M%S)"
    fi

    cat <<'EOF' > "$ZED_CONFIG_FILE"
{
  "theme": "One Dark",
  "ui_font_family": "JetBrains Mono",
  "buffer_font_family": "JetBrains Mono",
  "buffer_font_size": 14,
  "autosave": "on_focus_change",
  "format_on_save": "on",
  "telemetry": { "diagnostics": false, "metrics": false }
}
EOF

    success "Configurações do Zed aplicadas."
}

# --- ponto de entrada ---

# =============================================================================
# Ponto de entrada principal do script.
# =============================================================================
main() {
    detect_pkg_manager
    require_not_root
    require_sudo
    show_header
    
    ensure_pkg "curl"
    
    install_zed
    apply_zed_settings

    printf '\n%b\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    printf '%b\n' "   ${C4}Próximos passos:${RESET}"
    printf '%b\n' "   1. Execute: ${C3}zed .${RESET}"
    printf '%b\n\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
}

main "$@"
