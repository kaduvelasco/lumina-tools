#!/usr/bin/env bash
# =============================================================================
# Nome do Script : vscode-install.sh
# Descrição      : Instalação e configuração do VS Code para PHP/Moodle
# Versão         : 2.0.0
# =============================================================================

# --- opções do shell ---
set -euo pipefail

# --- constantes ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
readonly CODE_CMD="code"
readonly CONFIG_DIR="$HOME/.config/Code/User"
readonly CONFIG_FILE="$CONFIG_DIR/settings.json"

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
    printf '%b\n' "   ${C5}MÓDULO : ${C1}VS Code — PHP/Moodle Edition${RESET}"
    printf '%b\n\n' "   ${C5}Distro : ${C4}${PKG_MANAGER}${RESET}"
}

# --- funções de negócio ---

# =============================================================================
# Instala o VS Code via repositório oficial.
# =============================================================================
install_vscode() {
    if is_installed_cmd "$CODE_CMD"; then
        printf '%b\n' "${C2}✅ VS Code já está instalado.${RESET}"
        echo -ne "   Reinstalar / Atualizar? (s/${C3}N${RESET}): "
        read -r confirm
        [[ ! "$confirm" =~ ^[sS]$ ]] && return 0
    fi

    printf '%b\n' "${C6}⚙️  Instalando VS Code...${RESET}"

    case "$PKG_MANAGER" in
        apt)
            local keyring="/usr/share/keyrings/microsoft-archive-keyring.gpg"
            local sources_list="/etc/apt/sources.list.d/vscode.list"

            sudo rm -f /etc/apt/sources.list.d/vscode.sources
            
            [[ ! -f "$keyring" ]] && wget -qO - https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor | sudo dd of="$keyring" status=none
            
            if [[ ! -f "$sources_list" ]]; then
                echo "deb [arch=amd64,arm64 signed-by=$keyring] https://packages.microsoft.com/repos/code stable main" | sudo tee "$sources_list" > /dev/null
            fi

            sudo apt-get update -qq
            ensure_pkg "code"
            ;;
        dnf)
            if [[ ! -f "/etc/yum.repos.d/vscode.repo" ]]; then
                sudo rpm --import https://packages.microsoft.com/keys/microsoft.asc
                printf '[code]\nname=Visual Studio Code\nbaseurl=https://packages.microsoft.com/yumrepos/vscode\nenabled=1\ngpgcheck=1\ngpgkey=https://packages.microsoft.com/keys/microsoft.asc\n' | sudo tee /etc/yum.repos.d/vscode.repo > /dev/null
            fi
            ensure_pkg "code"
            ;;
        pacman)
            if is_installed_cmd "yay"; then
                yay -S --noconfirm visual-studio-code-bin
            elif is_installed_cmd "paru"; then
                paru -S --noconfirm visual-studio-code-bin
            else
                die "AUR helper não encontrado (yay ou paru)."
            fi
            ;;
    esac

    success "VS Code instalado com sucesso."
}

# =============================================================================
# Instala extensões recomendadas para VS Code.
# =============================================================================
install_vscode_extensions() {
    printf '%b\n' "${C6}⚙️  Instalando extensões...${RESET}"

    local extensions=(
        "bmewburn.vscode-intelephense-client"
        "MehediDracula.php-namespace-resolver"
        "imgildev.vscode-moodle-snippets"
        "junstyle.php-cs-fixer"
        "dawhite.mustache"
        "ms-azuretools.vscode-docker"
        "ms-vscode-remote.remote-containers"
        "k--kato.intellij-idea-keybindings"
        "narasimapandiyan.jetbrainsmono"
        "fogio.jetbrains-color-theme"
    )

    for ext in "${extensions[@]}"; do
        printf '%b\r' "${C6}   Instalando ${ext}...${RESET}"
        code --force --install-extension "$ext" &>/dev/null || true
    done
    printf '\n%b\n' "${C2}✅ Extensões processadas.${RESET}"
}

# =============================================================================
# Aplica configurações de interface (JetBrains style).
# =============================================================================
apply_vscode_settings() {
    printf '%b\n' "${C6}⚙️  Aplicando configurações...${RESET}"
    mkdir -p "$CONFIG_DIR"

    if [[ -f "$CONFIG_FILE" ]]; then
        echo -ne "   Configuração existente encontrada. Sobrescrever? (s/${C3}N${RESET}): "
        read -r confirm
        if [[ ! "$confirm" =~ ^[sS]$ ]]; then
            printf '%b\n' "${C4}↩️  Mantido.${RESET}"
            return 0
        fi
        cp "$CONFIG_FILE" "${CONFIG_FILE}.bak.$(date +%Y%m%d%H%M%S)"
    fi

    cat <<'EOF' > "$CONFIG_FILE"
{
    "telemetry.telemetryLevel": "off",
    "workbench.colorTheme": "JetBrains New UI Extended",
    "editor.fontFamily": "'JetBrains Mono', 'Fira Code', monospace",
    "editor.fontLigatures": true,
    "editor.fontSize": 14,
    "editor.lineHeight": 1.6,
    "files.autoSave": "onFocusChange",
    "workbench.editor.enablePreview": false,
    "explorer.compactFolders": false,
    "editor.formatOnSave": true,
    "php.validate.enable": false,
    "editor.minimap.enabled": false
}
EOF
    success "Configurações do VS Code aplicadas."
}

# --- ponto de entrada ---

# =============================================================================
# Ponto de entrada principal do script.
# =============================================================================
main() {
    detect_pkg_manager
    show_header
    
    ensure_pkg "wget"
    ensure_pkg "curl"
    ensure_pkg "gpg" || true

    install_vscode
    install_vscode_extensions
    apply_vscode_settings

    printf '\n%b\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    printf '%b\n' "   ${C4}Próximos passos:${RESET}"
    printf '%b\n' "   1. Abra o VS Code: ${C3}code .${RESET}"
    printf '%b\n\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
}

main "$@"
