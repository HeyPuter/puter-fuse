package kvdotgo

import (
	"sync"
	"time"

	"github.com/HeyPuter/puter-fuse-go/lang"
)

type CacheEntry[TVal any] struct {
	time.Time
	TTL   time.Duration
	Value TVal
}

type KVMap[TKey comparable, TVal any] struct {
	items                lang.IMap[TKey, CacheEntry[TVal]]
	cacheStampedeMapLock sync.RWMutex
	cacheStampedeMap     map[TKey]*sync.Mutex
}

func CreateKVMap[TKey comparable, TVal any]() *KVMap[TKey, TVal] {
	ins := &KVMap[TKey, TVal]{
		items:                lang.CreateSyncMap[TKey, CacheEntry[TVal]](nil),
		cacheStampedeMapLock: sync.RWMutex{},
		cacheStampedeMap:     map[TKey]*sync.Mutex{},
	}

	return ins
}

// Double-checked locking to get a mutex for updates to the specified key
func (m *KVMap[TKey, TVal]) getCacheStampedeMutex(key TKey) *sync.Mutex {
	m.cacheStampedeMapLock.RLock()
	mutex, exists := m.cacheStampedeMap[key]
	m.cacheStampedeMapLock.RUnlock()
	if !exists {
		m.cacheStampedeMapLock.Lock()
		mutex, exists = m.cacheStampedeMap[key]
		if !exists {
			mutex = &sync.Mutex{}
			m.cacheStampedeMap[key] = mutex
		}
		m.cacheStampedeMapLock.Unlock()
	}
	return mutex
}

func (m *KVMap[TKey, TVal]) GetOrSet(key TKey, ttl time.Duration, factory func() (TVal, error)) (TVal, error) {
	v, exists := m.items.Get(key)
	if exists && (v.Time.Add(v.TTL).After(time.Now()) || v.TTL == 0) {
		return v.Value, nil
	}

	// Lock the mutex for this key
	mutex := m.getCacheStampedeMutex(key)
	mutex.Lock()
	defer mutex.Unlock()

	// Check if the value was set while we were waiting
	v, exists = m.items.Get(key)
	if exists && (v.Time.Add(v.TTL).After(time.Now()) || v.TTL == 0) {
		return v.Value, nil
	}

	// Create the value and set it
	value, err := factory()
	m.items.Set(key, CacheEntry[TVal]{time.Now(), ttl, value})
	return value, err
}

func (m *KVMap[TKey, TVal]) Set(key TKey, value TVal, ttl time.Duration) {
	mutex := m.getCacheStampedeMutex(key)
	mutex.Lock()
	defer mutex.Unlock()
	m.items.Set(key, CacheEntry[TVal]{time.Now(), ttl, value})
}

func (m *KVMap[TKey, TVal]) SetAndLock(key TKey, value TVal, ttl time.Duration) *sync.Mutex {
	mutex := m.getCacheStampedeMutex(key)
	mutex.Lock()
	m.items.Set(key, CacheEntry[TVal]{time.Now(), ttl, value})
	return mutex
}

func (m *KVMap[TKey, TVal]) Get(key TKey) *TVal {
	v, exists := m.items.Get(key)
	if !exists || (v.TTL != 0 && v.Time.Add(v.TTL).Before(time.Now())) {
		return nil
	}
	value := v.Value
	return &value
}
