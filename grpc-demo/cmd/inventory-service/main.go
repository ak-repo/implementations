package main

import (
	"log"
	"net"

	"github.com/grpc-demo/internal/interceptor"
	"github.com/grpc-demo/internal/inventory"
	inventoryv1 "github.com/grpc-demo/proto/inventory/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.LoggingInterceptor()),
		grpc.StreamInterceptor(interceptor.LoggingStreamInterceptor()),
	)

	inventoryv1.RegisterInventoryServiceServer(s, inventory.NewServer())
	reflection.Register(s)

	log.Println("Inventory Service running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
