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

func writePipe(buf io.Writer, root string) error {
	gz := gzip.NewWriter(buf)
	defer gz.Close()

	aw := tarball.NewArchiveWriter(gz, tarball.VerboseWriter(true))
	defer aw.Close()

	if err := aw.AddFiles(root); err != nil {
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

type EncryptFlags struct {
	archiveFile string
	sourceDir   string
	password    []byte
}

func ParseEncryptFlags() EncryptFlags {
	ef := EncryptFlags{}
	flag.StringVar(&ef.sourceDir, "dir", "", "directory to archive")
	flag.StringVar(&ef.archiveFile, "out", "", "result file")
	flag.Parse()
	if ef.sourceDir == "" {
		panic("You have to specify directory to archive via -dir= option")
	}
	if ef.archiveFile == "" {
		panic("You have to specify result file via -out= option")
	}

	ef.password = requestPassword(true)

	return ef
}

func Encrypt(ef EncryptFlags) error {
	encrypter, err := crypto.NewEncrypter(ef.password)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	err = writePipe(&buf, ef.sourceDir)
	if err != nil {
		return err
	}

	encrypted, err := encrypter.EncryptMessage(buf.Bytes())
	if err != nil {
		return err
	}

	file, err := os.OpenFile(ef.archiveFile, os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

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
		password:    ef.password,
		resultDir:   "", // dry run
	})
}
