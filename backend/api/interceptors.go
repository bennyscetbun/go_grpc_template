package api

import (
	"context"
	"strings"
	"time"

	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/apihelpers"
	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/grpcerrors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ztrue/tracerr"
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
	prefix := "/xxxyourappyyy.apiproto.Api/"

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

func (g *GRPCServer) extractClaimsFromToken(_ context.Context, token string) (jwt.MapClaims, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, tracerr.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return g.TokenSecret, nil
	})
	if err != nil {
		return nil, grpcerrors.ErrorInvalidToken()
	}
	if claims, ok := t.Claims.(jwt.MapClaims); ok && t.Valid {
		return claims, nil
	}
	return nil, grpcerrors.ErrorInvalidToken()
}

func (g *GRPCServer) validateToken(ctx context.Context, token string) (string, error) {
	claims, err := g.extractClaimsFromToken(ctx, token)
	if err != nil {
		return "", err
	}
	id, ok := claims[tokenUserIDKey].(string)
	if !ok {
		return "", grpcerrors.ErrorInvalidToken()
	}

	return id, nil
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

	userID, err := g.validateToken(ctx, authTokens[0]) // Assuming a single token is sent in the header.
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, apihelpers.UserIdContextKey, userID)
	return handler(ctx, req)
}

func (g *GRPCServer) TimeOutInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx, cancelCtx := context.WithTimeout(ctx, grpcTimeout)
	defer cancelCtx()
	return handler(ctx, req)
}
