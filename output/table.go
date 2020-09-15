package output

import (
	"io"
	"os"

	table "github.com/olekukonko/tablewriter"
)

// TableOutputManager formats its output in a table
type TableOutputManager struct {
	table   *table.Table
	tracing bool
}

// NewDefaultTableOutputManager creates a new TableOutputManager using standard out
func NewDefaultTableOutputManager() *TableOutputManager {
	return NewTableOutputManager(os.Stdout)
}

// NewTableOutputManager creates a new TableOutputManager with a given Writer
func NewTableOutputManager(w io.Writer) *TableOutputManager {
	table := table.NewWriter(w)
	table.SetHeader([]string{"result", "file", "message"})
	return &TableOutputManager{
		table: table,
	}
}

// WithTracing adds tracing to the output.
func (t *TableOutputManager) WithTracing() OutputManager {
	t.tracing = true
	return t
}

// Put puts the result of the check to the manager in the managers buffer
func (t *TableOutputManager) Put(cr CheckResult) error {
	if t.tracing {
		table := table.NewWriter(os.Stdout)
		table.SetHeader([]string{"passed", "file", "query", "trace"})
		t.table = table

		for _, queryResult := range cr.Queries {
			for _, trace := range queryResult.Traces {
				var result string
				if queryResult.Passed() {
					result = "success"
				} else {
					result = "fail"
				}

				line := []string{result, cr.FileName, queryResult.Query, trace}
				table.Append(line)
			}
		}

		return nil
	}

	printResults := func(r Result, prefix string, filename string) {
		d := []string{prefix, filename, r.Message}
		t.table.Append(d)
	}

	for s := 0; s < cr.Successes; s++ {
		printResults(Result{}, "success", cr.FileName)
	}

	for _, r := range cr.Warnings {
		printResults(r, "warning", cr.FileName)
	}

	for _, r := range cr.Failures {
		printResults(r, "failure", cr.FileName)
	}

	return nil
}

// Flush writes the contents of the managers buffer to the console
func (t *TableOutputManager) Flush() error {
	if t.table.NumLines() > 0 {
		t.table.Render()
	}

	return nil
}
