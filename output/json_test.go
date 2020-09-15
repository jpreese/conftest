package output

import (
	"bytes"
	"log"
	"testing"
)

func TestJSON(t *testing.T) {
	type args struct {
		fileName string
		crs      []CheckResult
	}

	tests := []struct {
		msg  string
		args args
		exp  string
	}{
		{
			msg: "no Warnings or errors",
			args: args{
				crs: []CheckResult{
					{FileName: "examples/kubernetes/service.yaml"},
				},
			},
			exp: `[
	{
		"filename": "examples/kubernetes/service.yaml",
		"successes": 0,
		"warnings": [],
		"failures": [],
		"exceptions": []
	}
]
`,
		},
		{
			msg: "records failure and Warnings",
			args: args{
				crs: []CheckResult{
					{
						FileName: "examples/kubernetes/service.yaml",
						Warnings: []Result{{Message: "first warning"}},
						Failures: []Result{{Message: "first failure"}},
					},
				},
			},
			exp: `[
	{
		"filename": "examples/kubernetes/service.yaml",
		"successes": 0,
		"warnings": [
			{
				"msg": "first warning"
			}
		],
		"failures": [
			{
				"msg": "first failure"
			}
		],
		"exceptions": []
	}
]
`,
		},
		{
			msg: "mixed failure and Warnings",
			args: args{
				crs: []CheckResult{
					{
						FileName: "examples/kubernetes/service.yaml",
						Failures: []Result{
							{Message: "first failure"},
						},
					},
				},
			},
			exp: `[
	{
		"filename": "examples/kubernetes/service.yaml",
		"successes": 0,
		"warnings": [],
		"failures": [
			{
				"msg": "first failure"
			}
		],
		"exceptions": []
	}
]
`,
		},
		{
			msg: "handles stdin input",
			args: args{
				fileName: "-",
				crs: []CheckResult{
					{
						Failures: []Result{
							{Message: "first failure"},
						},
					},
				},
			},
			exp: `[
	{
		"filename": "",
		"successes": 0,
		"warnings": [],
		"failures": [
			{
				"msg": "first failure"
			}
		],
		"exceptions": []
	}
]
`,
		},
		{
			msg: "multiple check results",
			args: args{
				crs: []CheckResult{
					{FileName: "examples/kubernetes/service.yaml"},
					{FileName: "examples/kubernetes/deployment.yaml"},
				},
			},
			exp: `[
	{
		"filename": "examples/kubernetes/service.yaml",
		"successes": 0,
		"warnings": [],
		"failures": [],
		"exceptions": []
	},
	{
		"filename": "examples/kubernetes/deployment.yaml",
		"successes": 0,
		"warnings": [],
		"failures": [],
		"exceptions": []
	}
]
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := NewJSONOutputManager(log.New(buf, "", 0))

			for _, cr := range tt.args.crs {
				if err := s.Put(cr); err != nil {
					t.Fatalf("put output: %v", err)
				}
			}

			if err := s.Flush(); err != nil {
				t.Fatalf("flush output: %v", err)
			}

			actual := buf.String()

			if tt.exp != actual {
				t.Errorf("unexpected output. expected %v got %v", tt.exp, actual)
			}
		})
	}
}
