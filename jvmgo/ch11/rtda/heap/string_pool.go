package heap

import "unicode/utf16"


// 字符串常量池
// key是go字符串，value是java字符串
var internedStrings = map[string]*Object{}

// todo
// go string -> java.lang.String
// JString（）函数根据Go字符串返回相应的Java字符串实例
func JString(loader *ClassLoader, goStr string) *Object {
	// 如果Java字符串已经在池中，直接返回即可
	if internedStr, ok := internedStrings[goStr]; ok {
		return internedStr
	}

	// 先把Go字符串（UTF8格式）转换成Java字符数组（UTF16格式）
	chars := stringToUtf16(goStr)
	jChars := &Object{loader.LoadClass("[C"), chars, nil}

	// 然后创建一个Java字符串实例
	jStr := loader.LoadClass("java/lang/String").NewObject()
	// 把它的value变量设置成刚刚转换而来的字符数组
	jStr.SetRefVar("value", "[C", jChars)

	// 最后把Java字符串放入池中
	internedStrings[goStr] = jStr
	return jStr
}

// java.lang.String -> go string
func GoString(jStr *Object) string {
	charArr := jStr.GetRefVar("value", "[C")
	return utf16ToString(charArr.Chars())
}

// utf8 -> utf16
func stringToUtf16(s string) []uint16 {
	runes := []rune(s)         // utf32
	return utf16.Encode(runes) // func Encode(s []rune) []uint16
}

// utf16 -> utf8
func utf16ToString(s []uint16) string {
	runes := utf16.Decode(s) // func Decode(s []uint16) []rune
	return string(runes)
}

// todo
func InternString(jStr *Object) *Object {
	goStr := GoString(jStr)
	if internedStr, ok := internedStrings[goStr]; ok {
		return internedStr
	}

	internedStrings[goStr] = jStr
	return jStr
}
