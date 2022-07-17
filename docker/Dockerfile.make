# docker build --rm  -f docker/Dockerfile.server -t woodpeckerci/woodpecker-server .
FROM golang:1.18-alpine as golang_image
FROM node:16-alpine

RUN apk add make gcc musl-dev

# Build packages.
COPY --from=golang_image /usr/local/go /usr/local/go
ENV PATH=$PATH:/usr/local/go/bin

WORKDIR /build

CMD [ "sh" ]
