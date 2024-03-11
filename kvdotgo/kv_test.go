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
package kvdotgo

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestKVMap(t *testing.T) {
	t.Run("GetOrSet", func(t *testing.T) {
		m := CreateKVMap[string, string]()

		v0, _, _ := m.GetOrSet("key", time.Second, func() (string, bool, error) {
			return "this-one", true, nil
		})

		if v0 != "this-one" {
			t.Errorf("expected 'this-one', got '%s'", v0)
		}

		v, _, _ := m.GetOrSet("key", time.Second, func() (string, bool, error) {
			return "not-this-one", true, nil
		})

		if v != "this-one" {
			t.Errorf("expected 'this-one', got '%s'", v)
		}
	})

	t.Run("Stampede Protection", func(t *testing.T) {
		m := CreateKVMap[string, string]()

		calls := atomic.Uint64{}
		wg := sync.WaitGroup{}

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _, _ = m.GetOrSet("key", time.Second, func() (string, bool, error) {
					calls.Add(1)
					return "this-one", true, nil
				})
			}()
		}

		wg.Wait()

		if calls.Load() != 1 {
			t.Errorf("expected 1 call, got %d", calls.Load())
		}
	})
}
