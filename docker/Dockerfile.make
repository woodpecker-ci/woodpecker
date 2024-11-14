# docker build --rm  -f docker/Dockerfile.make -t woodpecker/make:local .
FROM docker.io/golang:1.23-alpine as golang_image
FROM docker.io/node:23-alpine

RUN apk add --no-cache --update make gcc binutils-gold musl-dev protoc && \
  corepack enable

# Build packages.
COPY --from=golang_image /usr/local/go /usr/local/go
COPY Makefile /
ENV PATH=$PATH:/usr/local/go/bin
ENV COREPACK_ENABLE_DOWNLOAD_PROMPT=0

# Cache tools
RUN GOBIN=/usr/local/go/bin make install-tools && \
    rm -rf /Makefile

ENV GOPATH=/tmp/go
ENV HOME=/tmp/home
ENV PATH=$PATH:/usr/local/go/bin:/tmp/go/bin

WORKDIR /build
RUN chmod -R 777 /root

CMD [ "/bin/sh" ]
