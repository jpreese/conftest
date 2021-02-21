package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// JSON represents an Outputter that outputs
// results in JSON format.
type JSON struct {
	Writer io.Writer
}

// NewJSON creates a new JSON with the given writer. 
func NewJSON(w io.Writer) *JSON {
	jsonOutput := JSON{
		Writer: w,
	}

	return &jsonOutput
}

// Output outputs the results.
func (j *JSON) Output(results []CheckResult) error {
	for r := range results {
		if results[r].FileName == "-" {
			results[r].FileName = ""
		}

		results[r].Queries = nil
	}

	b, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	var out bytes.Buffer
	if err := json.Indent(&out, b, "", "\t"); err != nil {
		return fmt.Errorf("indent: %w", err)
	}

	fmt.Fprintln(j.Writer, out.String())
	return nil
}
