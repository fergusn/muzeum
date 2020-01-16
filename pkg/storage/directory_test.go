package storage

import (
	"context"
	"testing"

	"github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/testdriver"
)

var (
	content = []byte{1, 2, 3, 4}
)

func TestName(t *testing.T) {
	tst := testdriver.New()
	dir := NewDirectoryDriver("qwerty", tst)

	if dir.Name() != name {
		t.Errorf("expected name %s got %s", name, dir.Name())
	}
}

func TestGetContent(t *testing.T) {
	tst := testdriver.New()
	dir := NewDirectoryDriver("qwerty", tst)
	tst.PutContent(context.TODO(), "/qwerty/abcd", content)

	if _, err := dir.GetContent(context.TODO(), "/abcd"); err != nil {
		t.Error("directoryDriver should use sub-directory")
	}
}

func TestPutContent(t *testing.T) {
	tst := testdriver.New()
	dir := NewDirectoryDriver("qwerty", tst)

	if err := dir.PutContent(context.TODO(), "/abcd", content); err != nil {
		t.Fatal(err)
	}

	assertExists(t, tst, "/qwerty/abcd")
}

func TestRead(t *testing.T) {
	tst := testdriver.New()
	dir := NewDirectoryDriver("qwerty", tst)
	tst.PutContent(context.TODO(), "/qwerty/abcd", content)

	rd, err := dir.Reader(context.TODO(), "/abcd", 0)
	if err != nil {
		t.Error(err)
	}
	rd.Close()
}

func TestWrite(t *testing.T) {
	tst := testdriver.New()
	dir := NewDirectoryDriver("qwerty", tst)

	wr, err := dir.Writer(context.TODO(), "/abcd", false)
	if err != nil {
		t.Error(err)
	}
	wr.Write(content)
	wr.Commit()

	assertExists(t, tst, "/qwerty/abcd")
}

func TestStat(t *testing.T) {
	tst := testdriver.New()
	dir := NewDirectoryDriver("qwerty", tst)
	tst.PutContent(context.TODO(), "/qwerty/abcd", content)

	info, err := dir.Stat(context.TODO(), "/abcd")

	if err != nil {
		t.Error(err)
	}

	if info.Path() != "/abcd" {
		t.Errorf("Stat FileInfo should return original path /abcd got %s", info.Path())
	}
}

func TestList(t *testing.T) {
	tst := testdriver.New()
	dir := NewDirectoryDriver("qwerty", tst)
	tst.PutContent(context.TODO(), "/qwerty/abcd", content)
	tst.PutContent(context.TODO(), "/qwerty/defg", content)

	xs, err := dir.List(context.TODO(), "/")

	if err != nil {
		t.Fatal(err)
	}

	if len(xs) != 2 || xs[0] != "/abcd" || xs[1] != "/defg" {
		t.Errorf("list expected [abcd, defg] got %v", xs)
	}
}

func TestMove(t *testing.T) {
	tst := testdriver.New()
	dir := NewDirectoryDriver("qwerty", tst)
	tst.PutContent(context.TODO(), "/qwerty/abcd", content)

	if err := dir.Move(context.TODO(), "/abcd", "/xyz"); err != nil {
		t.Fatal(err)
	}

	assertExists(t, tst, "/qwerty/xyz")
}

func TestDelete(t *testing.T) {
	tst := testdriver.New()
	dir := NewDirectoryDriver("qwerty", tst)
	tst.PutContent(context.TODO(), "/qwerty/abcd", content)

	if err := dir.Delete(context.TODO(), "/abcd"); err != nil {
		t.Fatal(err)
	}

	if _, err := tst.Stat(context.TODO(), "/qwerty/abcd"); err == nil {
		t.Error("file should be deleted")
	}
}

func assertExists(t *testing.T, dir driver.StorageDriver, path string) {
	if _, err := dir.GetContent(context.TODO(), path); err != nil {
		t.Error("directoryDriver should use sub-directory")
	}
}
