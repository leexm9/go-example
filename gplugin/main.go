package main

import (
	"fmt"
	"os"
	"plugin"
	"time"
)

// go build -buildmode=plugin -o outname.so
// go build -buildmode=plugin -gcflags="all=-N -l" -o outname.so  debug模式
func main() {
	path, _ := os.Getwd()
	p, err := plugin.Open(fmt.Sprintf("%s/gplugin/module/libtest.so", path))
	if err != nil {
		panic(err)
	}
	greet, err := p.Lookup("Greet")
	if err != nil {
		panic(err)
	}
	greet.(func(string))("Go")

	time.Sleep(60 * time.Second)

	p, err = plugin.Open(fmt.Sprintf("%s/gplugin/module/libtest.so", path))
	if err != nil {
		panic(err)
	}

	add, err := p.Lookup("Add")
	if err != nil {
		panic(err)
	}
	f := add.(func(int, int) int)(2, 4)
	fmt.Println(f)
}
