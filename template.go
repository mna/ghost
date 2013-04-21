package ghost

import (
	"io"
)

var compilers = make(map[string]TemplateCompiler)

// Register a template compiler for the specified extension. Extensions are case-sensitive.
// The extension must start with a dot (it is compared to the result of path.Ext() on a
// given file name).
//
// Registering is not thread-safe. Compilers should be registered before the http server
// is started.
func Register(ext string, c TemplateCompiler) {
	if c == nil {
		panic("ghost: Register TemplateCompiler is nil")
	}
	if _, dup := compilers[ext]; dup {
		panic("ghost: Register called twice for extension " + ext)
	}
	compilers[ext] = c
}

// Defines the interface that the template compiler must return. The Go native
// templates implement this interface.
type Templater interface {
	Execute(wr io.Writer, data interface{}) (err error)
}

// The interface that a template engine must implement to be used by Ghost.
type TemplateCompiler interface {
	Compile(fileName string) (Templater, error)
}
