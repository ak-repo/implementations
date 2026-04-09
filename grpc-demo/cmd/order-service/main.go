package main

import (
	"log"
	"net"

	"github.com/grpc-demo/internal/interceptor"
	"github.com/grpc-demo/internal/order"
	orderv1 "github.com/grpc-demo/proto/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":50050")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv, err := order.NewServer(
		"localhost:50051",
		"localhost:50052",
		"localhost:50053",
	)
	if err != nil {
		log.Fatalf("failed to create order server: %v", err)
	}
	defer srv.Close()

	s := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.LoggingInterceptor()),
		grpc.StreamInterceptor(interceptor.LoggingStreamInterceptor()),
	)

	orderv1.RegisterOrderServiceServer(s, srv)
	reflection.Register(s)

	log.Println("Order Service running on :50050")
	log.Println("Connected to:")
	log.Println("  - Inventory Service: localhost:50051")
	log.Println("  - Payment Service: localhost:50052")
	log.Println("  - Notification Service: localhost:50053")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
