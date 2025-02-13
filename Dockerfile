FROM golang:1.24-bookworm AS server-builder
RUN export GOBIN=$HOME/work/bin
WORKDIR /go/src/app
ADD server/ server/
ADD go.mod .
ADD go.sum .
WORKDIR /go/src/app/server
RUN go get -d -v ./...
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o main .

FROM gcr.io/distroless/static-debian12
LABEL org.opencontainers.image.source="https://github.com/seatsurfing/seatsurfing" \
      org.opencontainers.image.url="https://seatsurfing.io" \
      org.opencontainers.image.documentation="https://seatsurfing.io/docs/"
COPY --from=server-builder /go/src/app/server/main /app/
COPY server/res/ /app/res
ADD version.txt /app/
WORKDIR /app
EXPOSE 8080
USER 65532:65532
CMD ["./main"]