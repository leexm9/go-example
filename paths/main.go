package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	fmt.Printf("workingDir: %s \n", getWorkingDirPath())
	fmt.Printf("abs: %s \n", getAbsPath())
	fmt.Printf("exec: %s \n", getExecPath())
}

func getWorkingDirPath() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}

func getAbsPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return dir
}

func getExecPath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	dir := filepath.Dir(ex)
	return dir
}
