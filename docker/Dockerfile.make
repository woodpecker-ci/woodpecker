# docker build --rm  -f docker/Dockerfile.make -t woodpecker/make:local .
FROM golang:1.21-alpine@sha256:926f7f7e1ab8509b4e91d5ec6d5916ebb45155b0c8920291ba9f361d65385806 as golang_image
FROM node:21-alpine@sha256:df76a9449df49785f89d517764012e3396b063ba3e746e8d88f36e9f332b1864

RUN apk add --no-cache --update make gcc binutils-gold musl-dev && \
  corepack enable

# Build packages.
COPY --from=golang_image /usr/local/go /usr/local/go
COPY Makefile /
ENV PATH=$PATH:/usr/local/go/bin

# Cache tools
RUN make install-tools && \
  mv /root/go/bin/* /usr/local/go/bin/ && \
  chmod 755 /usr/local/go/bin/*

WORKDIR /build
RUN chmod -R 777 /root

CMD [ "/bin/sh" ]
