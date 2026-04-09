package inventory

import (
	"context"
	"sync"

	inventoryv1 "github.com/grpc-demo/proto/inventory/v1"
)

type Server struct {
	inventoryv1.UnimplementedInventoryServiceServer
	mu    sync.RWMutex
	stock map[string]*inventoryv1.StockItem
}

func NewServer() *Server {
	s := &Server{
		stock: make(map[string]*inventoryv1.StockItem),
	}
	s.seedData()
	return s
}

func (s *Server) seedData() {
	products := []struct {
		id       string
		name     string
		quantity int32
	}{
		{"PROD-001", "Laptop", 50},
		{"PROD-002", "Mouse", 200},
		{"PROD-003", "Keyboard", 150},
		{"PROD-004", "Monitor", 75},
		{"PROD-005", "Headphones", 100},
	}

	for _, p := range products {
		s.stock[p.id] = &inventoryv1.StockItem{
			ProductId:         p.id,
			Name:              p.name,
			AvailableQuantity: p.quantity,
			InStock:           p.quantity > 0,
		}
	}
}

func (s *Server) CheckStock(ctx context.Context, req *inventoryv1.CheckStockRequest) (*inventoryv1.CheckStockResponse, error) {
	s.mu.RLock()
	item, exists := s.stock[req.ProductId]
	s.mu.RUnlock()

	if !exists {
		return &inventoryv1.CheckStockResponse{
			Available:         false,
			AvailableQuantity: 0,
			Message:           "Product not found",
		}, nil
	}

	available := item.AvailableQuantity >= req.RequestedQuantity
	message := "In stock"
	if !available {
		message = "Insufficient stock"
	}

	return &inventoryv1.CheckStockResponse{
		Available:         available,
		AvailableQuantity: item.AvailableQuantity,
		Message:           message,
	}, nil
}

func (s *Server) UpdateStock(ctx context.Context, req *inventoryv1.UpdateStockRequest) (*inventoryv1.UpdateStockResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.stock[req.ProductId]
	if !exists {
		return &inventoryv1.UpdateStockResponse{
			Success:     false,
			NewQuantity: 0,
			Message:     "Product not found",
		}, nil
	}

	newQty := item.AvailableQuantity + req.QuantityChange
	if newQty < 0 {
		return &inventoryv1.UpdateStockResponse{
			Success:     false,
			NewQuantity: item.AvailableQuantity,
			Message:     "Insufficient stock",
		}, nil
	}

	item.AvailableQuantity = newQty
	item.InStock = newQty > 0

	return &inventoryv1.UpdateStockResponse{
		Success:     true,
		NewQuantity: newQty,
		Message:     "Stock updated successfully",
	}, nil
}

func (s *Server) GetStock(ctx context.Context, req *inventoryv1.GetStockRequest) (*inventoryv1.GetStockResponse, error) {
	s.mu.RLock()
	item, exists := s.stock[req.ProductId]
	s.mu.RUnlock()

	if !exists {
		return nil, nil
	}

	return &inventoryv1.GetStockResponse{Item: item}, nil
}
