package app

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type PreviewBlock struct {
	Kind string     `json:"kind"`
	Text string     `json:"text,omitempty"`
	Rows [][]string `json:"rows,omitempty"`
}
type PreviewSheet struct {
	Name string     `json:"name"`
	Rows [][]string `json:"rows"`
}
type FilePreview struct {
	Kind      string         `json:"kind"`
	Blocks    []PreviewBlock `json:"blocks,omitempty"`
	Sheets    []PreviewSheet `json:"sheets,omitempty"`
	Truncated bool           `json:"truncated,omitempty"`
}

func previewFile(filePath, mime string) (FilePreview, error) {
	switch {
	case strings.Contains(mime, "wordprocessingml"):
		return previewDOCX(filePath)
	case strings.Contains(mime, "spreadsheetml"):
		return previewXLSX(filePath)
	case strings.HasPrefix(mime, "text/csv"):
		return previewCSV(filePath)
	default:
		return FilePreview{}, errors.New("Этот формат просматривается в исходном приложении")
	}
}
func (a *App) servePreview(w http.ResponseWriter, v Version) {
	p, err := previewFile(filepath.Join(a.dir, "files", v.ID), v.MIME)
	if err != nil {
		problem(w, http.StatusUnprocessableEntity, "Не удалось показать содержимое. Скачайте исходный файл.")
		return
	}
	respond(w, 200, p)
}
func zipEntry(z *zip.ReadCloser, name string, limit int64) ([]byte, error) {
	for _, f := range z.File {
		if f.Name == name {
			if f.UncompressedSize64 > uint64(limit) {
				return nil, fmt.Errorf("preview XML too large: %s", name)
			}
			r, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer r.Close()
			b, err := io.ReadAll(io.LimitReader(r, limit+1))
			if err != nil {
				return nil, err
			}
			if int64(len(b)) > limit {
				return nil, fmt.Errorf("preview XML too large: %s", name)
			}
			return b, nil
		}
	}
	return nil, os.ErrNotExist
}
func xmlAttr(s xml.StartElement, name string) string {
	for _, a := range s.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}
func previewDOCX(filePath string) (FilePreview, error) {
	z, err := zip.OpenReader(filePath)
	if err != nil {
		return FilePreview{}, err
	}
	defer z.Close()
	b, err := zipEntry(z, "word/document.xml", 8<<20)
	if err != nil {
		return FilePreview{}, err
	}
	out := FilePreview{Kind: "document", Blocks: []PreviewBlock{}}
	dec := xml.NewDecoder(bytes.NewReader(b))
	var paragraph, style string
	var table [][]string
	var row []string
	var cell []string
	inTable, inCell, inText := false, false, false
	addBlock := func(block PreviewBlock) {
		if len(out.Blocks) < 200 {
			out.Blocks = append(out.Blocks, block)
		} else {
			out.Truncated = true
		}
	}
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return FilePreview{}, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "tbl":
				inTable = true
				table = [][]string{}
			case "tr":
				if inTable {
					row = []string{}
				}
			case "tc":
				if inTable {
					inCell = true
					cell = []string{}
				}
			case "p":
				paragraph = ""
				style = ""
			case "pStyle":
				style = xmlAttr(t, "val")
			case "t":
				inText = true
			case "tab":
				paragraph += "    "
			case "br":
				paragraph += "\n"
			}
		case xml.CharData:
			if inText {
				paragraph += string(t)
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				inText = false
			case "p":
				if value := strings.TrimSpace(paragraph); value != "" {
					if inCell {
						cell = append(cell, value)
					} else {
						kind := "paragraph"
						if strings.HasPrefix(strings.ToLower(style), "heading") || strings.HasPrefix(strings.ToLower(style), "заголовок") {
							kind = "heading"
						}
						addBlock(PreviewBlock{Kind: kind, Text: value})
					}
				}
			case "tc":
				if inCell {
					row = append(row, strings.Join(cell, "\n"))
					inCell = false
				}
			case "tr":
				if inTable {
					if len(table) < 100 {
						table = append(table, row)
					} else {
						out.Truncated = true
					}
				}
			case "tbl":
				if inTable {
					addBlock(PreviewBlock{Kind: "table", Rows: table})
					inTable = false
				}
			}
		}
		if out.Truncated {
			break
		}
	}
	if len(out.Blocks) == 0 {
		return FilePreview{}, errors.New("empty Word preview")
	}
	return out, nil
}
func previewCSV(filePath string) (FilePreview, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return FilePreview{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return FilePreview{}, err
	}
	b, err := io.ReadAll(io.LimitReader(f, 4<<20))
	if err != nil {
		return FilePreview{}, err
	}
	large := info.Size() > int64(len(b))
	if large {
		if end := bytes.LastIndexByte(b, '\n'); end >= 0 {
			b = b[:end+1]
		}
	}
	b = bytes.TrimPrefix(b, []byte{0xef, 0xbb, 0xbf})
	first := string(bytes.SplitN(b, []byte("\n"), 2)[0])
	delimiter := ','
	for _, candidate := range []rune{';', '\t'} {
		if strings.Count(first, string(candidate)) > strings.Count(first, string(delimiter)) {
			delimiter = candidate
		}
	}
	r := csv.NewReader(bytes.NewReader(b))
	r.Comma = delimiter
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	out := FilePreview{Kind: "spreadsheet", Sheets: []PreviewSheet{{Name: "CSV", Rows: [][]string{}}}, Truncated: large}
	for len(out.Sheets[0].Rows) < 200 {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return FilePreview{}, err
		}
		if len(row) > 30 {
			row = row[:30]
			out.Truncated = true
		}
		out.Sheets[0].Rows = append(out.Sheets[0].Rows, row)
	}
	if len(out.Sheets[0].Rows) == 200 {
		out.Truncated = true
	}
	return out, nil
}
func previewXLSX(filePath string) (FilePreview, error) {
	z, err := zip.OpenReader(filePath)
	if err != nil {
		return FilePreview{}, err
	}
	defer z.Close()
	shared := []string{}
	if b, e := zipEntry(z, "xl/sharedStrings.xml", 8<<20); e == nil {
		dec := xml.NewDecoder(bytes.NewReader(b))
		var value string
		inText := false
		for {
			tok, e := dec.Token()
			if e == io.EOF {
				break
			}
			if e != nil {
				return FilePreview{}, e
			}
			switch t := tok.(type) {
			case xml.StartElement:
				if t.Name.Local == "si" {
					value = ""
				}
				if t.Name.Local == "t" {
					inText = true
				}
			case xml.CharData:
				if inText {
					value += string(t)
				}
			case xml.EndElement:
				if t.Name.Local == "t" {
					inText = false
				}
				if t.Name.Local == "si" {
					shared = append(shared, value)
					if len(shared) > 100000 {
						return FilePreview{}, errors.New("too many shared strings")
					}
				}
			}
		}
	}
	rels := map[string]string{}
	if b, e := zipEntry(z, "xl/_rels/workbook.xml.rels", 2<<20); e == nil {
		dec := xml.NewDecoder(bytes.NewReader(b))
		for {
			tok, e := dec.Token()
			if e == io.EOF {
				break
			}
			if e != nil {
				return FilePreview{}, e
			}
			if t, ok := tok.(xml.StartElement); ok && t.Name.Local == "Relationship" {
				target := xmlAttr(t, "Target")
				if strings.HasPrefix(target, "/") {
					target = strings.TrimPrefix(target, "/")
				} else {
					target = path.Clean(path.Join("xl", target))
				}
				rels[xmlAttr(t, "Id")] = target
			}
		}
	}
	type sheetInfo struct{ name, file string }
	sheets := []sheetInfo{}
	if b, e := zipEntry(z, "xl/workbook.xml", 2<<20); e == nil {
		dec := xml.NewDecoder(bytes.NewReader(b))
		for {
			tok, e := dec.Token()
			if e == io.EOF {
				break
			}
			if e != nil {
				return FilePreview{}, e
			}
			if t, ok := tok.(xml.StartElement); ok && t.Name.Local == "sheet" {
				sheets = append(sheets, sheetInfo{xmlAttr(t, "name"), rels[xmlAttr(t, "id")]})
			}
		}
	}
	if len(sheets) == 0 {
		for _, f := range z.File {
			if strings.HasPrefix(f.Name, "xl/worksheets/sheet") && strings.HasSuffix(f.Name, ".xml") {
				sheets = append(sheets, sheetInfo{filepath.Base(f.Name), f.Name})
			}
		}
	}
	out := FilePreview{Kind: "spreadsheet", Sheets: []PreviewSheet{}}
	for i, info := range sheets {
		if i >= 5 {
			out.Truncated = true
			break
		}
		if info.file == "" {
			info.file = fmt.Sprintf("xl/worksheets/sheet%d.xml", i+1)
		}
		b, e := zipEntry(z, info.file, 8<<20)
		if e != nil {
			return FilePreview{}, e
		}
		rows, truncated, e := parseWorksheet(b, shared)
		if e != nil {
			return FilePreview{}, e
		}
		out.Truncated = out.Truncated || truncated
		out.Sheets = append(out.Sheets, PreviewSheet{Name: info.name, Rows: rows})
	}
	if len(out.Sheets) == 0 {
		return FilePreview{}, errors.New("empty workbook")
	}
	return out, nil
}
func columnIndex(ref string) int {
	n := 0
	for _, ch := range strings.ToUpper(ref) {
		if ch < 'A' || ch > 'Z' {
			break
		}
		n = n*26 + int(ch-'A'+1)
	}
	return n - 1
}
func parseWorksheet(b []byte, shared []string) ([][]string, bool, error) {
	dec := xml.NewDecoder(bytes.NewReader(b))
	rows := [][]string{}
	row := []string{}
	cellType, cellRef, value := "", "", ""
	inValue, inText, inCell := false, false, false
	truncated := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, false, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "row":
				row = []string{}
			case "c":
				inCell = true
				cellType = xmlAttr(t, "t")
				cellRef = xmlAttr(t, "r")
				value = ""
			case "v":
				if inCell {
					inValue = true
				}
			case "t":
				if inCell {
					inText = true
				}
			}
		case xml.CharData:
			if inValue || inText {
				value += string(t)
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "v":
				inValue = false
			case "t":
				inText = false
			case "c":
				if inCell {
					if cellType == "s" {
						idx, e := strconv.Atoi(strings.TrimSpace(value))
						if e == nil && idx >= 0 && idx < len(shared) {
							value = shared[idx]
						} else {
							value = ""
						}
					}
					if cellType == "b" {
						if value == "1" {
							value = "Да"
						} else {
							value = "Нет"
						}
					}
					col := columnIndex(cellRef)
					if col < 0 {
						col = len(row)
					}
					if col < 30 {
						for len(row) <= col {
							row = append(row, "")
						}
						row[col] = value
					} else {
						truncated = true
					}
					inCell = false
				}
			case "row":
				rows = append(rows, row)
				if len(rows) >= 200 {
					return rows, true, nil
				}
			}
		}
	}
	return rows, truncated, nil
}
