FROM docker.io/library/httpd:2-alpine3.17
# The github workflow don't build from the same place
ARG THIS_DIR

# mounts is in format "dirname-datapath:containerpath"
LABEL studentbox.config.mounts="html:/var/www/html"
EXPOSE 80

RUN apk upgrade --no-cache

COPY ${THIS_DIR}/httpd.conf /usr/local/apache2/conf/httpd.conf
COPY ${THIS_DIR}/to-php.apache.conf /usr/local/apache2/conf/to-php.apache.conf

RUN echo "Include /usr/local/apache2/conf/to-php.apache.conf" >> /usr/local/apache2/conf/httpd.conf
