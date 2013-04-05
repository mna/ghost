package ghost

import (
	"testing"
)

func TestTemplates(t *testing.T) {
	Run(&Options{
		"./testdata/pages/",
		"./testdata/public/",
	})
	t.Logf("%+v", pageRoutes)
}
