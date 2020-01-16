package debian

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/smira/go-xz"
)

func concat(parts ...string) (url string) {
	for i, x := range parts {
		if i > 0 && !strings.HasSuffix(url, "/") {
			url += "/"
		}
		part := strings.Trim(x, "/")

		if strings.HasPrefix(x, "./") {
			part = strings.TrimPrefix(part, "./")
		}

		if part != "." {
			url += part
		}
	}
	return
}

func write(w http.ResponseWriter) func(io.ReadCloser, error) {
	return func(r io.ReadCloser, err error) {
		if err != nil {
			w.WriteHeader(500)
			return
		}
		io.Copy(w, r)
	}
}

func decompress(r io.Reader, algo string) (io.Reader, error) {
	if algo == "gz" {
		return gzip.NewReader(r)
	} else if algo == "xz" {
		return xz.NewReader(r)
	} else {
		return nil, errors.New("unkown compression algorithm")
	}
}
