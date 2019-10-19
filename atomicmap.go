package main

import (
	"fmt"
	"sort"
	"sync"
)

type AtomicByteMap struct {
	data map[string]int64
	mu sync.Mutex
}

type Pair struct {
	Key []byte
	Value int64
}

type PairList []Pair

func (p PairList) Len() int { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }

func NewAtomicByteMap() AtomicByteMap {
	return AtomicByteMap{make(map[string]int64), sync.Mutex{}}
}

// GETTERS

func (r *AtomicByteMap) HasValue(b []byte) bool {
	_, okay := r.data[string(b)]
	return okay
}

func (r *AtomicByteMap) HasStringValue( s string) bool {
	_, okay := r.data[s]
	return okay
}

func (r *AtomicByteMap) GetByteCounter(b []byte) (int64, bool) {
	val, ok := r.data[string(b)]
	if !ok {
		val = -1
	}
	return val, ok
}

func (r *AtomicByteMap) GetStringCounter(s string) (int64, bool) {
	val, ok := r.data[s]
	if !ok {
		val = -1
	}
	return val, ok
}

func (r *AtomicByteMap) NumItems() int {
	return len(r.data)
}

func (r *AtomicByteMap) IsEmpty() bool {
	return r.NumItems() == 0
}

func (r *AtomicByteMap) Keys() [][]byte {
	keys := make([][]byte, len(r.data))
	i := 0
	for k := range r.data {
		keys[i] = []byte(k)
		i++
	}
	return keys
}

func (r *AtomicByteMap) StringKeys() []string {
	keys := make([]string, len(r.data))
	i := 0
	for k := range r.data {
		keys[i] = k
		i++
	}
	return keys
}

func (r *AtomicByteMap) Print() {
	for k, v := range r.data {
		fmt.Printf("%s: %d\n", k, v)
	}
}

// SETTERS

func (r *AtomicByteMap) AddOrIncrementByte(b []byte, i int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	val, okay := r.data[string(b)]
	if okay {
		r.data[string(b)] = val + 1
	} else {
		r.data[string(b)] = i
	}
}

func (r *AtomicByteMap) AddOrIncrementString(s string, i int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	val, okay := r.data[s]
	if okay {
		r.data[s] = val + 1
	} else {
		r.data[s] = i
	}
}

func (r *AtomicByteMap) IncrementByte(b []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	val, okay := r.data[string(b)]
	if okay {
		r.data[string(b)] = val + 1
	} else {
		r.data[string(b)] = 1
	}
}

func (r *AtomicByteMap) IncrementString(s string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	val, okay := r.data[s]
	if okay {
		r.data[s] = val + 1
	} else {
		r.data[s] = 1
	}
}

// Destroyers

func (r *AtomicByteMap) SortedValues() PairList {
	pl := make(PairList, len(r.data))
	i := 0
	for k, v := range r.data {
		pl[i] = Pair{[]byte(k), v}
		delete(r.data, k) // Save Memory
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

func (r* AtomicByteMap) Erase() {
	for k, _ := range r.data {
		delete(r.data, k)
	}
}