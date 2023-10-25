# docker build --rm  -f docker/Dockerfile.make -t woodpecker/make:local .
FROM golang:1.21-alpine3.18 as golang_image
FROM node:21-alpine3.18

# renovate: datasource=repology depName=alpine_3_18/make versioning=loose
ENV MAKE_VERSION="4.4.1-r1"
# renovate: datasource=repology depName=alpine_3_18/gcc versioning=loose
ENV GCC_VERSION="12.2.1_git20220924-r108"
# renovate: datasource=repology depName=alpine_3_18/binutils-gold versioning=loose
ENV BINUTILS_GOLD_VERSION="2.40-r7"
# renovate: datasource=repology depName=alpine_3_18/musl-dev versioning=loose
ENV MUSL_DEV_VERSION="1.2.4-r2"

RUN apk add --no-cache --update make=${MAKE_VERSION} gcc=${GCC_VERSION} binutils-gold=2.40-r7 musl-dev=${MUSL_DEV_VERSION} && \
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
