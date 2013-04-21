package gotpl

import (
	"html/template"

	"github.com/PuerkitoBio/ghost"
)

// The template compiler for native Go templates.
type GoTemplateCompiler struct{}

// Implementation of the TemplateCompiler interface.
func (this *GoTemplateCompiler) Compile(f string) (ghost.Templater, error) {
	return template.ParseFiles(f)
}

func init() {
	ghost.Register(".tmpl", new(GoTemplateCompiler))
}
