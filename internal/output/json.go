package output

import (
	"encoding/json"

	"github.com/hha-nguyen/canopy-cli/internal/api"
)

type JSONFormatter struct {
	pretty bool
}

func NewJSONFormatter(pretty bool) *JSONFormatter {
	return &JSONFormatter{pretty: pretty}
}

func (f *JSONFormatter) Format(result *api.ScanResult) ([]byte, error) {
	if f.pretty {
		return json.MarshalIndent(result, "", "  ")
	}
	return json.Marshal(result)
}
