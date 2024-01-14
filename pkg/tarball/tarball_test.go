package tarball

import (
	"bytes"
	"testing"
)

func TestFull(t *testing.T) {
	var buf bytes.Buffer
	aw := NewArchiveWriter(&buf)
	defer aw.Close()
	if err := aw.AddFiles("tarball_testdata"); err != nil {
		t.Fatal(err)
	}
	if err := aw.Close(); err != nil {
		t.Fatal(err)
	}

	ar := NewArchiveReader(&buf, VerboseReader(false), DryRun(true))
	if err := ar.ExtractFiles("tarball_testdata_result"); err != nil {
		t.Fatal(err)
	}
}
