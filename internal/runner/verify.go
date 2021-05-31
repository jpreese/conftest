package runner

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/open-policy-agent/conftest/output"
	"github.com/open-policy-agent/conftest/policy"
	"github.com/open-policy-agent/opa/cover"
	"github.com/open-policy-agent/opa/tester"
	"github.com/open-policy-agent/opa/topdown"
)

// VerifyRunner is the runner for the Verify command, executing
// Rego policy unit-tests.
type VerifyRunner struct {
	Coverage int
	Policy   []string
	Data     []string
	Output   string
	NoColor  bool `mapstructure:"no-color"`
	Trace    bool
}

// Run executes the Rego tests for the given policies.
func (r *VerifyRunner) Run(ctx context.Context) ([]output.CheckResult, error) {
	engine, err := policy.LoadWithData(ctx, r.Policy, r.Data)
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}

	if r.Trace {
		engine.EnableTracing()
	}

	cov := cover.New()

	runner := tester.NewRunner().
		SetCompiler(engine.Compiler()).
		SetStore(engine.Store()).
		SetModules(engine.Modules()).
		//EnableTracing(r.Trace).
		SetRuntime(engine.Runtime()).
		SetCoverageQueryTracer(cov)

	ch, err := runner.RunTests(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("running tests: %w", err)
	}

	/*
		var reporter tester.Reporter
		reporter = tester.JSONCoverageReporter{
			Cover:     cov,
			Modules:   engine.Modules(),
			Output:    os.Stdout,
			Threshold: 50,
		}*/

	var results []output.CheckResult
	for result := range ch {
		if result.Error != nil {
			return nil, fmt.Errorf("run test: %w", result.Error)
		}

		buf := new(bytes.Buffer)
		topdown.PrettyTrace(buf, result.Trace)
		var traces []string
		for _, line := range strings.Split(buf.String(), "\n") {
			if len(line) > 0 {
				traces = append(traces, line)
			}
		}

		var outputResult output.Result
		if result.Fail || result.Skip {
			outputResult.Message = result.Package + "." + result.Name
		}

		queryResult := output.QueryResult{
			Query:   result.Name,
			Results: []output.Result{outputResult},
			Traces:  traces,
		}

		checkResult := output.CheckResult{
			FileName: result.Location.File,
			Queries:  []output.QueryResult{queryResult},
		}
		if result.Fail {
			checkResult.Failures = []output.Result{outputResult}
		} else if result.Skip {
			checkResult.Skipped = []output.Result{outputResult}
		} else {
			checkResult.Successes++
		}

		results = append(results, checkResult)
	}

	if r.Coverage > 0 {

	}

	report := cov.Report(engine.Modules())

	fmt.Println("below")
	fmt.Println(cov.Report(engine.Modules()).Files["examples/kubernetes/policy/labels.rego"].Coverage)
	fmt.Println(report.Coverage)

	return results, nil
}
