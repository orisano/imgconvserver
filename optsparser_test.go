package imgconvserver_test

import (
	"strings"
	"testing"

	"github.com/akito0107/imgconvserver"
	"github.com/go-test/deep"
)

func Test_Parse(t *testing.T) {
	input := `mount = "./"

[[directives]]
urlpattern = "/resize/{:dx}/{:dy}/{:imagename}.{:extension}"
format = "png"
function = "resize"
parameters = {dx = ":dx", dy = ":dy", quality = 0.7}
`
	conf := imgconvserver.Parse(strings.NewReader(input))

	m := map[string]interface{}{
		"dx":      ":dx",
		"dy":      ":dy",
		"quality": 0.7,
	}
	d := imgconvserver.Directive{
		UrlPattern: "/resize/{:dx}/{:dy}/{:imagename}.{:extension}",
		Format:     "png",
		Function:   "resize",
		Parameters: m,
	}
	c := &imgconvserver.ServerConfig{
		Mount:      "./",
		Directives: []imgconvserver.Directive{d},
	}

	if diff := deep.Equal(conf, c); diff != nil {
		t.Error(diff)
	}

}
