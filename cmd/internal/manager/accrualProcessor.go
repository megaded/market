package manager

import (
	"context"
	"errors"
	"sync"
	"time"

	internal_error "github.com/megaded/market/cmd/internal/error"
	"github.com/megaded/market/cmd/internal/logger"
	"github.com/megaded/market/cmd/internal/storage"
	"go.uber.org/zap"
)

type AccrualProcessor struct {
	storage     storage.Storager
	client      AccrualClient
	workerCount int
}

func NewAccrualProcessor(storage storage.Storager, accrualClient *AccrualClient) *AccrualProcessor {
	return &AccrualProcessor{
		storage: storage,
		client:  *accrualClient,
	}
}

func (p *AccrualProcessor) Run(ctx context.Context, interval int, workerCount int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.processPendingOrders(ctx, workerCount)
		}
	}
}

func (p *AccrualProcessor) processPendingOrders(ctx context.Context, workerCount int) {
	orders, err := p.storage.GetProcessingOrders()
	if err != nil {
		if errors.Is(err, internal_error.ErrOrderNotFound) {
			logger.Log.Info("no pending orders found")
			return
		}
		logger.Log.Error("failed to get pending orders", zap.Error(err))
		return
	}

	jobs := make(chan storage.Order)
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for order := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
					p.processOrder(ctx, order)
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, order := range orders {
			select {
			case jobs <- order:
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
}

func (p *AccrualProcessor) processOrder(ctx context.Context, order storage.Order) {
	response := p.client.GetOrderStatus(ctx, order.Number)
	if response.Error != nil {
		logger.Log.Info("failed to get order info", zap.String("order_number", order.Number), zap.Error(response.Error))
		return
	}

	if response.Retry > 0 {
		logger.Log.Info("accrual service is busy, retrying later",
			zap.String("order_number", order.Number), zap.Int("retry_after", response.Retry))
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(response.Retry) * time.Second):
			return
		}
	}

	var newAccrual int
	if response.Response.Status == string(storage.OrderStatusProcessed) {
		newAccrual = response.Response.Accrual
	}

	if err := p.storage.UpdateOrder(order.Number, response.Response.Status, newAccrual); err != nil {
		logger.Log.Info("failed to update order accrual", zap.String("order_number", order.Number), zap.Error(err))
		return
	}
}
