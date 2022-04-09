/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package public

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"git.sr.ht/~arielcostas/new.vigo360.es/logger"
	goldmarkfigures "github.com/mdigger/goldmark-figures"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var parser goldmark.Markdown = goldmark.New(
	goldmark.WithExtensions(extension.Footnote),
	goldmark.WithExtensions(goldmarkfigures.Extension),
)

//go:embed html/*
var rawtemplates embed.FS

var t = func() *template.Template {
	t := template.New("")

	functions := template.FuncMap{
		"safeHTML": func(text string) template.HTML {
			return template.HTML(text)
		},
		// Converts a standard date returned by MySQL to a RFC3339 datetime
		"date3339": func(date string) (string, error) {
			t, err := time.Parse("2006-01-02 15:04:05", date)
			if err != nil {
				return "", err
			}
			return t.Format(time.RFC3339), nil
		},
		"markdown": func(text string) template.HTML {
			var buf bytes.Buffer
			parser.Convert([]byte(text), &buf)
			return template.HTML(buf.Bytes())
		},
	}

	entries, _ := rawtemplates.ReadDir("html")
	for _, de := range entries {
		filename := de.Name()
		contents, _ := rawtemplates.ReadFile("html/" + filename)

		_, err := t.New(filename).Funcs(functions).Parse(string(contents))
		if err != nil {
			logger.Critical("[public-main] error parsing template: %s", err.Error())
		}
	}

	return t
}()