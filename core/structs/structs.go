package structs

import (
	"time"
)

type Event struct {
	Method string
	Option map[string]struct{}
	Event string
	Reference []string
}

type LogData struct {
	Time time.Time
	Ghost string
	Status map[string]struct{}
	Script string
}
