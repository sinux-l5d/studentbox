FROM docker.io/library/mariadb:10.9
ARG THIS_DIR

LABEL studentbox.config.mounts="db:/var/lib/mysql"

# Env var defaults
ENV MARIADB_DATABASE=app MARIADB_USER=student
# Env var config (required and modifiers)
LABEL studentbox.config.envs="MARIADB_DATABASE:failempty,MARIADB_USER:failempty,MARIADB_PASSWORD:password(10),MARIADB_ROOT_PASSWORD:password(30)"
