// Ghostest is an interactive end-to-end Web site application to test
// the ghost packages. It serves the following URLs, with the specified
// features (handlers):
//
// / 										| panic;log;gzip;static; -> serve file index.html
// /public/styles.css 	| panic;log;gzip;StripPrefix;FileServer; -> serve directory public/
// /public/script.js 		| panic;log;gzip;StripPrefix;FileServer; -> serve directory public/
// /public/logo.jpg 		| panic;log;gzip;StripPrefix;FileServer; -> serve directory public/
// /session 						| panic;log;gzip;session;context;static; -> serve file session.html
// /session/auth 				| panic;log;gzip;basicauth;session;context;static; -> serve file auth.html
// /panic 							| panic;log;gzip;custom; -> panics
//
package main

import (
	"log"
	"net/http"

	"github.com/PuerkitoBio/ghost/handlers"
	"github.com/bmizerany/pat"
)

func main() {
	log.SetFlags(0)

	mux := pat.New()
	mux.Get("/", handlers.StaticFileHandler("./index.html"))
	mux.Get("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("/Users/martin/go/src/github.com/PuerkitoBio/ghost/ghostest/public/"))))
	h := handlers.PanicHandler(
		handlers.LogHandler(
			handlers.GZIPHandler(
				mux),
			handlers.NewLogOptions(nil, handlers.Ltiny)),
		nil)

	http.Handle("/", h)
	if err := http.ListenAndServe(":9000", nil); err != nil {
		panic(err)
	}
}
