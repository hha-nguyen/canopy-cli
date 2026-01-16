package output

import "github.com/hha-nguyen/canopy-cli/internal/api"

type Formatter interface {
	Format(result *api.ScanResult) ([]byte, error)
}

type ProgressFormatter interface {
	FormatProgress(percentage int, phase string) string
}

func NewFormatter(format string, noColor bool) Formatter {
	switch format {
	case "json":
		return NewJSONFormatter(false)
	case "sarif":
		return NewSARIFFormatter()
	default:
		return NewTextFormatter(noColor)
	}
}
