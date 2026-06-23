package docx

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TODO: probably break into separate buffer test and file/unzip test
func TestDocxBuild(t *testing.T) {
	docx := Docx{
		TemplatesDir:    "../../templates",
		Caption:         GetCaption("criminal"),
		ChangeFont:      true,
		ChangeCitations: true,
	}

	// buf := &bytes.Buffer{}
	//
	// err := docx.Build(buf)
	// if err != nil {
	// 	t.Fatalf("failed to build docx to buf: %v\n", err)
	// }
	//
	// zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	// if err != nil {
	// 	t.Fatalf("build did not produce a valid zip archive: %v\n", err)
	// }

	docPath := "./output/output.docx"

	built, err := os.Create(docPath)
	if err != nil {
		t.Fatalf("failed to create output docx file: %v\n", err)
	}

	docx.Build(built)

	zr, err := zip.OpenReader(docPath)
	if err != nil {
		t.Fatalf("build did not produce a valid zip archive: %v\n", err)
	}
	defer zr.Close()

	unzipDir := "./output/unzipped"

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}

		outFile := filepath.Join(unzipDir, filepath.Base(f.Name))

		rc, err := f.Open()
		if err != nil {
			t.Errorf("failed to open file %s: %v\n", f.Name, err)
			continue
		}

		dest, err := os.Create(outFile)
		if err != nil {
			rc.Close()
			t.Errorf("failed to create file %s: %v\n", outFile, err)
			continue
		}

		_, err = io.Copy(dest, rc)
		if err != nil {
			t.Errorf("failed to write to file %s: %v\n", outFile, err)
		}

		dest.Close()
		rc.Close()
	}
}
