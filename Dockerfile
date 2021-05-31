# docker build --rm -t drone/drone .

FROM drone/ca-certs
EXPOSE 8000 9000 80 443

ENV GODEBUG=netdns=go
ENV WOODPECKER_DATABASE_DATASOURCE=/var/lib/drone/drone.sqlite
ENV WOODPECKER_DATABASE_DRIVER=sqlite3
ENV WOODPECKER_LETS_ENCRYPT_PATH=/var/lib/drone

ADD release/drone-server /bin/

ENTRYPOINT ["/bin/drone-server"]
