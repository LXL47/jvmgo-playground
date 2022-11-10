package references

import "reflect"
import "jvmgo/ch11/instructions/base"
import "jvmgo/ch11/rtda"
import "jvmgo/ch11/rtda/heap"

// Throw exception or error
type ATHROW struct{ base.NoOperandsInstruction }

func (self *ATHROW) Execute(frame *rtda.Frame) {
	// 从操作数栈中弹出异常对象引用
	ex := frame.OperandStack().PopRef()
	if ex == nil {
		panic("java.lang.NullPointerException")
	}

	thread := frame.Thread()
	//
	if !findAndGotoExceptionHandler(thread, ex) {
		handleUncaughtException(thread, ex)
	}
}

// 从当前帧开始，遍历Java虚拟机栈，查找方法的异常处理表。
//假设遍历到帧F，如果在F对应的方法中找不到异常处理项，则把F
//弹出，继续遍历。反之如果找到了异常处理项，在跳转到异常处理
//代码之前，要先把F的操作数栈清空，然后把异常对象引用推入栈
//顶
func findAndGotoExceptionHandler(thread *rtda.Thread, ex *heap.Object) bool {
	for {
		frame := thread.CurrentFrame()
		pc := frame.NextPC() - 1

		handlerPC := frame.Method().FindExceptionHandler(ex.Class(), pc)
		if handlerPC > 0 {
			stack := frame.OperandStack()
			stack.Clear()
			stack.PushRef(ex)
			frame.SetNextPC(handlerPC)
			return true
		}

		thread.PopFrame()
		if thread.IsStackEmpty() {
			break
		}
	}
	return false
}


// 如果遍历完Java虚拟机栈还是找不到异常处理代码，则
//handleUncaughtException（）函数打印出Java虚拟机栈信息
// todo
func handleUncaughtException(thread *rtda.Thread, ex *heap.Object) {
	thread.ClearStack()

	jMsg := ex.GetRefVar("detailMessage", "Ljava/lang/String;")
	goMsg := heap.GoString(jMsg)
	println(ex.Class().JavaName() + ": " + goMsg)

	stes := reflect.ValueOf(ex.Extra())
	for i := 0; i < stes.Len(); i++ {
		ste := stes.Index(i).Interface().(interface {
			String() string
		})
		println("\tat " + ste.String())
	}
}
