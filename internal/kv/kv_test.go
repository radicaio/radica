package kv

import (
	"bytes"
	"testing"
)

// TestMap_GetSetDelete verifies the basic happy path for the three ops.
func TestMap_GetSetDelete(t *testing.T) {
	kv := NewMap()

	// Get on empty map.
	if _, ok := kv.Get([]byte("a")); ok {
		t.Fatal("expected miss on empty map")
	}

	// Set then Get.
	kv.Set([]byte("a"), []byte("alpha"))
	v, ok := kv.Get([]byte("a"))
	if !ok || !bytes.Equal(v, []byte("alpha")) {
		t.Fatalf("got=(%q,%v) want=(alpha,true)", v, ok)
	}

	// Overwrite.
	kv.Set([]byte("a"), []byte("beta"))
	v, _ = kv.Get([]byte("a"))
	if !bytes.Equal(v, []byte("beta")) {
		t.Fatalf("overwrite: got=%q want=beta", v)
	}

	// Delete present.
	if !kv.Delete([]byte("a")) {
		t.Fatal("delete: expected existed=true")
	}

	// Delete missing.
	if kv.Delete([]byte("a")) {
		t.Fatal("delete: expected existed=false on repeat")
	}

	// Get after Delete.
	if _, ok := kv.Get([]byte("a")); ok {
		t.Fatal("expected miss after delete")
	}
}

// TestMap_SetCopiesValue ensures Set defensively copies the value, so the
// caller can reuse the buffer.
func TestMap_SetCopiesValue(t *testing.T) {
	kv := NewMap()
	val := []byte("first")
	kv.Set([]byte("k"), val)

	// Mutate the caller's buffer; Map must be unaffected.
	val[0] = 'X'
	got, _ := kv.Get([]byte("k"))
	if !bytes.Equal(got, []byte("first")) {
		t.Fatalf("Map did not copy value on Set: got=%q", got)
	}
}
