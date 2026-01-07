package main

import (
	"flag"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := flag.String("password", "", "plaintext password to hash")
	flag.Parse()

	if *password == "" {
		fmt.Fprintf(os.Stderr, "usage: go run ./cmd/generate-hash --password <plaintext-password>\n")
		os.Exit(2)
	}

	// 生成 bcrypt 哈希值
	hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "生成哈希失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("明文密码: %s\n", *password)
	fmt.Printf("Bcrypt哈希: %s\n", string(hash))
}

