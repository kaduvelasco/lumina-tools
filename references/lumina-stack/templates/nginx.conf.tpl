# ==============================================================================
# LuminaStack - Configuração Nginx
# ==============================================================================
# Gerado automaticamente pelo install.sh. Não edite manualmente.
# Para alterar o PHP padrão, execute novamente a opção 4 do instalador.
# ==============================================================================

# --- COMPRESSÃO GZIP ---
gzip on;
gzip_vary on;
gzip_proxied any;
gzip_comp_level 6;
gzip_min_length 1000;
gzip_types
    text/plain
    text/css
    text/xml
    text/javascript
    application/json
    application/javascript
    application/xml+rss
    application/atom+xml
    image/svg+xml;

# --- BLOCO 1: DASHBOARD (localhost) ---
# Serve o index.php com a versão PHP padrão definida na instalação.
server {
    listen 80;
    server_name localhost 127.0.0.1;
    root /var/www/html;
    index index.php index.html;

    # --- Bloquear acesso a arquivos e diretórios sensíveis ---
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }
    location ~ \.(env|git|sql|bak|log|sh|ini|swp|dist|config)$ {
        deny all;
        access_log off;
        log_not_found off;
    }
    location ~ ^/(vendor|node_modules)/ {
        deny all;
        access_log off;
        log_not_found off;
    }

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        include fastcgi_params;
        fastcgi_pass {{DEFAULT_PHP}}:9000;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        fastcgi_param HTTP_X_FORWARDED_FOR $proxy_add_x_forwarded_for;
        fastcgi_param HTTP_X_FORWARDED_HOST $server_name;
    }

    # --- Headers de segurança HTTP ---
    add_header X-Content-Type-Options  "nosniff"                        always;
    add_header X-Frame-Options         "SAMEORIGIN"                     always;
    add_header X-XSS-Protection        "1; mode=block"                  always;
    add_header Referrer-Policy         "strict-origin-when-cross-origin" always;
}

# --- BLOCO 2: ROTEAMENTO DINÂMICO (phpXX.localhost) ---
# Roteia para o container PHP correspondente à versão do subdomínio.
# Exemplo: php81.localhost → container php81:9000
server {
    listen 80;
    server_name ~^php(?<p_ver>[0-9]+)\.localhost$;

    resolver 127.0.0.11 valid=30s;
    root /var/www/html;
    index index.php index.html;

    # --- Bloquear acesso a arquivos e diretórios sensíveis ---
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }
    location ~ \.(env|git|sql|bak|log|sh|ini|swp|dist|config)$ {
        deny all;
        access_log off;
        log_not_found off;
    }
    location ~ ^/(vendor|node_modules)/ {
        deny all;
        access_log off;
        log_not_found off;
    }

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
        fastcgi_param HTTP_X_FORWARDED_FOR $proxy_add_x_forwarded_for;
        fastcgi_param HTTP_X_FORWARDED_HOST $server_name;
    }

    # --- Headers de segurança HTTP ---
    add_header X-Content-Type-Options  "nosniff"                        always;
    add_header X-Frame-Options         "SAMEORIGIN"                     always;
    add_header X-XSS-Protection        "1; mode=block"                  always;
    add_header Referrer-Policy         "strict-origin-when-cross-origin" always;
}
