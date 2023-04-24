FROM docker.io/library/mariadb:10.9
ARG THIS_DIR

LABEL studentbox.config.mounts="db:/var/lib/mysql"

ENV MARIADB_DATABASE=app MARIADB_USER=student
LABEL studentbox.config.envs="MARIADB_DATABASE:failempty,MARIADB_USER,MARIADB_PASSWORD:password(30)"
