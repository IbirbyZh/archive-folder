package main

import (
	"github.com/IbirbyZh/sync-folder/pkg/cmd"
)

func main() {
	if err := cmd.Encrypt(cmd.ParseEncryptFlags()); err != nil {
		panic(err)
	}
}
