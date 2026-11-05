package events

import (
	"sync"

	sharedtypes "crona/shared/types"
)

type Listener func(sharedtypes.KernelEvent)

type Bus struct {
	mu        sync.RWMutex
	listeners map[int]Listener
	nextID    int

	muteMu sync.Mutex
	mutes  int
}

func NewBus() *Bus {
	return &Bus{
		listeners: make(map[int]Listener),
	}
}

func (b *Bus) Emit(event sharedtypes.KernelEvent) {
	b.muteMu.Lock()
	muted := b.mutes > 0
	b.muteMu.Unlock()
	if muted {
		return
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, listener := range b.listeners {
		listener(event)
	}
}

// Mute drops events until the returned release runs, for bulk work such as seeding that
// replaces the data wholesale and would otherwise send thousands of events. Mutes nest.
// After the outermost release the caller emits one event saying what changed.
func (b *Bus) Mute() func() {
	b.muteMu.Lock()
	b.mutes++
	b.muteMu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			b.muteMu.Lock()
			b.mutes--
			b.muteMu.Unlock()
		})
	}
}

func (b *Bus) Subscribe(listener Listener) func() {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := b.nextID
	b.nextID++
	b.listeners[id] = listener

	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		delete(b.listeners, id)
	}
}
