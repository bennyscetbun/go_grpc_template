package api

import (
	"context"
	"strings"
	"time"

	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/apihelpers"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/grpcerrors"
	"github.com/bennyscetbun/xxx_your_app_xxx/backend/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	grpcTimeout = time.Second * 15
)

func extractMethodName(fullMethod string) string {
	// Define the prefix to remove
	prefix := "/xxx_your_app_xxx.apiproto.Api/"

	// Use TrimPrefix to remove the prefix from the full method name
	method := strings.TrimPrefix(fullMethod, prefix)

	return method
}

func methodRequiresAuthentication(fullMethod string) bool {
	m := extractMethodName(fullMethod)
	m = strings.ToLower(m)

	// Define a list of methods that require authentication.

	NonAuthRequiredMethods := []string{
		"login",
		"signup",
		"verifyemail",
	}

	// Check if the requested method is in the list.
	for _, method := range NonAuthRequiredMethods {
		if m == method {
			return false
		}
	}

	return true
}

func (g *GRPCServer) AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if !methodRequiresAuthentication(info.FullMethod) {
		return handler(ctx, req)
	}
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "metadata not found")
	}

	authTokens := md[authorizationHeader]
	if len(authTokens) == 0 {
		return nil, grpcerrors.ErrorUnauthenticated()
	}

	token := authTokens[0] // Assuming a single token is sent in the header.
	ut := g.DBQueries.UserToken
	dbToken, err := ut.WithContext(ctx).Where(ut.ID.Eq(token)).First()
	if err != nil {
		logger.Errorln(err)
		return nil, err
	}
	if dbToken.ExpiredAt.Before(time.Now()) {
		return nil, grpcerrors.ErrorUnauthenticated()
	}

	ctx = context.WithValue(ctx, apihelpers.UserIdContextKey, dbToken.UserID)
	return handler(ctx, req)
}

func (g *GRPCServer) TimeOutInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx, cancelCtx := context.WithTimeout(ctx, grpcTimeout)
	defer cancelCtx()
	return handler(ctx, req)
}
