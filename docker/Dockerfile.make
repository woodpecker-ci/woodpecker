# docker build --rm  -f docker/Dockerfile.server -t woodpeckerci/woodpecker-server .
FROM golang:1.18-alpine as golang_image
FROM node:16-alpine

RUN apk add make gcc musl-dev

# Build packages.
COPY --from=golang_image /usr/local/go /usr/local/go
ENV PATH=$PATH:/usr/local/go/bin

# Cache tools
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
  go install github.com/rs/zerolog/cmd/lint@latest && \
  go install mvdan.cc/gofumpt@latest

WORKDIR /build

CMD [ "sh" ]
