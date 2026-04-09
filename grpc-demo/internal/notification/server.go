package notification

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	notificationv1 "github.com/grpc-demo/proto/notification/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	notificationv1.UnimplementedNotificationServiceServer
	mu            sync.RWMutex
	subscriptions map[string]chan *notificationv1.StatusUpdate
}

func NewServer() *Server {
	return &Server{
		subscriptions: make(map[string]chan *notificationv1.StatusUpdate),
	}
}

func (s *Server) SendNotification(ctx context.Context, req *notificationv1.SendNotificationRequest) (*notificationv1.SendNotificationResponse, error) {
	log.Printf("[Notification] Sending notification to customer %s for order %s: %s",
		req.CustomerId, req.OrderId, req.Message)

	notificationID := "NOTIF-" + time.Now().Format("20060102150405")

	return &notificationv1.SendNotificationResponse{
		Success:        true,
		NotificationId: notificationID,
	}, nil
}

func (s *Server) OrderStatusStream(req *notificationv1.SubscribeRequest, stream notificationv1.NotificationService_OrderStatusStreamServer) error {
	log.Printf("[Notification] Customer %s subscribed to order status stream", req.CustomerId)

	ch := make(chan *notificationv1.StatusUpdate, 100)
	s.mu.Lock()
	s.subscriptions[req.CustomerId] = ch
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.subscriptions, req.CustomerId)
		s.mu.Unlock()
		close(ch)
	}()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case update := <-ch:
			if err := stream.Send(update); err != nil {
				return err
			}
		case <-ticker.C:
			if err := stream.Send(&notificationv1.StatusUpdate{
				OrderId:   "DEMO-ORDER",
				Status:    "PROCESSING",
				Message:   "Your order is being processed",
				Timestamp: timestamppb.Now(),
			}); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

func (s *Server) LiveOrderTracking(stream notificationv1.NotificationService_LiveOrderTrackingServer) error {
	log.Printf("[Notification] Starting live order tracking stream")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		log.Printf("[Notification] Received tracking update for order %s: %s at %s",
			req.OrderId, req.Event, req.Location)

		response := &notificationv1.TrackingUpdate{
			OrderId:   req.OrderId,
			Event:     "CONFIRMED",
			Location:  "Distribution Center",
			Timestamp: timestamppb.Now(),
		}

		if err := stream.Send(response); err != nil {
			return err
		}
	}
}

func (s *Server) EmitStatusUpdate(customerID, orderID, status, message string) {
	s.mu.RLock()
	ch, exists := s.subscriptions[customerID]
	s.mu.RUnlock()

	if exists {
		update := &notificationv1.StatusUpdate{
			OrderId:   orderID,
			Status:    status,
			Message:   message,
			Timestamp: timestamppb.Now(),
		}
		select {
		case ch <- update:
		default:
		}
	}
}
