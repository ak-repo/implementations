package main

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/grpc-demo/internal/interceptor"
	notificationv1 "github.com/grpc-demo/proto/notification/v1"
	orderv1 "github.com/grpc-demo/proto/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	log.SetFlags(0)
	log.Println("=== gRPC Demo Client ===")

	conn, err := grpc.Dial("localhost:50050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	md := metadata.Pairs(interceptor.APIKeyHeader, interceptor.DefaultAPIKey)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	client := orderv1.NewOrderServiceClient(conn)

	log.Println("\n--- Test 1: Create Order (Unary RPC) ---")
	testCreateOrder(ctx, client)

	log.Println("\n--- Test 2: Bulk Create Orders (Client Streaming) ---")
	testBulkCreateOrders(ctx, client)

	log.Println("\n--- Test 3: Order Status Stream (Server Streaming) ---")
	testOrderStatusStream(ctx)

	log.Println("\n--- Test 4: Live Order Tracking (Bidirectional Streaming) ---")
	testLiveOrderTracking(ctx)

	log.Println("\n=== All tests completed ===")
}

func testCreateOrder(ctx context.Context, client orderv1.OrderServiceClient) {
	resp, err := client.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		CustomerId: "CUST-001",
		Items: []*orderv1.OrderItem{
			{ProductId: "PROD-001", Name: "Laptop", Quantity: 1, Price: 999.99},
			{ProductId: "PROD-002", Name: "Mouse", Quantity: 2, Price: 29.99},
		},
	})
	if err != nil {
		log.Printf("CreateOrder error: %v", err)
		return
	}
	log.Printf("Order created: ID=%s, Status=%s, Total=$%.2f", resp.Order.Id, resp.Order.Status, resp.Order.TotalAmount)
	log.Printf("Message: %s", resp.Message)
}

func testBulkCreateOrders(ctx context.Context, client orderv1.OrderServiceClient) {
	stream, err := client.BulkCreateOrders(ctx)
	if err != nil {
		log.Printf("BulkCreateOrders error: %v", err)
		return
	}

	orders := []*orderv1.CreateOrderRequest{
		{
			CustomerId: "CUST-002",
			Items: []*orderv1.OrderItem{
				{ProductId: "PROD-003", Name: "Keyboard", Quantity: 1, Price: 79.99},
			},
		},
		{
			CustomerId: "CUST-003",
			Items: []*orderv1.OrderItem{
				{ProductId: "PROD-004", Name: "Monitor", Quantity: 1, Price: 299.99},
			},
		},
		{
			CustomerId: "CUST-004",
			Items: []*orderv1.OrderItem{
				{ProductId: "PROD-005", Name: "Headphones", Quantity: 1, Price: 149.99},
			},
		},
	}

	for _, req := range orders {
		if err := stream.Send(req); err != nil {
			log.Printf("Send error: %v", err)
			return
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Printf("CloseAndRecv error: %v", err)
		return
	}

	log.Printf("Bulk orders result: Success=%d, Failed=%d", resp.SuccessCount, resp.FailureCount)
	for _, order := range resp.Orders {
		log.Printf("  - Order: %s, Status: %s", order.Id, order.Status)
	}
}

func testOrderStatusStream(ctx context.Context) {
	conn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to notification service: %v", err)
		return
	}
	defer conn.Close()

	md := metadata.Pairs(interceptor.APIKeyHeader, interceptor.DefaultAPIKey)
	streamCtx := metadata.NewOutgoingContext(ctx, md)

	client := notificationv1.NewNotificationServiceClient(conn)
	stream, err := client.OrderStatusStream(streamCtx, &notificationv1.SubscribeRequest{
		CustomerId: "CUST-001",
	})
	if err != nil {
		log.Printf("OrderStatusStream error: %v", err)
		return
	}

	log.Println("Listening for status updates (5 seconds)...")
	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-timeout:
			return
		default:
			update, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Printf("Stream error: %v", err)
				return
			}
			log.Printf("Status Update: Order=%s, Status=%s, Message=%s",
				update.OrderId, update.Status, update.Message)
		}
	}
}

func testLiveOrderTracking(ctx context.Context) {
	conn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to notification service: %v", err)
		return
	}
	defer conn.Close()

	md := metadata.Pairs(interceptor.APIKeyHeader, interceptor.DefaultAPIKey)
	streamCtx := metadata.NewOutgoingContext(ctx, md)

	client := notificationv1.NewNotificationServiceClient(conn)
	stream, err := client.LiveOrderTracking(streamCtx)
	if err != nil {
		log.Printf("LiveOrderTracking error: %v", err)
		return
	}

	updates := []*notificationv1.TrackingUpdate{
		{OrderId: "ORD-001", Event: "ORDER_PLACED", Location: "Online", Timestamp: timestamppb.Now()},
		{OrderId: "ORD-001", Event: "PAYMENT_CONFIRMED", Location: "Payment Gateway", Timestamp: timestamppb.Now()},
		{OrderId: "ORD-001", Event: "PREPARING", Location: "Warehouse", Timestamp: timestamppb.Now()},
		{OrderId: "ORD-001", Event: "SHIPPED", Location: "Shipping Center", Timestamp: timestamppb.Now()},
		{OrderId: "ORD-001", Event: "OUT_FOR_DELIVERY", Location: "Local Hub", Timestamp: timestamppb.Now()},
	}

	log.Println("Sending live tracking updates...")
	for _, update := range updates {
		if err := stream.Send(update); err != nil {
			log.Printf("Send error: %v", err)
			return
		}

		resp, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Printf("Recv error: %v", err)
			return
		}
		log.Printf("Tracking Response: Order=%s, Event=%s, Location=%s",
			resp.OrderId, resp.Event, resp.Location)
		time.Sleep(500 * time.Millisecond)
	}
}
