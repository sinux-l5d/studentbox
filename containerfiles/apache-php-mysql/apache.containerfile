FROM docker.io/library/httpd:2-alpine3.17

VOLUME [ "/var/www/html" ]
EXPOSE 80

RUN apk upgrade --no-cache

COPY ./httpd.conf /usr/local/apache2/conf/httpd.conf
COPY ./to-php.apache.conf /usr/local/apache2/conf/to-php.apache.conf

RUN echo "Include /usr/local/apache2/conf/to-php.apache.conf" >> /usr/local/apache2/conf/httpd.conf
