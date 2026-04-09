package main

import (
	"log"
	"net"

	"github.com/grpc-demo/internal/interceptor"
	"github.com/grpc-demo/internal/payment"
	paymentv1 "github.com/grpc-demo/proto/payment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.LoggingInterceptor()),
		grpc.StreamInterceptor(interceptor.LoggingStreamInterceptor()),
	)

	paymentv1.RegisterPaymentServiceServer(s, payment.NewServer())
	reflection.Register(s)

	log.Println("Payment Service running on :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
