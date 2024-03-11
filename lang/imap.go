/*
 * Copyright (C) 2024  Puter Technologies Inc.
 *
 * This file is part of puter-fuse.
 *
 * puter-fuse is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
package lang

import (
	"sync"
)

type IMap[TKey any, TVal any] interface {
	Get(key TKey) (val TVal, ok bool)
	Set(key TKey, val TVal)
	GetWithFactory(key TKey, factory func() (TVal, bool, error)) (TVal, bool, error)
	Has(key TKey) bool
	Del(key TKey)
	Keys() []TKey
	Values() []TVal
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

func (m *ProxyMap[TKey, TVal]) GetWithFactory(key TKey, factory func() (TVal, bool, error)) (TVal, bool, error) {
	return m.Delegate.GetWithFactory(key, factory)
}

func (m *ProxyMap[TKey, TVal]) Set(key TKey, val TVal) {
	m.Delegate.Set(key, val)
}

func (m *ProxyMap[TKey, TVal]) Has(key TKey) bool {
	return m.Delegate.Has(key)
}

func (m *ProxyMap[TKey, TVal]) Del(key TKey) {
	m.Delegate.Del(key)
}

func (m *ProxyMap[TKey, TVal]) Keys() []TKey {
	return m.Delegate.Keys()
}

func (m *ProxyMap[TKey, TVal]) Values() []TVal {
	return m.Delegate.Values()
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

func (m *Map[TKey, TVal]) GetWithFactory(key TKey, factory func() (TVal, bool, error)) (TVal, bool, error) {
	val, ok := m.Items[key]
	if !ok {
		val, ok, err := factory()
		if ok {
			m.Items[key] = val
		}
		return val, ok, err
	}
	return val, ok, nil
}

func (m *Map[TKey, TVal]) Has(key TKey) bool {
	_, ok := m.Items[key]
	return ok
}

func (m *Map[TKey, TVal]) Del(key TKey) {
	delete(m.Items, key)
}

func (m *Map[TKey, TVal]) Keys() []TKey {
	keys := make([]TKey, 0, len(m.Items))
	for k := range m.Items {
		keys = append(keys, k)
	}
	return keys
}

func (m *Map[TKey, TVal]) Values() []TVal {
	values := make([]TVal, 0, len(m.Items))
	for _, v := range m.Items {
		values = append(values, v)
	}
	return values
}

type SyncMap[TKey comparable, TVal any] struct {
	ProxyMap[TKey, TVal]
	lock    sync.RWMutex
	mapLock *CacheStampedeMap[TKey]
}

func CreateSyncMap[TKey comparable, TVal any](delegate IMap[TKey, TVal]) *SyncMap[TKey, TVal] {
	if delegate == nil {
		delegate = CreateMap[TKey, TVal]()
	}
	return &SyncMap[TKey, TVal]{
		ProxyMap: ProxyMap[TKey, TVal]{
			Delegate: delegate,
		},
		lock:    sync.RWMutex{},
		mapLock: CreateCacheStampedeMap[TKey](),
	}
}

func (m *SyncMap[TKey, TVal]) Get(key TKey) (TVal, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	val, ok := m.Delegate.Get(key)
	return val, ok
}

func (m *SyncMap[TKey, TVal]) Set(key TKey, val TVal) {
	m.lock.Lock()
	m.Delegate.Set(key, val)
	m.lock.Unlock()
}

func (m *SyncMap[TKey, TVal]) GetWithFactory(key TKey, factory func() (TVal, bool, error)) (TVal, bool, error) {
	v, ok := m.Get(key)
	if ok {
		return v, ok, nil
	}

	mutex := m.mapLock.Lock(key)
	defer mutex.Unlock()

	v, ok = m.Get(key)
	if ok {
		return v, ok, nil
	}

	value, ok, err := factory()
	if ok {
		m.Set(key, value)
	}
	return value, ok, err
}

func (m *SyncMap[TKey, TVal]) Has(key TKey) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.Delegate.Has(key)
}

func (m *SyncMap[TKey, TVal]) Del(key TKey) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Delegate.Del(key)
}

func (m *SyncMap[TKey, TVal]) Keys() []TKey {
	m.lock.RLock()
	defer m.lock.RUnlock()
	keys := m.Delegate.Keys()
	return keys
}

func (m *SyncMap[TKey, TVal]) Values() []TVal {
	m.lock.RLock()
	defer m.lock.RUnlock()
	values := m.Delegate.Values()
	return values
}
