FROM docker.io/library/mariadb:10.9
ARG THIS_DIR

LABEL studentbox.config.mounts="db:/var/lib/mysql"
