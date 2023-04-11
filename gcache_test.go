package gcache

import (
	"strconv"
	"testing"
	"time"
)

func TestInMemoryCache(t *testing.T) {
	cache := NewInMemoryCache[string](4, time.Second)
	cache.Set("foo", "bar", time.Second*2)

	// Wait for the item to expire
	time.Sleep(time.Second * 3)

	_, ok := cache.Get("foo")
	if ok {
		t.Errorf("Expected item to be evicted from cache, but it still exists")
	}

	cache.Set("foo", "bar", time.Second*2)

	val, ok := cache.Get("foo")
	if !ok {
		t.Errorf("Expected item to exist in cache, but it doesn't")
	}

	if val != "bar" {
		t.Errorf("Expected value to be 'bar', but got '%v'", val)
	}
	cache.Stop()
}

func BenchmarkInMemoryCache(b *testing.B) {
	cache := NewInMemoryCache[string](4, time.Second)
	defer cache.Stop()
	for i := 0; i < b.N; i++ {
		x := strconv.Itoa(i)
		cache.Set(x, x, time.Second*2)
		cache.Get(x)
	}
}
