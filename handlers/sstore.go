package handlers

import (
	"sync"
)

// SessionStore interface, must be implemented by any store to be used
// for session storage.
type SessionStore interface {
	Get(id string) (interface{}, error)    // Get the session from the store
	Set(id string, sess interface{}) error // Save the session in the store
	Delete(id string) error                // Delete the session from the store
	Clear() error                          // Delete all sessions from the store
	Len() int                              // Get the number of sessions in the store
}

// In-memory implementation of a session store. Not recommended for production
// use.
type MemoryStore struct {
	l    sync.RWMutex
	m    map[string]interface{}
	capc int
}

func NewMemoryStore(capc int) *MemoryStore {
	m := &MemoryStore{}
	m.capc = capc
	m.newMap()
	return m
}

func (this *MemoryStore) Len() int {
	return len(this.m)
}

func (this *MemoryStore) Get(id string) (interface{}, error) {
	this.l.RLock()
	defer this.l.RUnlock()
	return this.m[id], nil
}

func (this *MemoryStore) Set(id string, sess interface{}) error {
	this.l.Lock()
	defer this.l.Unlock()
	this.m[id] = sess
	return nil
}

func (this *MemoryStore) Delete(id string) error {
	this.l.Lock()
	defer this.l.Unlock()
	delete(this.m, id)
	return nil
}

func (this *MemoryStore) Clear() error {
	this.l.Lock()
	defer this.l.Unlock()
	this.newMap()
	return nil
}

func (this *MemoryStore) newMap() {
	this.m = make(map[string]interface{}, this.capc)
}
