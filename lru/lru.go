package lru

type LRU struct {
	// whatever fields you want here
}

func NewLru(limit int) *LRU {
	return new(LRU)
}

func (lru *LRU) MaxStorage() int {
	return 0
}

func (lru *LRU) RemainingStorage() int {
	return 0
}

func (lru *LRU) Get(key string) (value []byte, ok bool) {
	return nil, false
}

func (lru *LRU) Remove(key string) (value []byte, ok bool) {
	return nil, false
}

func (lru *LRU) Set(key string, value []byte) bool {
	return false
}

func (lru *LRU) Len() int {
	return 0
}
