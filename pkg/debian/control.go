package debian

import (
	"bufio"
	"io"
	"strings"
)

// Paragraph consist of a series of fields. Field names are unique for a paragraph.
type Paragraph map[string]string

// Version return the 'Version' field value from the paragraph
func (p Paragraph) Version() string {
	return p["Version"]
}

// Package return the 'Package' field value from the paragraph
func (p Paragraph) Package() string {
	return p["Package"]
}

// Filename return the 'Filename' field value from the paragraph
func (p Paragraph) Filename() string {
	return p["Filename"]
}

// ControlFileReader reads Paragraphs from a control file
type ControlFileReader struct {
	scanner *bufio.Scanner
}

// NewControlFileReader creates a new ControlFileReader
func NewControlFileReader(r io.Reader) *ControlFileReader {
	return &ControlFileReader{bufio.NewScanner(r)}
}

// Read the next paragraph from the control file
func (r *ControlFileReader) Read() (par Paragraph, ok bool) {
	par = Paragraph{}
	var key string
	for r.scanner.Scan() {
		ln := r.scanner.Text()
		if len(ln) == 0 {
			break
		} else if ln[0] == '#' {
			continue	
		} else if ln[0] == ' ' || ln[0] == '\t' {
			par[key] += "\n" + strings.TrimLeft(ln, " \t")
		} else {
			kv := strings.SplitN(ln, ":", 2)
			key = kv[0]
			par[key] = strings.TrimSpace(kv[1])
		}
		ok = true
	}
	return
}
