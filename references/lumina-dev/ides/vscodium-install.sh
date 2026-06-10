#!/usr/bin/env bash
# =============================================================================
# Nome do Script : vscodium-install.sh
# Descrição      : Instalação e configuração do VSCodium para PHP/Moodle
# Versão         : 2.0.0
# =============================================================================

# --- opções do shell ---
set -euo pipefail

# --- constantes ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
readonly CODIUM_CMD="codium"
readonly CONFIG_DIR="$HOME/.config/VSCodium/User"
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
    printf '%b\n' "   ${C5}MÓDULO : ${C1}VSCodium — PHP/Moodle Edition${RESET}"
    printf '%b\n\n' "   ${C5}Distro : ${C4}${PKG_MANAGER}${RESET}"
}

# --- funções de negócio ---

# =============================================================================
# Instala o VSCodium via repositório oficial.
# =============================================================================
install_vscodium() {
    if is_installed_cmd "$CODIUM_CMD"; then
        printf '%b\n' "${C2}✅ VSCodium já está instalado.${RESET}"
        echo -ne "   Reinstalar / Atualizar? (s/${C3}N${RESET}): "
        read -r confirm
        [[ ! "$confirm" =~ ^[sS]$ ]] && return 0
    fi

    printf '%b\n' "${C6}⚙️  Instalando VSCodium...${RESET}"

    case "$PKG_MANAGER" in
        apt)
            local keyring="/usr/share/keyrings/vscodium-archive-keyring.gpg"
            local sources_deb="/etc/apt/sources.list.d/vscodium.sources"
            
            sudo rm -f /etc/apt/sources.list.d/vscodium.list
            
            [[ ! -f "$keyring" ]] && wget -qO - https://gitlab.com/paulcarroty/vscodium-deb-rpm-repo/raw/master/pub.gpg | gpg --dearmor | sudo dd of="$keyring" status=none
            
            if [[ ! -f "$sources_deb" ]]; then
                printf 'Types: deb\nURIs: https://download.vscodium.com/debs\nSuites: vscodium\nComponents: main\nArchitectures: amd64 arm64\nSigned-by: /usr/share/keyrings/vscodium-archive-keyring.gpg\n' | sudo tee "$sources_deb" > /dev/null
            fi

            sudo apt-get update -qq
            ensure_pkg "codium"
            ;;
        dnf)
            if [[ ! -f "/etc/yum.repos.d/vscodium.repo" ]]; then
                sudo rpm --import https://gitlab.com/paulcarroty/vscodium-deb-rpm-repo/raw/master/pub.gpg
                printf '[gitlab.com_paulcarroty_vscodium_repo]\nname=download.vscodium.com\nbaseurl=https://download.vscodium.com/rpms/\nenabled=1\ngpgcheck=1\nrepo_gpgcheck=1\ngpgkey=https://gitlab.com/paulcarroty/vscodium-deb-rpm-repo/raw/master/pub.gpg\nmetadata_expire=1h\n' | sudo tee /etc/yum.repos.d/vscodium.repo > /dev/null
            fi
            ensure_pkg "codium"
            ;;
        pacman)
            if is_installed_cmd "yay"; then
                yay -S --noconfirm vscodium-bin
            elif is_installed_cmd "paru"; then
                paru -S --noconfirm vscodium-bin
            else
                die "AUR helper não encontrado (yay ou paru)."
            fi
            ;;
    esac

    configure_vscodium_marketplace
    success "VSCodium instalado com sucesso."
}

# =============================================================================
# Configura o Marketplace da Microsoft no VSCodium.
# =============================================================================
configure_vscodium_marketplace() {
    printf '%b\n' "${C6}⚙️  Configurando Marketplace...${RESET}"
    local local_product="$HOME/.config/VSCodium/product.json"
    mkdir -p "$HOME/.config/VSCodium"

    cat <<'EOF' > "$local_product"
{
  "extensionsGallery": {
    "serviceUrl": "https://marketplace.visualstudio.com/_apis/public/gallery",
    "itemUrl": "https://marketplace.visualstudio.com/items",
    "cacheUrl": "https://vscode.blob.core.windows.net/gallery/index",
    "controlUrl": ""
  }
}
EOF
}

# =============================================================================
# Instala extensões recomendadas para VSCodium.
# =============================================================================
install_vscodium_extensions() {
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
        codium --force --install-extension "$ext" &>/dev/null || true
    done
    printf '\n%b\n' "${C2}✅ Extensões processadas.${RESET}"
}

# =============================================================================
# Aplica configurações de interface (JetBrains style).
# =============================================================================
apply_vscodium_settings() {
    printf '%b\n' "${C6}⚙️  Aplicando configurações...${RESET}"
    mkdir -p "$CONFIG_DIR"

    if [[ -f "$CONFIG_FILE" ]]; then
        echo -ne "   Configuração existente encontrada. Sobrescrever? (s/${C3}N${RESET}): "
        read -r confirm
        [[ ! "$confirm" =~ ^[sS]$ ]] && return 0
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
    success "Configurações do VSCodium aplicadas."
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

    install_vscodium
    install_vscodium_extensions
    apply_vscodium_settings

    printf '\n%b\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    printf '%b\n' "   ${C4}Próximos passos:${RESET}"
    printf '%b\n' "   1. Abra o VSCodium: ${C3}codium .${RESET}"
    printf '%b\n\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
}

main "$@"
