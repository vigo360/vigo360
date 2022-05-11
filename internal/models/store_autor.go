/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package models

type AutorStore interface {
	Listar() ([]Autor, error)
	Obtener(string) (Autor, error)
	Buscar(string) ([]Autor, error)
}
