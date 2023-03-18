package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type StreamContextWrapper interface {
	grpc.ServerStream
	SetContext(context.Context)
}

type wrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrapper) Context() context.Context {
	return w.ctx
}

func (w *wrapper) SetContext(ctx context.Context) {
	w.ctx = ctx
}

func newStreamContextWrapper(inner grpc.ServerStream) StreamContextWrapper {
	ctx := inner.Context()
	return &wrapper{
		inner,
		ctx,
	}
}

type Authorizer struct {
	jwtManager *JWTManager
}

func NewAuthorizer(jwtManager *JWTManager) *Authorizer {
	return &Authorizer{jwtManager: jwtManager}
}

func (a *Authorizer) StreamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	_stream := newStreamContextWrapper(stream)

	newCtx, err := a.authorize(stream.Context(), info.FullMethod)
	if err != nil {
		return err
	}

	_stream.SetContext(newCtx)

	return handler(srv, _stream)
}

func (a *Authorizer) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	newCtx, err := a.authorize(ctx, info.FullMethod)
	if err != nil {
		return nil, err
	}
	return handler(newCtx, req)
}

func (a *Authorizer) authorize(ctx context.Context, fullMethod string) (context.Context, error) {
	// bypass auth for token endpoint
	if fullMethod == "/proto.WoodpeckerAuth/Auth" {
		return ctx, nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["token"]
	if len(values) == 0 {
		return ctx, status.Errorf(codes.Unauthenticated, "token is not provided")
	}

	accessToken := values[0]
	claims, err := a.jwtManager.Verify(accessToken)
	if err != nil {
		return ctx, status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}

	md.Append("agent_id", fmt.Sprintf("%d", claims.AgentID))

	return metadata.NewIncomingContext(ctx, md), nil
}
