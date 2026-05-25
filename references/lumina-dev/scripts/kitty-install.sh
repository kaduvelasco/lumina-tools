#!/usr/bin/env bash
# =============================================================================
# Nome do Script : kitty-install.sh
# Descrição      : Instalação e configuração do Kitty Terminal e Starship
# Versão         : 2.0.0
# =============================================================================

# --- opções do shell ---
set -euo pipefail

# --- constantes ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
readonly KITTY_CMD="kitty"
readonly KITTY_CONFIG_DIR="$HOME/.config/kitty"
readonly KITTY_CONFIG_FILE="$KITTY_CONFIG_DIR/kitty.conf"
readonly LOCAL_BIN="$HOME/.local/bin"
readonly FONT_DIR="$HOME/.local/share/fonts"
readonly NERD_FONT_VERSION="3.3.0"
readonly NERD_FONT_CHECK="JetBrainsMonoNerdFont-Regular.ttf"

# --- carregamento de dependências ---
if [[ ! -f "$SCRIPT_DIR/utils.sh" ]]; then
    printf '\033[0;31m❌ Erro fatal: utils.sh não encontrado. Abortando.\033[0m\n' >&2
    exit 1
fi
# shellcheck source=scripts/utils.sh
source "$SCRIPT_DIR/utils.sh"

# --- funções de interface ---

# =============================================================================
# Exibe o cabeçalho padrão com identificação do módulo.
# =============================================================================
show_header() {
    show_lumina_header
    printf '%b\n' "   ${C5}INSTALADOR KITTY TERMINAL${RESET}"
    printf '%b\n\n' "   ${C5}Distro : ${C4}${PKG_MANAGER}${RESET}"
}

# --- funções de negócio ---

# =============================================================================
# Instala JetBrains Mono Nerd Font.
# =============================================================================
install_nerd_font() {
    if [[ -f "$FONT_DIR/$NERD_FONT_CHECK" ]]; then
        printf '%b\n' "${C2}✅ JetBrainsMono Nerd Font já está instalada.${RESET}"
        echo -ne "   Deseja reinstalar / atualizar para a v${NERD_FONT_VERSION}? (${C3}s${RESET}/N): "
        read -r confirm
        [[ ! "$confirm" =~ ^[sS]$ ]] && return 0
    fi

    require_internet

    printf '%b\n' "${C6}⚙️  Baixando JetBrainsMono Nerd Font v${NERD_FONT_VERSION}...${RESET}"

    local temp_dir
    temp_dir=$(mktemp -d)
    trap 'rm -rf "$temp_dir"' EXIT

    local font_url="https://github.com/ryanoasis/nerd-fonts/releases/download/v${NERD_FONT_VERSION}/JetBrainsMono.zip"

    if ! curl -fsSL "$font_url" -o "$temp_dir/JetBrainsMono.zip"; then
        die "Falha ao baixar a Nerd Font."
    fi

    unzip -q "$temp_dir/JetBrainsMono.zip" -d "$temp_dir/JetBrainsMono"
    mkdir -p "$FONT_DIR"
    cp "$temp_dir/JetBrainsMono"/*.ttf "$FONT_DIR/"
    fc-cache -f

    printf '%b\n' "${C2}✅ JetBrainsMono Nerd Font instalada em ${C4}${FONT_DIR}${RESET}."

    rm -rf "$temp_dir"
    trap - EXIT
}

# =============================================================================
# Instala o Kitty Terminal via script oficial.
# =============================================================================
install_kitty() {
    if is_installed_cmd "$KITTY_CMD"; then
        local current_version
        current_version=$(kitty --version 2>/dev/null || echo "versão desconhecida")
        printf '%b\n' "${C2}✅ Kitty já está instalado (${current_version}).${RESET}"
        echo -ne "   Deseja reinstalar / atualizar? (${C3}s${RESET}/N): "
        read -r confirm
        [[ ! "$confirm" =~ ^[sS]$ ]] && return 0
    fi

    require_internet

    printf '%b\n' "${C6}⚙️  Baixando instalador oficial do Kitty...${RESET}"

    local kitty_installer
    kitty_installer=$(mktemp)
    trap 'rm -f "$kitty_installer"' EXIT

    if ! curl -fsSL https://sw.kovidgoyal.net/kitty/installer.sh -o "$kitty_installer"; then
        die "Falha ao baixar instalador do Kitty."
    fi
    if ! sh "$kitty_installer"; then
        die "Falha ao instalar o Kitty. Verifique sua conexão e tente novamente."
    fi

    rm -f "$kitty_installer"
    trap - EXIT

    mkdir -p "$LOCAL_BIN"
    ln -sf "$HOME/.local/kitty.app/bin/kitty"  "$LOCAL_BIN/kitty"
    ln -sf "$HOME/.local/kitty.app/bin/kitten" "$LOCAL_BIN/kitten"

    if [[ ":$PATH:" != *":$LOCAL_BIN:"* ]]; then
        ensure_local_bin_in_path
    fi

    local desktop_src="$HOME/.local/kitty.app/share/applications/kitty.desktop"
    local desktop_dst="$HOME/.local/share/applications/kitty.desktop"

    if [[ -f "$desktop_src" ]]; then
        mkdir -p "$HOME/.local/share/applications"
        cp "$desktop_src" "$desktop_dst"
        sed -i "s|Icon=kitty|Icon=$HOME/.local/kitty.app/share/icons/hicolor/256x256/apps/kitty.png|g" "$desktop_dst"
        sed -i "s|TryExec=kitty|TryExec=$LOCAL_BIN/kitty|g" "$desktop_dst"
        sed -i "s|^Exec=kitty$|Exec=$LOCAL_BIN/kitty|g" "$desktop_dst"
    fi

    if is_installed_cmd "update-desktop-database"; then
        update-desktop-database "$HOME/.local/share/applications" 2>/dev/null || true
    fi

    success "Kitty instalado com sucesso."
}

# =============================================================================
# Aplica configurações personalizadas do Kitty.
# =============================================================================
apply_kitty_settings() {
    printf '%b\n' "${C6}⚙️  Aplicando configurações personalizadas...${RESET}"

    mkdir -p "$KITTY_CONFIG_DIR"

    if [[ -f "$KITTY_CONFIG_FILE" ]]; then
        echo -ne "   Configuração existente encontrada. Sobrescrever? (${C3}s${RESET}/N): "
        read -r confirm
        if [[ ! "$confirm" =~ ^[sS]$ ]]; then
            printf '%b\n' "${C4}↩️  Configuração mantida.${RESET}"
            return 0
        fi

        local backup
        backup="${KITTY_CONFIG_FILE}.bak.$(date +%Y%m%d%H%M%S)"
        cp "$KITTY_CONFIG_FILE" "$backup"
    fi

    cat <<'EOF' > "$KITTY_CONFIG_FILE"
# kitty.conf — LuminaDev
repaint_delay       10
input_delay         3
sync_to_monitor     yes
font_family         JetBrainsMono Nerd Font
bold_font           auto
italic_font         auto
bold_italic_font    auto
font_size           11.0
disable_ligatures   never
remember_window_size        yes
initial_window_width        800
initial_window_height       500
background_opacity          0.95
background_blur             1
dynamic_background_opacity  yes
hide_window_decorations     no
window_padding_width        6
confirm_os_window_close     1
copy_on_select              yes
url_color                   #F2D5CF
cursor_shape                block
cursor_blink_interval       0
scrollback_lines    10000
enabled_layouts     tall:bias=50;full_size=1;mirrored=false,stack
tab_bar_style           powerline
tab_powerline_style     slanted
# BEGIN_KITTY_THEME
include current-theme.conf
# END_KITTY_THEME
EOF

    # Cria tema padrão apenas se não existir
    if [[ ! -e "$KITTY_CONFIG_DIR/current-theme.conf" ]]; then
        cat <<'EOF' > "$KITTY_CONFIG_DIR/current-theme.conf"
# Tema padrão Kitty — gerado pelo LuminaDev
# Substitua pelo tema preferido via: kitty +kitten themes
background            #1e1e2e
foreground            #cdd6f4
color0                #45475a
color8                #585b70
color1                #f38ba8
color9                #f38ba8
color2                #a6e3a1
color10               #a6e3a1
color3                #f9e2af
color11               #f9e2af
color4                #89b4fa
color12               #89b4fa
color5                #f5c2e7
color13               #f5c2e7
color6                #94e2d5
color14               #94e2d5
color7                #bac2de
color15               #a6adc8
EOF
    fi
    success "Configurações do Kitty aplicadas."
}

# =============================================================================
# Instalação e configuração do Starship prompt.
# =============================================================================
install_starship() {
    printf '\n%b\n' "${C6}⚙️  Configurando Starship Prompt...${RESET}"

    local do_install=true
    if is_installed_cmd "starship"; then
        echo -ne "   Starship já instalado. Reinstalar / Atualizar? (${C3}s${RESET}/N): "
        read -r confirm
        [[ ! "$confirm" =~ ^[sS]$ ]] && do_install=false
    fi

    if [[ "$do_install" == true ]]; then
        require_internet
        local starship_installer
        starship_installer=$(mktemp)
        trap 'rm -f "$starship_installer"' EXIT
        if ! curl -fsSL https://starship.rs/install.sh -o "$starship_installer"; then
            die "Falha ao baixar instalador do Starship."
        fi
        sh "$starship_installer" --yes --bin-dir "$LOCAL_BIN"
        rm -f "$starship_installer"
        trap - EXIT
    fi

    printf '%b\n' "TEMAS DO STARSHIP"
    printf '%b\n' ""
    printf '%b\n' "${C2}1.${RESET} gruvbox-rainbow"
    printf '%b\n' "${C2}2.${RESET} tokyo-night"
    printf '%b\n' "${C2}3.${RESET} pastel-powerline"
    printf '%b\n' "${C2}4.${RESET} pure-preset"
    printf '%b\n' ""
    echo -ne "Escolha um tema [1-4] (Enter = 1): "
    read -r preset_choice

    local preset_name
    case "$preset_choice" in
        2) preset_name="tokyo-night" ;;
        3) preset_name="pastel-powerline" ;;
        4) preset_name="pure-preset" ;;
        *) preset_name="gruvbox-rainbow" ;;
    esac

    "$LOCAL_BIN/starship" preset "$preset_name" -o "$HOME/.config/starship.toml"

    local init_line='eval "$(starship init bash)"'
    if ! grep -qF "$init_line" "$HOME/.bashrc" 2>/dev/null; then
        printf '%b\n' "\n# Starship prompt\n$init_line" >> "$HOME/.bashrc"
    fi

    success "Starship configurado (${preset_name})."
}

# =============================================================================
# Define o Kitty como terminal padrão (opcional).
# =============================================================================
set_default_terminal() {
    echo -ne "\nDefinir o Kitty como terminal padrão? (${C3}s${RESET}/N): "
    read -r confirm
    [[ ! "$confirm" =~ ^[sS]$ ]] && return 0

    local kitty_bin
    kitty_bin=$(command -v kitty 2>/dev/null || echo "$LOCAL_BIN/kitty")

    case "$PKG_MANAGER" in
        apt)
            if is_installed_cmd "update-alternatives"; then
                sudo update-alternatives --install /usr/bin/x-terminal-emulator x-terminal-emulator "$kitty_bin" 50
                sudo update-alternatives --set x-terminal-emulator "$kitty_bin"
            fi
            ;;
        dnf|pacman)
            ensure_pkg "xdg-utils"
            xdg-settings set default-terminal-emulator kitty.desktop 2>/dev/null || \
                warn "Não foi possível definir o terminal padrão via xdg-settings."
            ;;
    esac
    printf '%b\n' "${C2}✅ Kitty definido como terminal padrão.${RESET}"
}

# =============================================================================
# Insere ação "Abrir no Kitty" no Thunar via uca.xml.
# $1 = caminho absoluto do binário kitty
# =============================================================================
_install_thunar_action() {
    local kitty_bin="$1"
    local action_id="kitty-open-here"
    local uca_file="$HOME/.config/Thunar/uca.xml"

    mkdir -p "$HOME/.config/Thunar"

    if grep -q "$action_id" "$uca_file" 2>/dev/null; then
        printf '%b\n' "   ${C2}✅ Thunar: ação já configurada.${RESET}"
        return 0
    fi

    local action_block
    action_block="$(cat <<THUNAR_EOF
<action>
    <icon>utilities-terminal</icon>
    <name>Abrir no Kitty</name>
    <unique-id>${action_id}</unique-id>
    <command>${kitty_bin} --directory %f</command>
    <description>Abrir o terminal Kitty nesta pasta</description>
    <patterns>*</patterns>
    <startup-notify>false</startup-notify>
    <directories/>
</action>
THUNAR_EOF
)"

    if [[ ! -f "$uca_file" ]]; then
        cat > "$uca_file" <<THUNAR_EOF
<?xml version="1.0" encoding="UTF-8"?>
<actions>
${action_block}
</actions>
THUNAR_EOF
    else
        local tmp_file
        tmp_file=$(mktemp)
        while IFS= read -r line; do
            if [[ "$line" =~ \</actions\> ]]; then
                echo "$action_block"
            fi
            printf '%s\n' "$line"
        done < "$uca_file" > "$tmp_file"
        mv "$tmp_file" "$uca_file"
    fi
}

# =============================================================================
# Configura "Abrir no Kitty" no menu de contexto do gerenciador de arquivos.
# Suporta: Nemo, Nautilus, Dolphin, Thunar, COSMIC Files (orientação).
# =============================================================================
setup_context_menu() {
    echo -ne "\nConfigurar 'Abrir no Kitty' no menu de contexto? (${C3}s${RESET}/N): "
    read -r confirm
    [[ ! "$confirm" =~ ^[sS]$ ]] && return 0

    local kitty_bin
    kitty_bin=$(command -v kitty 2>/dev/null || echo "$LOCAL_BIN/kitty")
    local installed_any=false

    # --- Nemo (Linux Mint / Cinnamon) ---
    if is_installed_cmd "nemo"; then
        local nemo_dir="$HOME/.local/share/nemo/actions"
        mkdir -p "$nemo_dir"
        cat > "$nemo_dir/open_in_kitty.nemo_action" <<NEMO_EOF
[Nemo Action]
Active=true
Name=Abrir no Kitty
Comment=Abrir o terminal Kitty nesta pasta
Exec=${kitty_bin} --directory %P
Selection=none
Extensions=dir;
Icon-Name=terminal
NEMO_EOF
        printf '%b\n' "   ${C2}✅ Nemo (Linux Mint): ação configurada.${RESET}"
        installed_any=true
    fi

    # --- Nautilus (GNOME / CachyOS GNOME / Niri) ---
    if is_installed_cmd "nautilus"; then
        local nautilus_scripts_dir="$HOME/.local/share/nautilus/scripts"
        mkdir -p "$nautilus_scripts_dir"
        local script_path="$nautilus_scripts_dir/Abrir no Kitty"
        # Variáveis com \$ são expandidas em tempo de execução do script gerado.
        cat > "$script_path" <<NAUTILUS_EOF
#!/usr/bin/env bash
KITTY="${kitty_bin}"
if [[ -n "\$NAUTILUS_SCRIPT_SELECTED_FILE_PATHS" ]]; then
    folder=\$(echo "\$NAUTILUS_SCRIPT_SELECTED_FILE_PATHS" | head -1)
    [[ ! -d "\$folder" ]] && folder=\$(dirname "\$folder")
else
    folder="\${NAUTILUS_SCRIPT_CURRENT_URI#file://}"
    folder="\${folder//%20/ }"
fi
[[ -z "\$folder" ]] && folder="\$HOME"
exec "\$KITTY" --directory "\$folder"
NAUTILUS_EOF
        chmod +x "$script_path"
        printf '%b\n' "   ${C2}✅ Nautilus (GNOME): script configurado — clique direito → Scripts → Abrir no Kitty.${RESET}"
        installed_any=true
    fi

    # --- Dolphin (KDE Plasma) ---
    if is_installed_cmd "dolphin"; then
        local dolphin_dir="$HOME/.local/share/kio/servicemenus"
        mkdir -p "$dolphin_dir"
        cat > "$dolphin_dir/open_in_kitty.desktop" <<DOLPHIN_EOF
[Desktop Entry]
Type=Service
X-KDE-ServiceTypes=KonqPopupMenu/Plugin
MimeType=inode/directory;
Actions=openInKitty;

[Desktop Action openInKitty]
Name=Abrir no Kitty
Name[pt_BR]=Abrir no Kitty
Icon=utilities-terminal
Exec=${kitty_bin} --directory %f
DOLPHIN_EOF
        # Reconstrói cache de serviços do KDE para o menu aparecer imediatamente.
        if is_installed_cmd "kbuildsycoca6"; then
            kbuildsycoca6 2>/dev/null || true
        elif is_installed_cmd "kbuildsycoca5"; then
            kbuildsycoca5 2>/dev/null || true
        fi
        printf '%b\n' "   ${C2}✅ Dolphin (KDE Plasma): service menu configurado.${RESET}"
        installed_any=true
    fi

    # --- Thunar (Xfce / Niri) ---
    if is_installed_cmd "thunar"; then
        _install_thunar_action "$kitty_bin"
        printf '%b\n' "   ${C2}✅ Thunar: ação personalizada configurada.${RESET}"
        installed_any=true
    fi

    # --- COSMIC Files (Pop!_OS 24.04 / CachyOS COSMIC) ---
    # Ações de contexto via arquivo não são suportadas; a integração é feita
    # definindo o Kitty como terminal padrão nas configurações do sistema.
    if is_installed_cmd "cosmic-files"; then
        printf '%b\n' "   ${C4}⚠️  COSMIC Files: ações de contexto personalizadas não suportadas via arquivo.${RESET}"
        printf '%b\n' "   ${C3}→ Acesse: COSMIC Settings → Aplicações → Terminal padrão → Kitty${RESET}"
        printf '%b\n' "   ${C3}→ O botão 'Abrir Terminal' no COSMIC Files usará o Kitty automaticamente.${RESET}"
    fi

    if [[ "$installed_any" == false ]] && ! is_installed_cmd "cosmic-files"; then
        warn "Nenhum gerenciador de arquivos compatível encontrado (Nemo, Nautilus, Dolphin, Thunar)."
    fi
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
    ensure_pkg "unzip"

    install_nerd_font
    install_kitty
    apply_kitty_settings
    install_starship
    setup_context_menu
    set_default_terminal

    printf '\n%b\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
    printf '%b\n' "${C4}  Próximos passos:${RESET}"
    printf '%b\n' "  1. Abra o Kitty: ${C3}kitty${RESET}"
    printf '%b\n' "  2. Mude o tema: ${C3}ctrl+shift+F2${RESET}"
    printf '%b\n' "  3. Menu de contexto: clique direito em uma pasta no gerenciador de arquivos"
    printf '%b\n\n' "${C5}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
}

main "$@"
