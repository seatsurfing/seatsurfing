#!/bin/sh

# (Re)start mailhog
docker stop mailhog
docker rm mailhog
docker run --rm -d -p 1025:1025 -p 8025:8025 --name mailhog richarvey/mailhog

# Set environment variables
DEV=1
CRYPT_KEY=rC8REJftxMcdhzTvu9Tk6RqgygBRctZC
STATIC_UI_PATH=../ui/build
PLUGINS_SUB_PATH=../../plugins/build
SMTP_HOST=127.0.0.1
SMTP_PORT=1025 
PUBLIC_SCHEME=http
PUBLIC_PORT=8080
if [ -f .env ]; then
    echo "Reading additional environments from .env file..."
    export $(grep -v '^#' .env | xargs)
fi

# start Seatsurfing server
DEV=$DEV PUBLIC_SCHEME=$PUBLIC_SCHEME PUBLIC_PORT=$PUBLIC_PORT CRYPT_KEY=$CRYPT_KEY STATIC_UI_PATH=$STATIC_UI_PATH PLUGINS_SUB_PATH=$PLUGINS_SUB_PATH SMTP_HOST=$SMTP_HOST SMTP_PORT=$SMTP_PORT go run `ls *.go | grep -v _test.go`

# Stop mailhog
docker stop mailhog
docker rm mailhog
