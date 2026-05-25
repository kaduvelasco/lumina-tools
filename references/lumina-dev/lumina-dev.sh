#!/usr/bin/env bash
# =============================================================================
# Nome do Script : lumina-dev.sh
# Descrição      : Painel de Controle LuminaDev — Central de Instalação
# Versão         : 2.0.0
# =============================================================================

# --- opções do shell ---
set -euo pipefail

# --- constantes ---
BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly BASE_DIR

# --- carregamento de dependências ---
if [[ ! -f "$BASE_DIR/scripts/utils.sh" ]]; then
    printf '%b\n' "\033[0;31m❌ Erro fatal: scripts/utils.sh não encontrado. Abortando.\033[0m"
    exit 1
fi
# shellcheck source=scripts/utils.sh
source "$BASE_DIR/scripts/utils.sh"

# --- funções de interface ---

# =============================================================================
# Exibe o cabeçalho principal do painel LuminaDev.
# =============================================================================
show_header() {
    show_lumina_header "LuminaDev — Workstation Setup"
}

# =============================================================================
# Exibe o menu principal.
# =============================================================================
show_main_menu() {
    show_header
    printf '%b\n' "O que você deseja fazer?"
    printf '%b\n' ""
    printf '%b\n' "  ${C2}1.${RESET} Instalar Fontes JetBrains Mono"
    printf '%b\n' "  ${C2}2.${RESET} Instalar Git e libsecret"
    printf '%b\n' "  ${C2}3.${RESET} Instalar LLMs (IA)"
    printf '%b\n' "  ${C2}4.${RESET} Instalar IDEs"
    printf '%b\n' "  ${C2}5.${RESET} Instalar Servidores MCP"
    printf '%b\n' "  ${C2}6.${RESET} Instalar Kitty Terminal"
    printf '%b\n' "  ${C4}7.${RESET} Desinstalador"
    printf '%b\n' "  ${C1}0.${RESET} Sair"
    printf '%b\n' ""
    echo -ne "Selecione uma opção: "
}

# =============================================================================
# Exibe o menu de LLMs.
# =============================================================================
show_llm_menu() {
    while true; do
        show_header
        printf '%b\n' "Qual LLMS (IA) você deseja instalar?"
        printf '%b\n' ""
        printf '%b\n' "  ${C2}1.${RESET} Claude Code CLI"
        printf '%b\n' "  ${C2}2.${RESET} Gemini Code Assist CLI"
        printf '%b\n' "  ${C2}3.${RESET} OpenCode CLI"
        printf '%b\n' "  ${C2}4.${RESET} OpenCode Desktop"
        printf '%b\n' "  ${C2}5.${RESET} OpenAI Codex CLI"
        printf '%b\n' "  ${C1}0.${RESET} Voltar"
        printf '%b\n' ""
        echo -ne "Selecione uma opção: "
        read -r choice

        case "$choice" in
            1) run_script "scripts" "claude-install.sh" ;;
            2) run_script "scripts" "gemini-install.sh" ;;
            3) run_script "scripts" "opencode-install.sh" "cli" ;;
            4) run_script "scripts" "opencode-install.sh" "desktop" ;;
            5) run_script "scripts" "codex-install.sh" ;;
            0) break ;;
            *) printf '%b\n' "${C1}❌ Opção inválida.${RESET}" ;;
        esac
    done
}

# =============================================================================
# Exibe o menu de IDEs.
# =============================================================================
show_ide_menu() {
    while true; do
        show_header
        printf '%b\n' "Qual IDE/Editor você deseja instalar?"
        printf '%b\n' ""
        printf '%b\n' "  ${C2}1.${RESET} Zed Editor"
        printf '%b\n' "  ${C2}2.${RESET} VSCodium"
        printf '%b\n' "  ${C2}3.${RESET} VS Code"
        printf '%b\n' "  ${C2}4.${RESET} PHPStorm (Auxiliar)"
        printf '%b\n' "  ${C2}5.${RESET} Windsurf"
        printf '%b\n' "  ${C1}0.${RESET} Voltar"
        printf '%b\n' ""
        echo -ne "Selecione uma opção: "
        read -r choice

        case "$choice" in
            1) run_script "ides" "zed-install.sh" ;;
            2) run_script "ides" "vscodium-install.sh" ;;
            3) run_script "ides" "vscode-install.sh" ;;
            4) run_script "ides" "phpstorm-install.sh" ;;
            5) run_script "ides" "windsurf-install.sh" ;;
            0) break ;;
            *) printf '%b\n' "${C1}❌ Opção inválida.${RESET}" ;;
        esac
    done
}

# =============================================================================
# Exibe o menu de MCP.
# =============================================================================
show_mcp_menu() {
    while true; do
        show_header
        printf '%b\n' "Qual Servidor MCP você deseja instalar?"
        printf '%b\n' ""
        printf '%b\n' "  ${C2}1.${RESET} Moodle Dev MCP"
        printf '%b\n' "  ${C2}2.${RESET} Lumina AI Vault"
        printf '%b\n' "  ${C2}3.${RESET} Code Review Graph"
        printf '%b\n' "  ${C1}0.${RESET} Voltar"
        printf '%b\n' ""
        echo -ne "Selecione uma opção: "
        read -r choice

        case "$choice" in
            1) run_script "scripts" "mcp-install.sh" "moodle-dev-mcp" ;;
            2) run_script "scripts" "mcp-install.sh" "lumina-ai-vault" ;;
            3) run_script "scripts" "mcp-install.sh" "code-review-graph" ;;
            0) break ;;
            *) printf '%b\n' "${C1}❌ Opção inválida.${RESET}" ;;
        esac
    done
}

# --- funções de negócio ---

# =============================================================================
# Executa um script secundário. $1=pasta $2=script [$@=args extras].
# =============================================================================
run_script() {
    local folder="$1"
    local script_name="$2"
    shift 2
    local full_path="$BASE_DIR/$folder/$script_name"

    if [[ -f "$full_path" ]]; then
        printf '\n%b\n' "${C6}🚀 Executando $folder/$script_name...${RESET}"
        bash "$full_path" "$@"
    else
        die "Arquivo '$script_name' não encontrado em '$folder/'."
    fi
}

# =============================================================================
# Instala Git e configura libsecret.
# =============================================================================
install_git_libsecret() {
    printf '%b\n' "${C6}📦 Verificando Git e libsecret...${RESET}"

    case "$PKG_MANAGER" in
        apt)
            sudo apt-get update -qq
            ensure_pkg "git"
            ensure_pkg "libsecret-1-0"
            ensure_pkg "libsecret-1-dev"
            ensure_pkg "build-essential"
            ;;
        dnf)
            sudo dnf check-update -q || true
            ensure_pkg "git"
            ensure_pkg "libsecret"
            ensure_pkg "libsecret-devel"
            ensure_pkg "gcc"
            ensure_pkg "make"
            ;;
        pacman)
            sudo pacman -Sy --noconfirm
            ensure_pkg "git"
            ensure_pkg "libsecret"
            ensure_pkg "base-devel"
            ;;
    esac

    printf '%b\n' "${C6}⚙️  Configurando Git para usar libsecret...${RESET}"
    if [[ "$PKG_MANAGER" == "apt" ]]; then
        local libsecret_path="/usr/share/doc/git/contrib/credential/libsecret/git-credential-libsecret"
        local libsecret_dir="/usr/share/doc/git/contrib/credential/libsecret"
        if [[ ! -f "$libsecret_path" ]]; then
            printf '%b\n' "${C4}🛠️  Compilando o helper do libsecret...${RESET}"
            if [[ ! -d "$libsecret_dir" ]]; then
                die "Diretório do libsecret não encontrado.\n   Execute: sudo apt install git libsecret-1-dev build-essential"
            fi
            if ! sudo make -C "$libsecret_dir"; then
                die "Falha na compilação do libsecret. Verifique as dependências."
            fi
        fi
        git config --global credential.helper "$libsecret_path"
    else
        git config --global credential.helper libsecret
    fi

    success "Git e libsecret configurados."
}

# --- ponto de entrada ---

# =============================================================================
# Ponto de entrada principal do script.
# =============================================================================
main() {
    detect_pkg_manager
    require_not_root
    require_sudo

    printf '%b\n' "${C6}⚙️  Sincronizando permissões de execução...${RESET}"
    find "$BASE_DIR/scripts" -name "*.sh" -exec chmod +x {} + 2>/dev/null || true
    find "$BASE_DIR/ides"    -name "*.sh" -exec chmod +x {} + 2>/dev/null || true

    while true; do
        show_main_menu
        read -r choice
        case "$choice" in
            1) run_script "scripts" "fonts-install.sh" ;;
            2) install_git_libsecret ;;
            3) show_llm_menu ;;
            4) show_ide_menu ;;
            5) show_mcp_menu ;;
            6) run_script "scripts" "kitty-install.sh" ;;
            7) run_script "scripts" "uninstall.sh" ;;
            0)
                printf '\n%b\n\n' "${C6}👋 Até logo!${RESET}"
                exit 0
                ;;
            *)
                printf '\n%b\n' "${C1}❌ Opção inválida.${RESET}"
                ;;
        esac
    done
}

main "$@"
