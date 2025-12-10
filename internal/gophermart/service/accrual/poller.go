package accrual

import (
	"context"
	"errors"
	"time"

	"github.com/prbllm/go-loyalty-service/internal/gophermart/model"
	"github.com/prbllm/go-loyalty-service/internal/gophermart/repository"
	"github.com/prbllm/go-loyalty-service/internal/logger"
)

const (
	defaultPollingInterval = 1 * time.Second
)

type Poller struct {
	repo     repository.Repository
	client   Client
	logger   logger.Logger
	interval time.Duration
}

func NewPoller(repo repository.Repository, client Client, logger logger.Logger, interval time.Duration) *Poller {
	if interval <= 0 {
		interval = defaultPollingInterval
	}

	return &Poller{
		repo:     repo,
		client:   client,
		logger:   logger,
		interval: interval,
	}
}

func (p *Poller) Run(ctx context.Context) {
	delay := time.Duration(0)

	for {
		if delay > 0 {
			if !p.wait(ctx, delay) {
				return
			}
		}

		backoff := p.process(ctx)
		delay = p.interval
		if backoff > delay {
			delay = backoff
		}
	}
}

func (p *Poller) wait(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (p *Poller) process(ctx context.Context) time.Duration {
	var maxBackoff time.Duration
	statuses := []string{model.OrderStatusNew, model.OrderStatusProcessing}

	for _, status := range statuses {
		orders, err := p.repo.GetOrdersByStatus(ctx, status)
		if err != nil {
			p.logger.Errorf("poller: get orders by status %s: %v", status, err)
			continue
		}

		for _, order := range orders {
			backoff := p.handleOrder(ctx, order)
			if backoff > maxBackoff {
				maxBackoff = backoff
			}
		}
	}

	return maxBackoff
}

func (p *Poller) handleOrder(ctx context.Context, order *model.Order) time.Duration {
	resp, err := p.client.GetOrder(ctx, order.Number)
	if err != nil {
		switch {
		case errors.Is(err, ErrOrderNotRegistered):
			return 0
		default:
			var tmr *TooManyRequestsError
			if errors.As(err, &tmr) {
				return tmr.RetryAfter
			}
			p.logger.Errorf("poller: get order %s from accrual: %v", order.Number, err)
			return 0
		}
	}

	targetStatus, accrualAmount := mapStatus(resp.Status, resp.Accrual)
	if targetStatus == "" {
		return 0
	}

	if err := p.repo.UpdateOrderStatus(ctx, order.Number, targetStatus, accrualAmount); err != nil {
		p.logger.Errorf("poller: update order %s status to %s: %v", order.Number, targetStatus, err)
	}

	return 0
}

func mapStatus(accrualStatus string, accrual float64) (string, model.Amount) {
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
