#!/usr/bin/env bash

# =============================================================================
# Nome do Script : versions.sh
# Descrição      : Fonte única de verdade para versões de todos os componentes
# Versão         : 3.0.0
# =============================================================================

[[ -n "${LUMINA_VERSIONS_LOADED:-}" ]] && return 0
readonly LUMINA_VERSIONS_LOADED=1

export SUPPORTED_PHP_VERSIONS="7.4 8.0 8.1 8.2 8.3 8.4"
export NGINX_IMAGE="nginx:1.26-alpine"
export MARIADB_IMAGE="mariadb:11.4"
