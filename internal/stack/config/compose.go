package stackconfig

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

var supportedPHP = []string{"8.1", "8.2", "8.3", "8.4"}

// Compose generates docker-compose.yml and all supporting config files.
func Compose(_ context.Context, _ *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Criar Stack Docker")

	cfg, err := config.Load()
	if err != nil {
		ui.Err(stdout, "Falha ao carregar config: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("carregar config: %w", err)
	}

	workspace := cfg.WorkspacePath
	if workspace == "" {
		home, _ := os.UserHomeDir()
		workspace = filepath.Join(home, "workspace")
	}
	dockerDir := filepath.Join(workspace, "docker")

	// PHP versions
	fmt.Fprintf(stdout, "Versões PHP suportadas: %s\n", strings.Join(supportedPHP, " "))
	fmt.Fprint(stdout, "Versões desejadas (ex: 8.1 8.2): ")
	rawVersions := strings.TrimSpace(prompt.ReadLine())
	if rawVersions == "" {
		rawVersions = "8.1 8.2"
	}
	versions, err := validatePHP(rawVersions)
	if err != nil {
		ui.Err(stdout, err.Error())
		ui.WaitEnter(stdout)
		return err
	}

	// DB credentials
	fmt.Fprint(stdout, "\nUsuário do banco [admin]: ")
	dbUser := strings.TrimSpace(prompt.ReadLine())
	if dbUser == "" {
		dbUser = "admin"
	}

	fmt.Fprint(stdout, "Senha do banco (Enter para gerar): ")
	dbPass := strings.TrimSpace(prompt.ReadLine())
	if dbPass == "" {
		dbPass = genPassword()
		fmt.Fprintln(stdout, "  Senha gerada e salva no arquivo .env")
	}
	dbRoot := genPassword()

	// Create per-version PHP log directories (mounted by docker-compose).
	for _, v := range versions {
		logDir := filepath.Join(workspace, "logs", "php"+strings.ReplaceAll(v, ".", ""))
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			return fmt.Errorf("criar diretorio de log PHP %s: %w", v, err)
		}
	}

	// Generate files
	ui.Info(stdout, "Gerando arquivos em: "+dockerDir)

	for _, d := range []string{
		filepath.Join(dockerDir, "nginx"),
		filepath.Join(dockerDir, "php"),
		filepath.Join(dockerDir, "php-config"),
		filepath.Join(dockerDir, "mariadb", "init"),
		filepath.Join(dockerDir, "mariadb", "conf.d"),
	} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}

	defaultPHP := "php" + strings.ReplaceAll(versions[0], ".", "")
	defaultVer := strings.ReplaceAll(versions[0], ".", "")

	// docker-compose.yml
	compose := buildCompose(versions, workspace)
	if err := os.WriteFile(filepath.Join(dockerDir, "docker-compose.yml"), []byte(compose), 0o644); err != nil {
		return fmt.Errorf("escrever docker-compose.yml: %w", err)
	}

	// .env (sensitive — chmod 600)
	envPath := filepath.Join(dockerDir, ".env")
	envContent := fmt.Sprintf("DB_USER=%s\nDB_PASS=%s\nDB_ROOT_PASSWORD=%s\nPHP_VERSIONS=%s\n",
		dbUser, dbPass, dbRoot, strings.Join(versions, " "))
	f, err := os.OpenFile(envPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("criar .env: %w", err)
	}
	if _, err := f.WriteString(envContent); err != nil {
		f.Close()
		return fmt.Errorf("escrever .env: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("fechar .env: %w", err)
	}

	// nginx/default.conf
	nginxConf := strings.ReplaceAll(nginxConfTpl, "{{DEFAULT_PHP}}", defaultPHP)
	nginxConf = strings.ReplaceAll(nginxConf, "{{DEFAULT_PHP_VER}}", defaultVer)
	if err := os.WriteFile(filepath.Join(dockerDir, "nginx", "default.conf"), []byte(nginxConf), 0o644); err != nil {
		return fmt.Errorf("escrever nginx.conf: %w", err)
	}

	// php/Dockerfile
	if err := os.WriteFile(filepath.Join(dockerDir, "php", "Dockerfile"), []byte(phpDockerfile), 0o644); err != nil {
		return fmt.Errorf("escrever Dockerfile: %w", err)
	}

	// php-config/php.ini
	if err := os.WriteFile(filepath.Join(dockerDir, "php-config", "php.ini"), []byte(phpIni), 0o644); err != nil {
		return fmt.Errorf("escrever php.ini: %w", err)
	}

	// mariadb/init/01-permissions.sql
	sql := fmt.Sprintf("GRANT ALL PRIVILEGES ON *.* TO `%s`@'%%' WITH GRANT OPTION;\nFLUSH PRIVILEGES;\n", dbUser)
	if err := os.WriteFile(filepath.Join(dockerDir, "mariadb", "init", "01-permissions.sql"), []byte(sql), 0o644); err != nil {
		return fmt.Errorf("escrever sql init: %w", err)
	}

	// Persist settings in config
	cfg.DockerComposeDir = dockerDir
	cfg.Stack.PHPVersions = strings.Join(versions, " ")
	cfg.Stack.DBUser = dbUser
	cfg.Stack.DBPass = dbPass
	cfg.Stack.DBRootPass = dbRoot
	if err := config.Save(cfg); err != nil {
		ui.Warning(stdout, "Falha ao salvar configurações: "+err.Error())
	}

	ui.Success(stdout, "Stack gerada com sucesso.")
	ui.Info(stdout, "Versões PHP  : "+strings.Join(versions, " ")+
		"\nPHP padrão   : "+defaultPHP+
		"\nUsuário DB   : "+dbUser+
		"\nCredenciais  : "+envPath+" (chmod 600)")
	ui.WaitEnter(stdout)
	return nil
}

func validatePHP(raw string) ([]string, error) {
	supported := make(map[string]bool)
	for _, v := range supportedPHP {
		supported[v] = true
	}
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return nil, fmt.Errorf("informe ao menos uma versao PHP")
	}
	seen := make(map[string]bool)
	var unique []string
	for _, v := range fields {
		if !supported[v] {
			return nil, fmt.Errorf("versao PHP %q nao suportada. Disponiveis: %s", v, strings.Join(supportedPHP, " "))
		}
		if seen[v] {
			return nil, fmt.Errorf("versao PHP %q duplicada", v)
		}
		seen[v] = true
		unique = append(unique, v)
	}
	return unique, nil
}

func genPassword() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	s := base64.URLEncoding.EncodeToString(b)
	s = strings.Map(func(r rune) rune {
		if r == '=' || r == '+' || r == '/' || r == '-' || r == '_' {
			return -1
		}
		return r
	}, s)
	if len(s) > 16 {
		s = s[:16]
	}
	return s
}

func buildCompose(versions []string, workspace string) string {
	var phpServices, nginxDeps strings.Builder
	for _, v := range versions {
		name := "php" + strings.ReplaceAll(v, ".", "")
		phpServices.WriteString(fmt.Sprintf(`
  %s:
    container_name: %s
    build:
      context: .
      dockerfile: php/Dockerfile
      args:
        PHP_VERSION: %s
    restart: unless-stopped
    volumes:
      - %s/www/html:/var/www/html
      - %s/www/data:/var/www/data
      - %s/logs/%s:/var/log/php
      - ./php-config/php.ini:/usr/local/etc/php/php.ini
    environment:
      - XDEBUG_MODE=off
    extra_hosts:
      - "host.docker.internal:host-gateway"
    healthcheck:
      test: ["CMD-SHELL", "php-fpm -t > /dev/null 2>&1"]
      start_period: 10s
      interval: 15s
      timeout: 5s
      retries: 3
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 1G
    networks:
      - docker-php-network
`, name, name, v, workspace, workspace, workspace, name))

		nginxDeps.WriteString(fmt.Sprintf("      %s:\n        condition: service_healthy\n", name))
	}

	compose := dockerComposeTpl
	compose = strings.ReplaceAll(compose, "{{PHP_SERVICES}}", phpServices.String())
	compose = strings.ReplaceAll(compose, "{{NGINX_DEPENDS_ON}}", nginxDeps.String())
	compose = strings.ReplaceAll(compose, "{{NGINX_IMAGE}}", "nginx:1.26-alpine")
	compose = strings.ReplaceAll(compose, "{{MARIADB_IMAGE}}", "mariadb:11.4")
	compose = strings.ReplaceAll(compose, "{{WORKSPACE}}", workspace)
	return compose
}

// ── embedded templates ────────────────────────────────────────────────────────

const dockerComposeTpl = `x-logging: &default-logging
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"

services:
  nginx:
    container_name: nginx
    image: {{NGINX_IMAGE}}
    restart: unless-stopped
    ports:
      - "127.0.0.1:${NGINX_PORT:-80}:80"
    volumes:
      - {{WORKSPACE}}/www/html:/var/www/html
      - {{WORKSPACE}}/www/data:/var/www/data
      - {{WORKSPACE}}/logs/nginx:/var/log/nginx
      - ./nginx/default.conf:/etc/nginx/conf.d/default.conf
    depends_on:
      mariadb:
        condition: service_healthy
{{NGINX_DEPENDS_ON}}
    healthcheck:
      test: ["CMD-SHELL", "wget -q --spider http://127.0.0.1/ > /dev/null 2>&1"]
      start_period: 10s
      interval: 15s
      timeout: 5s
      retries: 3
    logging: *default-logging
    networks:
      - docker-php-network

{{PHP_SERVICES}}

  mariadb:
    container_name: mariadb
    image: {{MARIADB_IMAGE}}
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD}
      MYSQL_USER: ${DB_USER}
      MYSQL_PASSWORD: ${DB_PASS}
      MYSQL_DATABASE: dev_db
    ports:
      - "127.0.0.1:${MARIADB_PORT:-3306}:3306"
    volumes:
      - {{WORKSPACE}}/databases/mariadb:/var/lib/mysql
      - ./mariadb/conf.d:/etc/mysql/conf.d
      - ./mariadb/init:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD", "mariadb-admin", "ping", "-h", "localhost", "--silent"]
      start_period: 60s
      interval: 10s
      timeout: 10s
      retries: 10
    logging: *default-logging
    networks:
      - docker-php-network

networks:
  docker-php-network:
    name: docker-php-network
    driver: bridge
`

const nginxConfTpl = `server {
    listen 80;
    server_name localhost 127.0.0.1;
    root /var/www/html;
    index index.php index.html;

    location ~ /\. { deny all; }
    location ~ \.(env|sql|bak|log|sh|ini)$ { deny all; }

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        include fastcgi_params;
        fastcgi_pass {{DEFAULT_PHP}}:9000;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
    }

    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
}

server {
    listen 80;
    server_name ~^php(?<p_ver>[0-9]+)\.localhost$;
    resolver 127.0.0.11 valid=30s;
    root /var/www/html;
    index index.php index.html;

    location ~ /\. { deny all; }

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ [^/]\.php(/|$) {
        fastcgi_split_path_info ^(.+\.php)(/.+)$;
        include fastcgi_params;
        if ($p_ver = "") { set $p_ver {{DEFAULT_PHP_VER}}; }
        set $php_upstream php$p_ver:9000;
        fastcgi_pass $php_upstream;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        fastcgi_param PATH_INFO $fastcgi_path_info;
    }
}
`

const phpDockerfile = `ARG PHP_VERSION=8.1
FROM php:${PHP_VERSION}-fpm

RUN apt-get update && apt-get install -y --no-install-recommends \
    libzip-dev libpng-dev libicu-dev libxml2-dev libonig-dev \
    libjpeg-dev libfreetype6-dev libpq-dev libcurl4-openssl-dev libxslt-dev \
 && rm -rf /var/lib/apt/lists/*

RUN docker-php-ext-configure gd --with-freetype --with-jpeg

RUN docker-php-ext-install \
    pdo_mysql mysqli intl zip gd soap mbstring exif xml opcache

RUN pecl install redis || true && docker-php-ext-enable redis || true
RUN pecl install xdebug || true && docker-php-ext-enable xdebug || true

COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

RUN usermod -u 1000 www-data
WORKDIR /var/www/html
`

const phpIni = `log_errors = On
error_log = /var/log/php/error.log

memory_limit = 512M
max_execution_time = 300
upload_max_filesize = 256M
post_max_size = 256M
max_input_vars = 5000

realpath_cache_size = 4096K
realpath_cache_ttl = 600

opcache.enable = 1
opcache.memory_consumption = 256
opcache.interned_strings_buffer = 16
opcache.max_accelerated_files = 20000
opcache.revalidate_freq = 2
opcache.validate_timestamps = 1
opcache.save_comments = 1
opcache.enable_cli = 1

display_errors = On
display_startup_errors = On
error_reporting = E_ALL

date.timezone = America/Sao_Paulo
expose_php = Off
cgi.fix_pathinfo = 1

file_uploads = On
max_file_uploads = 20

session.cookie_httponly = 1
session.use_strict_mode = 1
session.gc_maxlifetime = 1440

xdebug.mode = off
xdebug.client_host = host.docker.internal
xdebug.client_port = 9003
xdebug.start_with_request = yes
xdebug.log = /var/log/php/xdebug.log
`
