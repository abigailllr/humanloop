package hub

import "sync"

type Hub struct {
	mu   sync.Mutex
	subs map[string][]chan []byte
}

var Default = &Hub{subs: make(map[string][]chan []byte)}

func (h *Hub) Subscribe(id string) chan []byte {
	ch := make(chan []byte, 8)
	h.mu.Lock()
	h.subs[id] = append(h.subs[id], ch)
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(id string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	list := h.subs[id]
	for i, c := range list {
		if c == ch {
			h.subs[id] = append(list[:i], list[i+1:]...)
			break
		}
	}
	if len(h.subs[id]) == 0 {
		delete(h.subs, id)
	}
}

func (h *Hub) Publish(id string, data []byte) {
	h.mu.Lock()
	list := make([]chan []byte, len(h.subs[id]))
	copy(list, h.subs[id])
	h.mu.Unlock()
	for _, ch := range list {
		select {
		case ch <- data:
		default:
		}
	}
}
