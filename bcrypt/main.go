package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	passwd := "hello world"
	buf, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Println("Hashed passwd:", string(buf))

	err = bcrypt.CompareHashAndPassword(buf, []byte(passwd))
	if err != nil {
		fmt.Println("passwd does not match.")
	} else {
		fmt.Println("passwd matches.")
	}
}
