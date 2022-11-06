package grpc

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Authorizer struct {
	jwtManager *JWTManager
}

func NewAuthorizer(jwtSecret string) *Authorizer {
	return &Authorizer{jwtManager: NewJWTManager(jwtSecret)}
}

func (a *Authorizer) StreamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := a.authorize(stream.Context(), info.FullMethod); err != nil {
		return err
	}
	return handler(srv, stream)
}

func (a *Authorizer) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if err := a.authorize(ctx, info.FullMethod); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func (a *Authorizer) authorize(ctx context.Context, fullMethod string) error {
	log.Println("grpc_fullMethod -->", fullMethod)

	// bypass auth for token endpoint
	if fullMethod == "/proto.WoodpeckerAuth/Auth" {
		return nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["token"]
	if len(values) == 0 {
		return status.Errorf(codes.Unauthenticated, "token is not provided")
	}

	accessToken := values[0]
	claims, err := a.jwtManager.Verify(accessToken)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}

	// add agent_id to context
	ctx = context.WithValue(ctx, "agent_id", claims.AgentID) // TODO: improve this
	return nil
}
