package domain

import "sync"

type Field string

type Changes struct {
	mu    sync.RWMutex
	dirty map[Field]struct{}
}

func NewChanges() *Changes {
	return &Changes{dirty: make(map[Field]struct{})}
}

func (c *Changes) Dirty(f Field) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.dirty[f]
	return ok
}

func (c *Changes) MarkDirty(f Field) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dirty[f] = struct{}{}
}

func (c *Changes) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	clear(c.dirty)
}
