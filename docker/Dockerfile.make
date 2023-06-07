# docker build --rm  -f docker/Dockerfile.make -t woodpecker/make:local .
FROM golang:1.20-alpine as golang_image
FROM node:18-alpine

RUN apk add --no-cache --update make gcc binutils-gold musl-dev && \
  corepack enable

# Build packages.
COPY --from=golang_image /usr/local/go /usr/local/go
ENV PATH=$PATH:/usr/local/go/bin

# Cache tools
RUN make install-tools && \
  mv /root/go/bin/* /usr/local/go/bin/ && \
  chmod 755 /usr/local/go/bin/*

WORKDIR /build
RUN chmod -R 777 /root

CMD [ "/bin/sh" ]
