package lang

import "sync"

type CacheStampedeMap[TKey comparable] struct {
	internalMapLock sync.RWMutex
	internalMap     map[TKey]*sync.Mutex
}

func CreateCacheStampedeMap[TKey comparable]() *CacheStampedeMap[TKey] {
	ins := &CacheStampedeMap[TKey]{
		internalMapLock: sync.RWMutex{},
		internalMap:     map[TKey]*sync.Mutex{},
	}

	return ins
}

func (m *CacheStampedeMap[TKey]) Lock(key TKey) *sync.Mutex {
	m.internalMapLock.RLock()
	mutex, exists := m.internalMap[key]
	m.internalMapLock.RUnlock()
	if !exists {
		m.internalMapLock.Lock()
		mutex, exists = m.internalMap[key]
		if !exists {
			mutex = &sync.Mutex{}
			m.internalMap[key] = mutex
		}
		m.internalMapLock.Unlock()
	}
	mutex.Lock()
	return mutex
}
