# docker build --rm -t woodpeckerci/woodpecker-server .

FROM drone/ca-certs
EXPOSE 8000 9000 80 443

ENV DATABASE_DRIVER=sqlite3
ENV DATABASE_CONFIG=/var/lib/drone/drone.sqlite
ENV GODEBUG=netdns=go
ENV XDG_CACHE_HOME /var/lib/drone

ADD release/woodpecker-server /bin/

ENTRYPOINT ["/bin/woodpecker-server"]
