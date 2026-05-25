#!/usr/bin/env bash

# =============================================================================
# Nome do Script : system.sh
# Descrição      : Detecta distro, instala pré-requisitos, instala Docker
#                  e gerencia roteamento local via /etc/hosts
# Versão         : 3.0.0
# =============================================================================

[[ -n "${LUMINA_SYSTEM_LOADED:-}" ]] && return 0
readonly LUMINA_SYSTEM_LOADED=1

detect_distro() {
    if [[ -f /etc/os-release ]]; then
        # shellcheck source=/dev/null
        . /etc/os-release
        export DISTRO="$ID"
        return 0
    fi
    warn "Não foi possível detectar a distribuição."
    return 1
}

install_prereqs() {
    printf '\n'
    info "Verificando e instalando pré-requisitos..."
    detect_distro || return 1

    case "$DISTRO" in
        ubuntu|linuxmint|pop|debian)
            sudo apt-get update -q
            sudo apt-get install -y -- curl git openssl lsof
            ;;
        fedora)
            sudo dnf install -y -- curl git openssl lsof
            ;;
        arch|manjaro)
            sudo pacman -S --noconfirm -- curl git openssl lsof
            ;;
        *)
            warn "Distribuição '$DISTRO' pode não ser totalmente suportada."
            ;;
    esac

    success "Pré-requisitos verificados."
}

install_docker() {
    printf '\n'
    info "Verificando Docker..."
    detect_distro || return 1

    if is_installed_cmd docker; then
        success "Docker já está instalado."
    else
        printf '%b\n' "${C3}📦 Instalando Docker para '${DISTRO}'...${RESET}"

        case "$DISTRO" in
            arch|manjaro)
                sudo pacman -S --noconfirm -- docker docker-compose
                sudo systemctl enable --now docker
                ;;
            *)
                printf '\n'
                printf '%b\n' "${C4}Como deseja instalar o Docker?${RESET}"
                printf '%b\n' "   ${C2}1.${RESET} Via package manager ${C3}(recomendado — mais seguro e rastreável)${RESET}"
                printf '%b\n' "   ${C2}2.${RESET} Via script oficial  ${C3}(get.docker.com — sempre a versão mais recente)${RESET}"
                printf '%s' "   Opção [1]: "
                read -r DOCKER_INSTALL_METHOD
                DOCKER_INSTALL_METHOD="${DOCKER_INSTALL_METHOD:-1}"

                if [[ "$DOCKER_INSTALL_METHOD" == "1" ]]; then
                    case "$DISTRO" in
                        ubuntu|debian|linuxmint|pop)
                            sudo apt-get update -q
                            sudo apt-get install -y -- docker.io docker-compose-v2
                            ;;
                        fedora)
                            sudo dnf install -y -- docker docker-compose
                            ;;
                        *)
                            warn "Distro '$DISTRO' não tem package manager configurado. Usando script oficial como fallback..."
                            DOCKER_INSTALL_METHOD="2"
                            ;;
                    esac
                fi

                if [[ "$DOCKER_INSTALL_METHOD" == "2" ]]; then
                    local installer
                    installer=$(mktemp)
                    trap 'rm -f "$installer"' EXIT
                    curl -fsSL https://get.docker.com -o "$installer"
                    if sudo sh "$installer"; then
                        rm -f "$installer"
                        trap - EXIT
                    else
                        die "Falha na instalação do Docker. Installer mantido em: $installer"
                    fi
                fi

                sudo systemctl enable --now docker
                ;;
        esac
    fi

    if ! groups "$USER" | grep -qw docker; then
        printf '%b\n' "${C3}👤 Adicionando ${USER} ao grupo docker...${RESET}"
        sudo usermod -aG docker "$USER"
        warn "Você precisará reiniciar a sessão para aplicar as permissões do grupo docker."
    fi

    if [[ "${DISTRO:-}" == "fedora" ]]; then
        printf '\n'
        warn "Fedora detectado: se containers não conseguirem ler/escrever nos volumes,"
        printf '%b\n' "    ${C3}execute: sudo setsebool -P container_manage_cgroup on${RESET}" >&2
    fi

    check_port_80
    success "Docker configurado com sucesso."
}

update_hosts() {
    printf '\n'
    info "Atualizando /etc/hosts para roteamento local..."

    local versions="${PHP_VERSIONS:-${SUPPORTED_PHP_VERSIONS:-"7.4 8.0 8.1 8.2 8.3 8.4"}}"
    local hosts_line="127.0.0.1"

    for v in $versions; do
        hosts_line="$hosts_line php${v/./}.localhost"
    done

    sudo sed -i '/# lumina-stack/d' /etc/hosts
    printf '%s\n' "$hosts_line # lumina-stack" | sudo tee -a /etc/hosts > /dev/null

    success "Arquivo /etc/hosts atualizado."
}

check_port_80() {
    if ! is_installed_cmd lsof; then
        warn "lsof não encontrado. Pulando verificação da porta 80."
        return 0
    fi

    if sudo lsof -i :80 > /dev/null 2>&1; then
        printf '\n'
        warn "A porta 80 já está em uso por outro processo:"
        sudo lsof -i :80 || printf '%b\n' "   ${C3}(Não foi possível listar os processos)${RESET}" >&2
        printf '%b\n' "${C3}    Solução: sudo systemctl stop apache2  ou  sudo systemctl stop nginx${RESET}" >&2
        return 1
    fi
    return 0
}
