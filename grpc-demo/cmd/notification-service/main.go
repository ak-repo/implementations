package main

import (
	"log"
	"net"

	"github.com/grpc-demo/internal/interceptor"
	"github.com/grpc-demo/internal/notification"
	notificationv1 "github.com/grpc-demo/proto/notification/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.LoggingInterceptor()),
		grpc.StreamInterceptor(interceptor.LoggingStreamInterceptor()),
	)

	notificationv1.RegisterNotificationServiceServer(s, notification.NewServer())
	reflection.Register(s)

	log.Println("Notification Service running on :50053")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
