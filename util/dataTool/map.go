package dataTool

import (
	"sync"
)

type Map[K comparable, V any] interface {
	Load(K) (V, bool)
	Store(K, V)
	LoadOrStore(K, V) (actual V, loaded bool)
	Delete(K)
	Range(func(K, V) (shouldContinue bool))
}

// SyncMap is type-safe sync.Map
type SyncMap[K comparable, V any] struct {
	Map sync.Map
}

func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{}
}
func (t *SyncMap[K, V]) Load(k K) (V, bool) {
	value, ok := t.Map.Load(k)
	if value != nil {
		return value.(V), ok
	}
	var v V
	return v, ok
}

func (t *SyncMap[K, V]) Store(k K, v V) {
	t.Map.Store(k, v)
}

func (t *SyncMap[K, V]) LoadOrStore(k K, v V) (actual V, loaded bool) {
	value, ok := t.Map.LoadOrStore(k, v)
	return value.(V), ok
}

func (t *SyncMap[K, V]) CompareAndSwap(k K, old, new V) (swapped bool) {
	return t.Map.CompareAndSwap(k, old, new)
}

func (t *SyncMap[K, V]) CompareAndDelete(old, new V) (deleted bool) {
	return t.Map.CompareAndDelete(old, new)
}

func (t *SyncMap[K, V]) Delete(key K) {
	t.Map.Delete(key)
}

func (t *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	t.Map.Range(
		func(key, value any) bool {
			return f(key.(K), value.(V))
		},
	)
}

type RWMutexMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

func NewRWMutexMap[K comparable, V any]() *RWMutexMap[K, V] {
	return &RWMutexMap[K, V]{
		m: make(map[K]V),
	}
}

func (sm *RWMutexMap[K, V]) Load(key K) (value V, ok bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	value, ok = sm.m[key]
	return
}

func (sm *RWMutexMap[K, V]) Store(key K, value V) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.m[key] = value
}

func (sm *RWMutexMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if actual, loaded = sm.m[key]; loaded {
		return actual, true
	}

	sm.m[key] = value
	return value, false
}

func (sm *RWMutexMap[K, V]) Delete(key K) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.m, key)
}

func (sm *RWMutexMap[K, V]) Range(f func(key K, value V) (shouldContinue bool)) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for k, v := range sm.m {
		if !f(k, v) {
			break
		}
	}
}

func (sm *RWMutexMap[K, V]) Len() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.m)
}
