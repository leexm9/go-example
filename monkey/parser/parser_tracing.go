package parser

import (
	"fmt"
	"strings"
)

var traceLevel = 0

const traceIdentPlaceholder string = "\t"

func trace(msg string) string {
	incIdent()
	tracePrint("BEGIN " + msg)
	return msg
}

func untrace(msg string) {
	tracePrint("END: " + msg)
	decIdent()
}

func incIdent() {
	traceLevel += 1
}

func decIdent() {
	traceLevel -= 1
}

func tracePrint(fs string) {
	fmt.Printf("%s%s\n", identLevel(), fs)
}

func identLevel() string {
	return strings.Repeat(traceIdentPlaceholder, traceLevel-1)
}
