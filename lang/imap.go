package lang

import "sync"

type IMap[TKey any, TVal any] interface {
	Get(key TKey) (val TVal, ok bool)
	Set(key TKey, val TVal)
	Del(key TKey)
}

type ProxyMap[TKey any, TVal any] struct {
	Delegate IMap[TKey, TVal]
}

func CreateProxyMap[TKey any, TVal any](delegate IMap[TKey, TVal]) *ProxyMap[TKey, TVal] {
	return &ProxyMap[TKey, TVal]{
		Delegate: delegate,
	}
}

func (m *ProxyMap[TKey, TVal]) Get(key TKey) (val TVal, ok bool) {
	return m.Delegate.Get(key)
}

func (m *ProxyMap[TKey, TVal]) Set(key TKey, val TVal) {
	m.Delegate.Set(key, val)
}

func (m *ProxyMap[TKey, TVal]) Del(key TKey) {
	m.Delegate.Del(key)
}

type Map[TKey comparable, TVal any] struct {
	Items map[TKey]TVal
}

func CreateMap[TKey comparable, TVal any]() *Map[TKey, TVal] {
	return &Map[TKey, TVal]{
		Items: map[TKey]TVal{},
	}
}

func (m *Map[TKey, TVal]) Get(key TKey) (TVal, bool) {
	val, ok := m.Items[key]
	return val, ok
}

func (m *Map[TKey, TVal]) Set(key TKey, val TVal) {
	m.Items[key] = val
}

func (m *Map[TKey, TVal]) Del(key TKey) {
	delete(m.Items, key)
}

type SyncMap[TKey comparable, TVal any] struct {
	ProxyMap[TKey, TVal]
	lock sync.RWMutex
}

func CreateSyncMap[TKey comparable, TVal any](delegate IMap[TKey, TVal]) *SyncMap[TKey, TVal] {
	if delegate == nil {
		delegate = CreateMap[TKey, TVal]()
	}
	return &SyncMap[TKey, TVal]{
		ProxyMap: ProxyMap[TKey, TVal]{
			Delegate: delegate,
		},
		lock: sync.RWMutex{},
	}
}

func (m *SyncMap[TKey, TVal]) Get(key TKey) (TVal, bool) {
	m.lock.RLock()
	val, ok := m.Delegate.Get(key)
	m.lock.RUnlock()
	return val, ok
}

func (m *SyncMap[TKey, TVal]) Set(key TKey, val TVal) {
	m.lock.Lock()
	m.Delegate.Set(key, val)
	m.lock.Unlock()
}

func (m *SyncMap[TKey, TVal]) Del(key TKey) {
	m.lock.Lock()
	m.Delegate.Del(key)
	m.lock.Unlock()
}
