// Package assert provides always-on assertions for Radica.
//
// Assertion failures indicate catastrophic bugs. They panic. They are not
// compiled out in release builds. Per the style guide, a wrong program is
// more dangerous than a dead one.
package assert

import "fmt"

// True panics if cond is false.
func True(cond bool, msg string) {
	if !cond {
		panic("assert: " + msg)
	}
}

// Eq panics if got != want.
func Eq[T comparable](got, want T, msg string) {
	if got != want {
		panic(fmt.Sprintf("assert: %s: got=%v want=%v", msg, got, want))
	}
}

// Nil panics if err is non-nil.
func Nil(err error) {
	if err != nil {
		panic(fmt.Sprintf("assert: expected nil error, got: %v", err))
	}
}

// NotNil panics if v is nil.
func NotNil(v any, msg string) {
	if v == nil {
		panic("assert: expected non-nil: " + msg)
	}
}

// Lt panics if a >= b.
func Lt[T int | int32 | int64 | uint | uint32 | uint64](a, b T, msg string) {
	if a >= b {
		panic(fmt.Sprintf("assert: %s: %v >= %v (expected <)", msg, a, b))
	}
}

// Implies panics if a is true and b is false.
func Implies(a, b bool, msg string) {
	if a && !b {
		panic("assert: implication failed: " + msg)
	}
}
