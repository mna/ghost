// Ghostest is an interactive end-to-end Web site application to test
// the ghost packages. It serves the following URLs, with the specified
// features (handlers):
//
// / 										| panic;log;gzip;static; -> serve file index.html
// /public/styles.css 	| panic;log;gzip;StripPrefix;FileServer; -> serve directory public/
// /public/script.js 		| panic;log;gzip;StripPrefix;FileServer; -> serve directory public/
// /public/logo.png 		| panic;log;gzip;StripPrefix;FileServer; -> serve directory public/
// /session 						| panic;log;gzip;session;context;Custom; -> serve dynamic Go template
// /session/auth 				| panic;log;gzip;session;context;basicAuth;Custom; -> serve dynamic template
// /panic 							| panic;log;gzip;Custom; -> panics
// /context 						| panic;log;gzip;context;Custom1;Custom2; -> serve dynamic Amber template
//
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/PuerkitoBio/ghost/handlers"
	"github.com/PuerkitoBio/ghost/templates"
	_ "github.com/PuerkitoBio/ghost/templates/amber"
	_ "github.com/PuerkitoBio/ghost/templates/gotpl"
	"github.com/bmizerany/pat"
)

const (
	sessionPageTitle     = "Session Page"
	sessionPageAuthTitle = "Authenticated Session Page"
	sessionPageKey       = "txt"
)

var (
	// Create the common session store and secret
	memStore = handlers.NewMemoryStore(1)
	secret   = "testimony of the ancients"

	// Create the common session handler function
	fnSsnH = http.HandlerFunc(
		// The custom handler that renders the dynamic page
		func(w http.ResponseWriter, r *http.Request) {
			ssn, ok := handlers.GetSession(w)
			if !ok {
				panic("no session")
			}
			var txt interface{}
			if r.Method == "GET" {
				txt = ssn.Data[sessionPageKey]
			} else {
				txt = r.FormValue(sessionPageKey)
				ssn.Data[sessionPageKey] = txt
			}
			var data sessionPageInfo
			var title string
			if r.URL.Path == "/session/auth" {
				title = sessionPageAuthTitle
			} else {
				title = sessionPageTitle
			}
			if txt != nil {
				data = sessionPageInfo{title, txt.(string)}
			} else {
				data = sessionPageInfo{title, "[nil]"}
			}
			err := templates.Render("templates/session.tmpl", w, data)
			if err != nil {
				panic(err)
			}
		})

	// The no-auth required handler
	hSsn = handlers.SessionHandler(
		handlers.ContextHandler(fnSsnH, 1),
		handlers.NewSessionOptions(memStore, secret))

	// The Auth-required handler
	hAuthSsn = handlers.BasicAuthHandler(hSsn,
		// The authentication function
		func(u, p string) (interface{}, bool) {
			if u == "user" && p == "pwd" {
				return u + p, true
			}
			return nil, false
		}, "")
)

// The struct used to pass data to the session template.
type sessionPageInfo struct {
	Title string
	Text  string
}

func main() {
	// Blank the default logger's prefixes
	log.SetFlags(0)

	// Compile the dynamic templates (native Go templates are registered via the
	// for-side-effects-only import of gotpl)
	err := templates.CompileDir("./templates/", true)
	if err != nil {
		panic(err)
	}

	// Set the simple routes for static files
	mux := pat.New()
	mux.Get("/", handlers.StaticFileHandler("./index.html"))
	mux.Get("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public/"))))

	// Set the more complex route for session handling and dynamic page (same
	// handler is used for both GET and POST).
	mux.Get("/session", hSsn)
	mux.Post("/session", hSsn)
	mux.Get("/session/auth", hAuthSsn)
	mux.Post("/session/auth", hAuthSsn)

	// Set the handler for the chained context route
	hCtx1 := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx, ok := handlers.GetContext(w)
			if !ok {
				panic("no context")
			}
			ctx["time"] = time.Now().String()
		})
	hCtx2 := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx, ok := handlers.GetContext(w)
			if !ok {
				panic("no context")
			}
			err := templates.Render("templates/amber/context.amber", w, &struct{ Val string }{ctx["time"].(string)})
			if err != nil {
				panic(err)
			}
		})
	mux.Get("/context", handlers.ContextHandler(handlers.ChainHandlers(hCtx1, hCtx2), 1))

	// Set the panic route, which simply panics
	mux.Get("/panic", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			panic("explicit panic")
		}))

	// Combine the top level handlers, that wrap around the muxer.
	// Panic is the outermost, so that any panic is caught and responded to with a code 500.
	// Log is next, so that every request is logged along with the URL, status code and response time.
	// GZIP is then applied, so that content is compressed.
	// Finally, the muxer finds the specific handler that applies to the route.
	h := handlers.PanicHandler(
		handlers.LogHandler(
			handlers.GZIPHandler(
				mux),
			handlers.NewLogOptions(nil, handlers.Ltiny)),
		nil)

	// Assign the combined handler to the server.
	http.Handle("/", h)

	// Start it up.
	if err := http.ListenAndServe(":9000", nil); err != nil {
		panic(err)
	}
}
