package ghost

import (
	"testing"
)

func TestTemplates(t *testing.T) {
	Run(&Options{"./testdata/pages/"})
	t.Logf("%+v", pageRoutes)
}
