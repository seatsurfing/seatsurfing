#!/bin/sh
docker stop mailhog
docker rm mailhog
docker run --rm -d -p 1025:1025 -p 8025:8025 --name mailhog richarvey/mailhog
DEV=1 CRYPT_KEY=rC8REJftxMcdhzTvu9Tk6RqgygBRctZC PLUGINS_SUB_PATH=../../plugins/build ORG_SIGNUP_ENABLED=1 ORG_SIGNUP_DELETE=1 PUBLIC_LISTEN_ADDR=0.0.0.0:8080 SMTP_HOST=127.0.0.1 SMTP_PORT=1025 POSTGRES_URL=postgres://postgres:root@localhost/seatsurfing?sslmode=disable STATIC_ADMIN_UI_PATH=../admin-ui/build STATIC_BOOKING_UI_PATH=../booking-ui/build PRINT_CONFIG=1 go run `ls *.go | grep -v _test.go`
docker stop mailhog
docker rm mailhog
