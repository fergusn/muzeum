package nuget

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/distribution/registry/storage/driver/testdriver"
)

func TestVersion(t *testing.T) {
	s := testdriver.New()
	s.PutContent(context.TODO(), path("pkgid", "1.2"), []byte{1, 2})
	s.PutContent(context.TODO(), path("pkgid", "2.3"), []byte{2, 3})

	repo := NewLocal(s)
	rsp := repo.Versions(context.TODO(), "pkgid")

	versions, err := rsp.Unmarshal()

	if err != nil {
		t.Error(err)
	}

	if len(versions) != 2 || versions[0] != "1.2" || versions[1] != "2.3" {
		t.Errorf("versions expected [1.2 2.3] got %v", versions)
	}
}

func TestDownload(t *testing.T) {
	s := testdriver.New()
	s.PutContent(context.TODO(), path("pkgid", "1.2"), []byte{1, 2})

	repo := NewLocal(s)
	rd, err := repo.Download(context.TODO(), "pkgid", "1.2")
	if err != nil {
		t.Error(err)
	}
	defer rd.Close()

	pkg, err := ioutil.ReadAll(rd)
	if err != nil {
		t.Error(err)
	}

	bytes.Compare(pkg, []byte{1, 2})
}

func TestUpload(t *testing.T) {
	s := testdriver.New()

	repo := NewLocal(s)

	rd := read(t, "xunit.2.4.1.nupkg")
	defer rd.Close()

	err := repo.Upload(context.TODO(), rd)
	if err != nil {
		t.Error(err)
	}

	pkg, err := s.GetContent(context.TODO(), path("xunit", "2.4.1"))

	if err != nil {
		t.Error(err)
	}

	if len(pkg) != 20733 {
		t.Error("Uploaded package should be 20733 bytes")
	}
}

func read(t *testing.T, name string) io.ReadCloser {
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return f
}
