package interceptor

import (
	"context"
	"log"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	APIKeyHeader  = "x-api-key"
	DefaultAPIKey = "grpc-demo-api-key-2024"
)

func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		log.Printf("[gRPC] Method: %s, Request: %+v", info.FullMethod, req)

		resp, err := handler(ctx, req)

		if err != nil {
			log.Printf("[gRPC] Method: %s, Error: %v", info.FullMethod, err)
		} else {
			log.Printf("[gRPC] Method: %s, Response: %+v", info.FullMethod, resp)
		}

		return resp, err
	}
}

func LoggingStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		log.Printf("[gRPC] Stream Method: %s", info.FullMethod)
		return handler(srv, ss)
	}
}

func AuthInterceptor(validAPIKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		apiKeys := md.Get(APIKeyHeader)
		if len(apiKeys) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing API key")
		}

		if apiKeys[0] != validAPIKey {
			return nil, status.Error(codes.Unauthenticated, "invalid API key")
		}

		return handler(ctx, req)
	}
}

func AuthStreamInterceptor(validAPIKey string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}

		apiKeys := md.Get(APIKeyHeader)
		if len(apiKeys) == 0 {
			return status.Error(codes.Unauthenticated, "missing API key")
		}

		if apiKeys[0] != validAPIKey {
			return status.Error(codes.Unauthenticated, "invalid API key")
		}

		return handler(srv, ss)
	}
}

func APIKeyValid(apiKey string) bool {
	return apiKey == DefaultAPIKey || strings.HasPrefix(apiKey, "grpc-demo-")
}
