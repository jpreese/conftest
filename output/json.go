package output

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
)

// JSONOutputManager formats its output to JSON.
type JSONOutputManager struct {
	logger  *log.Logger
	data    []CheckResult
	tracing bool
}

// NewDefaultJSONOutputManager creates a new JSONOutputManager using the default logger.
func NewDefaultJSONOutputManager() *JSONOutputManager {
	return NewJSONOutputManager(log.New(os.Stdout, "", 0))
}

// NewJSONOutputManager creates a new JSONOutputManager with a given logger instance.
func NewJSONOutputManager(l *log.Logger) *JSONOutputManager {
	return &JSONOutputManager{
		logger: l,
	}
}

// WithTracing adds tracing to the output.
func (j *JSONOutputManager) WithTracing() OutputManager {
	j.tracing = true
	return j
}

// Put puts the result of the check to the manager in the managers buffer.
func (j *JSONOutputManager) Put(cr CheckResult) error {
	if cr.FileName == "-" {
		cr.FileName = ""
	}

	checkResult := CheckResult{
		FileName:   cr.FileName,
		Successes:  cr.Successes,
		Warnings:   []Result{},
		Failures:   []Result{},
		Exceptions: []Result{},
	}

	for _, warning := range cr.Warnings {
		result := Result{
			Message:  warning.Message,
			Metadata: warning.Metadata,
		}

		checkResult.Warnings = append(checkResult.Warnings, result)
	}

	for _, failure := range cr.Failures {
		result := Result{
			Message:  failure.Message,
			Metadata: failure.Metadata,
		}

		checkResult.Failures = append(checkResult.Failures, result)
	}

	for _, exception := range cr.Exceptions {
		result := Result{
			Message:  exception.Message,
			Metadata: exception.Metadata,
		}

		checkResult.Exceptions = append(checkResult.Exceptions, result)
	}

	if j.tracing {
		for _, query := range cr.Queries {
			queryResult := QueryResult{
				Query:   query.Query,
				Results: query.Results,
				Traces:  query.Traces,
			}

			checkResult.Queries = append(checkResult.Queries, queryResult)
		}
	}

	j.data = append(j.data, checkResult)
	return nil
}

// Flush writes the contents of the managers buffer to the console.
func (j *JSONOutputManager) Flush() error {
	b, err := json.Marshal(j.data)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		return err
	}

	j.logger.Print(out.String())
	return nil
}
