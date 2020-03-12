package sse

import (
	"context"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type SSE struct {
	Events *Events
}

func New() *SSE {
	return &SSE{
		Events: NewEvents(),
	}
}

func (s *SSE) HandleSSE(ctx context.Context, w http.ResponseWriter, l *log.Entry, lei int) {
	ctx2, cancel := context.WithCancel(context.TODO())
	defer cancel()
	evts := s.Events.Subscribe(ctx2, lei)
	h := w.Header()
	// https://html.spec.whatwg.org/multipage/server-sent-events.html
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	l.Info("Starting SSE")
	var evt *Event
	for {
		select {
		case evt = <-evts:
			evt.Id = fmt.Sprintf("%d", lei)
			evt.Write(w)
			lei++
		case <-ctx.Done():
			cancel()
			break
		}
	}
}
