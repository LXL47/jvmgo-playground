package lang

import "runtime"
import "jvmgo/ch11/native"
import "jvmgo/ch11/rtda"
import "jvmgo/ch11/sandbox"

const jlRuntime = "java/lang/Runtime"

func init() {
	native.Register(jlRuntime, "availableProcessors", "()I", availableProcessors)
}

// public native int availableProcessors();
// ()I
func availableProcessors(frame *rtda.Frame) {
	numCPU := runtime.NumCPU()
	if sandbox.Enabled() {
		numCPU = 1
	}

	stack := frame.OperandStack()
	stack.PushInt(int32(numCPU))
}
