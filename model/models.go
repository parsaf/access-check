package model

import "time"

type AccessInterval struct {
	User string `json:"user"`
	Grantor string `json:"grantor"`
	Role string `json:"role"`
	// inclusive start of the interval
	StartTime time.Time
	// non-inclusive end of the interval
	EndTime time.Time
}

type EventLog struct {
	User string `json:"user"`
	Grantor string `json:"grantor"`
	Role string `json:"role"`
	Time time.Time
}