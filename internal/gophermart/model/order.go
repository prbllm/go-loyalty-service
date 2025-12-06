package model

import "time"

const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

type Order struct {
	ID         int64
	UserID     int64
	Number     string
	Status     string
	Accrual    float64
	UploadedAt time.Time
}
