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
