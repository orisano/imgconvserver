package imgconvserver

type Format int

const (
	PNG Format = iota
	JPEG
)

func FromString(str string) Format {
	switch str {
	case "png":
		return PNG
	case "jpeg", "JPEG", "jpg":
		return JPEG
	}
}
