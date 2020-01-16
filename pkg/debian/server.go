package debian

import (
	"io"
	"strings"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

// Server is a HTTP debian repository
type Server struct {
	name string
	url  *url.URL
	repo Repository
}

// NewServer creates a new server that delegate requests to a repository
func NewServer(name string, url *url.URL, repo Repository) *Server {
	return &Server{name, url, repo}
}

// Mount the server routes
func (srv *Server) Mount(route *mux.Route) {
	router := route.Subrouter()

	router.Methods(http.MethodGet).Path("/" + concat(srv.url.Path, "dists/{dist}/InRelease")).HandlerFunc(srv.release)
	router.Methods(http.MethodGet).Path("/" + concat(srv.url.Path, "dists/{dist}/{comp}/binary-{arch}/Packages.{compression}")).HandlerFunc(srv.index)

	router.Methods(http.MethodGet).Path("/" + concat(srv.url.Path, "dists/{dist}/{comp}/binary-{arch}/by-hash/{algorithm}/{hash}")).HandlerFunc(srv.byhash)
	

	router.Methods(http.MethodGet).HandlerFunc(srv.file)
}

func (srv *Server) release(w http.ResponseWriter, r *http.Request) {
	write(w)(srv.repo.Release(r.Context(), mux.Vars(r)["dist"]))
}

func (srv *Server) index(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	write(w)(srv.repo.Index(r.Context(), vars["dist"], vars["comp"], vars["arch"], vars["compression"]))
}

func (srv *Server) byhash(w http.ResponseWriter, r *http.Request) {
	// TODO: we need to read the index file and get the correct index
	vars := mux.Vars(r)
	write(w)(srv.repo.Index(r.Context(), vars["dist"], vars["comp"], vars["arch"], "gz"))
}

func (srv *Server) file(w http.ResponseWriter, r *http.Request) {
	rd, _, err := srv.repo.File(r.Context(), strings.TrimPrefix(r.URL.Path, srv.url.Path))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: Emit package pulled event. This is not currenly implemented because the local repository does not return package metadata

	io.Copy(w, rd)
}
