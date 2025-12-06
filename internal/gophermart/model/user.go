package model

import "time"

type User struct {
	ID           int64
	Login        string
	PasswordHash string
	Balance      float64
	Withdrawn    float64
	CreatedAt    time.Time
}
