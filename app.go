package ghost

import (
	"net/http"
	"sync"
)

// The App seems required to render templates. An experimental API could look
// like this:
//
// app := new(ghost.App)
// app.RegisterTemplateCompiler(ext string, c TemplateCompiler) - similar to gob.Register()
// app.CompileTemplates(path string, subdirs bool) - compile all templates in path
// app.ExecuteTemplate(path string, w io.Writer, data interface{}) error
//
// Internally, it keeps a map[string]TemplateCompiler for compilers, and a map[string]Templater
// of compiled templates. it uses locking, but best practice is to compile before starting
// the app.
//
// Now for the route handlers:
//
// app.Mux = pat|gorilla|trie|DefaultServeMux|whatever (pat modified for NotFound recommended)
//
// Automatically adds the AppProviderHandler, which replaces the ResponseWriter with an
// augmented one with an app field and GetApp(w) helper method.
type App struct {
	Env       string       // Env == "pprof" registers net/http/pprof handlers?
	H         http.Handler // Can be any handler or a Mux
	m         sync.RWMutex
	compilers map[string]TemplateCompiler
}

func (this *App) RegisterTemplateCompiler(ext string, c TemplateCompiler) {
	this.m.Lock()
	defer this.m.Unlock()
	m[ext] = c
}
