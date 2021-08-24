#! /busybox/sh

REGISTRY="docker.io"
IMAGE="woodpeckerci/woodpecker-$IMAGE_TYPE"
PATH_BINARY="./dist/cli_linux_amd64/woodpecker-$IMAGE_TYPE"
PATH_CONTEXT="./dist/docker-woodpecker-$IMAGE_TYPE"
PATH_DOCKERFILE="./docker/Dockerfile.cli"

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

echo "Building image ..."
/kaniko/executor \
  --context $PATH_CONTEXT/ \
  --destination $IMAGE:latest \
  --destination $IMAGE:$WOODPECKER_VERSION \

echo "Building alpine image ..."
/kaniko/executor \
  --context $PATH_CONTEXT-alpine/ \
  --destination $IMAGE:latest-alpine \
  --destination $IMAGE:$WOODPECKER_VERSION-alpine \

echo "Done"
