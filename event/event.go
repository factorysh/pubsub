package event

import (
	"time"
)

type Event struct {
	DataExists bool
	Data       string
	Id         string
	Event      string
	Retry      time.Duration
	Ending     bool
}
