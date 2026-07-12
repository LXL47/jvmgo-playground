package references

import "jvmgo/ch11/instructions/base"
import "jvmgo/ch11/rtda"
import "jvmgo/ch11/rtda/heap"

type INSTANCE_OF struct{ base.Index16Instruction }

// instanceof指令需要两个操作数。
// 第一个操作数是uint16索引，
// 从方法的字节码中获取，通过这个索引可以从当前类的运行时常量
// 池中找到一个类符号引用。
// 第二个操作数是对象引用，从操作数栈中弹出
func (self *INSTANCE_OF) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	ref := stack.PopRef()

	// 如果对象引用是null，放入0（意思是false），return
	if ref == nil {
		stack.PushInt(0)
		return
	}

	cp := frame.Method().Class().ConstantPool()
	classRef := cp.GetConstant(self.Index).(*heap.ClassRef)
	class := classRef.ResolvedClass()
	if ref.IsInstanceOf(class) {
		stack.PushInt(1)
	} else {
		stack.PushInt(0)
	}
}
