package api

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Shared-secret token authorization for the host<->plugin gRPC connections.
// HostAPI/SeatsurfingPlugin are plain network listeners, so anything that
// can reach the port could otherwise call them.

const tokenMetadataKey = "x-seatsurfing-token"

// tokenCredentials implements credentials.PerRPCCredentials, attaching a
// shared-secret token to every outbound RPC as metadata.
type tokenCredentials struct {
	token      string
	requireTLS bool
}

// NewTokenCredentials returns per-RPC credentials that attach token as
// metadata on every call. requireTLS should be true whenever the connection
// is NOT already secured by transport TLS at the dial-option level, so gRPC
// refuses to send the token in cleartext by accident; when the caller has
// already configured TLS transport credentials, pass false here to avoid a
// redundant requirement (grpc-go still enforces via the transport itself).
func NewTokenCredentials(token string, requireTLS bool) credentials.PerRPCCredentials {
	return tokenCredentials{token: token, requireTLS: requireTLS}
}

func (t tokenCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{tokenMetadataKey: t.token}, nil
}

func (t tokenCredentials) RequireTransportSecurity() bool {
	return t.requireTLS
}

// TokenAuthUnaryInterceptor returns a grpc.UnaryServerInterceptor that
// rejects any call not presenting the expected token in metadata.
func TokenAuthUnaryInterceptor(expectedToken string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		vals := md.Get(tokenMetadataKey)
		if len(vals) != 1 || vals[0] != expectedToken {
			return nil, status.Error(codes.Unauthenticated, "invalid or missing token")
		}
		return handler(ctx, req)
	}
}
