package heap

import "jvmgo/ch11/classfile"

type ExceptionTable []*ExceptionHandler

type ExceptionHandler struct {
	startPc   int //如果在 startPc 和 endPc 包围起来的那部分代码 抛出catchType类型异常，就交给handlerPc处理
	endPc     int
	handlerPc int
	catchType *ClassRef
}

// 把class文件中的异常处理表转换成ExceptionTable类型
// 异常处理项的catchType有可能是0。我们知道0是无效的常量池索引，但是在这里0并非表
// 示catch-none，而是表示catch-all
func newExceptionTable(entries []*classfile.ExceptionTableEntry, cp *ConstantPool) ExceptionTable {
	table := make([]*ExceptionHandler, len(entries))
	for i, entry := range entries {
		table[i] = &ExceptionHandler{
			startPc:   int(entry.StartPc()),
			endPc:     int(entry.EndPc()),
			handlerPc: int(entry.HandlerPc()),
			catchType: getCatchType(uint(entry.CatchType()), cp),
		}
	}

	return table
}

func getCatchType(index uint, cp *ConstantPool) *ClassRef {
	if index == 0 {
		return nil // catch all
	}
	return cp.GetConstant(index).(*ClassRef)
}

func (self ExceptionTable) findExceptionHandler(exClass *Class, pc int) *ExceptionHandler {
	for _, handler := range self {
		// jvms: The start_pc is inclusive and end_pc is exclusive
		if pc >= handler.startPc && pc < handler.endPc {
			if handler.catchType == nil {
				return handler
			}
			catchClass := handler.catchType.ResolvedClass()
			if catchClass == exClass || catchClass.IsSuperClassOf(exClass) {
				return handler
			}
		}
	}
	return nil
}
