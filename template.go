package ghost

import (
	"html/template"
	"io"
)

// Defines the interface that the template compiler must return. The Go native
// templates implement this interface.
type Templater interface {
	Execute(wr io.Writer, data interface{}) (err error)
}

// The interface that a template engine must implement to be used by Ghost.
type TemplateCompiler interface {
	Compile(fileName string) (Templater, error)
}

// The template compiler for native Go templates is provided by Ghost. More
// compilers can be found in ghost/templates.
type GoTemplateCompiler struct{}

// Implementation of the TemplateCompiler interface.
func (this *GoTemplateCompiler) Compile(f string) (Templater, error) {
	return template.ParseFiles(f)
}
