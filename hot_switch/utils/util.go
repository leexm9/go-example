package utils

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

const Charset = "abcdefghijklmnopqrstuvwxzyABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seedRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomString(length int) string {
	b := make([]byte, length)
	n := len(Charset)
	for i := range b {
		b[i] = Charset[seedRand.Intn(n)]
	}
	return string(b)
}

func IsDirectory(dir, alt string) error {
	if strings.TrimSpace(dir) == "" {
		return fmt.Errorf("%s cannont be empty", alt)
	}

	if stat, err := os.Stat(dir); err != nil {
		return err
	} else if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	} else {
		return nil
	}
}
