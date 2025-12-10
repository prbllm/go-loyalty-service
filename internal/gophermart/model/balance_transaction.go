package model

import "time"

type BalanceTransaction struct {
	UserID      int64
	OrderNumber string
	Sum         float64
	ProcessedAt time.Time
}
