#!/usr/bin/env bash

# =============================================================================
# Nome do Script : workspace.sh
# Descrição      : Cria a estrutura de diretórios do workspace, instala
#                  arquivos de interface e ajusta permissões iniciais
# Versão         : 3.0.0
# =============================================================================

[[ -n "${LUMINA_WORKSPACE_LOADED:-}" ]] && return 0
readonly LUMINA_WORKSPACE_LOADED=1

create_workspace() {
    local workspace_input
    printf '%s' "Local do workspace [/srv/workspace]: "
    read -r workspace_input
    workspace_input="${workspace_input#"${workspace_input%%[![:space:]]*}"}"
    workspace_input="${workspace_input%"${workspace_input##*[![:space:]]}"}"
    local workspace="${workspace_input:-/srv/workspace}"

    local template_dir
    template_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../templates" && pwd)"

    printf '\n'
    info "Criando estrutura de workspace em ${workspace}..."

    # Garante que o diretório base exista e o usuário tenha permissão
    if [[ ! -d "$workspace" ]]; then
        sudo mkdir -p "$workspace"
        sudo chown -R "$USER:$USER" "$workspace"
    fi

    if [[ -d "$workspace/www" ]]; then
        warn "O workspace já existe em ${workspace}"
        printf '%b\n' "   Continuar irá reinstalar os arquivos de interface (index.php, info.php)"
        printf '%b\n' "   e reajustar as permissões. Nenhum dado será apagado."
        printf '%b' "   Continuar? (${C3}s${RESET}/N): "
        read -r confirm
        if [[ ! "$confirm" =~ ^[sS]$ ]]; then
            warn "Operação cancelada."
            return 1
        fi
    fi

    mkdir -p "${workspace}"/{www/html,www/data,databases/mariadb,logs/nginx,docker,docker/mariadb/conf.d,backups}

    local versions="${PHP_VERSIONS:-${SUPPORTED_PHP_VERSIONS:-"7.4 8.1 8.2 8.3 8.4"}}"
    for v in $versions; do
        mkdir -p "${workspace}/logs/php${v//./}"
    done

    info "Instalando templates de interface..."
    if [[ ! -d "$template_dir" ]]; then
        warn "Pasta de templates não encontrada em ${template_dir}"
        return 1
    fi

    cp "$template_dir/info.php.tpl"  "${workspace}/www/html/info.php"
    cp "$template_dir/index.php.tpl" "${workspace}/www/html/index.php"
    success "index.php e info.php instalados."

    # 755: usuário e container conseguem ler/executar; escrita apenas pelo dono
    chmod -R 755 "${workspace}/www"

    # 775: container MariaDB precisa gravar moodle-performance.cnf
    chmod -R 775 "${workspace}/docker/mariadb"

    local lumina_dir="$HOME/.lumina"
    local config_file="${lumina_dir}/config.env"
    (umask 077; mkdir -p "$lumina_dir")
    (umask 177; touch "$config_file")
    {
        printf 'WORKSPACE=%q\n'    "${workspace}/docker"
        printf 'CONTAINER_NAME="mariadb"\n'
        printf 'BACKUP_DIR=%q\n'   "${workspace}/backups"
        printf 'BACKUPS_MANTER=3\n'
    } > "$config_file"
    success "Configuração salva em: ${config_file}"

    printf '\n'
    success "Workspace criado com sucesso em: ${workspace}"
    printf '%b\n' "   ${C4}Projetos PHP  :${RESET} ${workspace}/www/html"
    printf '%b\n' "   ${C4}Backups SQL   :${RESET} ${workspace}/backups  ${C3}(sincronizado via MegaSync)${RESET}"
    printf '%b\n' "   ${C4}Dados Moodle  :${RESET} ${workspace}/www/data  ${C3}(sincronizado via MegaSync)${RESET}"
    printf '%b\n' "   ${C4}Config MariaDB:${RESET} ${workspace}/docker/mariadb/conf.d"
    printf '\n'
    printf '%b\n' "   ${C6}💡 Dica: Use a opção 4 para gerar a stack Docker antes de iniciar o ambiente.${RESET}"
    printf '\n'
}
