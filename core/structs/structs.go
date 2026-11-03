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

type HttpOption struct {
	Accept string
	AcceptLanguage string
	Authorization string
	Cookie string
	ContentType string
	File string
	NoFile string
	Headers [][2]string
	Param []string
	ParamCharset string
	ParamFile string
}

type HttpResponse struct {
	Method string
	Url string
	Data string
	Status string
	Cookie string
	Headers string
}
