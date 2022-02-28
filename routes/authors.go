package routes

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

//#region /autores/id
type AutoresIdAutor struct {
	Id         string
	Nombre     string
	Email      string
	Rol        string
	Biografia  string
	Web_url    string
	Web_titulo string
}

type AutoresIdPublicacion struct {
	Id                  string
	Fecha_publicacion   string
	Fecha_actualizacion string
	Alt_portada         string
	Titulo              string
	Resumen             string
}

type AutoresIdParams struct {
	Autor AutoresIdAutor
	Posts []AutoresIdPublicacion
}

func AutoresIdPage(w http.ResponseWriter, r *http.Request) {
	req_author := mux.Vars(r)["id"]
	autor := AutoresIdAutor{}
	// TODO error handling
	err := db.QueryRowx("SELECT id, nombre, email, rol, biografia, web_url, web_titulo FROM autores WHERE id=?", req_author).StructScan(&autor)

	if err != nil {
		log.Fatalf(err.Error())
	}

	publicaciones := []AutoresIdPublicacion{}
	// TODO error handling
	err = db.Select(&publicaciones, `SELECT id, DATE_FORMAT(fecha_publicacion, '%d %b.') as fecha_publicacion, 
	DATE_FORMAT(fecha_actualizacion, '%d %b.') as fecha_actualizacion, alt_portada, titulo, resumen
	FROM publicaciones WHERE autor_id = 'ariel-costas' ORDER BY publicaciones.fecha_publicacion DESC;`)

	if err != nil {
		log.Fatalf(err.Error())
	}

	t.ExecuteTemplate(w, "authors-id.html", AutoresIdParams{
		Autor: autor,
		Posts: publicaciones,
	})
}

//#endregion

//#region /autores
type AutoresAutor struct {
	Id        string
	Nombre    string
	Rol       string
	Biografia string
}

type AutoresParams struct {
	Autores []AutoresAutor
}

func AutoresPage(w http.ResponseWriter, r *http.Request) {
	autores := []AutoresAutor{}
	db.Select(&autores, `SELECT id, nombre, rol, biografia FROM autores;`)

	t.ExecuteTemplate(w, "autores.html", AutoresParams{
		Autores: autores,
	})
}

//#endregion