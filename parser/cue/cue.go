package cue

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	cformat "cuelang.org/go/cue/format"
)

// Parser is a CUE parser.
type Parser struct{}

// Unmarshal unmarshals CUE files.
func (*Parser) Unmarshal(p []byte, v interface{}) error {
	out, err := cformat.Source(p)
	if err != nil {
		return fmt.Errorf("format cue: %w", err)
	}

	var r cue.Runtime
	instance, err := r.Compile("name", out)
	if err != nil {
		return fmt.Errorf("compile cue: %w", err)
	}

	j, err := instance.Value().MarshalJSON()
	if err != nil {
		return fmt.Errorf("marshal cue to json: %w", err)
	}

	if err := json.Unmarshal(j, v); err != nil {
		return fmt.Errorf("unmarshal cue json: %w", err)
	}

	return nil
}
