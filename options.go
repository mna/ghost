package ghost

type Options struct {
	TmplDir string // Defaults to ./tmpl
	PubDir  string
}

func (this *Options) fillWithDefault() {
	if this.TmplDir == "" {
		this.TmplDir = "./tmpl/pages/"
	}
	if this.PubDir == "" {
		this.PubDir = "./public/"
	}
}
