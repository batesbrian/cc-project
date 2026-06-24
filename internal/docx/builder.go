package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"slices"
	"strings"
)

type Generator struct {
	Templates fs.FS
}

func (g *Generator) Build(w io.Writer, docx Docx) error {
	b := build{fsys: g.Templates, doc: docx}
	return b.run(w)
}

type Docx struct {
	Caption         Caption
	Issues          []string
	ChangeFont      bool // INFO: default is Times New Roman, alt is Bookman Old Style
	ChangeCitations bool // INFO: defulat is italic, alt is underline
}

type build struct {
	fsys fs.FS
	doc  Docx
}

func (b *build) run(w io.Writer) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	file, err := fs.ReadFile(b.fsys, "caption.docx")
	if err != nil {
		return err
	}

	zr, err := zip.NewReader(bytes.NewReader(file), int64(len(file)))
	if err != nil {
		return err
	}

	for _, f := range zr.File {
		err = b.writeEntry(zw, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *build) writeEntry(zw *zip.Writer, f *zip.File) error {
	fw, err := zw.Create(f.Name)
	if err != nil {
		return err
	}

	fr, err := f.Open()
	if err != nil {
		return err
	}
	defer fr.Close()

	switch f.Name {
	case "word/styles.xml":
		return b.processStyles(fw, fr)

	case "word/document.xml":
		return b.processDocument(fw, fr)

	default:
		_, err = io.Copy(fw, fr)
		return err
	}
}

func (b *build) processStyles(w io.Writer, rc io.ReadCloser) error {
	if !b.doc.ChangeFont {
		_, err := io.Copy(w, rc)
		return err
	}

	d := xml.NewDecoder(rc)

	for {
		token, err := d.RawToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch t := token.(type) {
		case xml.StartElement:
			tagName := getTagName(t.Name)

			if tagName == "w:rFonts" {
				for i, attr := range t.Attr {
					if attr.Value == "Times New Roman" {
						t.Attr[i].Value = "Bookman Old Style"
					}
				}
			}

			writeStartElement(w, t)

		case xml.EndElement:
			fmt.Fprintf(w, "</%s>", getTagName(t.Name))

		case xml.CharData:
			w.Write(t)

		default:
			var buf bytes.Buffer
			e := xml.NewEncoder(&buf)
			e.EncodeToken(t)
			e.Flush()
			w.Write(buf.Bytes())
		}
	}

	return nil
}

func (b *build) processDocument(w io.Writer, rc io.ReadCloser) error {
	d := xml.NewDecoder(rc)

	phValues := map[string]string{
		"{county}": b.doc.Caption.County,
		"{party1}": b.doc.Caption.Party1,
		"{title1}": b.doc.Caption.Title1,
		"{party2}": b.doc.Caption.Party2,
		"{title2}": b.doc.Caption.Title2,
	}

	for {
		token, err := d.RawToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch t := token.(type) {
		case xml.StartElement:
			name := getTagName(t.Name)

			if name == "w:sectPr" && len(b.doc.Issues) > 0 {
				for _, issue := range b.doc.Issues {
					err = b.insertIssue(w, issue)
					if err != nil {
						return err
					}
				}
				// TODO: write service here
			}

			writeStartElement(w, t)

		case xml.EndElement:
			fmt.Fprintf(w, "</%s>", getTagName(t.Name))

		case xml.CharData:
			s := string(t)
			for k, v := range phValues {
				s = strings.ReplaceAll(s, k, v)
				t = []byte(s)
			}
			encodeAndWrite(w, t)

		default:
			encodeAndWrite(w, t)
		}
	}

	return nil
}

func (b *build) insertIssue(w io.Writer, issue string) error {
	file, err := fs.ReadFile(b.fsys, issue)
	if err != nil {
		return err
	}

	zr, err := zip.NewReader(bytes.NewReader(file), int64(len(file)))
	if err != nil {
		return err
	}

	for _, f := range zr.File {
		err := b.writeIssue(w, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *build) writeIssue(w io.Writer, f *zip.File) error {
	if f.Name == "word/document.xml" {
		fr, err := f.Open()
		if err != nil {
			return err
		}
		defer fr.Close()

		d := xml.NewDecoder(fr)

		writeMode := false

		targets := []string{"w:p", "w:pPr", "w:i", "w:r", "w:rPr", "w:t"}

		for {
			token, err := d.RawToken()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			switch t := token.(type) {
			case xml.StartElement:
				name := getTagName(t.Name)

				if name == "w:sectPr" {
					writeMode = false
				}

				if writeMode && slices.Contains(targets, name) {
					if name == "w:i" && b.doc.ChangeCitations {
						t.Name.Space = "w"
						t.Name.Local = "u"
						t.Attr = []xml.Attr{
							{
								Name:  xml.Name{Space: "w", Local: "val"},
								Value: "single",
							},
						}
					}

					writeStartElement(w, t)
				}

				if name == "w:body" {
					writeMode = true
				}

			case xml.EndElement:
				name := getTagName(t.Name)

				if writeMode && slices.Contains(targets, name) {
					if name == "w:i" && b.doc.ChangeCitations {
						t.Name.Space = "w"
						t.Name.Local = "u"
					}

					fmt.Fprintf(w, "</%s>", getTagName(t.Name))
				}

			case xml.CharData:
				if writeMode {
					var buf bytes.Buffer
					e := xml.NewEncoder(&buf)
					e.EncodeToken(t)
					e.Flush()
					w.Write(buf.Bytes())
				}
			}
		}
	}

	return nil
}

// Caption handles inserting values into caption placeholders
type Caption struct {
	County string
	Party1 string
	Title1 string
	Party2 string
	Title2 string
}

func GetCaption(caseType string) Caption {
	return caseTypeCaptions[caseType]
}

var caseTypeCaptions = map[string]Caption{
	"criminal": {Party1: "STATE OF FLORIDA", Title1: "Plaintiff", Party2: "", Title2: "Defendant"},
	"civil":    {Party1: "", Title1: "Plaintiff", Party2: "", Title2: "Defendant"},
	"appeal":   {Party1: "", Title1: "Appellant", Party2: "", Title2: "Appellee"},
	"writ":     {Party1: "", Title1: "Petitioner", Party2: "", Title2: "Respondent"},
}

// helper functions for decoding and encoding valid xml tags for Word
func getTagName(name xml.Name) string {
	if name.Space != "" && !bytes.HasPrefix([]byte(name.Space), []byte("http")) {
		return name.Space + ":" + name.Local
	}
	return name.Local
}

func writeStartElement(w io.Writer, tok xml.StartElement) {
	fmt.Fprintf(w, "<%s", getTagName(tok.Name))
	for _, attr := range tok.Attr {
		var buf bytes.Buffer
		xml.EscapeText(&buf, []byte(attr.Value))
		fmt.Fprintf(w, ` %s="%s"`, getTagName(attr.Name), buf.String())
	}
	fmt.Fprint(w, ">")
}

func encodeAndWrite(w io.Writer, t xml.Token) {
	var buf bytes.Buffer
	e := xml.NewEncoder(&buf)
	e.EncodeToken(t)
	e.Flush()
	w.Write(buf.Bytes())
}
