# docker build --rm  -f docker/Dockerfile.make -t woodpecker/make:local .
FROM docker.io/golang:1.23-alpine3.19 as golang_image
FROM docker.io/node:23-alpine3.19

# renovate: datasource=repology depName=alpine_3_19/make versioning=loose
ENV MAKE_VERSION="4.4.1-r2"
# renovate: datasource=repology depName=alpine_3_19/gcc versioning=loose
ENV GCC_VERSION="13.2.1_git20231014-r0"
# renovate: datasource=repology depName=alpine_3_19/binutils-gold versioning=loose
ENV BINUTILS_GOLD_VERSION="2.41-r0"
# renovate: datasource=repology depName=alpine_3_19/musl-dev versioning=loose
ENV MUSL_DEV_VERSION="1.2.4_git20230717-r4"
# renovate: datasource=repology depName=alpine_3_19/protoc versioning=loose
ENV PROTOC_VERSION="24.4-r0"

RUN apk add --no-cache --update make=${MAKE_VERSION} gcc=${GCC_VERSION} binutils-gold=${BINUTILS_GOLD_VERSION} musl-dev=${MUSL_DEV_VERSION} protoc=${PROTOC_VERSION} && \
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
