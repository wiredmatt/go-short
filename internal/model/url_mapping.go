package model

import "time"

type URLMapping struct {
	Code      string
	Original  string
	UserID    string
	CreatedAt time.Time
	ExpiresAt *time.Time
	Clicks    int
}
