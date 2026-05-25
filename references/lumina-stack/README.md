# 🚀 LuminaStack

> Ambiente de desenvolvimento PHP modular com roteamento dinâmico via Docker.

[![License](https://img.shields.io/badge/license-GPL--3.0-blue)](https://www.gnu.org/licenses/gpl-3.0)
[![Shell](https://img.shields.io/badge/shell-bash-green)](https://www.gnu.org/software/bash/)
[![Distros](https://img.shields.io/badge/distros-Ubuntu%20%7C%20Debian%20%7C%20Fedora%20%7C%20Arch-orange)](#️-compatibilidade)
[![GitHub](https://img.shields.io/badge/GitHub-kaduvelasco%2Flumina--stack-181717?logo=github)](https://github.com/kaduvelasco/lumina-stack)
[![CI](https://img.shields.io/github/actions/workflow/status/kaduvelasco/lumina-stack/lint.yml?label=lint%20%26%20smoke%20test)](https://github.com/kaduvelasco/lumina-stack/actions)

📄 Portuguese version: see [LEIAME.md](LEIAME.md)

O **LuminaStack** automatiza a criação de um ecossistema completo para desenvolvimento PHP local. Através de scripts Bash modulares, você configura **Nginx**, **múltiplas versões de PHP-FPM** e **MariaDB** em minutos, mantendo seu sistema operacional limpo e performático.

---

## ✨ Funcionalidades

| Recurso                    | Descrição                                                                          |
| -------------------------- | ---------------------------------------------------------------------------------- |
| **Setup Automático**       | Instalação de pré-requisitos e Docker com detecção automática de distro            |
| **Multi-versão PHP**       | Execute PHP 7.4 a 8.4 simultaneamente, cada um em seu próprio container            |
| **Roteamento Dinâmico**    | Acesse `http://phpXX.localhost` para alternar entre versões instantaneamente       |
| **Workspace Organizado**   | Estrutura de pastas padronizada e sincronizável via MegaSync                       |
| **Status em Tempo Real**   | Visualize saúde, uptime e uso de CPU/RAM de cada container diretamente no menu     |
| **Moodle Ready**           | Otimização de performance do MariaDB baseada na RAM disponível                     |
| **Backups Inteligentes**   | Dumps automáticos com timestamp, mantendo os 3 mais recentes localmente            |
| **Logs Segregados**        | Logs individuais por versão de PHP e por Nginx, com rotação automática             |
| **Permissões Automáticas** | Ajuste automático de permissões a cada inicialização, sem conflitos Host/Container |
| **Segurança**              | Portas bound em `127.0.0.1`, headers HTTP, bloqueio de arquivos sensíveis no Nginx |

---

## 📦 Tecnologias

- **Docker & Docker Compose** (Compose Spec)
- **Nginx 1.26 Alpine** — Reverse proxy com roteamento dinâmico por subdomínio
- **PHP-FPM** — Versões 7.4, 8.0, 8.1, 8.2, 8.3 e 8.4
- **MariaDB 11.4**
- **Bash 4.0+**

---

## 🖥️ Compatibilidade

| Distribuição                            | Suporte                         |
| --------------------------------------- | ------------------------------- |
| Ubuntu / Linux Mint / Pop!\_OS / Debian | ✅ Completo                     |
| Fedora                                  | ✅ Completo ⚠️ Ver nota SELinux |
| Arch Linux / Manjaro                    | ✅ Completo                     |

> **Fedora — SELinux:** Se os containers não conseguirem ler/escrever nos volumes, execute:
>
> ```bash
> sudo setsebool -P container_manage_cgroup on
> ```

---

## 📁 Estrutura do Projeto

```text
lumina-stack/
 ├── install.sh                  # Ponto de entrada — instalador interativo
 ├── clean-docker.sh             # Reset completo do ambiente Docker
 ├── lib/
 │   ├── utils.sh                # Cores, funções de saída e utilitários compartilhados
 │   ├── versions.sh             # Fonte única de versões suportadas
 │   ├── menu.sh                 # Menu do instalador
 │   ├── system.sh               # Detecção de distro, Docker e /etc/hosts
 │   ├── workspace.sh            # Criação da estrutura de pastas
 │   └── docker.sh               # Geração da stack Docker
 └── templates/
     ├── docker-compose.tpl      # Template da stack Docker
     ├── nginx.conf.tpl          # Configuração do proxy reverso
     ├── php.Dockerfile.tpl      # Dockerfile PHP-FPM
     ├── php.ini.tpl             # Configuração PHP
     ├── index.php.tpl           # Dashboard de desenvolvimento
     └── info.php.tpl            # Página phpinfo()
```

---

## 📂 Estrutura do Workspace

Após a instalação, o diretório `/srv/workspace` terá a seguinte estrutura:

```text
workspace/
 ├── www/
 │   ├── html/         # Seus projetos PHP            🔄 Sincronizado via MegaSync (opcional)
 │   └── data/         # Dados do Moodle (moodledata) 🔄 Sincronizado via MegaSync (opcional)
 ├── backups/          # Dumps SQL                     🔄 Sincronizado via MegaSync (opcional)
 ├── logs/             # Logs por versão de PHP e Nginx
 ├── databases/        # Arquivos binários do MariaDB  🚫 Não sincronizado
 └── docker/
     ├── docker-compose.yml
     ├── .env          # Credenciais (chmod 600)
     ├── nginx/        # Configuração do proxy
     ├── php/          # Dockerfile por versão
     ├── php-config/   # php.ini customizado
     └── mariadb/      # Configurações e otimizações
```

---

## 🚀 Instalação

**1. Clone o repositório:**

```bash
git clone https://github.com/kaduvelasco/lumina-stack.git
cd lumina-stack
```

**2. Dê permissão de execução e inicie o instalador:**

```bash
chmod +x install.sh
./install.sh
```

**3. Siga a ordem numérica do menu para um setup completo:**

| Passo | Opção                    | Descrição                                                     |
| ----- | ------------------------ | ------------------------------------------------------------- |
| 1     | Instalar pré-requisitos  | Detecta sua distro e instala `curl`, `git`, `openssl`, `lsof` |
| 2     | Instalar Docker          | Configura o engine e as permissões de grupo                   |
| 3     | Criar workspace          | Gera a estrutura de pastas em `/srv/workspace`                   |
| 4     | Gerar stack Docker       | Define versões PHP, usuário e senha do banco                  |

> **Após a opção 2**, pode ser necessário reiniciar a sessão para aplicar as permissões do grupo `docker`.

---

## 🌐 Roteamento Dinâmico

O Nginx roteia automaticamente cada subdomínio para o container PHP correspondente:

| URL                      | Container                                          |
| ------------------------ | -------------------------------------------------- |
| `http://localhost`       | Dashboard — PHP padrão (primeira versão instalada) |
| `http://php74.localhost` | Container PHP 7.4                                  |
| `http://php81.localhost` | Container PHP 8.1                                  |
| `http://php83.localhost` | Container PHP 8.3                                  |
| `http://php84.localhost` | Container PHP 8.4                                  |

> **Dica:** Digite sempre com `http://` explícito. Navegadores modernos podem tentar forçar HTTPS, resultando em erro de conexão.

---

## 📅 Fluxo de Trabalho Diário

```bash
# 1. Inicie a stack
docker compose -f /srv/workspace/docker/docker-compose.yml up -d

# 2. Desenvolva seus projetos em:
/srv/workspace/www/html/

# 3. Teste no navegador:
http://phpXX.localhost

# 4. Ao encerrar, derrube a stack
docker compose -f /srv/workspace/docker/docker-compose.yml down
```

---

## 🧨 Reset Completo do Ambiente

Para reinstalar o LuminaStack do zero — mantendo seus projetos e backups intactos — utilize o script `clean-docker.sh`:

```bash
chmod +x clean-docker.sh
./clean-docker.sh
```

O script solicita **duas confirmações** antes de executar — a segunda exige digitar `SIM` em maiúsculo — e informa claramente o que será e o que não será removido:

|                                                   | O que acontece                                 |
| ------------------------------------------------- | ---------------------------------------------- |
| 🗑️ Containers, imagens, volumes e networks Docker | **Removidos**                                  |
| 🗑️ `/srv/workspace/databases/`                       | **Removido** — MariaDB recria ao subir         |
| 🗑️ `/srv/workspace/docker/`                          | **Removido** — configs geradas pelo instalador |
| ✅ `/srv/workspace/www/html/`                        | **Preservado** — seus projetos PHP             |
| ✅ `/srv/workspace/www/data/`                        | **Preservado** — moodledata                    |
| ✅ `/srv/workspace/backups/`                         | **Preservado** — dumps SQL                     |

Após a limpeza, execute o instalador novamente a partir da **opção 3**:

```bash
./install.sh
# Opções: 3 → 4
```

> As opções 1 e 2 só precisam ser repetidas se você mudou de máquina ou reinstalou o sistema operacional.

---

## 🔒 Segurança

O LuminaStack aplica boas práticas de segurança por padrão:

- **Portas bound em `127.0.0.1`** — Nginx (80) e MariaDB (3306) acessíveis apenas localmente, não pela rede
- **`.env` com `chmod 600`** — credenciais do banco legíveis apenas pelo dono
- **Escrita atômica do `.env`** — arquivo temporário com `chmod 600` antes de mover, sem janela de exposição
- **`MYSQL_PWD`** — senha do banco passada via variável de ambiente, não via argumento de linha de comando
- **Nginx bloqueia arquivos sensíveis** — `.env`, `.git/`, `vendor/`, `node_modules/` e extensões como `.sql`, `.log`, `.sh`
- **Headers HTTP de segurança** — `X-Content-Type-Options`, `X-Frame-Options`, `X-XSS-Protection` e `Referrer-Policy`
- **Validação de entradas** — usuário e senha do banco validados antes de qualquer operação
- **Identificadores SQL com crases** — nomes de usuário no SQL usam `` `backticks` `` (identificadores corretos)

---

## 🔧 Solução de Problemas

**Subdomínio não resolve (erro de conexão)**

Verifique se o `/etc/hosts` foi atualizado com as entradas `# lumina-stack`. Se usar Firefox ou Zen Browser, desative o **DNS sobre HTTPS (DoH)** nas configurações de rede do navegador.

**PHP não consegue gravar arquivos**

Execute o script de correção de permissões ou rode `docker compose` com o usuário correto. Isso é especialmente comum após sincronização via MegaSync, que não preserva permissões de arquivo.

**Porta 80 em uso**

Outro serviço (Apache ou Nginx local) está ocupando a porta. Desative-o antes de subir a stack:

```bash
sudo systemctl stop apache2
# ou
sudo systemctl stop nginx
```

Verifique a porta 80 antes de subir os containers.

**Containers sobem mas não leem os volumes (Fedora)**

O SELinux está bloqueando o acesso. Execute:

```bash
sudo setsebool -P container_manage_cgroup on
```

**Preciso reiniciar a sessão após instalar o Docker?**

Sim. A opção 2 adiciona seu usuário ao grupo `docker`, mas a mudança só é aplicada após fazer logout e login novamente.

**Ambiente demora muito para iniciar (primeira vez)**

Na primeira execução, o Docker baixa as imagens base e compila os containers PHP — isso pode levar alguns minutos dependendo da conexão. Nas execuções seguintes, o cache de layers reduz significativamente o tempo.

**Xdebug não conecta ao IDE (Linux nativo)**

O `extra_hosts: host.docker.internal:host-gateway` é configurado automaticamente em cada container PHP, tornando `host.docker.internal` resolvível em Linux sem Docker Desktop. Certifique-se de que o IDE está escutando na porta `9003`.

---

## ⚠️ Requisitos

- Sistema operacional Linux (Ubuntu, Fedora ou Arch — ver tabela de compatibilidade)
- Porta **80** disponível no host
- Usuário com permissão de `sudo`
- Conexão com a internet durante a instalação
- Bash **4.0** ou superior

---

## 🤝 Contribuindo

Contribuições são bem-vindas! Para contribuir:

1. Faça um fork do repositório
2. Crie uma branch: `git checkout -b feature/minha-melhoria`
3. Siga o padrão dos scripts existentes (cabeçalho, `source lib/utils.sh`, `set -euo pipefail`)
4. Certifique-se de que o ShellCheck passa sem warnings: `shellcheck -x seu-script.sh`
5. Abra um Pull Request descrevendo o que foi alterado

---

## 📋 Changelog

### v3.0.0

**Bug fixes**

- MariaDB 11.4+ compatibility: removed `RENAME USER` from init script — MariaDB 11.4 creates `MYSQL_USER` with `%` host by default (not `localhost`), causing the container to crash on first initialization
- Fixed `update_hosts` fallback to use `SUPPORTED_PHP_VERSIONS` from `versions.sh` — previously missing PHP 8.0
- Fixed `${v/./}` → `${v//./}` in `workspace.sh` for consistent log directory naming across all PHP versions
- Fixed `clean-docker.sh` hardcoded paths — now reads workspace location from `~/.lumina/config.env`

**Maintainability**

- Docker image versions (`NGINX_IMAGE`, `MARIADB_IMAGE`) now resolved from `versions.sh` at stack generation time — update once, propagate everywhere
- Load guards added to all library files (`menu.sh`, `system.sh`, `workspace.sh`, `docker.sh`) — prevents re-definition on multiple source calls

**Style**

- Removed `sleep 1` from invalid option handler in `install.sh`
- Fixed stray `│` character on menu item "Sair"

---

### v2.1.0

**Segurança**

- Portas do Nginx (80) e MariaDB (3306) agora fazem bind em `127.0.0.1` — não expostas na rede local
- Nginx bloqueia acesso a `.env`, `.git/`, `vendor/`, `node_modules/` e arquivos sensíveis (`.sql`, `.log`, `.sh`, etc.)
- Headers HTTP de segurança adicionados em todos os virtual hosts: `X-Content-Type-Options`, `X-Frame-Options`, `X-XSS-Protection`, `Referrer-Policy`
- Escrita atômica do `.env` via arquivo temporário com `chmod 600` antes do `mv` — elimina janela de exposição
- Identificadores SQL de usuário trocados de aspas simples para crases (sintaxe correta para identificadores MariaDB)
**Novas funcionalidades**

- `lib/versions.sh` — fonte única de versões (`SUPPORTED_PHP_VERSIONS`, `NGINX_IMAGE`, `MARIADB_IMAGE`)
- Instalação do Docker oferece escolha entre package manager (`apt`/`dnf`) e script oficial (`get.docker.com`)
- Descrição contextual em cada opção do menu do instalador (`↳ ...`)
- Dupla confirmação em `clean-docker.sh` — segunda confirmação exige digitar `SIM` em maiúsculo

**Infraestrutura Docker**

- Logging com rotação automática em todos os containers: `json-file`, max 10MB × 3 arquivos
- Resource limits configurados: Nginx 1CPU/256M, MariaDB 2CPU/2G, PHP 2CPU/1G
- Healthcheck adicionado ao Nginx (`wget --spider`)
- Nginx aguarda todos os containers PHP via `depends_on: condition: service_healthy`
- Healthcheck dos containers PHP trocado de `php -v` para `php-fpm -t` — verifica FPM real
- `start_period` do MariaDB aumentado para 60s com 10 retries — evita falso `unhealthy` na primeira inicialização
- `extra_hosts: host.docker.internal:host-gateway` em cada container PHP — resolve Xdebug no Linux nativo

**Performance**

- BuildKit cache mount para `apt-get` no Dockerfile PHP — evita re-download em rebuilds
- `docker compose down --timeout 5 --remove-orphans` — shutdown até 10× mais rápido
- Compressão gzip habilitada no Nginx para text, CSS, JS, JSON, XML e SVG

**Correções de bugs**

- `ler_credenciais()` valida usuário e senha com até 3 tentativas; propaga falha para os chamadores
- `executar_restore()` valida existência do arquivo no disco antes de restaurar
- `remover_bancos_de_dados()` separa captura e filtragem da saída do `docker exec` — falhas de conexão são reportadas
- `docker compose up` e `docker compose down` verificam código de retorno — falhas não passam silenciosamente
- `limpar_backups_antigos()` usa process substitution — erros de `rm` são detectados e reportados
- `detect_distro()` exporta `DISTRO` — valor persiste entre chamadas de funções no mesmo processo
- `check_port_80()` verifica disponibilidade de `lsof` antes de usá-lo
- Instalação do Docker verifica retorno antes de remover `get-docker.sh`
- Template `docker-compose.tpl` validado antes do `awk`; saída verificada após geração
- `create_workspace()` detecta workspace existente e pede confirmação antes de sobrescrever
- `PHP_VERSIONS` normalizado com `xargs` — espaços extras não geram serviços com nome vazio
- `stat -c` com fallback para `date -r` — portabilidade em sistemas sem GNU coreutils
- `unset MAP` antes de `declare -A MAP` no loop de logs — mapa reiniciado a cada iteração

**Manutenibilidade**

- `otimizar_mariadb_moodle()` dividida em `detect_system_ram()`, `prompt_buffer_pool_allocation()` e `write_mariadb_config()`
- CI inclui `lib/versions.sh` no ShellCheck e na lista de arquivos obrigatórios

---

### v2.0.0

**Segurança**

- Credenciais MariaDB passadas via `MYSQL_PWD` (evita exposição no `ps aux`)
- `chmod 600` aplicado no `.env` gerado pelo instalador
- Senha root do banco gerada dinamicamente (não mais hardcoded)
- Validação de caracteres no nome de usuário do banco

**Novas funcionalidades**

- Adicionado `lib/utils.sh` com paleta ANSI centralizada e funções de saída padronizadas
- `fix_permissions` executado automaticamente a cada inicialização
- Limpeza automática de backups: mantém os 3 mais recentes localmente
- Restore do banco com seleção numerada (sem digitação manual do nome)
- Detecção automática de RAM em `otimizar_mariadb_moodle`
- Placeholder `{{DEFAULT_PHP}}` no `nginx.conf` elimina `php81` hardcoded
- `depends_on` com `condition: service_healthy` no `docker-compose`
- `update_hosts` cirúrgico com marcador `# lumina-stack`
- Feedback de conclusão com pausa após cada opção do instalador
- Script `clean-docker.sh` para reset completo do ambiente

**Correções**

- Corrigido `cut -d= -f2` → `f2-` para senhas com caracteres `=`
- Removido `-t` do `docker exec` em `verificar_tabelas`
- Fallbacks inseguros removidos do `docker-compose.yml`

**Padronização**

- Cabeçalho de documentação adicionado em todos os arquivos `.sh`
- Cores e mensagens padronizadas em todos os scripts
- Aviso de SELinux adicionado para usuários Fedora
- README reescrito com tabelas, estrutura do projeto e seção de reset

---

## 📜 Licença

Distribuído sob a licença **GPL-3.0**. Consulte o arquivo [LICENSE](LICENSE) para mais informações.

---

---

Made with ❤️ and AI by [Kadu Velasco](https://github.com/kaduvelasco)
