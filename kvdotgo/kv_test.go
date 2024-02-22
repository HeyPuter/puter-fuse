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

		v0, _ := m.GetOrSet("key", time.Second, func() (string, error) {
			return "this-one", nil
		})

		if v0 != "this-one" {
			t.Errorf("expected 'this-one', got '%s'", v0)
		}

		v, _ := m.GetOrSet("key", time.Second, func() (string, error) {
			return "not-this-one", nil
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
				_, _ = m.GetOrSet("key", time.Second, func() (string, error) {
					calls.Add(1)
					return "this-one", nil
				})
			}()
		}

		wg.Wait()

		if calls.Load() != 1 {
			t.Errorf("expected 1 call, got %d", calls.Load())
		}
	})
}
