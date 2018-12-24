package syncex

import (
	"sync"
)

// A ReentrantMutex is particular type of mutual exclusion (mutex) that
// may be locked multiple times by the same goroutine.
//
// A ReentrantMutex must not be copied after first use.
type ReentrantMutex struct {
	mu sync.Mutex
	c  chan struct{}
	v  int32
	id uint64
}

// Lock locks m.
// If m is already owned by different goroutine, it waits
// until the ReentrantMutex is available.
func (m *ReentrantMutex) Lock() {
	id := getGID()
	for {
		m.mu.Lock()
		if m.c == nil {
			m.c = make(chan struct{}, 1)
		}
		if m.v == 0 || m.id == id {
			m.v++
			m.id = id
			m.mu.Unlock()
			break
		}
		m.mu.Unlock()
		<-m.c
	}
}

// Unlock unlocks m.
// It panics if m is not locked on entry to Unlock.
func (m *ReentrantMutex) Unlock() {
	m.mu.Lock()
	if m.c == nil {
		m.c = make(chan struct{}, 1)
	}
	if m.v <= 0 {
		m.mu.Unlock()
		panic(ErrNotLocked)
	}
	m.v--
	if m.v == 0 {
		m.id = 0
	}
	m.mu.Unlock()
	select {
	case m.c <- struct{}{}:
	default:
	}
}
