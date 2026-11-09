package app

import (
	"context"
	"net"
	"os"
	"sync"

	aiproviders "riders.agency/pulse/api/internal/modules/ai/providers"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	aiagentv1 "riders.agency/pulse/api/internal/gen/ai_agent/v1"
	aimodel "riders.agency/pulse/api/internal/modules/ai/modelgateway"
	airepo "riders.agency/pulse/api/internal/modules/ai/repository"
	aihttp "riders.agency/pulse/api/internal/modules/ai/transport/http"
)

func (a *Application) StartToolGatewayGRPC(ctx context.Context) (func(), error) {
	if a == nil || a.ToolGateway == nil || a.Config.ToolGRPCAddr == "" {
		return func() {}, nil
	}
	listener, err := net.Listen("tcp", a.Config.ToolGRPCAddr)
	if err != nil {
		return nil, err
	}
	token := os.Getenv("AI_AGENT_SERVICE_TOKEN")
	server := grpc.NewServer(
		grpc.UnaryInterceptor(toolGatewayTokenInterceptor(token)),
		grpc.StreamInterceptor(toolGatewayStreamTokenInterceptor(token)),
	)
	aiagentv1.RegisterPulseToolServiceServer(server, aihttp.NewPulseToolServiceServer(a.ToolGateway))
	var once sync.Once
	stop := func() {
		once.Do(func() {
			server.GracefulStop()
			_ = listener.Close()
		})
	}
	go func() {
		<-ctx.Done()
		stop()
	}()
	go func() {
		if err := server.Serve(listener); err != nil {
			_ = listener.Close()
		}
	}()
	return stop, nil
}

func (a *Application) StartModelGatewayGRPC(ctx context.Context) (func(), error) {
	if a == nil || a.Config.ModelGRPCAddr == "" {
		return func() {}, nil
	}
	modelGateway, err := aimodel.NewServerFromEnv(
		a.Config.Environment,
		a.Deps.HTTPClient,
		aimodel.WithConfigResolver(&airepo.AI{
			P:           a.Deps.DB,
			MasterHEX:   a.Deps.MasterKey,
			ModelClient: &aiproviders.ModelClient{HTTP: a.Deps.HTTPClient},
		}),
	)
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", a.Config.ModelGRPCAddr)
	if err != nil {
		return nil, err
	}
	token := os.Getenv("AI_AGENT_SERVICE_TOKEN")
	server := grpc.NewServer(
		grpc.UnaryInterceptor(toolGatewayTokenInterceptor(token)),
		grpc.StreamInterceptor(toolGatewayStreamTokenInterceptor(token)),
	)
	aiagentv1.RegisterModelGatewayServiceServer(server, modelGateway)
	var once sync.Once
	stop := func() {
		once.Do(func() {
			server.GracefulStop()
			_ = listener.Close()
		})
	}
	go func() {
		<-ctx.Done()
		stop()
	}()
	go func() {
		if err := server.Serve(listener); err != nil {
			_ = listener.Close()
		}
	}()
	return stop, nil
}

func toolGatewayTokenInterceptor(token string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if token == "" {
			return handler(ctx, req)
		}
		md, _ := metadata.FromIncomingContext(ctx)
		for _, value := range md.Get("x-pulse-service-token") {
			if value == token {
				return handler(ctx, req)
			}
		}
		return nil, status.Error(codes.Unauthenticated, "invalid service token")
	}
}

func toolGatewayStreamTokenInterceptor(token string) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if token == "" {
			return handler(srv, stream)
		}
		md, _ := metadata.FromIncomingContext(stream.Context())
		for _, value := range md.Get("x-pulse-service-token") {
			if value == token {
				return handler(srv, stream)
			}
		}
		return status.Error(codes.Unauthenticated, "invalid service token")
	}
}
