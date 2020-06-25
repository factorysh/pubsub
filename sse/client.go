package sse

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	_event "github.com/factorysh/pubsub/event"
)

type SSEReader struct {
	scanner *bufio.Scanner
}

func NewSSEReader(r io.Reader) *SSEReader {
	return &SSEReader{bufio.NewScanner(r)}
}

func (s *SSEReader) Read() (*_event.Event, error) {
	evt := &_event.Event{}
	for s.scanner.Scan() {
		line := s.scanner.Text()
		if line == "" {
			return evt, nil
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		switch len(parts) {
		case 1:
			event(evt, parts[0], "")
		case 2:
			event(evt, parts[0], parts[1][:len(parts[1])])
		}
	}
	return nil, io.EOF
}

func event(evt *_event.Event, key, value string) {
	value = strings.TrimPrefix(value, " ")
	if strings.HasSuffix(value, "\r") {
		value = value[:len(value)-1]
	}
	switch key {
	case "id":
		evt.Id = value
	case "retry":
		retry, err := strconv.Atoi(value)
		if err == nil {
			// like Mozilla, we doesn't throw an error
			evt.Retry = time.Duration(retry) * time.Millisecond
		}
	case "event":
		evt.Event = value
	case "data":
		if evt.Data != "" {
			evt.Data = fmt.Sprintf("%s\n%s", evt.Data, value)
		} else {
			evt.Data = value
		}
	}
}
