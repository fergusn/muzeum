package debian

import (
	"strings"
	"testing"
)

func TestReadMultiParagraphs(t *testing.T) {
	data := "Package: A\nVersion: 1\n\nPackage: B\nFilename: /d.deb\n"

	rd := NewControlFileReader(strings.NewReader(data))

	p1, ok := rd.Read()
	if !ok {
		t.Fatal("control data - 1st paragraph expected")
	}

	p2, ok := rd.Read()
	if !ok {
		t.Fatal("control data - 2nd paragraph expected")
	}

	_, ok = rd.Read()
	if ok {
		t.Errorf("only 2 paragraphs expected")
	}

	if p1.Package() != "A" || p1.Version() != "1" {
		t.Errorf("paragraph 1 expected map[Package:A Version:1], got %v", p1)
	}
	if p2.Package() != "B" || p2.Filename() != "/d.deb" {
		t.Errorf("paragraph 2 expected map[Package:B Filename:/d.deb], got %v", p2)
	}
}

func TestMultilineField(t *testing.T) {
	data := "Package: test\nDescription: the first line\n the second line\n the third line"

	rd := NewControlFileReader(strings.NewReader(data))

	p1, ok := rd.Read()

	if !ok {
		t.Fatal("1 paragraph expected")
	}

	if p1["Description"] != "the first line\nthe second line\nthe third line" {
		t.Errorf("descriptions expected 'the first line\nthe second line\nthe third line' got %v", p1["Description"])
	}

}