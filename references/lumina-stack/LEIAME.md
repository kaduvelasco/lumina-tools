# 🚀 LuminaStack

> Ambiente de desenvolvimento PHP modular com roteamento dinâmico via Docker.

[![License](https://img.shields.io/badge/license-GPL--3.0-blue)](https://www.gnu.org/licenses/gpl-3.0)
[![Shell](https://img.shields.io/badge/shell-bash-green)](https://www.gnu.org/software/bash/)
[![Distros](https://img.shields.io/badge/distros-Ubuntu%20%7C%20Debian%20%7C%20Fedora%20%7C%20Arch-orange)](#️-compatibilidade)
[![GitHub](https://img.shields.io/badge/GitHub-kaduvelasco%2Flumina--stack-181717?logo=github)](https://github.com/kaduvelasco/lumina-stack)
[![CI](https://img.shields.io/github/actions/workflow/status/kaduvelasco/lumina-stack/lint.yml?label=lint%20%26%20smoke%20test)](https://github.com/kaduvelasco/lumina-stack/actions)

📄 English version: see [README.md](README.md)

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
| 3     | Criar workspace          | Gera a estrutura de pastas em `/srv/workspace`                |
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
| 🗑️ `/srv/workspace/databases/`                    | **Removido** — MariaDB recria ao subir         |
| 🗑️ `/srv/workspace/docker/`                       | **Removido** — configs geradas pelo instalador |
| ✅ `/srv/workspace/www/html/`                     | **Preservado** — seus projetos PHP             |
| ✅ `/srv/workspace/www/data/`                     | **Preservado** — moodledata                    |
| ✅ `/srv/workspace/backups/`                      | **Preservado** — dumps SQL                     |

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

Contribuições são bem-vindas! Consulte o arquivo [CONTRIBUINDO.md](CONTRIBUINDO.md) para o guia completo.

---

## 📜 Licença

Distribuído sob a licença **GPL-3.0**. Consulte o arquivo [LICENSE](LICENSE) para mais informações.

---

Made with ❤️ and AI by [Kadu Velasco](https://github.com/kaduvelasco)
