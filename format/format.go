package main

import "fmt"

type People struct {
	Name string
}

func main() {
	p := People{Name: "Jack"}

	fmt.Printf("%%v: %v \n", p)   // {Jack}
	fmt.Printf("%%+v: %+v \n", p) // {Name:Jack}
	fmt.Printf("%%#v: %#v \n", p) // main.People{Name:"Jack"}
	fmt.Printf("%%T: %T \n", p)   // main.People
	fmt.Printf("%%p: %p \n", &p)  // 指针，十六进制表示，0x 前缀
	fmt.Println()

	// 布尔占位符
	fmt.Printf("%%t: %t \n", true)
	fmt.Println()

	// 整数占位符
	fmt.Printf("%%b: %b \n", 13) // 二进制
	fmt.Printf("%%o: %o \n", 13) // 八进制
	fmt.Printf("%%d: %d \n", 13) // 十进制
	fmt.Printf("%%x: %x \n", 13) // 十六进制，小写字母
	fmt.Printf("%%X: %X \n", 13) // 十六进制，大写字母
	fmt.Println()

	fmt.Printf("%%q: %q \n", 123) // 单引号围绕的字符字面量，由 go 语法安全的转义
	fmt.Printf("%%c: %c \n", 123) // 相应的 unicode 码点表示的字符
	fmt.Printf("%%U: %U \n", 123) // unicode 格式：U+1234，等同于 "U+%04X"
	fmt.Printf("U+%04X \n", 123)
	fmt.Println()

	// 浮点数
	fmt.Printf("%%e: %e \n", 10.2) // 1.020000e+01
	fmt.Printf("%%E: %E \n", 10.2) // 1.020000E+01
	fmt.Printf("%%f: %f \n", 10.2) // 10.200000
	// 根据情况选择 %e 或 %f 以产生更紧凑的（无末尾的 0）输出
	fmt.Printf("%%g: %g \n", 10.20) // 10.2
	// 根据情况选择 %E 或 %f 以产生更紧凑的（无末尾的 0）输出
	fmt.Printf("%%G: %G \n", 10.20+2.0i) // (10.2+2i)
	fmt.Println()

	// 字符串
	fmt.Printf("%%s: %s \n", "golang") // golang
	// 双引号围绕的字符串，由 go 安全转义
	fmt.Printf("%%q: %q \n", "golang") // "golang"
	// 十六进制表示
	fmt.Printf("%%x: %x \n", "golang") // 676f6c616e67
	fmt.Printf("%%X: %X \n", "golang") // 676F6C616E67
	fmt.Println()

	fmt.Printf("%%+q: %+q \n", "中") // "\u4e2d"
	fmt.Printf("%%#U: %#U \n", '中') // U+4E2D '中'
}
