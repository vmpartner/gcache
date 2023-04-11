# GCache
Simple in memory generic cache. No dependency.

Usage:  
```go
cache := NewInMemoryCache[string](4, time.Second * 30)
defer cache.Stop()
cache.Set("foo", "bar", time.Second*60)
val, ok := cache.Get("foo")
fmt.Println(val, ok) // bar, true
```