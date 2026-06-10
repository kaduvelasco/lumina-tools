log_errors = On
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
opcache.fast_shutdown = 1
opcache.save_comments = 1
opcache.enable_cli = 1

; Ambiente de desenvolvimento local — não use em produção
display_errors = On
display_startup_errors = On
error_reporting = E_ALL

date.timezone = America/Sao_Paulo

cgi.fix_pathinfo = 1
expose_php = Off

file_uploads = On
max_file_uploads = 20

session.cookie_httponly = 1
session.use_strict_mode = 1
session.gc_maxlifetime = 1440

xdebug.mode = off
; Funciona com extra_hosts "host.docker.internal:host-gateway" no docker-compose (Linux e Docker Desktop)
xdebug.client_host = host.docker.internal
xdebug.client_port = 9003
xdebug.start_with_request = yes
xdebug.log = /var/log/php/xdebug.log
