FROM --platform=$BUILDPLATFORM docker.io/tonistiigi/xx AS xx

FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.24-bookworm AS server-builder
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

FROM gcr.io/distroless/base-debian12
LABEL org.opencontainers.image.source="https://github.com/seatsurfing/seatsurfing" \
      org.opencontainers.image.url="https://seatsurfing.io" \
      org.opencontainers.image.documentation="https://seatsurfing.io/docs/"
COPY --from=server-builder /go/src/app/main /app/
COPY server/res/ /app/res
ADD version.txt /app/
WORKDIR /app
EXPOSE 8080
USER 65532:65532
CMD ["./main"]