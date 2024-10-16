package dataTool

import (
	"sync"
)

// SyncMap is type-safe sync.Map
type SyncMap[K comparable, V any] struct {
	sync.Map
}

func (t *SyncMap[K, V]) Load(k K) (V, bool) {
	value, ok := t.Map.Load(k)
	return value.(V), ok
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

func (t *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	t.Map.Range(
		func(key, value any) bool {
			return f(key.(K), value.(V))
		},
	)
}

func (t *SyncMap[K, V]) IsEmpty() (IsEmpty bool) {
	IsEmpty = true
	t.Map.Range(
		func(key, value any) bool {
			IsEmpty = false
			return false
		},
	)
	return
}
