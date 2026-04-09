# gRPC Microservices Demo

A production-structured gRPC-based microservices demo in Go demonstrating real-world service-to-service communication.

## Architecture

```
                    ┌──────────────────┐
                    │    Client        │
                    └────────┬─────────┘
                             │ gRPC
                             ▼
┌──────────────────────────────────────────────────────────────┐
│                     Order Service (:50050)                    │
│  - CreateOrder (Unary RPC)                                   │
│  - GetOrder (Unary RPC)                                      │
│  - BulkCreateOrders (Client Streaming RPC)                   │
└──────┬──────────────────┬──────────────────┬───────────────┘
       │ gRPC              │ gRPC              │ gRPC
       ▼                  ▼                  ▼
┌──────────────┐  ┌──────────────┐  ┌────────────────────────┐
│  Inventory    │  │   Payment    │  │    Notification        │
│  Service      │  │   Service    │  │    Service (:50053)     │
│  (:50051)     │  │   (:50052)   │  │                         │
│               │  │              │  │  - OrderStatusStream     │
│  - CheckStock │  │  - Process   │  │    (Server Streaming)   │
│  - UpdateStock│  │    Payment   │  │                         │
│  - GetStock   │  │              │  │  - LiveOrderTracking     │
└──────────────┘  └──────────────┘  │    (Bidi Streaming)     │
                                     └─────────────────────────┘
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| Order | 50050 | Main service coordinating order processing |
| Inventory | 50051 | Manages product stock |
| Payment | 50052 | Processes payments (85% success rate) |
| Notification | 50053 | Sends notifications and streams updates |

## RPC Types Demonstrated

1. **Unary RPC**: `CreateOrder`, `GetOrder`
2. **Server Streaming**: `OrderStatusStream`
3. **Client Streaming**: `BulkCreateOrders`
4. **Bidirectional Streaming**: `LiveOrderTracking`

## Project Structure

```
grpc-demo/
├── cmd/
│   ├── order-service/          # Order service entry point
│   ├── inventory-service/      # Inventory service entry point
│   ├── payment-service/        # Payment service entry point
│   ├── notification-service/   # Notification service entry point
│   └── client/                 # Test client
├── internal/
│   ├── order/                  # Order business logic
│   ├── inventory/              # Inventory business logic
│   ├── payment/                # Payment business logic
│   ├── notification/           # Notification business logic
│   └── interceptor/            # Logging & Auth interceptors
├── proto/
│   ├── order/v1/               # Order service proto
│   ├── inventory/v1/           # Inventory service proto
│   ├── payment/v1/            # Payment service proto
│   └── notification/v1/        # Notification service proto
├── frontend/                   # Minimal HTML UI
├── go.mod
└── README.md
```

## Quick Start

### 1. Generate Proto Files (if modified)

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
export PATH="$PATH:$(go env GOPATH)/bin"

protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/*/v1/*.proto
```

### 2. Start All Services

Terminal 1 - Inventory Service:
```bash
go run cmd/inventory-service/main.go
```

Terminal 2 - Payment Service:
```bash
go run cmd/payment-service/main.go
```

Terminal 3 - Notification Service:
```bash
go run cmd/notification-service/main.go
```

Terminal 4 - Order Service:
```bash
go run cmd/order-service/main.go
```

### 3. Run the Client

```bash
go run cmd/client/main.go
```

## grpcurl Commands

### Check if services are running
```bash
grpcurl localhost:50050 list
grpcurl localhost:50051 list
grpcurl localhost:50052 list
grpcurl localhost:50053 list
```

### Create Order (Unary RPC)
```bash
grpcurl -d '{
  "customer_id": "CUST-001",
  "items": [
    {"product_id": "PROD-001", "name": "Laptop", "quantity": 1, "price": 999.99},
    {"product_id": "PROD-002", "name": "Mouse", "quantity": 2, "price": 29.99}
  ]
}' localhost:50050 order.v1.OrderService/CreateOrder
```

### Get Order
```bash
grpcurl -d '{"order_id": "ORD-xxx"}' localhost:50050 order.v1.OrderService/GetOrder
```

### Check Stock (Inventory Service)
```bash
grpcurl -d '{"product_id": "PROD-001", "requested_quantity": 5}' localhost:50051 inventory.v1.InventoryService/CheckStock
```

### Process Payment (Payment Service)
```bash
grpcurl -d '{
  "order_id": "ORD-001",
  "customer_id": "CUST-001",
  "amount": 999.99,
  "method": "PAYMENT_METHOD_CREDIT_CARD"
}' localhost:50052 payment.v1.PaymentService/ProcessPayment
```

### Subscribe to Order Status (Server Streaming)
```bash
grpcurl -d '{"customer_id": "CUST-001"}' localhost:50053 notification.v1.NotificationService/OrderStatusStream
```

### Send Notification
```bash
grpcurl -d '{
  "order_id": "ORD-001",
  "customer_id": "CUST-001",
  "message": "Your order has been shipped!",
  "type": "NOTIFICATION_TYPE_EMAIL"
}' localhost:50053 notification.v1.NotificationService/SendNotification
```

### With Authentication Header
```bash
grpcurl -H "x-api-key: grpc-demo-api-key-2024" \
       -d '{"product_id": "PROD-001"}' \
       localhost:50051 inventory.v1.InventoryService/GetStock
```

## Evans CLI

```bash
# Install evans
go install github.com/ktr0731/evans@latest

# Use evans REPL
evans -p 50050 -H "x-api-key: grpc-demo-api-key-2024"

# In evans REPL:
package order.v1
service OrderService
call CreateOrder
```

## Interceptors

### Logging Interceptor
Logs all unary and streaming RPC calls with request/response data.

### Auth Interceptor
Validates `x-api-key` header in metadata.
- Default API key: `grpc-demo-api-key-2024`
- Also accepts keys with `grpc-demo-` prefix

## Testing

### Run all services and test manually:
1. Start services in 4 terminals
2. Run client: `go run cmd/client/main.go`
3. Or use grpcurl commands above

### View frontend:
```bash
# Open in browser
open frontend/index.html
```

## Error Handling

Services return proper gRPC status codes:
- `InvalidArgument` - Invalid request data
- `NotFound` - Resource not found
- `Internal` - Internal service error
- `Unauthenticated` - Invalid API key

## Mock Data

### Products (Inventory)
| Product ID | Name | Stock |
|------------|------|-------|
| PROD-001 | Laptop | 50 |
| PROD-002 | Mouse | 200 |
| PROD-003 | Keyboard | 150 |
| PROD-004 | Monitor | 75 |
| PROD-005 | Headphones | 100 |

## License

MIT
