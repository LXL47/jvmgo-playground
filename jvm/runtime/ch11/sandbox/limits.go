package sandbox

import (
	"fmt"
	"sync/atomic"
)

type Limits struct {
	Enabled         bool
	MaxInstructions uint64
	MaxHeapBytes    uint64
	MaxArrayLength  uint64
	MaxOutputBytes  uint64
}

var current Limits
var instructions uint64
var heapBytes uint64
var outputBytes uint64

func Configure(limits Limits) {
	current = limits
	atomic.StoreUint64(&instructions, 0)
	atomic.StoreUint64(&heapBytes, 0)
	atomic.StoreUint64(&outputBytes, 0)
}

func Enabled() bool {
	return current.Enabled
}

func CountInstruction() {
	if !current.Enabled || current.MaxInstructions == 0 {
		return
	}
	used := atomic.AddUint64(&instructions, 1)
	if used > current.MaxInstructions {
		panic(fmt.Sprintf("sandbox: instruction budget exceeded (%d)", current.MaxInstructions))
	}
}

func ReserveHeap(bytes uint64) {
	if !current.Enabled || current.MaxHeapBytes == 0 {
		return
	}
	used := atomic.AddUint64(&heapBytes, bytes)
	if used > current.MaxHeapBytes {
		panic(fmt.Sprintf("sandbox: heap budget exceeded (%d bytes)", current.MaxHeapBytes))
	}
}

func CheckArrayLength(length uint64) {
	if current.Enabled && current.MaxArrayLength > 0 && length > current.MaxArrayLength {
		panic(fmt.Sprintf("sandbox: array length exceeded (%d)", current.MaxArrayLength))
	}
}

func CountOutput(bytes uint64) {
	if !current.Enabled || current.MaxOutputBytes == 0 {
		return
	}
	used := atomic.AddUint64(&outputBytes, bytes)
	if used > current.MaxOutputBytes {
		panic(fmt.Sprintf("sandbox: output budget exceeded (%d bytes)", current.MaxOutputBytes))
	}
}

func Deny(capability string) {
	if current.Enabled {
		panic("sandbox: capability denied: " + capability)
	}
}
