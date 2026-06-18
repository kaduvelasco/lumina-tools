ARG PHP_VERSION=8.1
FROM php:${PHP_VERSION}-fpm

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update \
 && apt-get install -y --no-install-recommends \
    libzip-dev \
    libpng-dev \
    libicu-dev \
    libxml2-dev \
    libonig-dev \
    libjpeg-dev \
    libfreetype6-dev \
    libpq-dev \
    libcurl4-openssl-dev \
    libxslt-dev \
 && rm -rf /var/lib/apt/lists/*

RUN docker-php-ext-configure gd \
    --with-freetype \
    --with-jpeg

RUN docker-php-ext-install \
pdo_mysql \
mysqli \
intl \
zip \
gd \
soap \
mbstring \
exif \
xml \
opcache

RUN pecl install redis || true \
 && docker-php-ext-enable redis || true

RUN pecl install xdebug || true \
 && docker-php-ext-enable xdebug || true

# Instalar Composer (Cópia direta do binário oficial)
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

# Instalar PHP-CS-Fixer (versão pinada para build reproduzível)
RUN curl -L https://github.com/PHP-CS-Fixer/PHP-CS-Fixer/releases/download/v3.65.0/php-cs-fixer.phar \
    -o /usr/local/bin/php-cs-fixer \
    && chmod +x /usr/local/bin/php-cs-fixer

RUN usermod -u 1000 www-data

WORKDIR /var/www/html
