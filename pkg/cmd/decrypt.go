package cmd

import (
	"bytes"
	"compress/gzip"
	"flag"
	"io"
	"os"

	"github.com/IbirbyZh/sync-folder/pkg/crypto"
	"github.com/IbirbyZh/sync-folder/pkg/tarball"
)

func readPipe(buf io.Reader, df DecryptFlags) error {
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return err
	}
	defer gz.Close()

	ar := tarball.NewArchiveReader(
		gz,
		tarball.VerboseReader(df.verbose),
		tarball.DryRun(df.resultDir == ""),
	)

	if err := ar.ExtractFiles(df.resultDir); err != nil {
		return err
	}

	if err := gz.Close(); err != nil {
		return err
	}
	return nil
}

type DecryptFlags struct {
	archiveFile string
	resultDir   string
	keyFunc     func(crypto.Salt, bool) []byte
	verbose     bool
}

func ParseDecryptFlags() DecryptFlags {
	df := DecryptFlags{}
	flag.StringVar(&df.archiveFile, "in", "", "archive file")
	flag.StringVar(&df.resultDir, "dir", "", "result directory")
	flag.Parse()
	if df.archiveFile == "" {
		panic("You have to specify archive via -in= option")
	}
	if df.resultDir == "" {
		panic("You have to specify result directory via -dir= option")
	}
	df.verbose = true
	password := requestPassword(false)
	df.keyFunc = func(s crypto.Salt, verbose bool) []byte { return crypto.GenerateArgonKey(password, s, verbose) }

	return df
}

func Decrypt(df DecryptFlags) error {
	file, err := os.Open(df.archiveFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var salt crypto.Salt
	if _, err := io.ReadFull(file, salt[:]); err != nil {
		return err
	}

	encrypter, err := crypto.NewEncrypter(df.keyFunc(salt, df.verbose))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		return err
	}

	decrypted, err := encrypter.DecryptMessage(buf.Bytes())
	if err != nil {
		return err
	}

	if err = readPipe(bytes.NewBuffer(decrypted), df); err != nil {
		return err
	}

	return nil
}
