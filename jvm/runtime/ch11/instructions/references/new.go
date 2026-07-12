package references

import "jvmgo/ch11/instructions/base"
import "jvmgo/ch11/rtda"
import "jvmgo/ch11/rtda/heap"

// new指令专门用来创建类实例。数组由专门的指令创建
type NEW struct{ base.Index16Instruction }

// new指令的操作数是一个uint16索引，来自字节码。通过这个索
// 引，可以从当前类的运行时常量池中找到一个类符号引用。解析这
// 个类符号引用，拿到类数据，然后创建对象，并把对象引用推入栈
// 顶，new指令的工作就完成了
func (self *NEW) Execute(frame *rtda.Frame) {
	cp := frame.Method().Class().ConstantPool()
	classRef := cp.GetConstant(self.Index).(*heap.ClassRef)
	class := classRef.ResolvedClass()
	if !class.InitStarted() {
		frame.RevertNextPC()

		// 初始化类
		base.InitClass(frame.Thread(), class)
		return
	}

	if class.IsInterface() || class.IsAbstract() {
		panic("java.lang.InstantiationError")
	}

	ref := class.NewObject()
	frame.OperandStack().PushRef(ref)
}
