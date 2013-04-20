package ghost

import (
	"log"
	"net/http"
	"path"
	"strings"
	"sync"
)

// Logging function, defaults to Go's native log.Printf function. The idea to use
// this instead of a *log.Logger struct is that it can be set to any of log.{Printf,Fatalf, Panicf},
// but also to more flexible userland loggers like SeeLog (https://github.com/cihub/seelog).
// It could be set, for example, to SeeLog's Debugf function. Any function with the
// signature func(fmt string, params ...interface{}).
var LogFn = log.Printf

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
//
// TODO : OR, the App may not be necessary if the template rendering provider is itself
// a handler? Like GetTemplater(w, path) Templater?
type App struct {
	*http.Server              // TODO: Embeds a native http Server but should hide parts
	Env          string       // Env == "pprof" registers net/http/pprof handlers?
	H            http.Handler // Can be any handler or a Mux

	// Internal fields
	mc        sync.RWMutex
	mt        sync.RWMutex
	compilers map[string]TemplateCompiler
	templates map[string]Templater
}

func (this *App) Run() {

}

func (this *App) RegisterTemplateCompiler(ext string, c TemplateCompiler) {
	this.mc.Lock()
	defer this.mc.Unlock()
	this.compilers[ext] = c
}

func (this *App) CompileTemplates(path string, recursive bool) error {
	return nil
}

func (this *App) compileTemplate(p string) error {
	this.mc.RLock()
	defer this.mc.RUnlock()
	ext := strings.ToLower(path.Ext(p))
	c, ok := this.compilers[ext]
	if !ok {
		return nil // ErrNoTemplateCompiler
	}
	t, err := c.Compile(p)
	if err != nil {
		return err
	}
	this.addTemplater(strings.ToLower(p), t)
	return nil
}

func (this *App) addTemplater(p string, t Templater) {
	this.mt.Lock()
	defer this.mt.Unlock()
	this.templates[strings.ToLower(p)] = t
}
