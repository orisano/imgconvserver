package imgconvserver

import (
	"log"
	"net/http"
	"strconv"

	"github.com/akito0107/imgconvserver/engine"
	"github.com/go-chi/chi"
)

func MakeHandler(d *Directive) func(w http.ResponseWriter, r *http.Request) {
	eng, ok := engine.Engines[d.Engine]
	if !ok {
		log.Fatalf("Unsupported Engine %s", d.Engine)
	}

	resizer, ok := eng.(engine.Resizer)
	for _, c := range d.Converts {
		if !ok && c.Function == "resize" {
			log.Fatalf("Unsupported Resize function on engine %s", d.Engine)
		}
	}

	encoder, ok := eng.(engine.Encoder)
	if !ok {
		log.Println("using Default Encoder")
		encoder = &engine.DefaultEncoder{}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		filepath := chi.URLParam(r, ":filepath")
		im, err := openImage(filepath)
		if err != nil {
			http.Error(w, http.StatusText(400), 400)
			return
		}
		for _, c := range d.Converts {
			switch c.Function {
			case "Resize":
				wid := chi.URLParam(r, ":dw")
				hgt := chi.URLParam(r, ":dh")

				dw, err := strconv.Atoi(wid)
				if err != nil {
					http.Error(w, http.StatusText(400), 400)
					return
				}
				dh, err := strconv.Atoi(hgt)
				if err != nil {
					http.Error(w, http.StatusText(400), 400)
					return
				}
				im, err = resizer.Resize(im, dw, dh)
				if err != nil {
					http.Error(w, http.StatusText(500), 500)
					return
				}
			default:
				log.Printf("unsupported function pattern %s\n", c.Function)
				http.Error(w, http.StatusText(500), 500)
				return
			}
		}
		of := d.Format
		op := &engine.EncodeOptions{
			Format: FromString(of),
		}
		encoder.Encode(w, im, op)
	}
}
