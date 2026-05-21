package output

import (
	"io"

	"gopkg.in/yaml.v3"
)

func formatYAML(w io.Writer, data any) error {
	enc := yaml.NewEncoder(w)
	defer enc.Close()
	return enc.Encode(data)
}
