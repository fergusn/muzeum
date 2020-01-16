package nuget

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func nuspec(archive *zip.Reader) (*Package, []byte, error) {
	file, err := find(archive.File, func(f *zip.File) bool {
		return strings.HasSuffix(f.Name, ".nuspec")
	})
	if err != nil {
		return nil, nil, err
	}
	reader, err := file.Open()
	if err != nil {
		return nil, nil, err
	}
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)

	spec := &Package{}
	err = xml.Unmarshal(data, spec)

	return spec, data, err
}

func find(xs []*zip.File, p func(*zip.File) bool) (*zip.File, error) {
	for _, x := range xs {
		if p(x) {
			return x, nil
		}
	}
	return nil, errors.New("No found")
}

// JSON marchal object to json and set Content-Type
func JSON(w http.ResponseWriter, p interface{}) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}
func path(id, version string) string {
	return fmt.Sprintf("/%s/%s/%s.%s.nupkg", id, version, id, version)
}

type httpError struct {
	code    int
	message string
}

var (
	errNotImplemented      = httpError{http.StatusNotImplemented, "Not Implemented"}
	errNotFound            = httpError{http.StatusNotFound, "Not Found"}
	errInternalServerError = httpError{http.StatusInternalServerError, "Internal Server Error"}
)

func (err httpError) Error() string {
	return err.message
}
