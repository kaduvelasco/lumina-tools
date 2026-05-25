#!/usr/bin/env bash

# =============================================================================
# Nome do Script : clean-docker.sh
# Descrição      : Remove todos os containers, imagens, volumes e networks do
#                  Docker. Use para resetar completamente o ambiente.
# Versão         : 3.0.0
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR

if [[ ! -f "$SCRIPT_DIR/lib/utils.sh" ]]; then
    printf '\033[0;31m❌ Erro fatal: lib/utils.sh não encontrado.\033[0m\n' >&2
    exit 1
fi
# shellcheck source=/dev/null
source "$SCRIPT_DIR/lib/utils.sh"

# Derive workspace paths from config (fallback to default)
_workspace_base="/srv/workspace"
if [[ -f "$HOME/.lumina/config.env" ]]; then
    # shellcheck source=/dev/null
    source "$HOME/.lumina/config.env"
    _workspace_base="$(dirname "${WORKSPACE:-/srv/workspace/docker}")"
fi
readonly DATABASES_DIR="${_workspace_base}/databases"
readonly DOCKER_DIR="${_workspace_base}/docker"

show_lumina_header "LuminaStack — Linux + Nginx + PHP + MariaDb + Docker"

# --- confirmação de segurança ---
printf '%b\n' "${C1}⚠️  ATENÇÃO — Esta operação irá remover:${RESET}"
printf '%b\n' "   • Todos os containers (rodando ou parados)"
printf '%b\n' "   • Todas as imagens Docker"
printf '%b\n' "   • Todos os volumes não utilizados"
printf '%b\n' "   • Todas as networks não utilizadas"
printf '%b\n' "   • ${DATABASES_DIR}/  (dados binários do MariaDB)"
printf '%b\n' "   • ${DOCKER_DIR}/     (docker-compose.yml, .env e configs)"
printf '\n'
printf '%b\n' "${C2}   Os seguintes dados do usuário NÃO serão apagados:${RESET}"
printf '%b\n' "   ✅ ${_workspace_base}/www/html    (seus projetos PHP)"
printf '%b\n' "   ✅ ${_workspace_base}/www/data    (moodledata)"
printf '%b\n' "   ✅ ${BACKUP_DIR:-${_workspace_base}/backups}/    (dumps SQL)"
printf '\n'
printf '%b' "   Tem certeza que deseja continuar? (${C1}s${RESET}/N): "
read -r confirm
if [[ ! "$confirm" =~ ^[sS]$ ]]; then
    printf '\n%b\n\n' "${C2}Operação cancelada.${RESET}"
    exit 0
fi

printf '\n%b\n' "${C1}⚠️  Confirmação final: digite ${C3}SIM${C1} (em maiúsculo) para prosseguir:${RESET}"
printf '%s' "   > "
read -r double_check
if [[ "$double_check" != "SIM" ]]; then
    printf '\n%b\n\n' "${C2}Operação cancelada.${RESET}"
    exit 0
fi

printf '\n'
info "Iniciando limpeza do Docker..."
printf '\n'

# --- limpeza ---
printf '%b\n' "${C3}⏹️  Parando containers em execução...${RESET}"
mapfile -t containers < <(docker ps -aq)
if [[ ${#containers[@]} -gt 0 ]]; then
    docker stop "${containers[@]}" 2>/dev/null || true
    success "Containers parados."
else
    printf '%b\n' "   Nenhum container em execução."
fi

printf '\n%b\n' "${C3}🗑️  Removendo containers...${RESET}"
if [[ ${#containers[@]} -gt 0 ]]; then
    docker rm "${containers[@]}" 2>/dev/null || true
    success "Containers removidos."
else
    printf '%b\n' "   Nenhum container para remover."
fi

printf '\n%b\n' "${C3}🗑️  Removendo imagens...${RESET}"
mapfile -t images < <(docker images -q)
if [[ ${#images[@]} -gt 0 ]]; then
    docker rmi "${images[@]}" -f 2>/dev/null || true
    success "Imagens removidas."
else
    printf '%b\n' "   Nenhuma imagem para remover."
fi

printf '\n%b\n' "${C3}🗑️  Removendo volumes não utilizados...${RESET}"
docker volume prune -f 2>/dev/null || true
success "Volumes limpos."

printf '\n%b\n' "${C3}🗑️  Removendo networks não utilizadas...${RESET}"
docker network prune -f 2>/dev/null || true
success "Networks limpas."

printf '\n%b\n' "${C3}🗑️  Removendo dados binários do banco...${RESET}"
if [[ -d "$DATABASES_DIR" ]]; then
    rm -rf "$DATABASES_DIR"
    success "${DATABASES_DIR}/ removido."
else
    printf '%b\n' "   Pasta não encontrada, nada a remover."
fi

printf '\n%b\n' "${C3}🗑️  Removendo configurações da stack...${RESET}"
if [[ -d "$DOCKER_DIR" ]]; then
    rm -rf "$DOCKER_DIR"
    success "${DOCKER_DIR}/ removido."
else
    printf '%b\n' "   Pasta não encontrada, nada a remover."
fi

# --- resumo ---
printf '\n'
success "Limpeza concluída com sucesso!"
printf '%b\n' "${C4}──────────────────────────────────────────────────${RESET}"
printf '%b\n' "   Para recriar o ambiente, execute ${C3}./install.sh${RESET}"
printf '%b\n' "   e siga as opções ${C3}3 → 4${RESET} do menu."
printf '%b\n' "${C4}──────────────────────────────────────────────────${RESET}"
printf '\n'
