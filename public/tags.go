package public

import (
	"log"
	"net/http"

	"git.sr.ht/~arielcostas/new.vigo360.es/common"
	"github.com/gorilla/mux"
)

type TagsIdPost struct {
	Id                string
	Fecha_publicacion string
	Alt_portada       string
	Titulo            string
	Autor_nombre      string `db:"nombre"`
}

type TagsIdTag struct {
	Titulo string
}

type TagsIdParams struct {
	Tag   TagsIdTag
	Posts []TagsIdPost
	Meta  common.PageMeta
}

func TagsIdPage(w http.ResponseWriter, r *http.Request) {
	req_tagid := mux.Vars(r)["tagid"]

	tag := TagsIdTag{}
	err := db.QueryRowx("SELECT nombre as titulo FROM tags WHERE id=?;", req_tagid).StructScan(&tag)
	if err != nil {
		log.Fatalf(err.Error())
	}

	posts := []TagsIdPost{}
	db.Select(&posts, `SELECT publicaciones.id, DATE_FORMAT(publicaciones.fecha_publicacion, '%d %b. %Y') as fecha_publicacion, publicaciones.alt_portada, publicaciones.titulo,
	autores.nombre FROM publicaciones_tags
	LEFT JOIN publicaciones ON publicaciones_tags.publicacion_id = publicaciones.id
    LEFT JOIN autores ON publicaciones.autor_id = autores.id
    WHERE tag_id = ? ORDER BY publicaciones.fecha_publicacion DESC;`, req_tagid)

	t.ExecuteTemplate(w, "tags-id.html", TagsIdParams{
		Tag:   tag,
		Posts: posts,
		Meta: common.PageMeta{
			Titulo:      tag.Titulo,
			Descripcion: "Publicaciones en Vigo360 sobre " + tag.Titulo,
			Canonica:    FullCanonica("/tags/" + req_tagid),
		},
	})
}