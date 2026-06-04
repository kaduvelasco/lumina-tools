#!/usr/bin/env bash
# =============================================================================
# Script Name : install.sh
# Description : Installs or updates lumina-tools from GitHub Releases
# Version     : 1.0.0
# =============================================================================
set -Eeuo pipefail

readonly GITHUB_REPO="kaduvelasco/lumina-tools"
readonly BINARY="lumina"
readonly INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
readonly GITHUB_API="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"

# --- Colors ---
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly RESET='\033[0m'

info()    { printf '%b\n' "${YELLOW}-> ${1}${RESET}"; }
success() { printf '%b\n' "${GREEN}+ ${1}${RESET}"; }
die()     { printf '%b\n' "${RED}x ${1}${RESET}" >&2; exit 1; }

# --- Detect architecture ---
detect_arch() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64)         echo "amd64" ;;
        aarch64|arm64)  echo "arm64" ;;
        *)              die "Arquitetura '${arch}' nao suportada." ;;
    esac
}

# --- Get latest version from GitHub ---
get_latest_version() {
    local version
    if command -v jq &>/dev/null; then
        version=$(curl -fsSL "$GITHUB_API" | jq -r '.tag_name // empty')
    else
        version=$(curl -fsSL "$GITHUB_API" \
            | grep '"tag_name"' \
            | sed 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/')
    fi
    if [[ -z "$version" ]]; then
        die "Nao foi possivel obter a versao mais recente. Verifique sua conexao."
    fi
    echo "$version"
}

# --- Main ---
main() {
    printf '\n'
    info "Instalando lumina-tools..."

    # Check dependencies
    command -v curl &>/dev/null || die "curl e necessario para instalar."

    local arch version download_url tmp_file
    arch=$(detect_arch)
    version=$(get_latest_version)
    download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/lumina-linux-${arch}"

    info "Versao: ${version} | Plataforma: linux/${arch}"
    info "URL: ${download_url}"

    tmp_file=$(mktemp)
    trap 'rm -f "$tmp_file"' EXIT

    info "Baixando binario..."
    if ! curl -fsSL "$download_url" -o "$tmp_file"; then
        die "Falha ao baixar o binario. Verifique se a release '${version}' possui o asset 'lumina-linux-${arch}'."
    fi

    chmod +x "$tmp_file"

    info "Instalando em ${INSTALL_DIR}/${BINARY}..."
    if [[ -w "$INSTALL_DIR" ]]; then
        mv -- "$tmp_file" "${INSTALL_DIR}/${BINARY}"
    else
        sudo mv -- "$tmp_file" "${INSTALL_DIR}/${BINARY}"
    fi
    trap - EXIT

    success "lumina ${version} instalado em ${INSTALL_DIR}/${BINARY}"

    # Install completions (optional)
    local bash_comp_dir="/etc/bash_completion.d"
    if [[ -d "$bash_comp_dir" ]]; then
        info "Instalando completion bash..."
        if curl -fsSL "https://raw.githubusercontent.com/${GITHUB_REPO}/main/completions/lumina.bash" \
            | sudo tee "${bash_comp_dir}/lumina" > /dev/null 2>&1; then
            success "Bash completion instalado."
        fi
    fi

    printf '\n'
    printf '%b\n' "${GREEN}Instalacao concluida!${RESET}"
    printf '%b\n' "Execute: ${YELLOW}lumina${RESET}"
    printf '\n'
}

main "$@"
