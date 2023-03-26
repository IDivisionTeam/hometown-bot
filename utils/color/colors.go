package color

// Discord color palette
const (
	White           int = 16777215
	Greyple         int = 10070709
	Black           int = 2303786
	DarkButNotBlack int = 2895667
	NotQuiteBlack   int = 2303786
	Blurple         int = 5793266
	Green           int = 5763719
	Yellow          int = 16705372
	Fuchsia         int = 15418782
	Red             int = 15548997
)

type ColorType int

const (
	Default ColorType = iota
	Success
	Warning
	Failure
)

func GetColorFrom(ct ColorType) int {
	switch ct {
	case Success:
		return Green
	case Warning:
		return Yellow
	case Failure:
		return Red
	default:
		return Greyple
	}
}
