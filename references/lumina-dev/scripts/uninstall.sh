#!/usr/bin/env bash
# =============================================================================
# Nome do Script : uninstall.sh
# Descrição      : Ferramenta de Remoção LuminaDev
# Versão         : 2.0.0
# =============================================================================

# --- opções do shell ---
set -euo pipefail

# --- constantes ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR

# --- carregamento de dependências ---
if [[ ! -f "$SCRIPT_DIR/utils.sh" ]]; then
    printf '%b\n' "\033[0;31m❌ Erro fatal: scripts/utils.sh não encontrado. Abortando.\033[0m"
    exit 1
fi
# shellcheck source=scripts/utils.sh
source "$SCRIPT_DIR/utils.sh"

# --- funções de interface ---

# =============================================================================
# Exibe o cabeçalho padrão com identificação do módulo.
# =============================================================================
show_header() {
    show_lumina_header "LuminaDev — Workstation Setup - DESINSTALAÇÃO"
}

# =============================================================================
# Exibe o menu principal de desinstalação.
# =============================================================================
show_uninstall_menu() {
    show_header
    printf '%b\n' "O que você deseja DESINSTALAR?"
    printf '%b\n' ""
    printf '%b\n' "  ${C3}1.${RESET} Fontes JetBrains Mono"
    printf '%b\n' "  ${C3}2.${RESET} IDEs"
    printf '%b\n' "  ${C3}3.${RESET} LLMs (IA)"
    printf '%b\n' "  ${C4}4.${RESET} Servidores MCP"
    printf '%b\n' "  ${C4}5.${RESET} Kitty Terminal"
    printf '%b\n' "  ${C1}0.${RESET} Voltar"
    printf '%b\n' ""
    printf '%s' "Selecione uma opção: "
}

# =============================================================================
# Exibe o submenu de desinstalação de IDEs.
# =============================================================================
show_ide_uninstall_menu() {
    while true; do
        show_header
        printf '%b\n' "Qual IDE/Editor você deseja DESINSTALAR?"
        printf '%b\n' ""
        printf '%b\n' "  ${C3}1.${RESET} Zed Editor"
        printf '%b\n' "  ${C3}2.${RESET} VSCodium"
        printf '%b\n' "  ${C3}3.${RESET} VS Code"
        printf '%b\n' "  ${C3}4.${RESET} PHPStorm"
        printf '%b\n' "  ${C3}5.${RESET} Windsurf"
        printf '%b\n' "  ${C1}0.${RESET} Voltar"
        printf '%b\n' ""
        printf '%s' "Selecione uma opção: "
        read -r choice

        case "$choice" in
            1) remove_zed ;;
            2) remove_vscodium ;;
            3) remove_vscode ;;
            4) remove_phpstorm ;;
            5) remove_windsurf ;;
            0) return 0 ;;
            *) printf '%b\n' "${C1}❌ Opção inválida.${RESET}" ;;
        esac
        if [[ "$choice" =~ ^[1-5]$ ]]; then
            printf '%b\n' "${C5}─────────────────────────────────────────${RESET}"
            read -r -p "   Pressione Enter para continuar..."
        fi
    done
}

# =============================================================================
# Exibe o submenu de desinstalação de LLMs.
# =============================================================================
show_llm_uninstall_menu() {
    while true; do
        show_header
        printf '%b\n' "Qual LLM (IA) você deseja DESINSTALAR?"
        printf '%b\n' ""
        printf '%b\n' "  ${C3}1.${RESET} Claude Code CLI"
        printf '%b\n' "  ${C3}2.${RESET} Gemini Code Assist CLI"
        printf '%b\n' "  ${C3}3.${RESET} OpenCode CLI"
        printf '%b\n' "  ${C3}4.${RESET} OpenCode Desktop"
        printf '%b\n' "  ${C3}5.${RESET} OpenAI Codex CLI"
        printf '%b\n' "  ${C1}0.${RESET} Voltar"
        printf '%b\n' ""
        printf '%s' "Selecione uma opção: "
        read -r choice

        case "$choice" in
            1) remove_claude ;;
            2) remove_gemini ;;
            3) remove_opencode_cli ;;
            4) remove_opencode_desktop ;;
            5) remove_codex ;;
            0) return 0 ;;
            *) printf '%b\n' "${C1}❌ Opção inválida.${RESET}" ;;
        esac
        if [[ "$choice" =~ ^[1-5]$ ]]; then
            printf '%b\n' "${C5}─────────────────────────────────────────${RESET}"
            read -r -p "   Pressione Enter para continuar..."
        fi
    done
}

# --- funções auxiliares ---

# =============================================================================
# Confirma a remoção de um item.
# =============================================================================
confirm_removal() {
    local label="$1"
    printf '%b\n' "${C3}⚠️  Ação Destrutiva: ${C4}${label}${RESET}"
    echo -ne "   Tem certeza que deseja remover? (${C1}s${RESET}/N): "
    read -r confirm
    [[ "$confirm" =~ ^[sS]$ ]]
}

# =============================================================================
# Remove um pacote do sistema usando o gerenciador detectado.
# =============================================================================
uninstall_pkg() {
    local pkg="$1"
    case "$PKG_MANAGER" in
        apt)
            is_installed_pkg "$pkg" && sudo apt-get purge -y "$pkg" && sudo apt-get autoremove -y || true
            ;;
        dnf)
            is_installed_pkg "$pkg" && sudo dnf remove -y "$pkg" || true
            ;;
        pacman)
            is_installed_pkg "$pkg" && sudo pacman -Rns --noconfirm "$pkg" || true
            ;;
    esac
}

# --- funções de negócio ---

# =============================================================================
# Remove o VS Code.
# =============================================================================
remove_vscode() {
    confirm_removal "VS Code" || return 0
    printf '%b\n' "${C6}🗑️  Removendo VS Code...${RESET}"
    uninstall_pkg "code"

    case "$PKG_MANAGER" in
        apt)
            sudo rm -f /etc/apt/sources.list.d/vscode.list
            sudo rm -f /usr/share/keyrings/microsoft-archive-keyring.gpg
            sudo apt-get update -qq
            ;;
        dnf)
            sudo rm -f /etc/yum.repos.d/vscode.repo
            ;;
    esac

    rm -rf "$HOME/.vscode" "$HOME/.config/Code"
    success "VS Code removido."
}

# =============================================================================
# Remove o VSCodium.
# =============================================================================
remove_vscodium() {
    confirm_removal "VSCodium" || return 0
    printf '%b\n' "${C6}🗑️  Removendo VSCodium...${RESET}"
    uninstall_pkg "codium"

    if [[ "$PKG_MANAGER" == "apt" ]]; then
        sudo rm -f /etc/apt/sources.list.d/vscodium.sources
        sudo rm -f /usr/share/keyrings/vscodium-archive-keyring.gpg
        sudo apt-get update -qq
    elif [[ "$PKG_MANAGER" == "dnf" ]]; then
        sudo rm -f /etc/yum.repos.d/vscodium.repo
    fi

    rm -rf "$HOME/.vscode-oss" "$HOME/.config/VSCodium"
    success "VSCodium removido."
}

# =============================================================================
# Remove o Zed Editor.
# =============================================================================
remove_zed() {
    confirm_removal "Zed Editor" || return 0
    printf '%b\n' "${C6}🗑️  Removendo Zed Editor...${RESET}"
    rm -rf "$HOME/.local/bin/zed" "$HOME/.local/zed.app" "$HOME/.config/zed" "$HOME/.local/share/zed"
    success "Zed Editor removido."
}

# =============================================================================
# Remove o PHPStorm.
# =============================================================================
remove_phpstorm() {
    confirm_removal "PHPStorm" || return 0
    printf '%b\n' "${C6}🗑️  Removendo PHPStorm...${RESET}"
    sudo rm -rf "/opt/phpstorm" "/usr/local/bin/phpstorm" "/usr/bin/phpstorm" "/usr/share/applications/phpstorm.desktop"

    if is_installed_cmd "update-desktop-database"; then
        sudo update-desktop-database /usr/share/applications &>/dev/null || true
    fi
    success "PHPStorm removido."
}

# =============================================================================
# Remove o Windsurf.
# =============================================================================
remove_windsurf() {
    confirm_removal "Windsurf" || return 0
    printf '%b\n' "${C6}🗑️  Removendo Windsurf...${RESET}"

    uninstall_pkg "windsurf"

    case "$PKG_MANAGER" in
        apt)
            sudo rm -f /etc/apt/keyrings/windsurf-stable.gpg
            sudo rm -f /etc/apt/sources.list.d/windsurf.list
            sudo apt-get update -qq
            ;;
        dnf)
            sudo rm -f /etc/yum.repos.d/windsurf.repo
            ;;
    esac

    rm -rf "$HOME/.codeium/windsurf" "$HOME/.config/windsurf"
    success "Windsurf removido."
}

# =============================================================================
# Remove o Claude Code CLI.
# =============================================================================
remove_claude() {
    confirm_removal "Claude Code CLI" || return 0
    printf '%b\n' "${C6}🗑️  Removendo Claude Code CLI...${RESET}"
    rm -f  "$HOME/.local/bin/claude"
    rm -rf "$HOME/.local/share/claude"
    rm -rf "$HOME/.config/claude"
    rm -rf "$HOME/.cache/claude"
    if is_installed_cmd "claude"; then
        warn "claude ainda encontrado em: $(type -P claude)"
    else
        success "Claude Code CLI removido."
    fi
    info "Utilize nano ~/.bashrc para remover as variáveis de ambiente"
}

# =============================================================================
# Remove o Gemini Code Assist CLI.
# =============================================================================
remove_gemini() {
    confirm_removal "Gemini Code Assist CLI" || return 0
    printf '%b\n' "${C6}🗑️  Removendo Gemini Code Assist CLI...${RESET}"
    if is_installed_cmd "npm"; then
        sudo npm uninstall -g @google/gemini-cli 2>/dev/null || true
    fi
    if is_installed_cmd "gemini"; then
        warn "gemini ainda encontrado em: $(type -P gemini)"
    else
        success "Gemini Code Assist CLI removido."
    fi
}

# =============================================================================
# Remove o OpenCode CLI.
# =============================================================================
remove_opencode_cli() {
    confirm_removal "OpenCode CLI" || return 0
    printf '%b\n' "${C6}🗑️  Removendo OpenCode CLI...${RESET}"
    if is_installed_cmd "npm"; then
        sudo npm uninstall -g opencode-ai 2>/dev/null || true
    fi
    rm -rf "$HOME/.opencode"
    rm -rf "$HOME/.config/opencode"
    rm -rf "$HOME/.cache/opencode"
    if is_installed_cmd "opencode"; then
        warn "opencode ainda encontrado em: $(type -P opencode)"
    else
        success "OpenCode CLI removido."
    fi
}

# =============================================================================
# Remove o OpenCode Desktop.
# =============================================================================
remove_opencode_desktop() {
    confirm_removal "OpenCode Desktop" || return 0
    printf '%b\n' "${C6}🗑️  Removendo OpenCode Desktop...${RESET}"
    case "$PKG_MANAGER" in
        apt|dnf)
            uninstall_pkg "opencode-desktop"
            ;;
        pacman)
            rm -f "$HOME/.local/bin/opencode-desktop"
            ;;
    esac
    success "OpenCode Desktop removido."
}

# =============================================================================
# Remove o OpenAI Codex CLI.
# =============================================================================
remove_codex() {
    confirm_removal "OpenAI Codex CLI" || return 0
    printf '%b\n' "${C6}🗑️  Removendo OpenAI Codex CLI...${RESET}"
    if is_installed_cmd "npm"; then
        sudo npm uninstall -g @openai/codex 2>/dev/null || true
    fi
    rm -rf "$HOME/.codex"
    if is_installed_cmd "codex"; then
        warn "codex ainda encontrado em: $(type -P codex)"
    else
        success "OpenAI Codex CLI removido."
    fi
}

# =============================================================================
# Remove Servidores MCP.
# =============================================================================
remove_mcp_servers() {
    confirm_removal "Servidores MCP" || return 0
    printf '%b\n' "${C6}🗑️  Removendo servidores MCP...${RESET}"

    if is_installed_cmd "npm"; then
        sudo npm uninstall -g moodle-dev-mcp lumina-ai-vault 2>/dev/null || true

        if is_installed_cmd "claude"; then
            claude mcp remove moodle-dev-mcp 2>/dev/null || true
            claude mcp remove lumina-aivault 2>/dev/null || true
        fi
    fi
    success "Servidores MCP removidos."
}

# =============================================================================
# Remove Fontes.
# =============================================================================
remove_fonts() {
    confirm_removal "JetBrains Mono" || return 0
    printf '%b\n' "${C6}🗑️  Removendo fontes JetBrains Mono...${RESET}"
    find "$HOME/.local/share/fonts" -name "JetBrainsMono*" -delete 2>/dev/null || true
    fc-cache -f
    success "Fontes JetBrains Mono removidas."
}

# =============================================================================
# Remove Kitty.
# =============================================================================
remove_kitty() {
    confirm_removal "Kitty Terminal" || return 0
    printf '%b\n' "${C6}🗑️  Removendo Kitty Terminal...${RESET}"
    rm -rf "$HOME/.local/kitty.app" "$HOME/.local/bin/kitty" "$HOME/.local/bin/kitten" \
           "$HOME/.local/share/applications/kitty.desktop" "$HOME/.config/kitty"

    case "$PKG_MANAGER" in
        apt)
            if is_installed_cmd "update-alternatives"; then
                sudo update-alternatives --remove x-terminal-emulator "$HOME/.local/bin/kitty" 2>/dev/null || true
            fi
            ;;
        dnf|pacman)
            warn "Se o Kitty era seu terminal padrão, configure outro via: xdg-settings set default-terminal-emulator <outro.desktop>"
            ;;
    esac
    success "Kitty Terminal removido."
}

# --- ponto de entrada ---

# =============================================================================
# Ponto de entrada principal do script.
# =============================================================================
main() {
    detect_pkg_manager
    require_not_root
    require_sudo

    while true; do
        show_uninstall_menu
        read -r choice
        case "$choice" in
            1) remove_fonts ;;
            2) show_ide_uninstall_menu ;;
            3) show_llm_uninstall_menu ;;
            4) remove_mcp_servers ;;
            5) remove_kitty ;;
            0) return 0 ;;
            *) printf '%b\n' "${C1}❌ Opção inválida.${RESET}" ;;
        esac

        if [[ "$choice" =~ ^[145]$ ]]; then
            printf '%b\n' "${C5}─────────────────────────────────────────${RESET}"
            read -r -p "   Pressione Enter para continuar..."
        fi
    done
}

main "$@"
