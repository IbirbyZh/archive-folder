package main

import (
	"github.com/IbirbyZh/sync-folder/pkg/cmd"
)

func main() {
	if err := cmd.Decrypt(cmd.ParseDecryptFlags()); err != nil {
		panic(err)
	}
}
