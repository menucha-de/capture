package capture

import (
	"sync"
)

type pubsub struct {
	mu     sync.RWMutex
	subs   map[string]map[chan CaptureData]bool
	closed bool
}

var Pub pubsub

func (ps *pubsub) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if !ps.closed {
		ps.closed = true
		for _, subs := range ps.subs {
			for ch := range subs {
				close(ch)
			}
		}
	}
}
func (ps *pubsub) subscribe(topic string) chan CaptureData {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan CaptureData, 1)
	if ps.subs == nil {
		ps.subs = make(map[string]map[chan CaptureData]bool)
	}
	if ps.subs[topic] == nil {
		ps.subs[topic] = make(map[chan CaptureData]bool)
	}
	ps.subs[topic][ch] = true

	return ch
}
func (ps *pubsub) Publish(topic string, msg CaptureData) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ps.closed {
		return
	}

	for ch := range ps.subs[topic] {

		//go func(ch chan string) {
		if ps.subs[topic][ch] {
			go func(ch chan CaptureData) {
				//non-blocking send
				select {
				case ch <- msg:
				default:
				}
			}(ch)
		}

	}
}
func (ps *pubsub) unSubscribe(topic string, ch chan CaptureData) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.subs[topic][ch] = false

	delete(ps.subs[topic], ch)

}
