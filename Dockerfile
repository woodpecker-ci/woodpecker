# docker build --rm -t woodpeckerci/woodpecker-server .

# use golang image to copy ssl certs later
FROM golang:1.16

FROM scratch

# copy certs from golang:1.16 image
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 8000 9000 80 443

ENV GODEBUG=netdns=go
ENV DATABASE_DRIVER=sqlite3
ENV DATABASE_CONFIG=/var/lib/drone/drone.sqlite
ENV XDG_CACHE_HOME /var/lib/drone

ADD release/drone-server /bin/

ENTRYPOINT ["/bin/drone-server"]
