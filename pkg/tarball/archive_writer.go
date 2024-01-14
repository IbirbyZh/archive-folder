package tarball

import (
	"archive/tar"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	md5InfoFile = "_md5_info.json"
)

var blackListFiles = map[string]struct{}{
	".DS_Store": {},
}

type md5InfoType map[string][md5.Size]byte

type ArchiveWriter struct {
	verbose   bool
	tarWriter *tar.Writer
	isClosed  bool
	md5Info   md5InfoType
}

func NewArchiveWriter(w io.Writer, options ...func(*ArchiveWriter)) *ArchiveWriter {
	aw := &ArchiveWriter{
		verbose:   false,
		tarWriter: tar.NewWriter(w),
		md5Info:   make(map[string][16]byte),
	}

	for _, o := range options {
		o(aw)
	}

	return aw
}

func VerboseWriter(verbose bool) func(*ArchiveWriter) {
	return func(e *ArchiveWriter) {
		e.verbose = verbose
	}
}

func (aw *ArchiveWriter) Close() error {
	if aw.isClosed {
		return nil
	}

	aw.isClosed = true

	if err := aw.saveMD5Info(); err != nil {
		return err
	}

	return aw.tarWriter.Close()
}

func (aw *ArchiveWriter) saveMD5Info() error {
	md5InfoContent, err := json.Marshal(aw.md5Info)
	if err != nil {
		return err
	}

	md5InfoHeader := &tar.Header{
		Name: md5InfoFile,
		Mode: 0600,
		Size: int64(len(md5InfoContent)),
	}
	if err := aw.tarWriter.WriteHeader(md5InfoHeader); err != nil {
		return err
	}

	if _, err = aw.tarWriter.Write(md5InfoContent); err != nil {
		return err
	}
	return nil
}

func (aw *ArchiveWriter) AddFiles(root string) error {
	return filepath.Walk(root, aw.createAddToArchiveFunc(root))
}

func (aw *ArchiveWriter) createAddToArchiveFunc(root string) filepath.WalkFunc {
	if !strings.HasSuffix(root, "/") {
		root += "/"
	}
	return func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip files from blacklist
		if _, ok := blackListFiles[filepath.Base(path)]; ok {
			return nil
		}

		p := strings.TrimPrefix(path, root)

		// Check duplicates
		if _, ok := aw.md5Info[p]; ok {
			return fmt.Errorf("file %v is already in archive", p)
		}

		if aw.verbose {
			fmt.Printf("%v %v\n", p, info.Size())
		}

		// Write header
		header := &tar.Header{
			Name: p,
			Mode: 0400,
			Size: info.Size(),
		}
		if err := aw.tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Write content
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		content := make([]byte, info.Size())
		if _, err = file.Read(content); err != nil {
			return err
		}

		if _, err = aw.tarWriter.Write(content); err != nil {
			return err
		}

		aw.md5Info[p] = md5.Sum(content)

		return nil
	}
}
