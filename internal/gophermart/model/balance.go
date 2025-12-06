package model

import "time"

type Balance struct {
	Current   float64
	Withdrawn float64
}

type Withdrawal struct {
	OrderNumber string
	Sum         float64
	ProcessedAt time.Time
}
