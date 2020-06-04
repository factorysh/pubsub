package event

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvents(t *testing.T) {
	evts := NewEvents()
	wait := &sync.WaitGroup{}
	wait.Add(2)
	for i := 0; i < 2; i++ {
		go func(n int) {
			ctx := context.TODO()
			events := evts.Subscribe(ctx, 0)
			cpt := 0
			for {
				evt := <-events
				fmt.Println(n, "evt : ", evt)
				cpt++
				if cpt == 2 {
					break
				}
			}
			wait.Done()
			fmt.Println("Waiting done")
		}(i)
	}
	evts.Append(&Event{
		Data: "Pim",
	})
	evts.Append(&Event{
		Data: "Pam",
	})
	fmt.Println("Waiting")
	wait.Wait()
	fmt.Println("Stop waiting")
	evts.Close()
	evts.block.RLock()
	defer evts.block.RUnlock()
	assert.Len(t, evts.broadcast, 0)
	assert.Equal(t, 2, evts.Size())
	assert.Len(t, evts.Since(0), 2)
}

func TestMoreEvents(t *testing.T) {
	evts := NewEvents()
	evts.SetPrems(func(ctx context.Context) *Event {
		return &Event{
			Data: "prems",
		}
	})
	wait := &sync.WaitGroup{}
	wait.Add(4)
	ctx1 := context.TODO()
	client1 := evts.Subscribe(ctx1, 0)
	go func() {
		evt := <-client1
		assert.Equal(t, "prems", evt.Data)
		wait.Done()
		evt = <-client1
		assert.Equal(t, "hop", evt.Data)
		wait.Done()
	}()

	evts.Append(&Event{
		Data: "hop",
	})

	ctx2 := context.TODO()
	client2 := evts.Subscribe(ctx2, 0)
	go func() {
		evt := <-client2
		assert.Equal(t, "prems", evt.Data)
		wait.Done()
		evt = <-client2
		assert.Equal(t, "hop", evt.Data)
		wait.Done()
	}()

	wait.Wait()
}
