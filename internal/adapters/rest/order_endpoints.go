package rest

import (
	"context"

	"github.com/miguelaht/microservices/order/internal/application/core/domain"
)

type CreateOrderRequest struct {
	UserId     int64       `json:"user_id"`
	OrderItems []OrderItem `json:"order_items"`
}

type OrderItem struct {
	ProductCode string  `json:"product_code"`
	UnitPrice   float32 `json:"unit_price"`
	Quantity    int32   `json:"quantity"`
}

type CreateOrderResponse struct {
	OrderId int64 `json:"order_id"`
}

func (a Adapter) Create(ctx context.Context, request *CreateOrderRequest) (*CreateOrderResponse, error) {
	var orderItems []domain.OrderItem
	for _, orderItem := range request.OrderItems {
		orderItems = append(orderItems, domain.OrderItem{
			ProductCode: orderItem.ProductCode,
			UnitPrice:   orderItem.UnitPrice,
			Quantity:    orderItem.Quantity,
		})
	}
	newOrder := domain.NewOrder(request.UserId, orderItems)
	result, err := a.api.PlaceOrder(newOrder)
	if err != nil {
		return nil, err
	}
	return &CreateOrderResponse{OrderId: result.ID}, nil
}
