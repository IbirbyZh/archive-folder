package tarball

import (
	"archive/tar"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

type ArchiveReader struct {
	verbose   bool
	dryRun    bool
	tarReader *tar.Reader
}

func NewArchiveReader(r io.Reader, options ...func(*ArchiveReader)) *ArchiveReader {
	ar := &ArchiveReader{
		verbose:   false,
		dryRun:    false,
		tarReader: tar.NewReader(r),
	}

	for _, o := range options {
		o(ar)
	}

	return ar
}

func VerboseReader(verbose bool) func(*ArchiveReader) {
	return func(r *ArchiveReader) {
		r.verbose = verbose
	}
}

func DryRun(dryRun bool) func(*ArchiveReader) {
	return func(r *ArchiveReader) {
		r.dryRun = dryRun
	}
}

func (ar *ArchiveReader) ExtractFiles(resultFolder string) error {
	if !ar.dryRun {
		if _, err := os.Stat(resultFolder); err == nil {
			return fmt.Errorf("path %v already exists", resultFolder)
		}
		if err := os.MkdirAll(filepath.Dir(resultFolder), 0750); err != nil {
			return err
		}
		if err := os.Mkdir(resultFolder, 0700); err != nil {
			return err
		}
	}

	md5Info := make(md5InfoType)
	var archiveMD5Info md5InfoType = nil
	for {
		hdr, err := ar.tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		n, err := io.Copy(&buf, ar.tarReader)
		if err != nil {
			return err
		}
		if n != hdr.Size {
			return fmt.Errorf("wrong size of file %v", hdr.Name)
		}

		if hdr.Name == md5InfoFile {
			if archiveMD5Info != nil {
				return fmt.Errorf("more than one %v in archive", md5InfoFile)
			}
			archiveMD5Info = make(md5InfoType)
			json.Unmarshal(buf.Bytes(), &archiveMD5Info)
			continue
		}

		md5Info[hdr.Name] = md5.Sum(buf.Bytes())

		resultPath := filepath.Join(resultFolder, hdr.Name)
		resultDir := filepath.Dir(resultPath)
		if ar.dryRun {
			continue
		}

		if err = os.MkdirAll(resultDir, 0700); err != nil {
			return err
		}
		file, err := os.OpenFile(resultPath, os.O_CREATE|os.O_WRONLY, hdr.FileInfo().Mode().Perm())
		if err != nil {
			return err
		}
		defer file.Close()
		if _, err = file.Write(buf.Bytes()); err != nil {
			return err
		}
	}

	if archiveMD5Info == nil {
		return fmt.Errorf("no %v in archive", md5InfoFile)
	}
	if !reflect.DeepEqual(archiveMD5Info, md5Info) {
		return fmt.Errorf(
			"md5 info of files is corrupted\nExpected: %v\nArchive: %v",
			md5Info,
			archiveMD5Info,
		)
	}

	return nil
}
