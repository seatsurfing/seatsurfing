FROM --platform=$BUILDPLATFORM docker.io/tonistiigi/xx AS xx

FROM --platform=$BUILDPLATFORM node:lts-alpine AS ui-builder
RUN apk add --no-cache jq bash
ARG CI_VERSION
ENV NEXT_PUBLIC_PRODUCT_VERSION=$CI_VERSION
ENV NODE_ENV=production
ADD ui /app/
WORKDIR /app
RUN ./add-missing-translations.sh
RUN npm ci
RUN npm run build

FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.25.4-bookworm AS server-builder
RUN apt-get update && apt-get install -y clang lld
COPY --from=xx / /
ARG TARGETPLATFORM
RUN xx-apt install -y libc6-dev binutils gcc libc6-dev
RUN export GOBIN=$HOME/work/bin
WORKDIR /go/src/app
ADD server/ .
WORKDIR /go/src/app
RUN go get -d -v .
RUN CGO_ENABLED=1 xx-go build -ldflags="-w -s" -o main && xx-verify main

FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.25.4-bookworm AS healthcheck-builder
COPY --from=xx / /
ARG TARGETPLATFORM
RUN xx-apt install -y libc6-dev binutils gcc libc6-dev
WORKDIR /go/src/healthcheck
ADD healthcheck/ .
RUN xx-go build -ldflags="-w -s" -o healthcheck . && xx-verify healthcheck

FROM gcr.io/distroless/base-debian12
LABEL org.opencontainers.image.source="https://github.com/seatsurfing/seatsurfing" \
      org.opencontainers.image.url="https://seatsurfing.io" \
      org.opencontainers.image.documentation="https://seatsurfing.io/docs/"
COPY --from=server-builder /go/src/app/main /app/
COPY --from=healthcheck-builder /go/src/healthcheck/healthcheck /app/
COPY --from=ui-builder /app/build/ /app/ui
COPY server/res/ /app/res
ADD version.txt /app/
WORKDIR /app
EXPOSE 8080
USER 65532:65532
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD ["/app/healthcheck"]
CMD ["./main"]
