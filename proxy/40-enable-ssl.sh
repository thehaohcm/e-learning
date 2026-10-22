#!/bin/sh
# Turn on the HTTPS server block only if the certificate files are present.
set -e
if [ -f /etc/nginx/certs/origin.pem ] && [ -f /etc/nginx/certs/origin.key ]; then
    cp /etc/nginx/ssl.conf.template /etc/nginx/conf.d/ssl.conf
    echo "40-enable-ssl.sh: certificate found, HTTPS on port 443 enabled"
else
    rm -f /etc/nginx/conf.d/ssl.conf
    echo "40-enable-ssl.sh: no certificate in ./proxy/certs, serving HTTP on port 80 only"
fi
