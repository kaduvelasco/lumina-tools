#!/usr/bin/env bash
# =============================================================================
# Nome do Script : utils.sh
# Descrição      : Utilitários compartilhados do LuminaDev
# Versão         : 2.1.0
# =============================================================================

# --- opções do shell ---
set -euo pipefail

# Proteção contra carregamento duplo
[[ -n "${LUMINA_UTILS_LOADED:-}" ]] && return 0
readonly LUMINA_UTILS_LOADED=1

# --- paleta de cores ---
export C1="\033[0;31m"   # Erros, ações destrutivas
export C2="\033[0;32m"   # Sucesso, operações normais
export C3="\033[0;33m"   # Avisos, manutenção
export C4="\033[0;34m"   # Informativas, consulta
export C5="\033[0;35m"   # Rótulos, bordas
export C6="\033[0;36m"   # Dicas, decorativos
export H1="\033[1;32m"   # Logo — linha principal
export H2="\033[0;32m"   # Logo — subtítulo
export TS=""              # Type Something — sem cor definida
export RESET="\033[0m"

# --- constantes ---
readonly NODE_MIN_VERSION=18

# --- funções de saída padronizadas ---

# Encerra com mensagem de erro. $1=mensagem [$2=exit_code]
die() {
    local message="$1"
    local exit_code="${2:-1}"
    printf '%b\n' "${C1}❌ Erro: ${message}${RESET}" >&2
    exit "$exit_code"
}

# Exibe aviso (stderr). $1=mensagem
warn() {
    printf '%b\n' "${C3}⚠️  ${1}${RESET}" >&2
}

# Exibe informação. $1=mensagem
info() {
    printf '%b\n' "${C4}ℹ️  ${1}${RESET}"
}

# Exibe sucesso. $1=mensagem
success() {
    printf '%b\n' "${C2}✅ ${1}${RESET}"
}

# --- funções auxiliares ---

# =============================================================================
# Detecta o gerenciador de pacotes do sistema. Idempotente: retorna se já
# detectado. Torna PKG_MANAGER/PKG_INSTALL/PKG_UPDATE readonly após detecção.
# =============================================================================
detect_pkg_manager() {
    [[ -n "${PKG_MANAGER:-}" ]] && return 0

    if command -v apt-get &>/dev/null; then
        PKG_MANAGER="apt"
        PKG_INSTALL="sudo apt-get install -y"
        PKG_UPDATE="sudo apt-get update -qq"
    elif command -v dnf &>/dev/null; then
        PKG_MANAGER="dnf"
        PKG_INSTALL="sudo dnf install -y"
        PKG_UPDATE="sudo dnf check-update -q || true"
    elif command -v pacman &>/dev/null; then
        PKG_MANAGER="pacman"
        PKG_INSTALL="sudo pacman -S --noconfirm"
        PKG_UPDATE="sudo pacman -Sy --noconfirm"
    else
        die "Gerenciador de pacotes não suportado. Use apt, dnf ou pacman."
    fi

    export PKG_MANAGER PKG_INSTALL PKG_UPDATE
    readonly PKG_MANAGER PKG_INSTALL PKG_UPDATE
}

# =============================================================================
# Verifica se um binário existe no PATH (ignora aliases e funções).
# =============================================================================
is_installed_cmd() {
    local cmd="$1"
    type -P "$cmd" &>/dev/null
}

# =============================================================================
# Verifica se um pacote está instalado no sistema.
# =============================================================================
is_installed_pkg() {
    local pkg="$1"
    case "${PKG_MANAGER:-}" in
        apt)    dpkg -s "$pkg" &>/dev/null 2>&1 ;;
        dnf)    rpm -q "$pkg" &>/dev/null 2>&1 ;;
        pacman) pacman -Qi "$pkg" &>/dev/null 2>&1 ;;
        *)      return 1 ;;
    esac
}

# =============================================================================
# Garante a instalação de um pacote. Detecta o pkg manager se necessário.
# =============================================================================
ensure_pkg() {
    local pkg="$1"

    if [[ -z "${PKG_MANAGER:-}" ]]; then
        detect_pkg_manager
    fi

    if is_installed_pkg "$pkg"; then
        printf '%b\n' "${C3}✅ ${pkg} já está instalado. Pulando.${RESET}"
        return 0
    fi

    printf '%b\n' "${C4}📥 Instalando ${pkg}...${RESET}"

    case "$PKG_MANAGER" in
        apt)    sudo apt-get install -y -- "$pkg" ;;
        dnf)    sudo dnf install -y -- "$pkg" ;;
        pacman) sudo pacman -S --noconfirm -- "$pkg" ;;
    esac
}

# =============================================================================
# Verifica se a versão do Node.js atende ao requisito mínimo.
# =============================================================================
check_node_version() {
    local version
    if ! is_installed_cmd "node"; then
        return 1
    fi

    version=$(node -e "console.log(process.version.slice(1).split('.')[0])" 2>/dev/null)

    if [[ -z "$version" ]]; then
        return 1
    fi

    if [[ "$version" -lt "$NODE_MIN_VERSION" ]]; then
        warn "Node.js v${version} encontrado, mas é necessário v${NODE_MIN_VERSION}+."
        return 1
    fi

    success "Node.js v${version} encontrado."
    return 0
}

# =============================================================================
# Instala o Node.js via NodeSource (LTS mais recente).
# =============================================================================
install_node() {
    printf '%b\n' "${C4}📥 Instalando Node.js LTS via NodeSource...${RESET}"

    case "${PKG_MANAGER:-}" in
        apt)
            ensure_pkg "curl"
            curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
            ensure_pkg "nodejs"
            ;;
        dnf)
            curl -fsSL https://rpm.nodesource.com/setup_lts.x | sudo bash -
            ensure_pkg "nodejs"
            ;;
        pacman)
            ensure_pkg "nodejs"
            ensure_pkg "npm"
            ;;
    esac

    if ! is_installed_cmd "node"; then
        die "Falha ao instalar o Node.js. Verifique sua conexão e tente novamente."
    fi

    success "Node.js $(node -v) instalado com sucesso."
}

# =============================================================================
# Verifica se o script está sendo executado como root.
# =============================================================================
require_not_root() {
    if [[ "$EUID" -eq 0 ]]; then
        die "Não execute este script como root. Use seu usuário normal.\n   O script solicitará sudo quando necessário."
    fi
}

# =============================================================================
# Verifica se sudo está disponível e funcional.
# =============================================================================
require_sudo() {
    if ! type -P sudo &>/dev/null; then
        die "sudo não encontrado. Instale-o e tente novamente."
    fi

    if ! sudo -v &>/dev/null; then
        die "Falha ao obter permissões sudo. Verifique suas credenciais."
    fi
}

# =============================================================================
# Verifica se há conexão com a internet.
# =============================================================================
require_internet() {
    printf '%b\n' "${C6}🌐 Verificando conexão com a internet...${RESET}"
    if ! curl -fsSL --max-time 5 https://1.1.1.1 &>/dev/null; then
        die "Sem conexão com a internet. Verifique sua rede."
    fi
    success "Conexão OK."
}

# =============================================================================
# Garante que ~/.local/bin está no PATH (idempotente).
# =============================================================================
ensure_local_bin_in_path() {
    local path_line='export PATH="$HOME/.local/bin:$PATH"'

    grep -qxF "$path_line" "$HOME/.bashrc" 2>/dev/null || printf '%s\n' "$path_line" >> "$HOME/.bashrc"

    if [[ -f "$HOME/.zshrc" ]]; then
        grep -qxF "$path_line" "$HOME/.zshrc" 2>/dev/null || printf '%s\n' "$path_line" >> "$HOME/.zshrc"
    fi

    export PATH="$HOME/.local/bin:$PATH"
    success "PATH atualizado: ~/.local/bin incluído."
}

# =============================================================================
# Imprime a versão de um comando de forma segura.
# =============================================================================
print_version() {
    local cmd="$1"
    local version_flag="${2:---version}"

    if is_installed_cmd "$cmd"; then
        local version
        version=$("$cmd" "$version_flag" 2>/dev/null | head -1)
        printf '%b\n' "${C2}   ${cmd}: ${version}${RESET}"
    else
        printf '%b\n' "${C4}   ${cmd}: não instalado${RESET}"
    fi
}

# =============================================================================
# Exibe o cabeçalho ASCII padrão Lumina. $1 = subtítulo (opcional).
# =============================================================================
show_lumina_header() {
    local subtitle="${1:-LUMINA CLI ENGINE}"
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
