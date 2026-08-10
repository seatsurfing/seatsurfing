#!/usr/bin/env bash
# Regenerates the Go stubs for the plugin<->host gRPC contract.
#
# Requires: protoc, protoc-gen-go, protoc-gen-go-grpc on PATH, matching the
# google.golang.org/protobuf and google.golang.org/grpc versions in go.mod.
#
#   go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
#   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
#
# Generated *.pb.go files are committed to the repo - this script only needs
# to be re-run when a .proto file changes, not as part of the normal build.
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

API_DIR=".."

mkdir -p "${API_DIR}/commonpb" "${API_DIR}/pluginpb" "${API_DIR}/hostapipb"

protoc --proto_path=. \
  --go_out="${API_DIR}/commonpb" --go_opt=paths=source_relative \
  common.proto

protoc --proto_path=. \
  --go_out="${API_DIR}/pluginpb" --go_opt=paths=source_relative \
  --go-grpc_out="${API_DIR}/pluginpb" --go-grpc_opt=paths=source_relative \
  plugin.proto

protoc --proto_path=. \
  --go_out="${API_DIR}/hostapipb" --go_opt=paths=source_relative \
  --go-grpc_out="${API_DIR}/hostapipb" --go-grpc_opt=paths=source_relative \
  hostapi.proto
