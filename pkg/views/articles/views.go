package articles

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

const (
	templateExtension = ".gohtml"
)

func NewView(layout string, files ...string) *View {
	var templateDir, layoutDir string

	if os.Getenv("DEPLOYMENT") == "TRUE" {
		layoutDir = "views/layouts/"
		templateDir = "views/"
	} else {
		layoutDir = "pkg/views/layouts/"
		templateDir = "pkg/views/"
	}

	for i := range files {
		files[i] = templateDir + files[i] + templateExtension
	}
	layoutFiles, _ := filepath.Glob(layoutDir + "*" + templateExtension)
	files = append(files, layoutFiles...)

	t, err := template.ParseFiles(files...)
	must(err)

	return &View{
		Template: t,
		Layout:   layout,
	}

}

type View struct {
	Template *template.Template
	Layout   string
}

func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html; charset=utf8")

	return v.Template.ExecuteTemplate(w, v.Layout, data)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
