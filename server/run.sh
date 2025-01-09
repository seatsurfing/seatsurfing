#!/bin/sh
docker stop mailhog
docker rm mailhog
docker run --rm -d -p 1025:1025 -p 8025:8025 --name mailhog richarvey/mailhog
DEV=1 CRYPT_KEY=rC8REJftxMcdhzTvu9Tk6RqgygBRctZC ORG_SIGNUP_ENABLED=1 ORG_SIGNUP_DELETE=1 FRONTEND_URL=https://seatsurfing.loca.lt PUBLIC_LISTEN_ADDR=0.0.0.0:8080 PUBLIC_URL=https://seatsurfing.loca.lt SMTP_HOST=127.0.0.1:1025 POSTGRES_URL=postgres://postgres:root@localhost/flexspace?sslmode=disable STATIC_ADMIN_UI_PATH=../admin-ui/build STATIC_BOOKING_UI_PATH=../booking-ui/build PRINT_CONFIG=1 go run `ls *.go | grep -v _test.go`
docker stop mailhog
docker rm mailhog
