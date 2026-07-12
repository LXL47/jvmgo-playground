package heap

import "jvmgo/ch11/sandbox"

func (self *Class) IsArray() bool {
	return self.name[0] == '['
}

func (self *Class) ComponentClass() *Class {
	componentClassName := getComponentClassName(self.name)
	return self.loader.LoadClass(componentClassName)
}

// 创建新数组对象
func (self *Class) NewArray(count uint) *Object {
	if !self.IsArray() {
		panic("Not array class: " + self.name)
	}
	sandbox.CheckArrayLength(uint64(count))
	sandbox.ReserveHeap(32 + uint64(count)*self.arrayElementSize())
	switch self.Name() {
	// 布尔数组是使用字节数组来表示的
	case "[Z":
		return &Object{self, make([]int8, count), nil}
	case "[B":
		return &Object{self, make([]int8, count), nil}
	case "[C":
		return &Object{self, make([]uint16, count), nil}
	case "[S":
		return &Object{self, make([]int16, count), nil}
	case "[I":
		return &Object{self, make([]int32, count), nil}
	case "[J":
		return &Object{self, make([]int64, count), nil}
	case "[F":
		return &Object{self, make([]float32, count), nil}
	case "[D":
		return &Object{self, make([]float64, count), nil}
	default:
		return &Object{self, make([]*Object, count), nil}
	}
}

func (self *Class) arrayElementSize() uint64 {
	switch self.Name() {
	case "[Z", "[B":
		return 1
	case "[C", "[S":
		return 2
	case "[I", "[F":
		return 4
	default:
		return 8
	}
}

func NewByteArray(loader *ClassLoader, bytes []int8) *Object {
	sandbox.CheckArrayLength(uint64(len(bytes)))
	sandbox.ReserveHeap(32 + uint64(len(bytes)))
	return &Object{loader.LoadClass("[B"), bytes, nil}
}
