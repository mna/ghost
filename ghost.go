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
	opts.fillWithDefault()
	loadTemplates(opts)
	mux := buildRoutes(opts)
	mux.Get("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(opts.PubDir))))
	http.Handle("/", mux)
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func buildRoutes(opts *Options) *pat.PatternServeMux {
	mux := pat.New()
	for k, tpl := range pageRoutes {
		func(nm string, t *template.Template) {
			mux.Get(fmt.Sprintf("/%s", nm), http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				err := t.Execute(w, nil)
				if err != nil {
					log.Fatal("buildRoutes:", err)
				}
			}))
		}(k, tpl)
	}

	return mux
}

/*
The templates are loaded from a directory specified by the Options. It is possible
with amber to inherit or include portions of the page from another amber file. These
incomplete but reusable amber files should *not* be in the same directory as the 
"real" pages, because then Ghost will generate useless templates for them.

The recommended directory hierarchy is as follows (master and header are examples
of reusable, non-pages amber files, * denotes a directory, - denotes a file):
* ./
  * tmpl
    * pages
      * users
        * id
          - user.amber
        - users.amber
      - default.amber
    - master.amber
    - header.amber
  * public
    - site.js
    - styles.css
    - logo.png
    - favicon.ico
*/
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
