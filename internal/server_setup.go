// SPDX-FileCopyrightText: 2022 Ariel Costas <ariel@vigo360.es>
//
// SPDX-License-Identifier: MPL-2.0
package internal

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/thanhpk/randstr"
)

func (s *Server) JsonifyRoutes(router *mux.Router, path string) *mux.Router {
	var newrouter = router
	newrouter.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var isJsonRoute = strings.HasPrefix(r.URL.Path, path)
			if isJsonRoute {
				w.Header().Add("Content-Type", "application/json")
			}
			h.ServeHTTP(w, r)
			if isJsonRoute {
				fmt.Fprintf(w, "\n")
			}
		})
	})
	return newrouter
}

func (s *Server) IdentifyRequests(router *mux.Router) *mux.Router {
	var newrouter = router
	newrouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var rid = randstr.String(10)
			fmt.Printf("<6>[%s] [%s] %s %s\n", r.Header.Get("X-Forwarded-For"), rid, r.Method, r.URL.Path)
			newContext := context.WithValue(r.Context(), ridContextKey("rid"), rid)
			r = r.WithContext(newContext)
			w.Header().Add("vigo360-rid", rid)
			next.ServeHTTP(w, r)
		})
	})
	return newrouter
}

func (s *Server) SetupApiRoutes(router *mux.Router) *mux.Router {
	var newrouter = router

	newrouter.HandleFunc("/api/v1/comentarios", s.withJsonAuth(s.handle_api_listar_comentarios)).Methods(http.MethodGet)

	return newrouter
}

func (s *Server) SetupWebRoutes(router *mux.Router) *mux.Router {
	var newrouter = router

	newrouter.Handle("/admin/", http.RedirectHandler("/admin/login", http.StatusFound)).Methods(http.MethodGet)
	newrouter.HandleFunc("/admin/login", s.handleAdminLoginPage("")).Methods(http.MethodGet)
	newrouter.HandleFunc("/admin/login", s.handleAdminLoginAction()).Methods(http.MethodPost)
	newrouter.HandleFunc("/admin/logout", s.withAuth(s.handleAdminLogoutAction())).Methods(http.MethodGet)

	newrouter.HandleFunc("/admin/comentarios", s.withAuth(s.handleAdminListComentarios())).Methods(http.MethodGet)
	newrouter.HandleFunc("/admin/comentarios/aprobar", s.withAuth(s.handleAdminAprobarComentario())).Methods(http.MethodGet)
	newrouter.HandleFunc("/admin/comentarios/rechazar", s.withAuth(s.handleAdminRechazarComentario())).Methods(http.MethodGet)

	newrouter.HandleFunc("/admin/dashboard", s.withAuth(s.handleAdminDashboardPage())).Methods(http.MethodGet)
	newrouter.HandleFunc("/admin/post", s.withAuth(s.handleAdminListPost())).Methods(http.MethodGet)
	newrouter.HandleFunc("/admin/post", s.withAuth(s.handleAdminCreatePost())).Methods(http.MethodPost)

	newrouter.HandleFunc("/admin/post/{id}", s.withAuth(s.handleAdminEditPage())).Methods(http.MethodGet)
	newrouter.HandleFunc("/admin/post/{id}", s.withAuth(s.handleAdminEditAction())).Methods(http.MethodPost)
	newrouter.HandleFunc("/admin/post/{postid}/delete", s.withAuth(s.handleAdminDeletePost())).Methods(http.MethodGet)

	newrouter.HandleFunc("/admin/series", s.withAuth(s.handleAdminListSeries())).Methods(http.MethodGet)
	newrouter.HandleFunc("/admin/series", s.withAuth(s.handleAdminCreateSeries())).Methods(http.MethodPost)

	newrouter.HandleFunc("/admin/perfil", s.withAuth(s.handleAdminPerfilView())).Methods(http.MethodGet)
	newrouter.HandleFunc("/admin/perfil", s.withAuth(s.handleAdminPerfilEdit())).Methods(http.MethodPost)

	newrouter.HandleFunc("/admin/preview", s.withAuth(s.handleAdminPreviewPage())).Methods(http.MethodPost)

	newrouter.HandleFunc("/admin/async/fotosExtra", s.withAuth(s.handleAdminListarFotoExtra())).Methods(http.MethodGet)
	newrouter.HandleFunc("/admin/async/fotosExtra", s.withAuth(s.handleAdminCrearFotoExtra())).Methods(http.MethodPost)
	newrouter.HandleFunc("/admin/async/fotosExtra", s.withAuth(s.handleAdminDeleteFotoExtra())).Methods(http.MethodDelete)

	newrouter.HandleFunc(`/post/{postid}`, s.handlePublicPostPage()).Methods(http.MethodGet)
	newrouter.HandleFunc(`/post/{postid}`, s.handlePublicEnviarComentario()).Methods(http.MethodPost)

	newrouter.HandleFunc(`/tags`, s.handlePublicListTags()).Methods(http.MethodGet)
	newrouter.HandleFunc(`/tags/{tagid}/`, s.handlePublicTagPage()).Methods(http.MethodGet)
	newrouter.HandleFunc(`/trabajos`, s.handlePublicListTrabajos()).Methods(http.MethodGet)
	newrouter.HandleFunc(`/trabajos/{trabajoid}`, s.handlePublicTrabajoPage()).Methods(http.MethodGet)
	newrouter.HandleFunc(`/autores/{id}`, s.handlePublicAutorPage()).Methods(http.MethodGet)
	newrouter.HandleFunc(`/autores`, s.handlePublicListAutores()).Methods(http.MethodGet)

	newrouter.HandleFunc(`/legal`, s.handlePublicNodbPage()).Methods(http.MethodGet)
	newrouter.HandleFunc(`/contacto`, s.handlePublicNodbPage()).Methods(http.MethodGet)

	newrouter.HandleFunc(`/atom.xml`, s.handlePublicIndexAtom()).Methods(http.MethodGet)

	newrouter.HandleFunc(`/sitemap.xml`, s.handlePublicSitemap()).Methods(http.MethodGet)
	newrouter.HandleFunc("/buscar", s.handlePublicBusqueda()).Methods(http.MethodGet)

	var indexnowkeyurl = fmt.Sprintf("/%s.txt", os.Getenv("INDEXNOW_KEY"))
	newrouter.HandleFunc(indexnowkeyurl, s.handlePublicIndexnowKey()).Methods(http.MethodGet)

	newrouter.HandleFunc("/", s.handlePublicIndex()).Methods(http.MethodGet)
	return newrouter
}