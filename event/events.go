package event

import (
	"context"
	"fmt"
	"sync"
)

// Events handles a flow of Event
type Events struct {
	prems     func(context.Context) *Event
	events    []*Event
	bid       int64 //Broadcast id
	eid       int64 //Event id
	broadcast map[int64]*subscriber
	block     sync.RWMutex
	lock      sync.RWMutex
	elock     sync.Mutex
}

type subscriber struct {
	events chan *Event
	ctx    context.Context
}

// NewEvents return a new Events
func NewEvents() *Events {
	return &Events{
		events:    make([]*Event, 0),
		broadcast: make(map[int64]*subscriber),
	}
}

// SetPrems sent the initial Event when a client subscribe
func (e *Events) SetPrems(prems func(context.Context) *Event) {
	e.prems = prems
}

func (e *Events) NextEventId() string {
	e.elock.Lock()
	defer e.elock.Unlock()
	e.eid++
	return fmt.Sprintf("%d", e.eid)
}

func (e *Events) nextBid() int64 {
	e.block.Lock()
	defer e.block.Unlock()
	e.bid++
	return e.bid
}

// Since returns all Event since an id
func (e *Events) Since(since int) []*Event {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.events[since:]
}

// Subscribe for future Event
func (e *Events) Subscribe(ctx context.Context) chan *Event {
	return e.SubscribeSince(ctx, -1)
}

// SubscribeSince subscribes to a list of Event, past and future
func (e *Events) SubscribeSince(ctx context.Context, since int) chan *Event {
	bid := e.nextBid()
	s := &subscriber{
		events: make(chan *Event, 100),
		ctx:    ctx,
	}
	if e.prems != nil {
		s.events <- e.prems(ctx)
	}
	e.block.Lock()
	e.broadcast[bid] = s
	e.block.Unlock()
	if since >= 0 {
		e.lock.RLock()
		for i := since; i < len(e.events); i++ {
			s.events <- e.events[i]
		}
		e.lock.RUnlock()
	}
	go func() {
		<-ctx.Done()
		e.block.Lock()
		delete(e.broadcast, bid)
		e.block.Unlock()
	}()
	return s.events
}

// Append an Event
func (e *Events) Append(evt *Event) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.events = append(e.events, evt)
	e.block.RLock()
	defer e.block.RUnlock()
	for _, s := range e.broadcast {
		s.events <- evt
	}
}

// Close this Events
func (e *Events) Close() {
	e.block.Lock()
	for k, s := range e.broadcast {
		close(s.events)
		delete(e.broadcast, k)
	}
	e.block.Unlock()
}

// Size of this Events collection
func (e *Events) Size() int {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return len(e.events)
}
