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

func readPipe(buf io.Reader, resultDir string) error {
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return err
	}
	defer gz.Close()

	ar := tarball.NewArchiveReader(
		gz,
		tarball.VerboseReader(true),
		tarball.DryRun(resultDir == ""),
	)

	if err := ar.ExtractFiles(resultDir); err != nil {
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
	password    []byte
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

	df.password = requestPassword(false)

	return df
}

func Decrypt(df DecryptFlags) error {
	encrypter, err := crypto.NewEncrypter(df.password)
	if err != nil {
		return err
	}

	file, err := os.Open(df.archiveFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		return err
	}

	decrypted, err := encrypter.DecryptMessage(buf.Bytes())
	if err != nil {
		return err
	}

	if err = readPipe(bytes.NewBuffer(decrypted), df.resultDir); err != nil {
		return err
	}

	return nil
}
