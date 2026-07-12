package heap

import "jvmgo/ch11/classfile"

type Method struct {
	ClassMember
	maxStack                uint
	maxLocals               uint
	code                    []byte
	exceptionTable          ExceptionTable                      // todo: rename
	lineNumberTable         *classfile.LineNumberTableAttribute // 行号
	exceptions              *classfile.ExceptionsAttribute      // todo: rename
	parameterAnnotationData []byte                              // RuntimeVisibleParameterAnnotations_attribute
	annotationDefaultData   []byte                              // AnnotationDefault_attribute
	parsedDescriptor        *MethodDescriptor
	argSlotCount            uint // 参数槽的数量，不一定和方法的参数数量一样（long和double类型的参数占两个槽）（实例方法多一个this参数）
}

func newMethods(class *Class, cfMethods []*classfile.MemberInfo) []*Method {
	methods := make([]*Method, len(cfMethods))
	for i, cfMethod := range cfMethods {
		methods[i] = newMethod(class, cfMethod)
	}
	return methods
}

func newMethod(class *Class, cfMethod *classfile.MemberInfo) *Method {
	method := &Method{}
	method.class = class
	method.copyMemberInfo(cfMethod)
	method.copyAttributes(cfMethod)
	md := parseMethodDescriptor(method.descriptor)
	method.parsedDescriptor = md
	method.calcArgSlotCount(md.parameterTypes)

	// 如果是本地方法，添加一些信息
	if method.IsNative() {
		method.injectCodeAttribute(md.returnType)
	}
	return method
}

func (self *Method) copyAttributes(cfMethod *classfile.MemberInfo) {
	if codeAttr := cfMethod.CodeAttribute(); codeAttr != nil {
		self.maxStack = codeAttr.MaxStack()
		self.maxLocals = codeAttr.MaxLocals()
		self.code = codeAttr.Code()
		self.lineNumberTable = codeAttr.LineNumberTableAttribute()
		self.exceptionTable = newExceptionTable(codeAttr.ExceptionTable(),
			self.class.constantPool)
	}
	self.exceptions = cfMethod.ExceptionsAttribute()
	self.annotationData = cfMethod.RuntimeVisibleAnnotationsAttributeData()
	self.parameterAnnotationData = cfMethod.RuntimeVisibleParameterAnnotationsAttributeData()
	self.annotationDefaultData = cfMethod.AnnotationDefaultAttributeData()
}

// 计算参数槽的数量，不一定和方法的参数数量一样（long和double类型的参数占两个槽）（实例方法多一个this参数）
func (self *Method) calcArgSlotCount(paramTypes []string) {
	for _, paramType := range paramTypes {
		self.argSlotCount++
		if paramType == "J" || paramType == "D" {
			self.argSlotCount++
		}
	}
	if !self.IsStatic() {
		self.argSlotCount++ // `this` reference
	}
}

// 本地方法帧的操作数栈至少要能容纳返回值，
// 为了简化代码，暂时给maxStack字段赋值为4。
// 因为本地方法帧的局部变量表只用来存放参数值，所以把argSlotCount赋给maxLocals字段刚好。
// 至于code字段，也就是本地方法的字节码，第一条指令
// 都是0xFE，第二条指令则根据函数的返回值选择相应的返回指令
func (self *Method) injectCodeAttribute(returnType string) {
	self.maxStack = 4 // todo
	self.maxLocals = self.argSlotCount
	switch returnType[0] {
	case 'V':
		self.code = []byte{0xfe, 0xb1} // return
	case 'L', '[':
		self.code = []byte{0xfe, 0xb0} // areturn
	case 'D':
		self.code = []byte{0xfe, 0xaf} // dreturn
	case 'F':
		self.code = []byte{0xfe, 0xae} // freturn
	case 'J':
		self.code = []byte{0xfe, 0xad} // lreturn
	default:
		self.code = []byte{0xfe, 0xac} // ireturn
	}
}

func (self *Method) IsSynchronized() bool {
	return 0 != self.accessFlags&ACC_SYNCHRONIZED
}
func (self *Method) IsBridge() bool {
	return 0 != self.accessFlags&ACC_BRIDGE
}
func (self *Method) IsVarargs() bool {
	return 0 != self.accessFlags&ACC_VARARGS
}
func (self *Method) IsNative() bool {
	return 0 != self.accessFlags&ACC_NATIVE
}
func (self *Method) IsAbstract() bool {
	return 0 != self.accessFlags&ACC_ABSTRACT
}
func (self *Method) IsStrict() bool {
	return 0 != self.accessFlags&ACC_STRICT
}

// getters
func (self *Method) MaxStack() uint {
	return self.maxStack
}
func (self *Method) MaxLocals() uint {
	return self.maxLocals
}
func (self *Method) Code() []byte {
	return self.code
}
func (self *Method) ParameterAnnotationData() []byte {
	return self.parameterAnnotationData
}
func (self *Method) AnnotationDefaultData() []byte {
	return self.annotationDefaultData
}
func (self *Method) ParsedDescriptor() *MethodDescriptor {
	return self.parsedDescriptor
}
func (self *Method) ArgSlotCount() uint {
	return self.argSlotCount
}

func (self *Method) FindExceptionHandler(exClass *Class, pc int) int {
	handler := self.exceptionTable.findExceptionHandler(exClass, pc)
	if handler != nil {
		return handler.handlerPc
	}
	return -1
}

// 找行号，本地方法没有字节码，返回-2，没有行号返回-1
func (self *Method) GetLineNumber(pc int) int {
	if self.IsNative() {
		return -2
	}
	if self.lineNumberTable == nil {
		return -1
	}
	return self.lineNumberTable.GetLineNumber(pc)
}

func (self *Method) isConstructor() bool {
	return !self.IsStatic() && self.name == "<init>"
}
func (self *Method) isClinit() bool {
	return self.IsStatic() && self.name == "<clinit>"
}

// reflection
func (self *Method) ParameterTypes() []*Class {
	if self.argSlotCount == 0 {
		return nil
	}

	paramTypes := self.parsedDescriptor.parameterTypes
	paramClasses := make([]*Class, len(paramTypes))
	for i, paramType := range paramTypes {
		paramClassName := toClassName(paramType)
		paramClasses[i] = self.class.loader.LoadClass(paramClassName)
	}

	return paramClasses
}
func (self *Method) ReturnType() *Class {
	returnType := self.parsedDescriptor.returnType
	returnClassName := toClassName(returnType)
	return self.class.loader.LoadClass(returnClassName)
}
func (self *Method) ExceptionTypes() []*Class {
	if self.exceptions == nil {
		return nil
	}

	exIndexTable := self.exceptions.ExceptionIndexTable()
	exClasses := make([]*Class, len(exIndexTable))
	cp := self.class.constantPool

	for i, exIndex := range exIndexTable {
		classRef := cp.GetConstant(uint(exIndex)).(*ClassRef)
		exClasses[i] = classRef.ResolvedClass()
	}

	return exClasses
}
