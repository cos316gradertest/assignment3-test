# HTTP Routing Framework

In this project, you'll build an in-memory cache using a least-recently-used (LRU)
eviction scheme.

## API

Your solution must implement the following API:

```go
type LRU struct {
	// whatever fields you want here
}

// Return a new LRU with capacity to store limit bytes.
func NewLru(limit int) *LRU

// Return the maximum number of bytes that your LRU can store.
func (lru *LRU) MaxStorage() int

// Return the number of bytes of storage remaining in your LRU.
func (lru *LRU) RemainingStorage() int

// Return the number of bindings currently stored in the LRU.
func (lru *LRU) Len() int

// Add a binding to the LRU with the specified key and value.
//
// Use len(key) and len(value) to give the number of bytes that this new binding
// would consume, and ensure that there is enough space in the LRU to
// accommodate it.
//
// If the LRU is not large enough to accommodate the binding, return false.
//
// Otherwise, evict the least-recently-used binding as many times as
// necessary until there is room to insert the new binding.
func (lru *LRU) Set(key string, value []byte) bool

// Return the value associated with the specified key from the LRU.
//
// Use `ok=true` to indicate the value was returned successfully, or `ok=false`
// to indicate some issue (e.g. no binding exists for that key)
//
// Additionally, each call to `Get` should update the binding in some way
// to ensure that it is the most-recently-used.
func (lru *LRU) Get(key string) (value []byte, ok bool)

// Remove the binding with specified key from the LRU, and return the value
// that was bound to it. Use `ok=true` to indicate the value was removed
// and returned successfully, or `ok=false` to indicate some issue
// (e.g. no binding exists for that key)
func (lru *LRU) Remove(key string) (value []byte, ok bool)
```

## Additional Specifications


## Sample Application

TBD. For now, you may test the lru using the provided `lru_test.go` file and
the `go test` command.
