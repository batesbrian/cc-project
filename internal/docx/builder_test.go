package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"testing"
)

func TestDocxBuild(t *testing.T) {
	templates := os.DirFS("./../../templates")

	gen := Generator{
		Templates: templates,
	}

	caption, ok := GetCaption("criminal")
	if !ok {
		t.Errorf("no caption found for case type: criminal")
	}

	docx := Docx{
		Caption:         caption,
		Issues:          []string{"facial_sufficiency.docx"},
		ChangeFont:      true,
		ChangeCitations: true,
	}

	var buf bytes.Buffer
	err := gen.Build(&buf, docx)
	if err != nil {
		t.Fatalf("failed to build docx: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("build not a valid zip: %v", err)
	}

	for _, f := range zr.File {
		if f.Name == "word/document.xml" || f.Name == "word/styles.xml" {
			r, err := f.Open()
			if err != nil {
				t.Fatalf("failed to open %s: %v", f.Name, err)
			}
			assertWellFormed(t, r)
		}
	}
}

func assertWellFormed(t *testing.T, r io.Reader) {
	d := xml.NewDecoder(r)

	for {
		_, err := d.Token()
		if err == io.EOF {
			return
		}
		if err != nil {
			t.Fatalf("malformed XML: %v", err)
		}
	}
}
