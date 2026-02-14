# docker build --rm -f docker/Dockerfile.make -t woodpecker/make:local .
FROM docker.io/golang:1.26-alpine AS golang_image
FROM docker.io/node:24-alpine

RUN apk add --no-cache --update make gcc binutils-gold musl-dev && \
    apk add --no-cache --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main protoc && \
  corepack enable

# Build packages.
COPY --from=golang_image /usr/local/go /usr/local/go
COPY Makefile /
ENV PATH=$PATH:/usr/local/go/bin
ENV COREPACK_ENABLE_DOWNLOAD_PROMPT=0
ENV COREPACK_ENABLE_AUTO_PIN=0

# Cache tools
RUN GOBIN=/usr/local/go/bin make install-tools && \
    rm -rf /Makefile

ENV GOPATH=/tmp/go
ENV HOME=/tmp/home
ENV PATH=$PATH:/usr/local/go/bin:/tmp/go/bin

WORKDIR /build
RUN chmod -R 777 /root

CMD [ "/bin/sh" ]
