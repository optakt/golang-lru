package simplelru

import (
	"testing"
)

func TestLRU(t *testing.T) {
	t.Run("without aborted eviction", func(t *testing.T) {
		evictCounter := 0
		onEvicted := func(k, v interface{}, used int) bool {
			if k != v {
				t.Fatalf("Evict values not equal (%v!=%v)", k, v)
			}
			evictCounter++
			return true
		}
		l, err := NewLRU(128, onEvicted)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		for i := 0; i < 256; i++ {
			l.Add(i, i)
		}
		if l.Len() != 128 {
			t.Fatalf("bad len: %v", l.Len())
		}

		if evictCounter != 128 {
			t.Fatalf("bad evict count: %v", evictCounter)
		}

		for i, k := range l.Keys() {
			if v, ok := l.Get(k); !ok || v != k || v != i+128 {
				t.Fatalf("bad key: %v", k)
			}
		}
		for i := 0; i < 128; i++ {
			_, ok := l.Get(i)
			if ok {
				t.Fatalf("should be evicted")
			}
		}
		for i := 128; i < 256; i++ {
			_, ok := l.Get(i)
			if !ok {
				t.Fatalf("should not be evicted")
			}
		}
		for i := 128; i < 192; i++ {
			ok := l.Remove(i)
			if !ok {
				t.Fatalf("should be contained")
			}
			ok = l.Remove(i)
			if ok {
				t.Fatalf("should not be contained")
			}
			_, ok = l.Get(i)
			if ok {
				t.Fatalf("should be deleted")
			}
		}

		l.Get(192) // expect 192 to be last key in l.Keys()

		for i, k := range l.Keys() {
			if (i < 63 && k != i+193) || (i == 63 && k != 192) {
				t.Fatalf("out of order key: %v", k)
			}
		}

		l.Purge()
		if l.Len() != 0 {
			t.Fatalf("bad len: %v", l.Len())
		}
		if _, ok := l.Get(200); ok {
			t.Fatalf("should contain nothing")
		}
	})

	t.Run("with aborted eviction", func(t *testing.T) {
		abortEviction := true
		evictCounter := 0
		onEvicted := func(k, v interface{}, used int) bool {
			if k != v {
				t.Fatalf("Evict values not equal (%v!=%v)", k, v)
			}
			evictCounter++

			// Abort the first eviction.
			// The next entry should be evicted instead, then.
			if abortEviction {
				abortEviction = false
				return false
			}
			return true
		}
		l, err := NewLRU(128, onEvicted)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		for i := 0; i < 256; i++ {
			l.Add(i, i)
		}
		if l.Len() != 128 {
			t.Fatalf("bad len: %v", l.Len())
		}

		if evictCounter != 129 {
			t.Fatalf("bad evict count: %v", evictCounter)
		}

		for i, k := range l.Keys() {
			v, ok := l.Get(k)
			if !ok {
				t.Fatalf("should contain key: %v", k)
			}

			// Oldest value's eviction should have been aborted.
			if i == 0 && v != k {
				t.Fatalf("bad key: got %v want 0", k)
			}

			// From then on, we expect normal key values.
			if i != 0 && (v != k || v != i+128) {
				t.Fatalf("bad key: got %v want %d", k, i+128)
			}
		}
		for i := 1; i < 129; i++ {
			_, ok := l.Get(i)
			if ok {
				t.Fatalf("should be evicted")
			}
		}

		_, ok := l.Get(0)
		if !ok {
			t.Fatalf("should not be evicted")
		}
		for i := 129; i < 256; i++ {
			_, ok := l.Get(i)
			if !ok {
				t.Fatalf("should not be evicted")
			}
		}
		for i := 129; i < 193; i++ {
			ok := l.Remove(i)
			if !ok {
				t.Fatalf("should be contained")
			}
			ok = l.Remove(i)
			if ok {
				t.Fatalf("should not be contained")
			}
			_, ok = l.Get(i)
			if ok {
				t.Fatalf("should be deleted")
			}
		}

		l.Get(193) // expect 192 to be last key in l.Keys()

		for i, k := range l.Keys() {
			if i == 0 && k == 0 {
				continue
			}
			if (i < 63 && k != i+193) || (i == 63 && k != 193) {
				t.Fatalf("out of order key: %v", k)
			}
		}

		l.Purge()
		if l.Len() != 0 {
			t.Fatalf("bad len: %v", l.Len())
		}
		if _, ok := l.Get(200); ok {
			t.Fatalf("should contain nothing")
		}
	})
}

func TestLRU_GetOldest_RemoveOldest(t *testing.T) {
	l, err := NewLRU(128, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	for i := 0; i < 256; i++ {
		l.Add(i, i)
	}
	k, _, ok := l.GetOldest()
	if !ok {
		t.Fatalf("missing")
	}
	if k.(int) != 128 {
		t.Fatalf("bad: %v", k)
	}

	k, _, ok = l.RemoveOldest()
	if !ok {
		t.Fatalf("missing")
	}
	if k.(int) != 128 {
		t.Fatalf("bad: %v", k)
	}

	k, _, ok = l.RemoveOldest()
	if !ok {
		t.Fatalf("missing")
	}
	if k.(int) != 129 {
		t.Fatalf("bad: %v", k)
	}
}

// Test that Add returns true/false if an eviction occurred
func TestLRU_Add(t *testing.T) {
	t.Run("without aborted eviction", func(t *testing.T) {
		t.Parallel()

		evictCounter := 0
		onEvicted := func(k, v interface{}, used int) bool {
			evictCounter++
			return true
		}

		l, err := NewLRU(1, onEvicted)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		if l.Add(1, 1) == true || evictCounter != 0 {
			t.Errorf("should not have an eviction")
		}
		if l.Add(2, 2) == false || evictCounter != 1 {
			t.Errorf("should have an eviction")
		}
	})

	t.Run("with aborted eviction", func(t *testing.T) {
		t.Parallel()

		abortEviction := true
		evictCounter := 0
		onEvicted := func(k, v interface{}, used int) bool {
			evictCounter++

			// Abort the first eviction.
			// The next entry should be evicted instead, then.
			if abortEviction {
				abortEviction = false
				return false
			}
			return true
		}

		l, err := NewLRU(1, onEvicted)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		if l.Add(1, 1) == true || evictCounter != 0 {
			t.Errorf("should not have an eviction")
		}
		if l.Add(2, 2) == false || evictCounter != 2 {
			t.Errorf("should have an eviction")
		}
	})
}

// Test that Contains doesn't update recent-ness
func TestLRU_Contains(t *testing.T) {
	l, err := NewLRU(2, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.Add(1, 1)
	l.Add(2, 2)
	if !l.Contains(1) {
		t.Errorf("1 should be contained")
	}

	l.Add(3, 3)
	if l.Contains(1) {
		t.Errorf("Contains should not have updated recent-ness of 1")
	}
}

// Test that Peek doesn't update recent-ness
func TestLRU_Peek(t *testing.T) {
	l, err := NewLRU(2, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.Add(1, 1)
	l.Add(2, 2)
	if v, ok := l.Peek(1); !ok || v != 1 {
		t.Errorf("1 should be set to 1: %v, %v", v, ok)
	}

	l.Add(3, 3)
	if l.Contains(1) {
		t.Errorf("should not have updated recent-ness of 1")
	}
}

// Test that Resize can upsize and downsize
func TestLRU_Resize(t *testing.T) {
	t.Run("without aborted evictions", func(t *testing.T) {
		t.Parallel()

		onEvictCounter := 0
		onEvicted := func(k, v interface{}, used int) bool {
			onEvictCounter++
			return true
		}
		l, err := NewLRU(2, onEvicted)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		// Downsize
		l.Add(1, 1)
		l.Add(2, 2)
		evicted := l.Resize(1)
		if evicted != 1 {
			t.Errorf("1 element should have been evicted: %v", evicted)
		}
		if onEvictCounter != 1 {
			t.Errorf("onEvicted should have been called once: %v", onEvictCounter)
		}

		l.Add(3, 3)
		if l.Contains(1) {
			t.Errorf("Element 1 should have been evicted")
		}

		// Upsize
		evicted = l.Resize(2)
		if evicted != 0 {
			t.Errorf("0 elements should have been evicted: %v", evicted)
		}

		l.Add(4, 4)
		if !l.Contains(3) || !l.Contains(4) {
			t.Errorf("Cache should have contained 2 elements")
		}
	})

	t.Run("with aborted evictions", func(t *testing.T) {
		t.Parallel()

		onEvictCounter := 0
		abortEviction := true
		onEvicted := func(k, v interface{}, used int) bool {
			onEvictCounter++

			// Abort the first eviction.
			// The next entry should be evicted instead, then.
			if abortEviction {
				abortEviction = false
				return false
			}
			return true
		}
		l, err := NewLRU(2, onEvicted)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		// Downsize
		l.Add(1, 1)
		l.Add(2, 2)
		evicted := l.Resize(1)
		if evicted != 1 {
			t.Errorf("1 element should have been evicted: %v", evicted)
		}
		if onEvictCounter != 2 {
			t.Errorf("onEvicted should have been called twice: %v", onEvictCounter)
		}
		if !l.Contains(1) {
			t.Errorf("Element 1 should not have been evicted")
		}
		if l.Contains(2) {
			t.Errorf("Element 2 should have been evicted")
		}

		l.Add(3, 3)
		if l.Contains(1) {
			t.Errorf("Element 1 should have been evicted")
		}
		if l.Contains(2) {
			t.Errorf("Element 2 should have been evicted")
		}
		if !l.Contains(3) {
			t.Errorf("Element 3 should not have been evicted")
		}

		// Upsize
		evicted = l.Resize(2)
		if evicted != 0 {
			t.Errorf("0 elements should have been evicted: %v", evicted)
		}

		l.Add(4, 4)
		if !l.Contains(3) || !l.Contains(4) {
			t.Errorf("Cache should have contained 2 elements")
		}
	})
}
