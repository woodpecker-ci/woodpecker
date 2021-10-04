#! /bin/sh

set -e

REGISTRY="docker.io"
IMAGE="woodpeckerci/woodpecker-$IMAGE_TYPE"
PATH_BINARY="./dist/${IMAGE_TYPE}_linux_amd64/woodpecker-$IMAGE_TYPE"
PATH_CONTEXT="./dist/docker-woodpecker-$IMAGE_TYPE"
PATH_DOCKERFILE="./docker/Dockerfile.$IMAGE_TYPE"

echo "Building $IMAGE_TYPE => $IMAGE:$WOODPECKER_VERSION ..."

# setup credentials
mkdir -p /kaniko/.docker
echo "{\"auths\":{\"$DOCKER_REGISTRY\":{\"username\":\"$DOCKER_USERNAME\",\"password\":\"$DOCKER_PASSWORD\"}}}" \
  > /kaniko/.docker/config.json

echo "Preparing build context ..."
mkdir -p $PATH_CONTEXT/
cp $PATH_BINARY $PATH_CONTEXT/
cp $PATH_DOCKERFILE $PATH_CONTEXT/Dockerfile

# prepare alpine version
mkdir -p $PATH_CONTEXT-alpine/
cp $PATH_BINARY $PATH_CONTEXT-alpine/
cp $PATH_DOCKERFILE.alpine $PATH_CONTEXT-alpine/Dockerfile

if "$WOODPECKER_VERSION" == "next"; then
  echo "Building pre-release (next) image ..."
  /kaniko/executor \
    --context $PATH_CONTEXT/ \
    --destination $IMAGE:next

  echo "Building pre-release (next) alpine image ..."
  /kaniko/executor \
    --context $PATH_CONTEXT-alpine/ \
    --destination $IMAGE:next-alpine
else
  echo "Building image ..."
  /kaniko/executor \
    --context $PATH_CONTEXT/ \
    --destination $IMAGE:latest \
    --destination $IMAGE:$WOODPECKER_VERSION

  echo "Building alpine image ..."
  /kaniko/executor \
    --context $PATH_CONTEXT-alpine/ \
    --destination $IMAGE:latest-alpine \
    --destination $IMAGE:$WOODPECKER_VERSION-alpine
fi

echo "Done"
