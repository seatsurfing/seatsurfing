# Seatsurfing Backend

[![](https://img.shields.io/github/v/release/seatsurfing/seatsurfing)](https://github.com/seatsurfing/seatsurfing/releases)
[![](https://img.shields.io/github/release-date/seatsurfing/seatsurfing)](https://github.com/seatsurfing/seatsurfing/releases)
[![](https://img.shields.io/github/actions/workflow/status/seatsurfing/seatsurfing/release.yml?branch=main)](https://github.com/seatsurfing/seatsurfing/actions)
[![](https://img.shields.io/github/license/seatsurfing/seatsurfing)](https://github.com/seatsurfing/seatsurfing/blob/main/LICENSE)

Seatsurfing is a software which enables your organisation's employees to book seats, desks and rooms.

This repository contains the Backend, which consists of:
* The Server (REST API Backend) written in Go
* User Self-Service Booking Web Interface ("Booking UI"), built as a Progressive Web Application (PWA) which can be installed on mobile devices
* Admin Web Interface ("Admin UI")
* Common TypeScript files for the two TypeScript/React web frontends

**[Visit project's website for more information.](https://seatsurfing.app)**

## Screenshots

### Web Admin UI
![Seatsurfing Web Admin UI](https://raw.githubusercontent.com/seatsurfing/seatsurfing/main/.github/admin-ui.png)

### Web Booking UI
![Seatsurfing Web Booking UI](https://raw.githubusercontent.com/seatsurfing/seatsurfing/main/.github/booking-ui.png)

## Quick reference
* **Maintained by:** [Seatsurfing.app](https://seatsurfing.app/)
* **Where to get help:** [Documentation](https://seatsurfing.app/docs/)
* **Docker architectures:** [amd64, arm64](https://github.com/seatsurfing?tab=packages&repo_name=seatsurfing)
* **License:** [GPL 3.0](https://github.com/seatsurfing/seatsurfing/blob/main/LICENSE)

## How to use the Docker image
### Start using Docker Compose
```
version: '3.7'

services:
  server:
    image: ghcr.io/seatsurfing/backend
    restart: always
    networks:
      sql:
      http:
    ports:
      - 8080:8080
    environment:
      POSTGRES_URL: 'postgres://seatsurfing:DB_PASSWORD@db/seatsurfing?sslmode=disable'
      JWT_SIGNING_KEY: 'some_random_string'
      BOOKING_UI_BACKEND: 'booking-ui:3001'
      ADMIN_UI_BACKEND: 'admin-ui:3000'
      PUBLIC_URL: 'https://seatsurfing.your-domain.com'
      FRONTEND_URL: 'https://seatsurfing.your-domain.com'
      CRYPT_KEY: 'some-random-32-bytes-long-string'
  booking-ui:
    image: ghcr.io/seatsurfing/booking-ui
    restart: always
    networks:
      http:
    environment:
      FRONTEND_URL: 'https://seatsurfing.your-domain.com'
  admin-ui:
    image: ghcr.io/seatsurfing/admin-ui
    restart: always
    networks:
      http:
    environment:
      FRONTEND_URL: 'https://seatsurfing.your-domain.com'
  db:
    image: postgres:16
    restart: always
    networks:
      sql:
    volumes:
      - db:/var/lib/postgresql/data
    environment:
      POSTGRES_PASSWORD: DB_PASSWORD
      POSTGRES_USER: seatsurfing
      POSTGRES_DB: seatsurfing

volumes:
  db:

networks:
  sql:
  http:
```

This starts...
* a PostgreSQL database with data stored on Docker volume "db"
* a Seatsurfing Backend instance with port 8080 exposed.
* a Seatsurfing Booking UI instance which is accessible through the Backend instance at: :8080/ui/
* a Seatsurfing Admin UI instance which is accessible through the Backend instance at: :8080/admin/

### Running on Kubernetes
Please refer to our [Kubernetes documentation](https://seatsurfing.app/docs/kubernetes/).

## Environment variables
Please check out the [documentation](https://seatsurfing.app/docs/config) for information on available environment variables and further guidance.

**Hint**: When running in an IPV6-only Docker/Podman environment with multiple network interfaces bound to the Frontend containers, setting the ```LISTEN_ADDR``` environment variable can be necessary as NextJS binds to only one network interface by default. Set it to ```::``` to bind to any address.