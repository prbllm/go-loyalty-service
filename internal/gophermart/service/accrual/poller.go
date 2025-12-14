package accrual

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/repository"
	"github.com/prbllm/go-loyalty-service/internal/logger"
)

const (
	defaultPollingInterval = 1 * time.Second
	DefaultWorkers         = 5
)

type WorkerPool struct {
	repo          repository.Repository
	client        Client
	logger        logger.Logger
	interval      time.Duration
	workers       int
	jobs          chan *model.Order
	rateLimitChan chan time.Duration
	wg            sync.WaitGroup
}

func NewWorkerPool(repo repository.Repository, client Client, logger logger.Logger, interval time.Duration, workers int) *WorkerPool {
	if interval <= 0 {
		interval = defaultPollingInterval
	}
	if workers <= 0 {
		workers = DefaultWorkers
	}

	return &WorkerPool{
		repo:          repo,
		client:        client,
		logger:        logger,
		interval:      interval,
		workers:       workers,
		jobs:          make(chan *model.Order, workers*2),
		rateLimitChan: make(chan time.Duration, workers),
	}
}

func (wp *WorkerPool) Run(ctx context.Context) {
	wp.wg.Add(1)
	go wp.runPuller(ctx)

	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.runWorker(ctx, i)
	}
}

func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}

func (wp *WorkerPool) runPuller(ctx context.Context) {
	defer wp.wg.Done()
	defer close(wp.jobs)

	delay := time.Duration(0)
	ticker := time.NewTicker(wp.interval)
	defer ticker.Stop()

	for {
		if delay > 0 {
			if !wp.wait(ctx, delay) {
				return
			}
			delay = 0
		}

		select {
		case <-ctx.Done():
			return
		case retryAfter := <-wp.rateLimitChan:
			if retryAfter > delay {
				delay = retryAfter
			}
			continue
		case <-ticker.C:
		}

		backoff := wp.fetchAndQueueOrders(ctx)
		if backoff > delay {
			delay = backoff
		}
	}
}

func (wp *WorkerPool) wait(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (wp *WorkerPool) fetchAndQueueOrders(ctx context.Context) time.Duration {
	statuses := []model.OrderStatus{model.OrderStatusNew, model.OrderStatusProcessing}

	for _, status := range statuses {
		orders, err := wp.repo.GetOrdersByStatus(ctx, status)
		if err != nil {
			wp.logger.Errorf("poller: get orders by status %s: %v", status, err)
			continue
		}

		for _, order := range orders {
			select {
			case <-ctx.Done():
				return 0
			case wp.jobs <- order:
			}
		}
	}

	return 0
}

func (wp *WorkerPool) runWorker(ctx context.Context, id int) {
	defer wp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case order, ok := <-wp.jobs:
			if !ok {
				return
			}
			wp.handleOrder(ctx, order, id)
		}
	}
}

func (wp *WorkerPool) handleOrder(ctx context.Context, order *model.Order, workerID int) {
	resp, err := wp.client.GetOrder(ctx, order.Number)
	if err != nil {
		switch {
		case errors.Is(err, ErrOrderNotRegistered):
			return
		default:
			var tmr *TooManyRequestsError
			if errors.As(err, &tmr) {
				select {
				case wp.rateLimitChan <- tmr.RetryAfter:
				case <-ctx.Done():
				default:
				}
				return
			}
			wp.logger.Errorf("worker %d: get order %s from accrual: %v", workerID, order.Number, err)
			return
		}
	}

	targetStatus, accrualAmount := mapStatus(resp.Status, resp.Accrual)
	if targetStatus == "" {
		return
	}

	if err := wp.repo.UpdateOrderStatus(ctx, order.Number, targetStatus, accrualAmount); err != nil {
		wp.logger.Errorf("worker %d: update order %s status to %s: %v", workerID, order.Number, targetStatus, err)
	}
}

func mapStatus(accrualStatus string, accrual float64) (model.OrderStatus, model.Amount) {
	switch accrualStatus {
	case StatusRegistered, StatusProcessing:
		return model.OrderStatusProcessing, model.Amount(0)
	case StatusInvalid:
		return model.OrderStatusInvalid, model.Amount(0)
	case StatusProcessed:
		return model.OrderStatusProcessed, model.FromFloat64(accrual)
	default:
		return "", model.Amount(0)
	}
}
