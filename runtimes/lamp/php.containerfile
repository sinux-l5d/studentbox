FROM docker.io/library/php:8-fpm-alpine3.17
ARG THIS_DIR

LABEL studentbox.config.mounts="html:/var/www/html"

RUN apk upgrade --no-cache && docker-php-ext-install mysqli