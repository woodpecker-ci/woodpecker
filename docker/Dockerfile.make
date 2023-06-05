# docker build --rm  -f docker/Dockerfile.make -t woodpecker/make:local .
FROM golang:1.20-alpine as golang_image
FROM node:18-alpine

RUN apk add --no-cache --update make gcc binutils-gold musl-dev && \
  corepack enable

# Build packages.
COPY --from=golang_image /usr/local/go /usr/local/go
ENV PATH=$PATH:/usr/local/go/bin

# Cache tools
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
  go install github.com/rs/zerolog/cmd/lint@latest && \
  go install mvdan.cc/gofumpt@latest && \
  mv /root/go/bin/golangci-lint /usr/local/go/bin/golangci-lint && \
  mv /root/go/bin/lint /usr/local/go/bin/lint && \
  mv /root/go/bin/gofumpt /usr/local/go/bin/gofumpt && \
  chmod 755 /usr/local/go/bin/*

WORKDIR /build
RUN chmod -R 777 /root

CMD [ "/bin/sh" ]
