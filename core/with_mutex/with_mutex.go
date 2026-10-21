package with_mutex

import (
	"sync"
)

type WithMutex[T comparable, U any] struct {
	mutex *sync.RWMutex
	value map[T]U
}

func NewWithMutex[T comparable, U any]() WithMutex[T, U] {
	return WithMutex[T, U]{
		mutex: &sync.RWMutex{},
		value: map[T]U{},
	}
}

func (wm WithMutex[T, U]) Lock(f func(map[T]U)) {
	defer wm.mutex.Unlock()
	wm.mutex.Lock()
	f(wm.value)
}

func (wm WithMutex[T, U]) RLock(f func(map[T]U)) {
	defer wm.mutex.RUnlock()
	wm.mutex.RLock()
	f(wm.value)
}
