package cmd

import (
	"crypto/rand"
	"io"
	"path/filepath"
	"testing"

	"github.com/IbirbyZh/sync-folder/pkg/crypto"
)

func TestEncrypt(t *testing.T) {
	dir := t.TempDir()
	var key [32]byte
	if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
		t.Fatal(err)
	}

	if err := Encrypt(EncryptFlags{
		archiveFile: filepath.Join(dir, "archive"),
		sourceDir:   "../../testdata",
		keyFunc: func(s crypto.Salt, _ bool) []byte {
			return key[:]
		},
		verbose: false,
	}); err != nil {
		t.Fatal(err)
	}
}
