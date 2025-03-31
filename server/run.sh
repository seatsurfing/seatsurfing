#!/bin/sh
docker stop mailhog
docker rm mailhog
docker run --rm -d -p 1025:1025 -p 8025:8025 --name mailhog richarvey/mailhog
DEV=1 CRYPT_KEY=rC8REJftxMcdhzTvu9Tk6RqgygBRctZC PLUGINS_SUB_PATH=../../plugins/build SMTP_HOST=127.0.0.1 SMTP_PORT=1025 go run `ls *.go | grep -v _test.go`
docker stop mailhog
docker rm mailhog
