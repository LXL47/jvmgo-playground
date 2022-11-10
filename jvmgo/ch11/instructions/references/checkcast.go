package references

import "jvmgo/ch11/instructions/base"
import "jvmgo/ch11/rtda"
import "jvmgo/ch11/rtda/heap"

// 判断对象是否为某个类
type CHECK_CAST struct{ base.Index16Instruction }


// checkcast指令和instanceof指令很像，区别在于：instanceof指令
//会改变操作数栈（弹出对象引用，推入判断结果）；checkcast则不改
//变操作数栈（如果判断失败，直接抛出ClassCastException异常）

func (self *CHECK_CAST) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	ref := stack.PopRef()
	stack.PushRef(ref)
	if ref == nil {
		return
	}

	cp := frame.Method().Class().ConstantPool()
	classRef := cp.GetConstant(self.Index).(*heap.ClassRef)
	class := classRef.ResolvedClass()
	if !ref.IsInstanceOf(class) {
		panic("java.lang.ClassCastException")
	}
}
