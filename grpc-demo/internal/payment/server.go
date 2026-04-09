package payment

import (
	"context"
	"math/rand"
	"sync"
	"time"

	paymentv1 "github.com/grpc-demo/proto/payment/v1"
)

type Server struct {
	paymentv1.UnimplementedPaymentServiceServer
	mu       sync.RWMutex
	payments map[string]*paymentv1.PaymentResponse
}

func NewServer() *Server {
	return &Server{
		payments: make(map[string]*paymentv1.PaymentResponse),
	}
}

func (s *Server) ProcessPayment(ctx context.Context, req *paymentv1.PaymentRequest) (*paymentv1.PaymentResponse, error) {
	time.Sleep(100 * time.Millisecond)

	paymentID := generatePaymentID()
	success := rand.Float32() < 0.85

	status := paymentv1.PaymentStatus_PAYMENT_STATUS_COMPLETED
	message := "Payment processed successfully"
	if !success {
		status = paymentv1.PaymentStatus_PAYMENT_STATUS_FAILED
		message = "Payment failed. Please try again."
	}

	payment := &paymentv1.PaymentResponse{
		PaymentId: paymentID,
		OrderId:   req.OrderId,
		Status:    status,
		Amount:    req.Amount,
		Message:   message,
	}

	s.mu.Lock()
	s.payments[paymentID] = payment
	s.mu.Unlock()

	return payment, nil
}

func (s *Server) RefundPayment(ctx context.Context, req *paymentv1.RefundRequest) (*paymentv1.RefundResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	payment, exists := s.payments[req.PaymentId]
	if !exists {
		return &paymentv1.RefundResponse{
			Success: false,
			Message: "Payment not found",
		}, nil
	}

	if payment.Status != paymentv1.PaymentStatus_PAYMENT_STATUS_COMPLETED {
		return &paymentv1.RefundResponse{
			Success: false,
			Message: "Payment not eligible for refund",
		}, nil
	}

	payment.Status = paymentv1.PaymentStatus_PAYMENT_STATUS_REFUNDED

	return &paymentv1.RefundResponse{
		Success: true,
		Message: "Refund processed successfully",
	}, nil
}

func generatePaymentID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
