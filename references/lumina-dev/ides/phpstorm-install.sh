#!/usr/bin/env bash
# =============================================================================
# Nome do Script : phpstorm-install.sh
# Descrição      : Instalador auxiliar do PHPStorm via tar.gz
# Versão         : 2.0.0
# =============================================================================

# --- opções do shell ---
set -euo pipefail

# --- constantes ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
readonly INSTALL_DIR="/opt/phpstorm"
readonly SYMLINK="/usr/local/bin/phpstorm"
readonly DESKTOP_FILE="/usr/share/applications/phpstorm.desktop"

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
    printf '%b\n' "   ${C5}INSTALADOR AUXILIAR PHPSTORM${RESET}"
    printf '%b\n\n' "   ${C5}Distro : ${C4}${PKG_MANAGER}${RESET}"
}

# --- funções auxiliares ---

# =============================================================================
# Solicita e valida o caminho do arquivo .tar.gz.
# =============================================================================
get_phpstorm_package_path() {
    local input_path
    printf '%b\n' "${C6}📂 Arraste o arquivo .tar.gz para cá ou digite o caminho completo:${RESET}"
    echo -ne "> "
    read -r input_path

    input_path="${input_path//\'/}"
    input_path="${input_path//\"/}"
    input_path="${input_path#"${input_path%%[![:space:]]*}"}"
    input_path="${input_path%"${input_path##*[![:space:]]}"}"

    if [[ -z "$input_path" ]]; then
        die "Nenhum caminho informado."
    fi

    if [[ ! -f "$input_path" ]]; then
        die "Arquivo não encontrado: '${input_path}'"
    fi

    if [[ "$input_path" != *.tar.gz ]]; then
        die "O arquivo precisa ser um pacote .tar.gz oficial da JetBrains."
    fi

    local version
    version=$(basename "$input_path" | grep -oP '\d+\.\d+(\.\d+)?' | head -1 || echo "")
    if [[ -n "$version" ]]; then
        printf '%b\n' "${C2}✅ Pacote reconhecido: PHPStorm ${version}${RESET}"
    fi
    
    echo "$input_path"
}

# --- funções de negócio ---

# =============================================================================
# Instala o PHPStorm a partir do pacote fornecido.
# =============================================================================
install_phpstorm() {
    local package_path="$1"

    if [[ -d "$INSTALL_DIR" ]]; then
        warn "PHPStorm já está instalado em ${INSTALL_DIR}."
        echo -ne "   Deseja substituir? (${C3}s${RESET}/N): "
        read -r confirm
        [[ ! "$confirm" =~ ^[sS]$ ]] && return 0
        
        sudo rm -rf "$INSTALL_DIR"
    fi

    printf '%b\n' "${C6}⚙️  Extraindo pacote para ${INSTALL_DIR}...${RESET}"

    local temp_dir
    temp_dir=$(mktemp -d)
    trap 'sudo rm -rf "$temp_dir"' EXIT

    if ! sudo tar -xzf "$package_path" -C "$temp_dir" --strip-components=1; then
        die "Falha ao extrair o pacote."
    fi

    sudo mv "$temp_dir" "$INSTALL_DIR"
    sudo chmod +x "$INSTALL_DIR/bin/phpstorm.sh"

    sudo ln -sf "$INSTALL_DIR/bin/phpstorm.sh" "$SYMLINK"

    local icon_path="$INSTALL_DIR/bin/phpstorm.svg"
    [[ ! -f "$icon_path" ]] && icon_path="$INSTALL_DIR/bin/phpstorm.png"

    cat <<EOF | sudo tee "$DESKTOP_FILE" > /dev/null
[Desktop Entry]
Version=1.0
Type=Application
Name=PHPStorm
Icon=${icon_path}
Exec="${INSTALL_DIR}/bin/phpstorm.sh" %f
Comment=The Lightning-smart PHP IDE
Categories=Development;IDE;
Terminal=false
StartupWMClass=jetbrains-phpstorm
EOF

    if is_installed_cmd "update-desktop-database"; then
        sudo update-desktop-database /usr/share/applications &>/dev/null
    fi

    success "PHPStorm instalado com sucesso."
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
    
    local package_path
    package_path=$(get_phpstorm_package_path)
    
    install_phpstorm "$package_path"

    printf '\n%b\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    printf '%b\n' "   ${C4}Próximos passos:${RESET}"
    printf '%b\n' "   1. Inicie pelo terminal: ${C3}phpstorm${RESET}"
    printf '%b\n' "   2. Ou pelo menu: busque por ${C3}PHPStorm${RESET}"
    printf '%b\n\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
}

main "$@"
