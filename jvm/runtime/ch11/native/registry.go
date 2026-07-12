package native

import "jvmgo/ch11/rtda"

type NativeMethod func(frame *rtda.Frame)

// 本地方法表
var registry = map[string]NativeMethod{}

func emptyNativeMethod(frame *rtda.Frame) {
	// do nothing
}

func Register(className, methodName, methodDescriptor string, method NativeMethod) {
	key := className + "~" + methodName + "~" + methodDescriptor
	registry[key] = method
}

// java.lang.Object等类是通过一个叫作registerNatives（）的本地方法来
// 注册其他本地方法的。在本章和后面的章节中，将自己注册所有的
// 本地方法实现。所以像registerNatives（）这样的方法就没有太大的用
// 处。为了避免重复代码，这里统一处理，如果遇到这样的本地方
// 法，就返回一个空的实现，
func FindNativeMethod(className, methodName, methodDescriptor string) NativeMethod {
	key := className + "~" + methodName + "~" + methodDescriptor
	if method, ok := registry[key]; ok {
		return method
	}
	if methodDescriptor == "()V" {
		if methodName == "registerNatives" || methodName == "initIDs" {
			return emptyNativeMethod
		}
	}
	return nil
}
