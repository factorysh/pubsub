package event

import (
	"time"
)

type Event struct {
	Data   string
	Id     string
	Event  string
	Retry  time.Duration
	Ending bool
}
