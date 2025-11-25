package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/szjason72/zervigo/shared/core/utils"
)

func main() {
	hash := flag.String("hash", "", "bcrypt hash to verify")
	passwords := flag.String("passwords", "", "comma-separated list of plaintext passwords to try")
	flag.Parse()

	if *hash == "" || *passwords == "" {
		fmt.Fprintf(os.Stderr, "usage: go run ./cmd/check-password --hash <bcrypt-hash> --passwords <pwd1,pwd2,...>\n")
		os.Exit(2)
	}

	list := strings.Split(*passwords, ",")
	ok := false
	for _, pwd := range list {
		pwd = strings.TrimSpace(pwd)
		if pwd == "" {
			continue
		}

		if utils.CheckPassword(pwd, *hash) {
			fmt.Printf("密码 '%s' 匹配！\n", pwd)
			ok = true
		} else {
			fmt.Printf("密码 '%s' 不匹配。\n", pwd)
		}
	}

	if !ok {
		os.Exit(1)
	}
}
