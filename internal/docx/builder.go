package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"
)

type Docx struct {
	TemplatesDir    string
	Caption         Caption
	Issues          []string
	ChangeFont      bool // INFO: default is Times New Roman, alt is Bookman Old Style
	ChangeCitations bool // INFO: defulat is italic, alt is underline
}

func processStyles(w io.Writer, r io.ReadCloser, change bool) error {
	if !change {
		_, err := io.Copy(w, r)
		return err
	}

	d := xml.NewDecoder(r)

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

// TODO: finish refactor
// func processDocument(w io.Writer, r io.Reader) error {
// 	return nil
// }

func insertIssues(w io.Writer, issues []string, changeCitations bool) error {
	if len(issues) == 0 {
		return nil
	}

	for _, issue := range issues {
		r, err := zip.OpenReader(issue)
		if err != nil {
			return err
		}

		for _, f := range r.File {
			if f.Name == "word/document.xml" {
				rInner, err := f.Open()
				if err != nil {
					return err
				}

				d := xml.NewDecoder(rInner)
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
							if name == "w:i" && changeCitations {
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
							if name == "w:i" && changeCitations {
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
				rInner.Close()
				break
			}
		}
		r.Close()
	}
	return nil
}

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

func (docx *Docx) Build(w io.Writer) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	captionPath := filepath.Join(docx.TemplatesDir, "caption.docx")

	rCaption, err := zip.OpenReader(captionPath)
	if err != nil {
		return err
	}
	defer rCaption.Close()

	for _, f := range rCaption.File {
		fw, err := zw.Create(f.Name)
		if err != nil {
			return err
		}

		fr, err := f.Open()
		if err != nil {
			return err
		}

		switch f.Name {
		case "word/styles.xml":
			err = processStyles(fw, fr, docx.ChangeFont)
			if err != nil {
				fr.Close()
				return err
			}

		case "word/document.xml":
			// TODO: factor this out to a processDocument func
			d := xml.NewDecoder(fr)

			phValues := map[string]string{
				"{county}": docx.Caption.County,
				"{party1}": docx.Caption.Party1,
				"{title1}": docx.Caption.Title1,
				"{party2}": docx.Caption.Party2,
				"{title2}": docx.Caption.Title2,
			}

			for {
				token, err := d.RawToken()
				if err == io.EOF {
					break
				}
				if err != nil {
					fr.Close()
					return err
				}

				switch t := token.(type) {
				case xml.StartElement:
					name := getTagName(t.Name)

					if name == "w:sectPr" {
						err = insertIssues(fw, docx.Issues, docx.ChangeCitations)
						if err != nil {
							fr.Close()
							return err
						}
						// TODO: write service here
					}

					writeStartElement(fw, t)

				case xml.EndElement:
					fmt.Fprintf(fw, "</%s>", getTagName(t.Name))

				case xml.CharData:
					s := string(t)
					for k, v := range phValues {
						s = strings.ReplaceAll(s, k, v)
						t = []byte(s)
					}
					encodeAndWrite(fw, t)

				default:
					encodeAndWrite(fw, t)
				}
			}

		default:
			_, err = io.Copy(fw, fr)
			if err != nil {
				fr.Close()
				return err
			}
		}

		fr.Close()
	}

	return nil
}

func encodeAndWrite(w io.Writer, t xml.Token) {
	var buf bytes.Buffer
	e := xml.NewEncoder(&buf)
	e.EncodeToken(t)
	e.Flush()
	w.Write(buf.Bytes())
}
