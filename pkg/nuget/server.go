package nuget

import (
	"fmt"
	"io"
	"net/http"

	"github.com/fergusn/muzeum/pkg/events"
	"github.com/fergusn/muzeum/pkg/model"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Server expose a repository on HTTP
type Server struct {
	name       string
	repository Repository
}

// Mount the server on a mux router
func (srv Server) Mount(route *mux.Route) {
	router := route.Subrouter()
	router.HandleFunc("/index.json", srv.index(route)).Methods(http.MethodGet)

	router.HandleFunc("/content/{id}/{version}/{file}.nupkg", srv.download).Methods(http.MethodGet)
	router.HandleFunc("/content/{id}/index.json", srv.versions).Methods(http.MethodGet)

	router.HandleFunc("/package/", srv.publish).Methods(http.MethodPut)
	router.HandleFunc("/package/{id}/{version}", srv.relist).Methods(http.MethodPost)
	router.HandleFunc("/package/{id}/{version}", srv.delete).Methods(http.MethodDelete)

	router.HandleFunc("/search/", srv.search).Methods(http.MethodGet)
}

func (srv *Server) index(route *mux.Route) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		path, _ := route.URLPath()

		scheme := "http"

		if r.TLS != nil {
			scheme = "https"
		}

		JSON(w, ServiceIndex{
			Version: "3.0.0",
			Resources: []Resource{
				Resource{
					ID:   fmt.Sprintf("%s://%s%s/content/", scheme, r.Host, path.Path),
					Type: PackageBaseAddress,
				},
				Resource{
					ID:   fmt.Sprintf("%s://%s%s/package/", scheme, r.Host, path.Path),
					Type: PackagePublish,
				},
				Resource{
					ID:   fmt.Sprintf("%s://%s%s/search/", scheme, r.Host, path.Path),
					Type: SearchQueryService,
				},
				Resource{
					ID:   fmt.Sprintf("%s://%s%s/registration/", scheme, r.Host, path.Path),
					Type: RegistrationsBaseUrl,
				},
			},
		})
	}
}

func (srv *Server) versions(w http.ResponseWriter, r *http.Request) {
	versions := srv.repository.Versions(r.Context(), mux.Vars(r)["id"])

	if versions.Status > 0 {
		w.WriteHeader(versions.Status)
		return
	}
	defer versions.Close()

	io.Copy(w, versions)
}

func (srv *Server) download(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	version := mux.Vars(r)["version"]

	nupkg, err := srv.repository.Download(r.Context(), id, version)
	if err, ok := err.(httpError); ok {
		http.Error(w, err.message, err.code)
		return
	}
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/octet-stream")

	n, err := io.Copy(w, nupkg)
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	events.Package.Pulled.Emit(&events.Pulled{
		Registry: srv.name,
		Package: &model.Package{
			Type:    "nuget",
			Name:    id,
			Version: version,
		},
		Location: r.RemoteAddr,
		Size:     n,
	})
}

func (srv *Server) publish(w http.ResponseWriter, r *http.Request) {
	parts, err := r.MultipartReader()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	part, err := parts.NextPart()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = srv.repository.Upload(r.Context(), part)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (srv *Server) delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	srv.repository.Delete(r.Context(), vars["id"], vars["version"])
}

func (srv *Server) relist(w http.ResponseWriter, r *http.Request) {

}

func (srv *Server) search(w http.ResponseWriter, r *http.Request) {

}
