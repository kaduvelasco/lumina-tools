#!/usr/bin/env bash

# =============================================================================
# Nome do Script : docker.sh
# Descrição      : Valida versões PHP e credenciais do banco, gera o
#                  docker-compose.yml a partir dos templates e configura
#                  o roteamento local via update_hosts
# Versão         : 3.0.0
# =============================================================================

[[ -n "${LUMINA_DOCKER_LOADED:-}" ]] && return 0
readonly LUMINA_DOCKER_LOADED=1

validate_php_versions() {
    local versions="$1"
    local supported="${SUPPORTED_PHP_VERSIONS:-"7.4 8.0 8.1 8.2 8.3 8.4"}"
    for v in $versions; do
        if [[ ! " $supported " =~ \ $v\  ]]; then
            warn "Versão PHP '$v' não é suportada."
            printf '%b\n' "   Disponíveis: ${C3}${supported}${RESET}"
            return 1
        fi
    done
    return 0
}

validate_db_credentials() {
    local user="$1" pass="$2"
    if [[ -z "$user" ]]; then
        warn "O nome de usuário não pode ser vazio."
        return 1
    fi
    if [[ ! "$user" =~ ^[a-zA-Z0-9_]+$ ]]; then
        warn "O usuário '${user}' contém caracteres inválidos. Use apenas letras, números e _."
        return 1
    fi
    if [[ ${#pass} -lt 8 ]]; then
        warn "A senha deve ter pelo menos 8 caracteres."
        return 1
    fi
    return 0
}

generate_secure_password() {
    openssl rand -base64 16 | tr -d "=+/" | cut -c1-16
}

generate_docker_stack() {
    local workspace="/srv/workspace"
    if [[ -f "$HOME/.lumina/config.env" ]]; then
        # shellcheck source=/dev/null
        source "$HOME/.lumina/config.env"
        workspace="$(dirname "${WORKSPACE:-/srv/workspace/docker}")"
    fi
    local docker_dir="${workspace}/docker"
    local template_dir
    template_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../templates" && pwd)"

    if [[ ! -d "$workspace" ]]; then
        warn "Workspace não encontrado. Execute a opção 3 primeiro."
        return 1
    fi

    printf '\n'
    info "Gerando Stack Docker LuminaStack..."

    mkdir -p "${docker_dir}"/{nginx,php,php-config,mariadb/init,mariadb/conf.d}

    if [[ ! -f "${template_dir}/docker-compose.tpl" ]]; then
        warn "Template não encontrado: ${template_dir}/docker-compose.tpl"
        return 1
    fi

    # --- versões do PHP ---
    local php_versions
    while true; do
        printf '%s' "Versões do PHP (ex: 7.4 8.1 8.2): "
        read -r php_versions
        php_versions="${php_versions#"${php_versions%%[![:space:]]*}"}"
        php_versions="${php_versions%"${php_versions##*[![:space:]]}"}"
        [[ -z "$php_versions" ]] && warn "Informe ao menos uma versão." && continue
        validate_php_versions "$php_versions" && break
    done
    PHP_VERSIONS="$php_versions"

    # --- credenciais do banco ---
    local db_user db_pass
    while true; do
        printf '%s' "Usuário do banco [admin]: "
        read -r db_user
        db_user="${db_user:-admin}"

        printf '%s' "Senha do banco (Enter para gerar automaticamente): "
        read -r -s db_pass
        printf '\n'

        if [[ -z "$db_pass" ]]; then
            db_pass=$(generate_secure_password)
            printf '%b\n' "   ${C3}🔐 Senha gerada: ${db_pass}${RESET}"
        fi

        validate_db_credentials "$db_user" "$db_pass" && break
    done

    local db_root_password
    db_root_password=$(generate_secure_password)
    printf '%b\n' "   ${C3}🔐 Senha root gerada: ${db_root_password}${RESET}"

    # --- script de permissões MariaDB ---
    # MariaDB 11.4+ creates MYSQL_USER with '%' host (not 'localhost'), so RENAME USER is not needed.
    cat > "${docker_dir}/mariadb/init/01-permissions.sql" << EOF
-- Grant full access from any host. MariaDB 11.4+ creates users with '%' host by default.
GRANT ALL PRIVILEGES ON *.* TO \`${db_user}\`@'%' WITH GRANT OPTION;
FLUSH PRIVILEGES;
EOF

    # --- gera os serviços PHP e o bloco depends_on do Nginx ---
    local php_services="" nginx_depends_on="" first_version=""

    for v in $PHP_VERSIONS; do
        local name="php${v//./}"
        [[ -z "$first_version" ]] && first_version="$name"
        mkdir -p "${workspace}/logs/${name}"

        nginx_depends_on="${nginx_depends_on}      ${name}:
        condition: service_healthy
"
        php_services="${php_services}
  ${name}:
    container_name: ${name}
    build:
      context: .
      dockerfile: php/Dockerfile
      args:
        PHP_VERSION: ${v}
    restart: unless-stopped
    volumes:
      - ../www/html:/var/www/html
      - ../www/data:/var/www/data
      - ../logs/${name}:/var/log/php
      - ./php-config/php.ini:/usr/local/etc/php/php.ini
    environment:
      - XDEBUG_MODE=off
    extra_hosts:
      - \"host.docker.internal:host-gateway\"
    healthcheck:
      test: [\"CMD-SHELL\", \"php-fpm -t > /dev/null 2>&1\"]
      start_period: 10s
      interval: 15s
      timeout: 5s
      retries: 3
    logging:
      driver: \"json-file\"
      options:
        max-size: \"10m\"
        max-file: \"3\"
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 1G
    networks:
      - docker-php-network
"
    done

    # --- gera docker-compose.yml ---
    awk -v php_services="$php_services" \
        -v default_php="$first_version" \
        -v nginx_deps="$nginx_depends_on" \
        -v nginx_image="${NGINX_IMAGE:-nginx:1.26-alpine}" \
        -v mariadb_image="${MARIADB_IMAGE:-mariadb:11.4}" '{
        if ($0 ~ /{{PHP_SERVICES}}/)          { print php_services }
        else if ($0 ~ /{{DEFAULT_PHP}}/)      { gsub(/{{DEFAULT_PHP}}/, default_php); print }
        else if ($0 ~ /{{NGINX_DEPENDS_ON}}/) { print nginx_deps }
        else if ($0 ~ /{{NGINX_IMAGE}}/)      { gsub(/{{NGINX_IMAGE}}/, nginx_image); print }
        else if ($0 ~ /{{MARIADB_IMAGE}}/)    { gsub(/{{MARIADB_IMAGE}}/, mariadb_image); print }
        else                                  { print }
    }' "${template_dir}/docker-compose.tpl" > "${docker_dir}/docker-compose.yml"

    if [[ ! -s "${docker_dir}/docker-compose.yml" ]]; then
        warn "Falha ao gerar docker-compose.yml (arquivo vazio)."
        rm -f "${docker_dir}/docker-compose.yml"
        return 1
    fi

    cp "${template_dir}/nginx.conf.tpl"     "${docker_dir}/nginx/default.conf"
    cp "${template_dir}/php.Dockerfile.tpl" "${docker_dir}/php/Dockerfile"
    cp "${template_dir}/php.ini.tpl"        "${docker_dir}/php-config/php.ini"

    local first_ver_num="${first_version#php}"
    sed -i \
        -e "s/{{DEFAULT_PHP}}/${first_version}/g" \
        -e "s/{{DEFAULT_PHP_VER}}/${first_ver_num}/g" \
        "${docker_dir}/nginx/default.conf"

    # --- gera .env com permissão restrita ---
    local env_tmp
    env_tmp=$(mktemp "${docker_dir}/.env.XXXXXX")
    trap 'rm -f "$env_tmp"' EXIT
    chmod 600 "$env_tmp"
    cat > "$env_tmp" << EOF
DB_USER=${db_user}
DB_PASS=${db_pass}
DB_ROOT_PASSWORD=${db_root_password}
PHP_VERSIONS=${PHP_VERSIONS}
EOF
    mv "$env_tmp" "${docker_dir}/.env"
    trap - EXIT

    export PHP_VERSIONS
    update_hosts

    printf '\n'
    success "Stack criada em: ${docker_dir}"
    printf '%b\n' "   ${C4}Versões PHP   :${RESET} ${PHP_VERSIONS}"
    printf '%b\n' "   ${C4}PHP padrão    :${RESET} ${first_version} (usado em localhost)"
    printf '%b\n' "   ${C4}Usuário DB    :${RESET} ${db_user}"
    printf '%b\n' "   ${C4}Credenciais   :${RESET} ${docker_dir}/.env ${C3}(chmod 600)${RESET}"
}
