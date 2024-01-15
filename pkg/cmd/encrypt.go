package cmd

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"syscall"

	"golang.org/x/term"

	"github.com/IbirbyZh/sync-folder/pkg/crypto"
	"github.com/IbirbyZh/sync-folder/pkg/tarball"
)

func requestPassword(doubleCheck bool) []byte {
	fmt.Println("Enter password:")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}
	if doubleCheck {
		fmt.Println("One more time:")
		passwordAgain, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			panic(err)
		}
		if !bytes.Equal(password, passwordAgain) {
			panic("passwords are not equal")
		}
	}

	return password
}

func writePipe(buf io.Writer, ef EncryptFlags) error {
	gz := gzip.NewWriter(buf)
	defer gz.Close()

	aw := tarball.NewArchiveWriter(gz, tarball.VerboseWriter(ef.verbose))
	defer aw.Close()

	if err := aw.AddFiles(ef.sourceDir); err != nil {
		return err
	}

	if err := aw.Close(); err != nil {
		return err
	}

	if err := gz.Close(); err != nil {
		return err
	}
	return nil
}

func loadArgon2Settings(filePath string) crypto.Argon2Settings {
	if filePath == "" {
		return crypto.DefaultArgon2Settings()
	} else {
		a2s, err := crypto.LoadArgon2Settings(filePath)
		if err != nil {
			panic(err)
		}
		return a2s
	}
}

type EncryptFlags struct {
	archiveFile string
	sourceDir   string
	salt        crypto.Salt
	keyFunc     func(crypto.Salt, bool) []byte
	verbose     bool
}

func ParseEncryptFlags() EncryptFlags {
	ef := EncryptFlags{verbose: true}
	flag.StringVar(&ef.sourceDir, "dir", "", "directory to archive")
	flag.StringVar(&ef.archiveFile, "out", "", "result file")
	a2sFile := flag.String("a2s", "", "file with argon2 settings")
	flag.Parse()
	if ef.sourceDir == "" {
		panic("You have to specify directory to archive via -dir= option")
	}
	if ef.archiveFile == "" {
		panic("You have to specify result file via -out= option")
	}

	salt, err := crypto.GenerateSalt()
	if err != nil {
		panic(err)
	}
	ef.salt = salt

	a2s := loadArgon2Settings(*a2sFile)

	password := requestPassword(true)
	ef.keyFunc = func(s crypto.Salt, verbose bool) []byte {
		return crypto.GenerateArgonKey(password, s, a2s, verbose)
	}

	return ef
}

func Encrypt(ef EncryptFlags) error {
	encrypter, err := crypto.NewEncrypter(ef.keyFunc(ef.salt, ef.verbose))
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	err = writePipe(&buf, ef)
	if err != nil {
		return err
	}

	encrypted, err := encrypter.EncryptMessage(buf.Bytes())
	if err != nil {
		return err
	}

	file, err := os.OpenFile(ef.archiveFile, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(ef.salt[:]); err != nil {
		return err
	}
	if _, err := file.Write(encrypted); err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return testDecrypt(ef)
}

func testDecrypt(ef EncryptFlags) error {
	return Decrypt(DecryptFlags{
		archiveFile: ef.archiveFile,
		keyFunc:     ef.keyFunc,
		resultDir:   "", // dry run
	})
}
