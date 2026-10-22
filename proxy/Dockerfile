FROM nginx:1.27-alpine
COPY default.conf /etc/nginx/conf.d/default.conf
COPY cloudflare-realip.conf proxy_params.conf /etc/nginx/snippets/
COPY ssl.conf.template /etc/nginx/ssl.conf.template
COPY 40-enable-ssl.sh /docker-entrypoint.d/40-enable-ssl.sh
RUN sed -i 's/\r$//' /docker-entrypoint.d/40-enable-ssl.sh && chmod +x /docker-entrypoint.d/40-enable-ssl.sh
EXPOSE 80 443
