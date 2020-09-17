package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/open-policy-agent/conftest/output"
	"github.com/open-policy-agent/conftest/parser"
	"github.com/open-policy-agent/conftest/policy"
)

type TestRunner struct {
	Trace         bool
	Policy        []string
	Data          []string
	Update        []string
	Ignore        string
	Input         string
	Namespace     []string
	AllNamespaces bool `mapstructure:"all-namespaces"`
	FailOnWarn    bool `mapstructure:"fail-on-warn"`
	NoColor       bool `mapstructure:"no-color"`
	Combine       bool
	Output        string
}

// Run executes the TestRunner, verifying all Rego policies against the given
// list of configuration files.
func (t *TestRunner) Run(ctx context.Context, fileList []string) ([]output.CheckResult, error) {
	files, err := parseFileList(fileList, t.Ignore)
	if err != nil {
		return nil, fmt.Errorf("parse files: %w", err)
	}

	var configurations map[string]interface{}
	if t.Input != "" {
		configurations, err = parser.ParseConfigurationsAs(files, t.Input)
	} else {
		configurations, err = parser.ParseConfigurations(files)
	}
	if err != nil {
		return nil, fmt.Errorf("get configurations: %w", err)
	}

	loader := policy.Loader{
		DataPaths:   t.Data,
		PolicyPaths: t.Policy,
		URLs:        t.Update,
	}
	engine, err := loader.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}

	namespaces := t.Namespace
	if t.AllNamespaces {
		namespaces = engine.Namespaces()
	}

	var results []output.CheckResult
	for _, namespace := range namespaces {
		if t.Combine {
			result, err := engine.CheckCombined(ctx, configurations, namespace)
			if err != nil {
				return nil, fmt.Errorf("check combined: %w", err)
			}

			results = append(results, result)
		} else {
			result, err := engine.Check(ctx, configurations, namespace)
			if err != nil {
				return nil, fmt.Errorf("query rule: %w", err)
			}

			results = append(results, result...)
		}
	}

	return results, nil
}

func parseFileList(fileList []string, ignoreRegex string) ([]string, error) {
	var files []string
	for _, file := range fileList {
		if file == "" {
			continue
		}

		if file == "-" {
			files = append(files, "-")
			continue
		}

		fileInfo, err := os.Stat(file)
		if err != nil {
			return nil, fmt.Errorf("get file info: %w", err)
		}

		if fileInfo.IsDir() {
			directoryFiles, err := getFilesFromDirectory(file, ignoreRegex)
			if err != nil {
				return nil, fmt.Errorf("get files from directory: %w", err)
			}

			files = append(files, directoryFiles...)
		} else {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found")
	}

	return files, nil
}

func getFilesFromDirectory(directory string, ignoreRegex string) ([]string, error) {
	regexp, err := regexp.Compile(ignoreRegex)
	if err != nil {
		return nil, fmt.Errorf("given regexp couldn't be parsed :%w", err)
	}

	var files []string
	walk := func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk path: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		if ignoreRegex != "" && regexp.MatchString(currentPath) {
			return nil
		}

		for _, input := range parser.ValidInputs() {
			currentInput := strings.ToLower(input)

			if strings.HasSuffix(info.Name(), currentInput) {
				files = append(files, currentPath)
			}
		}

		return nil
	}

	err = filepath.Walk(directory, walk)
	if err != nil {
		return nil, err
	}

	return files, nil
}
