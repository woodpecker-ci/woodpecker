#!/bin/sh
set -e

# Create woodpecker group if it doesn't exist
if ! getent group woodpecker > /dev/null 2>&1; then
    groupadd --system woodpecker
fi

# Create woodpecker user if it doesn't exist
if ! getent passwd woodpecker > /dev/null 2>&1; then
    useradd \
        --system \
        --gid woodpecker \
        --no-create-home \
        --home-dir /var/lib/woodpecker \
        --shell /sbin/nologin \
        woodpecker
fi
