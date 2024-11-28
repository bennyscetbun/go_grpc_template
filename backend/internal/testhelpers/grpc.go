package testhelpers

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader = "authorization"
)

func AddTokenToContext(ctx context.Context, token string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{authorizationHeader: token})
	} else {
		md.Set(authorizationHeader, token)
	}
	return metadata.NewOutgoingContext(ctx, md)
}
