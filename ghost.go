package ghost

import (
	"fmt"
	"github.com/bmizerany/pat"
	"github.com/eknkc/amber"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

var (
	pageRoutes map[string]*template.Template
)

func Run(opts *Options) {
	loadTemplates(opts)
	mux := buildRoutes(opts)
	http.Handle("/", mux)
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func buildRoutes(opts *Options) *pat.PatternServeMux {
	mux := pat.New()
	for k, tpl := range pageRoutes {
		mux.Get(fmt.Sprintf("/%s", k), http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			err := tpl.Execute(w, nil)
			if err != nil {
				log.Fatal("buildRoutes:", err)
			}
		}))
	}

	return mux
}

func loadTemplates(opts *Options) {
	// TODO : Recursive...

	dir, err := os.Open(opts.TmplDir)
	if err != nil {
		log.Fatal("os.Open:", err)
	}
	defer dir.Close()

	names, err := dir.Readdirnames(0)
	if err != nil {
		log.Fatal("os.Readdirnames:", err)
	}

	cmp := amber.New()
	pageRoutes = make(map[string]*template.Template, len(names))
	for _, nm := range names {
		if strings.ToLower(path.Ext(nm)) == ".amber" {
			base := path.Base(nm)
			base = base[:len(base)-len(".amber")]
			err := cmp.ParseFile(path.Join(opts.TmplDir, nm))
			if err != nil {
				log.Fatal("amber.ParseFile:", err)
			}
			tpl, err := cmp.Compile()
			if err != nil {
				log.Fatal("amber.Compile:", err)
			}
			pageRoutes[base] = tpl
		}
	}
}
