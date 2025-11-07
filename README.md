# Seatsurfing

[![](https://img.shields.io/github/v/release/seatsurfing/seatsurfing)](https://github.com/seatsurfing/seatsurfing/releases)
[![](https://img.shields.io/github/release-date/seatsurfing/seatsurfing)](https://github.com/seatsurfing/seatsurfing/releases)
[![](https://img.shields.io/github/actions/workflow/status/seatsurfing/seatsurfing/release.yml?branch=main)](https://github.com/seatsurfing/seatsurfing/actions)
[![](https://img.shields.io/github/license/seatsurfing/seatsurfing)](https://github.com/seatsurfing/seatsurfing/blob/main/LICENSE)

## 🚀 Seatsurfing SaaS available!

We offer [Seatsurfing](https://seatsurfing.io/) as a fully-hosted Software-as-a-Service (SaaS) at. [Start for free now](https://seatsurfing.io/sign-up)!

- **No installation required** - Get started immediately
- **Microsoft Teams integration** - See [Microsoft AppSource marketplace](https://appsource.microsoft.com/product/office/WA200008773)
- **Get it free** - Free for up to 10 users
- **Automatic updates** - Always enjoy the latest features
- **Managed infrastructure** - Servers in Germany (EU)

## 📖 Introduction

Seatsurfing is a software which enables your organization's employees to book seats, desks and rooms.

This repository contains the Backend, which consists of:

- The Server (REST API Backend) written in Go
- User Self-Service Booking Web Interface ("Booking UI"), built as a Progressive Web Application (PWA) which can be installed on mobile devices
- Admin Web Interface ("Admin UI")
- Common TypeScript files for the two TypeScript/React web frontends

**[Visit project's website for more information.](https://seatsurfing.io)**

## 📷 Screenshots

### Web Admin UI

![Seatsurfing Web Admin UI](https://raw.githubusercontent.com/seatsurfing/seatsurfing/main/.github/admin-ui.png)

### Web Booking UI

![Seatsurfing Web Booking UI](https://raw.githubusercontent.com/seatsurfing/seatsurfing/main/.github/booking-ui.png)

## 🗸 Quick reference

- **Maintained by:** [seatsurfing.io](https://seatsurfing.io/)
- **Where to get help:** [Documentation](https://seatsurfing.io/docs/)
- **Docker architectures:** [amd64, arm64](https://github.com/seatsurfing/seatsurfing/pkgs/container/backend)
- **License:** [GPL 3.0](https://github.com/seatsurfing/seatsurfing/blob/main/LICENSE)

## 🐋 How to use the Docker image

### Start using Docker Compose

```
services:
  server:
    image: ghcr.io/seatsurfing/backend
    restart: always
    networks:
      sql:
    ports:
      - 8080:8080
    environment:
      POSTGRES_URL: 'postgres://seatsurfing:DB_PASSWORD@db/seatsurfing?sslmode=disable'
      CRYPT_KEY: 'some-random-32-bytes-long-string'
  db:
    image: postgres:17
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
```

This starts...

- a PostgreSQL database with data stored on Docker volume "db"
- a Seatsurfing instance with port 8080 exposed

The Seatsurfing Booking UI is accessible at :8080/ui/search/ and the Seatsurfing Admin UI instance at :8080/ui/admin/.

To login, use the default admin login (user `admin@seatsurfing.local` and password `12345678`) or set the [environment variables](https://seatsurfing.io/docs/self-hosted/config) `INIT_ORG_USER` and `INIT_ORG_PASS` to customize the admin login.

### Running on Kubernetes

Please refer to our [Kubernetes documentation](https://seatsurfing.io/docs/self-hosted/kubernetes/).

## ⚙️ Environment variables

Please check out the [documentation](https://seatsurfing.io/docs/self-hosted/config) for information on available environment variables and further guidance.

**Hint**: When running in an IPV6-only Docker/Podman environment with multiple network interfaces bound to the Frontend containers, setting the `LISTEN_ADDR` environment variable can be necessary as NextJS binds to only one network interface by default. Set it to `::` to bind to any address.
