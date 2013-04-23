package ghost

import (
	"fmt"
	"io"
	"os"
	"path"
)

var (
	compilers  = make(map[string]TemplateCompiler)
	templaters = make(map[string]Templater)
)

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

func CompileDir(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	fis, err := f.Readdir(0)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		// Ignore directories
		if !fi.IsDir() {
			err = compileTemplate(path.Join(dir, fi.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func compileTemplate(p string) error {
	ext := path.Ext(p)
	c, ok := compilers[ext]
	// Ignore file if no template compiler exist for this extension
	if ok {
		t, err := c.Compile(p)
		if err != nil {
			return err
		}
		LogFn("ghost : storing template for file %s", p)
		templaters[p] = t
	}
	return nil
}

func Execute(tplName string, w io.Writer, data interface{}) error {
	t, ok := templaters[tplName]
	if !ok {
		return fmt.Errorf("no template found for file %s", tplName)
	}
	return t.Execute(w, data)
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
