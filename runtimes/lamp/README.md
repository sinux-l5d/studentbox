# LAMP

This is a simple apache+mariadb+php runtime. It allows .htaccess.

## Usage

```bash
export USER=user-name
export PROJECT=project-name
export PROJECT_DIR=./$USER/$PROJECT

mkdir -p $PROJECT_DIR/{code,db}
cp -r ./examples/lamp/{.,}* $PROJECT_DIR/code

podman pod create --name $USER-$PROJECT -p 8080:80
podman run -d --rm --pod $USER-$PROJECT --name $USER-$PROJECT-mysql -v $PROJECT_DIR/db:/var/lib/mysql -e MARIADB_DATABASE=$PROJECT -e MARIADB_RANDOM_ROOT_PASSWORD="yes" -e MARIADB_USER=$USER -e MARIADB_PASSWORD="testtest" ghcr.io/sinux-l5d/studentbox/runtime/lamp.mysql
podman run -d --rm --pod $USER-$PROJECT --name $USER-$PROJECT-php -v $PROJECT_DIR/code:/var/www/html ghcr.io/sinux-l5d/studentbox/runtime/lamp.php
podman run -d --rm --pod $USER-$PROJECT --name $USER-$PROJECT-apache -v $PROJECT_DIR/code:/var/www/html ghcr.io/sinux-l5d/studentbox/runtime/lamp.apache

curl http://localhost:8080 -L
```