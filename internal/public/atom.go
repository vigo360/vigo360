/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package public

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"vigo360.es/new/internal/database"
	"vigo360.es/new/internal/logger"
	"vigo360.es/new/internal/model"
)

type AtomEntry struct {
	Id                  string
	Fecha_publicacion   string
	Fecha_actualizacion string

	Titulo       string
	Resumen      string
	Autor_id     string
	Autor_nombre string
	Autor_email  string
	Tag_id       sql.NullString
	Raw_tags     sql.NullString
	Tags         []string
}

// DEPRECATED
// TODO: Get rid of this
type FeedParams struct {
	BaseURL      string
	Id           string
	Nombre       string
	LastUpdate   string
	GeneratorURI string
	Entries      []AtomEntry
}

type AtomParams struct {
	Dominio    string
	Path       string
	Titulo     string
	Subtitulo  string
	LastUpdate string
	Entries    model.Publicaciones
}

func PostsAtomFeed(w http.ResponseWriter, r *http.Request) *appError {
	rps := model.NewPublicacionStore(database.GetDB())
	pp, err := rps.Listar()
	if err != nil {
		return &appError{Error: err, Message: "error obtaining public posts", Response: "Error obteniendo datos", Status: 500}
	}
	pp = pp.FiltrarPublicas()

	lastUpdate, err := pp.ObtenerUltimaActualizacion()
	if err != nil {
		return &appError{Error: err, Message: "error parsing date", Response: "Error obteniendo datos", Status: 500}
	}

	var result bytes.Buffer
	err = t.ExecuteTemplate(&result, "atom.xml", AtomParams{
		Dominio:    os.Getenv("DOMAIN"),
		Path:       "/atom.xml",
		Titulo:     "Publicaciones",
		Subtitulo:  "Últimas publicaciones en el sitio web de Vigo360",
		LastUpdate: lastUpdate.Format(time.RFC3339),
		Entries:    pp,
	})
	if err != nil {
		return &appError{Error: err, Message: "error rendering template", Response: "Error produciendo feed", Status: 500}
	}
	w.Write(result.Bytes())
	return nil
}

func TrabajosAtomFeed(w http.ResponseWriter, r *http.Request) {
	trabajos := []AtomEntry{}
	err := db.Select(&trabajos, `SELECT trabajos.id, fecha_publicacion, fecha_actualizacion, titulo, resumen, autor_id, autores.nombre as autor_nombre, autores.email as autor_email FROM TrabajosPublicos trabajos LEFT JOIN autores ON trabajos.autor_id = autores.id`)

	// An unexpected error
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Warning("[atom] unexpected error selecting trabajos: %s", err.Error())
	}

	writeFeed(w, r, "trabajos-atom.xml", trabajos, "Trabajos", "")
}

func TagsAtomFeed(w http.ResponseWriter, r *http.Request) {
	tagid := mux.Vars(r)["tagid"]
	trabajos := []AtomEntry{}
	err := db.Select(&trabajos, `SELECT publicaciones.id, publicaciones.fecha_publicacion, publicaciones.fecha_actualizacion, publicaciones.titulo, publicaciones.resumen, publicaciones.autor_id, autores.nombre as autor_nombre, autores.email as autor_email FROM publicaciones_tags LEFT JOIN publicaciones ON publicaciones_tags.publicacion_id = publicaciones.id LEFT JOIN autores ON publicaciones.autor_id = autores.id WHERE publicaciones_tags.tag_id = ? AND fecha_publicacion < NOW();`, tagid)

	// An unexpected error
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Warning("[atom] unexpected error selecting posts for tag %s: %s", tagid, err.Error())
	}

	var tagnombre string
	err = db.QueryRowx(`SELECT nombre FROM tags WHERE id = ?;`, tagid).Scan(&tagnombre)

	// An unexpected error
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Warning("[atom] unexpected error fetching tags: %s", err.Error())
	}

	// TODO: Use tagid
	writeFeed(w, r, "tags-id-atom.xml", trabajos, tagnombre, tagid)
}

func AutorAtomFeed(w http.ResponseWriter, r *http.Request) {
	autorid := mux.Vars(r)["autorid"]

	var autor_nombre string
	err := db.QueryRowx("SELECT nombre FROM autores WHERE id=?", autorid).Scan(&autor_nombre)

	if errors.Is(err, sql.ErrNoRows) {
		logger.Error("[atom] autor not found")
		NotFoundHandler(w, r)
		return
	}

	tags := []Tag{}
	tagMap := map[string]string{}
	err = db.Select(&tags, `SELECT * FROM tags`)

	// An unexpected error
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Warning("[atom] unexpected error selecting tags: %s", err.Error())
	}

	for _, tag := range tags {
		tagMap[tag.Id] = tag.Nombre
	}

	posts := []AtomEntry{}
	// TODO: Clean this query
	err = db.Select(&posts, `SELECT pp.id, fecha_publicacion, fecha_actualizacion, titulo, resumen, autor_id, autores.nombre as autor_nombre, autores.email as autor_email, tag_id, GROUP_CONCAT(tag_id) as raw_tags FROM PublicacionesPublicas pp LEFT JOIN publicaciones_tags ON pp.id = publicaciones_tags.publicacion_id LEFT JOIN autores ON pp.autor_id = autores.id WHERE autor_id=? GROUP BY id;`, autorid)

	// An unexpected error
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Warning("[atom] unexpected error selecting posts by author %s: %s", autorid, err.Error())
	}

	for i := 0; i < len(posts); i++ {
		p := posts[i]
		p.Tags = []string{}

		for _, tag := range strings.Split(p.Raw_tags.String, ",") {
			p.Tags = append(p.Tags, tagMap[tag])
		}

		posts[i] = p
	}

	writeFeed(w, r, "autores-atom.xml", posts, "Ariel Costas", autorid)
}

func writeFeed(w http.ResponseWriter, r *http.Request, feedName string, items []AtomEntry, nombre string, id string) {
	// TODO: Refactor line above
	var lastUpdate time.Time

	for i := 0; i < len(items); i++ {
		p := &items[i]

		t, err := time.Parse("2006-01-02 15:04:05", p.Fecha_actualizacion)
		if err != nil {
			logger.Error("unexpected error parsing fecha_actualizacion: %s", err.Error())
			InternalServerErrorHandler(w, r)
		}

		if lastUpdate.Before(t) {
			lastUpdate = t
		}

		p.Id = url.PathEscape(p.Id)
	}

	w.Header().Add("Content-Type", "application/atom+xml; charset=utf-8")
	err := t.ExecuteTemplate(w, feedName, &FeedParams{
		BaseURL:      os.Getenv("DOMAIN"),
		LastUpdate:   lastUpdate.Format(time.RFC3339),
		Entries:      items,
		Nombre:       nombre,
		Id:           id,
		GeneratorURI: os.Getenv("SOURCE_URL"),
	})

	if err != nil {
		logger.Error("unexpected error rendering feed %s: %s", feedName, err.Error())
		InternalServerErrorHandler(w, r)
	}
}
