package templates

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/PuerkitoBio/ghost"
)

var (
	compilers  = make(map[string]TemplateCompiler)
	templaters = make(map[string]Templater)
)

// TODO : How to manage Go nested templates?
// TODO : Support Go's port of the mustache template?

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

// Compile all templates that have a matching compiler (based on their extension) in the
// specified directory.
func CompileDir(dir string) error {
	return filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if fi == nil {
			return fmt.Errorf("directory %s does not exist", path)
		}
		if !fi.IsDir() {
			err = compileTemplate(path)
			if err != nil {
				ghost.LogFn("ghost.templates : error compiling template %s : %s", path, err)
				return err
			}
		}
		return nil
	})
}

// Compile the specified template file if there is a matching compiler.
func compileTemplate(p string) error {
	ext := path.Ext(p)
	c, ok := compilers[ext]
	// Ignore file if no template compiler exist for this extension
	if ok {
		t, err := c.Compile(p)
		if err != nil {
			return err
		}
		ghost.LogFn("ghost : storing template for file %s", p)
		templaters[p] = t
	}
	return nil
}

// Execute the template.
func Execute(tplName string, w io.Writer, data interface{}) error {
	t, ok := templaters[tplName]
	if !ok {
		return fmt.Errorf("no template found for file %s", tplName)
	}
	return t.Execute(w, data)
}

// Render is the same as Execute, except that it takes a http.ResponseWriter
// instead of a generic io.Writer, and sets the Content-Type to text/html.
func Render(tplName string, w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	return Execute(tplName, w, data)
}

// Defines the interface that the template compiler must return. The Go native
// templates implement this interface.
type Templater interface {
	Execute(wr io.Writer, data interface{}) error
}

// The interface that a template engine must implement to be used by Ghost.
type TemplateCompiler interface {
	Compile(fileName string) (Templater, error)
}
