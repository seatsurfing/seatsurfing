#!/bin/sh

# (Re)start postgres and create databases (if they do not exist)
if docker ps --format '{{.Names}}' | grep -q '^postgres-seatsurfing$'; then
    echo "Postgres database already running, skipping start …"
else
    echo "Starting Postgres database …"
    docker stop postgres-seatsurfing 2>/dev/null
    docker rm postgres-seatsurfing 2>/dev/null
    docker run --rm -d \
        -p 5432:5432 \
        -e POSTGRES_PASSWORD=root \
        -e POSTGRES_USER=postgres \
        -v postgres-seatsurfing:/var/lib/postgresql/data \
        --name postgres-seatsurfing \
        postgres:17-alpine
    until docker exec postgres-seatsurfing pg_isready -U postgres > /dev/null 2>&1; do sleep 1; done
fi
docker exec postgres-seatsurfing psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname='seatsurfing'" | grep -q 1 || docker exec postgres-seatsurfing psql -U postgres -c "CREATE DATABASE seatsurfing"
docker exec postgres-seatsurfing psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname='seatsurfing_test'" | grep -q 1 || docker exec postgres-seatsurfing psql -U postgres -c "CREATE DATABASE seatsurfing_test"

# (Re)start mailhog
echo "Starting mailhog …"
docker stop mailhog-seatsurfing 2>/dev/null
docker rm mailhog-seatsurfing 2>/dev/null
docker run --rm -d -p 1025:1025 -p 8025:8025 --name mailhog-seatsurfing richarvey/mailhog

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
    echo "Reading additional settings from .env file …"
    export $(grep -v '^#' .env | xargs)
fi

# start Seatsurfing server
echo "Starting Seatsurfing …"
DEV=$DEV PUBLIC_SCHEME=$PUBLIC_SCHEME PUBLIC_PORT=$PUBLIC_PORT CRYPT_KEY=$CRYPT_KEY STATIC_UI_PATH=$STATIC_UI_PATH PLUGINS_SUB_PATH=$PLUGINS_SUB_PATH SMTP_HOST=$SMTP_HOST SMTP_PORT=$SMTP_PORT go run `ls *.go | grep -v _test.go`

# Stop mailhog and postgres
echo "Shutting down …"
docker stop mailhog-seatsurfing 2>/dev/null
docker rm mailhog-seatsurfing 2>/dev/null
docker stop postgres-seatsurfing 2>/dev/null
docker rm postgres-seatsurfing 2>/dev/null
