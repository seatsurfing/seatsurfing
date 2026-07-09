FROM --platform=$BUILDPLATFORM docker.io/tonistiigi/xx AS xx

FROM --platform=$BUILDPLATFORM node:25-alpine AS ui-builder
RUN apk add --no-cache jq bash
ARG CI_VERSION
ENV NEXT_PUBLIC_PRODUCT_VERSION=$CI_VERSION
ENV NODE_ENV=production
COPY ui/package.json ui/package-lock.json /app/
WORKDIR /app
RUN --mount=type=cache,target=/root/.npm \
    npm ci
COPY ui/ /app/
RUN ./add-missing-translations.sh
RUN npm run build

FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.27rc2-bookworm AS server-builder
RUN apt-get update && apt-get install -y clang lld
COPY --from=xx / /
ARG TARGETPLATFORM
RUN xx-apt install -y libc6-dev binutils gcc
WORKDIR /go/src/app
COPY server/go.mod server/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY server/ .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 xx-go build -ldflags="-w -s" -o main && xx-verify main

FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.27rc2-bookworm AS healthcheck-builder
COPY --from=xx / /
ARG TARGETPLATFORM
RUN xx-apt install -y libc6-dev binutils gcc
WORKDIR /go/src/healthcheck
COPY healthcheck/ .
RUN --mount=type=cache,target=/root/.cache/go-build \
    xx-go build -ldflags="-w -s" -o healthcheck . && xx-verify healthcheck

FROM gcr.io/distroless/base-debian13@sha256:57c1e4c72feb5925c4763ae4f6bd2013ad3854f57eff5b60dd9acb1ce0abc66e
LABEL org.opencontainers.image.source="https://github.com/seatsurfing/seatsurfing" \
      org.opencontainers.image.url="https://seatsurfing.io" \
      org.opencontainers.image.documentation="https://seatsurfing.io/docs/"
COPY --from=server-builder /go/src/app/main /app/
COPY --from=healthcheck-builder /go/src/healthcheck/healthcheck /app/
COPY --from=ui-builder /app/build/ /app/ui
COPY server/res/ /app/res
COPY version.txt /app/
WORKDIR /app
EXPOSE 8080
USER 65532:65532
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD ["/app/healthcheck"]
CMD ["./main"]
