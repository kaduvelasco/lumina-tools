#!/usr/bin/env bash
# =============================================================================
# Nome do Script : fonts-install.sh
# Descrição      : Instalação da fonte JetBrains Mono
# Versão         : 2.0.0
# =============================================================================

# --- opções do shell ---
set -euo pipefail

# --- constantes ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
readonly FONT_NAME="JetBrains Mono"
readonly FONT_VERSION="2.304"
readonly FONT_URL="https://github.com/JetBrains/JetBrainsMono/releases/download/v${FONT_VERSION}/JetBrainsMono-${FONT_VERSION}.zip"
readonly FONT_DIR="$HOME/.local/share/fonts"
readonly FONT_CHECK="JetBrainsMono-Regular.ttf"

# --- carregamento de dependências ---
# shellcheck source=scripts/utils.sh
source "$SCRIPT_DIR/utils.sh"

# --- funções auxiliares ---

# =============================================================================
# Verifica e instala dependências necessárias.
# =============================================================================
check_font_dependencies() {
    local missing=()

    for cmd in curl unzip fc-cache; do
        if ! is_installed_cmd "$cmd"; then
            missing+=("$cmd")
        fi
    done

    if [[ ${#missing[@]} -gt 0 ]]; then
        echo -e "${C6}⚙️  Instalando dependências ausentes: ${C4}${missing[*]}${RESET}"
        detect_pkg_manager
        for pkg in "${missing[@]}"; do
            case "$pkg" in
                fc-cache) ensure_pkg "fontconfig" ;;
                *)        ensure_pkg "$pkg" ;;
            esac
        done
    fi
}

# --- funções de negócio ---

# =============================================================================
# Baixa e instala a fonte JetBrains Mono.
# =============================================================================
install_fonts() {
    if [[ -f "$FONT_DIR/$FONT_CHECK" ]]; then
        echo -e "${C2}✅ ${FONT_NAME} já está instalada em ${C4}${FONT_DIR}${RESET}."
        echo -ne "   Deseja reinstalar / atualizar para a v${FONT_VERSION}? (${C3}s${RESET}/N): "
        read -r confirm
        [[ ! "$confirm" =~ ^[sS]$ ]] && return 0
    fi

    check_font_dependencies
    require_internet

    echo -e "${C6}⚙️  Iniciando instalação da fonte ${C4}${FONT_NAME} v${FONT_VERSION}${RESET}..."

    mkdir -p "$FONT_DIR"

    local temp_dir
    temp_dir=$(mktemp -d)
    trap 'rm -rf "$temp_dir"' EXIT

    echo -e "${C6}📥 Baixando fontes oficiais...${RESET}"
    if ! curl -fsSL "$FONT_URL" -o "$temp_dir/fonts.zip"; then
        die "Erro ao baixar as fontes. Verifique sua conexão."
    fi

    echo -e "${C6}📦 Extraindo arquivos...${RESET}"
    if ! unzip -q "$temp_dir/fonts.zip" -d "$temp_dir"; then
        die "Erro ao extrair o arquivo zip."
    fi

    local count
    count=$(find "$temp_dir" -name "*.ttf" | wc -l)

    if [[ "$count" -eq 0 ]]; then
        die "Nenhum arquivo .ttf encontrado no pacote baixado."
    fi

    find "$temp_dir" -maxdepth 3 -type f -name "*.ttf" -exec cp -- {} "$FONT_DIR/" \;
    echo -e "${C2}✅ ${count} arquivos copiados para ${C4}${FONT_DIR}${RESET}."

    echo -e "${C6}🔄 Atualizando cache de fontes do sistema...${RESET}"
    fc-cache -f

    success "Fonte ${FONT_NAME} instalada com sucesso."
}

# --- ponto de entrada ---

# =============================================================================
# Ponto de entrada principal do script.
# =============================================================================
main() {
    detect_pkg_manager
    install_fonts
}

main "$@"
