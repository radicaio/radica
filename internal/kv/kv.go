// Package kv provides the storage interface used by the state machine
// and one MVP implementation backed by a plain Go map.
//
// The interface exists so the storage can be swapped for a real KV engine
// later without touching the state machine or the replica.
package kv

// KV is the storage swap point. All methods take and return byte slices;
// callers must not retain the returned slice across mutations.
type KV interface {
	Get(key []byte) (value []byte, ok bool)
	Set(key, value []byte)
	Delete(key []byte) (existed bool)
	Len() int
}

// Map is the MVP implementation: a plain in-memory map.
// It is not safe for concurrent use; the replica owns it and accesses it
// only from the loop goroutine.
type Map struct {
	m map[string][]byte
}

// NewMap returns an empty Map.
func NewMap() *Map {
	return &Map{m: make(map[string][]byte)}
}

// Get returns the value for key and whether it was present.
// The returned slice is owned by the Map; do not retain it.
func (kv *Map) Get(key []byte) ([]byte, bool) {
	v, ok := kv.m[string(key)]
	return v, ok
}

// Set stores value under key, replacing any previous value.
// The value is copied; the caller may reuse the buffer.
func (kv *Map) Set(key, value []byte) {
	cp := make([]byte, len(value))
	copy(cp, value)
	kv.m[string(key)] = cp
}

// Delete removes the key. Returns true if the key was present.
func (kv *Map) Delete(key []byte) bool {
	k := string(key)
	if _, ok := kv.m[k]; !ok {
		return false
	}
	delete(kv.m, k)
	return true
}

// Len returns the number of entries.
func (kv *Map) Len() int { return len(kv.m) }
