package order

import (
	"context"
	"io"
	"sync"
	"time"

	inventoryv1 "github.com/grpc-demo/proto/inventory/v1"
	notificationv1 "github.com/grpc-demo/proto/notification/v1"
	orderv1 "github.com/grpc-demo/proto/order/v1"
	paymentv1 "github.com/grpc-demo/proto/payment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	orderv1.UnimplementedOrderServiceServer
	mu                 sync.RWMutex
	orders             map[string]*orderv1.Order
	inventoryClient    inventoryv1.InventoryServiceClient
	paymentClient      paymentv1.PaymentServiceClient
	notificationClient notificationv1.NotificationServiceClient
	inventoryConn      *grpc.ClientConn
	paymentConn        *grpc.ClientConn
	notificationConn   *grpc.ClientConn
}

func NewServer(inventoryAddr, paymentAddr, notificationAddr string) (*Server, error) {
	s := &Server{
		orders: make(map[string]*orderv1.Order),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	s.inventoryConn, err = grpc.DialContext(ctx, inventoryAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	s.inventoryClient = inventoryv1.NewInventoryServiceClient(s.inventoryConn)

	s.paymentConn, err = grpc.DialContext(ctx, paymentAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	s.paymentClient = paymentv1.NewPaymentServiceClient(s.paymentConn)

	s.notificationConn, err = grpc.DialContext(ctx, notificationAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	s.notificationClient = notificationv1.NewNotificationServiceClient(s.notificationConn)

	return s, nil
}

func (s *Server) Close() {
	if s.inventoryConn != nil {
		s.inventoryConn.Close()
	}
	if s.paymentConn != nil {
		s.paymentConn.Close()
	}
	if s.notificationConn != nil {
		s.notificationConn.Close()
	}
}

func (s *Server) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "order must contain at least one item")
	}

	orderID := generateOrderID()
	order := &orderv1.Order{
		Id:         orderID,
		CustomerId: req.CustomerId,
		Items:      req.Items,
		Status:     orderv1.OrderStatus_ORDER_STATUS_PENDING,
		CreatedAt:  timestamppb.Now(),
		UpdatedAt:  timestamppb.Now(),
	}

	for _, item := range req.Items {
		order.TotalAmount += item.Price * float64(item.Quantity)

		stockResp, err := s.inventoryClient.CheckStock(ctx, &inventoryv1.CheckStockRequest{
			ProductId:         item.ProductId,
			RequestedQuantity: item.Quantity,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to check stock: %v", err)
		}

		if !stockResp.Available {
			order.Status = orderv1.OrderStatus_ORDER_STATUS_FAILED
			s.saveOrder(order)
			return &orderv1.CreateOrderResponse{
				Order:   order,
				Message: "Insufficient stock for product " + item.ProductId,
			}, nil
		}
	}

	order.Status = orderv1.OrderStatus_ORDER_STATUS_CONFIRMED
	s.saveOrder(order)

	paymentResp, err := s.paymentClient.ProcessPayment(ctx, &paymentv1.PaymentRequest{
		OrderId:    orderID,
		CustomerId: req.CustomerId,
		Amount:     order.TotalAmount,
		Method:     paymentv1.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to process payment: %v", err)
	}

	if paymentResp.Status != paymentv1.PaymentStatus_PAYMENT_STATUS_COMPLETED {
		order.Status = orderv1.OrderStatus_ORDER_STATUS_CANCELLED
		s.saveOrder(order)
		return &orderv1.CreateOrderResponse{
			Order:   order,
			Message: "Payment failed: " + paymentResp.Message,
		}, nil
	}

	order.Status = orderv1.OrderStatus_ORDER_STATUS_PROCESSING
	order.UpdatedAt = timestamppb.Now()
	s.saveOrder(order)

	for _, item := range req.Items {
		s.inventoryClient.UpdateStock(ctx, &inventoryv1.UpdateStockRequest{
			ProductId:      item.ProductId,
			QuantityChange: -item.Quantity,
		})
	}

	_, _ = s.notificationClient.SendNotification(ctx, &notificationv1.SendNotificationRequest{
		OrderId:    orderID,
		CustomerId: req.CustomerId,
		Message:    "Your order " + orderID + " has been confirmed!",
		Type:       notificationv1.NotificationType_NOTIFICATION_TYPE_EMAIL,
	})

	return &orderv1.CreateOrderResponse{
		Order:   order,
		Message: "Order created successfully",
	}, nil
}

func (s *Server) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	s.mu.RLock()
	order, exists := s.orders[req.OrderId]
	s.mu.RUnlock()

	if !exists {
		return nil, status.Error(codes.NotFound, "order not found")
	}

	return &orderv1.GetOrderResponse{Order: order}, nil
}

func (s *Server) BulkCreateOrders(stream orderv1.OrderService_BulkCreateOrdersServer) error {
	var orders []*orderv1.Order
	successCount := 0
	failureCount := 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&orderv1.BulkCreateOrdersResponse{
				Orders:       orders,
				SuccessCount: int32(successCount),
				FailureCount: int32(failureCount),
			})
		}
		if err != nil {
			return err
		}

		resp, err := s.CreateOrder(stream.Context(), req)
		if err != nil {
			failureCount++
			continue
		}

		orders = append(orders, resp.Order)
		if resp.Order.Status == orderv1.OrderStatus_ORDER_STATUS_PROCESSING {
			successCount++
		} else {
			failureCount++
		}
	}
}

func (s *Server) saveOrder(order *orderv1.Order) {
	s.mu.Lock()
	s.orders[order.Id] = order
	s.mu.Unlock()
}

func generateOrderID() string {
	return "ORD-" + time.Now().Format("20060102150405") + "-" + randomString(4)
}

func randomString(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}
