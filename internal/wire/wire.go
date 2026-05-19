// Package wire defines the operation set and reply codes shared by the
// state machine, transport, and client. There is no on-the-wire framing in
// this MVP iteration — the binary protocol will land alongside the
// standalone TCP listener. For now, Op and Reply are passed in-process.
package wire

// OpCode identifies the kind of operation.
type OpCode uint8

const (
	OpGet OpCode = iota + 1
	OpSet
	OpDelete
)

// Status is the result of an Apply call.
type Status uint8

const (
	StatusOK Status = iota
	// StatusMiss is returned by Get when the key is absent and by Delete
	// when the key did not exist.
	StatusMiss
	StatusInvalid
)

// Op is one decoded operation, fed to StateMachine.Apply.
type Op struct {
	Code  OpCode
	Key   []byte
	Value []byte // unused for Get and Delete
}

// Reply is the result of one Apply call.
type Reply struct {
	Status Status
	Value  []byte // populated by Get on a hit
}
