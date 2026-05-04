package models

import "time"

type PC struct {
	ID        int
	Name      string
	Spec      string
	Price     int
	EndTime   *time.Time
	UserEmail string
}
