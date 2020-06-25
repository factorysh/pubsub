package sse

import (
	"context"
	"fmt"
	"net/http"

	_event "github.com/factorysh/pubsub/event"
	log "github.com/sirupsen/logrus"
)

func HandleSSE(ctx context.Context, e *_event.Events, w http.ResponseWriter, l *log.Entry, lei int) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	ctx2, cancel := context.WithCancel(context.TODO())
	defer cancel()
	evts := e.SubscribeSince(ctx2, lei)
	h := w.Header()
	// https://html.spec.whatwg.org/multipage/server-sent-events.html
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	flusher.Flush()
	l.Info("Starting SSE")
	var evt *_event.Event
	for {
		select {
		case evt = <-evts:
			evt.Id = fmt.Sprintf("%d", lei)
			err := WriteEvent(w, evt)
			if err != nil {
				log.WithError(err).Error()
				// lets kill the http connection
				return
			}
			flusher.Flush()
			lei++
			if evt.Ending {
				break
			}
		case <-ctx.Done():
			break
		}
	}
}
