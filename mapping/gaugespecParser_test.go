package mapping

import (
	"os"
	"strings"
	"testing"

	"github.com/SAP/quality-continuous-traceability-monitor/testreport"
	"github.com/SAP/quality-continuous-traceability-monitor/utils"
)

var testGaugeSpecs = []testMapping{
	{
		input:          ``,
		expectedResult: []TestBacklog{},
	},
	{
		input: `
Trace: Jira:MYPROJECT-1

# Trace definitions outside a spec shall be ignored
        `,
		expectedResult: []TestBacklog{},
	},
	{
		input: `
# Just a spec, single requirement
Trace: Jira:MYPROJECT-13
        `,
		expectedResult: []TestBacklog{
			{
				Test: Test{
					ClassName: "Just a spec, single requirement",
					FileURL:   "testFile.spec",
					Method:    "",
				},
				BacklogItem: []BacklogItem{
					{ID: "MYPROJECT-13", Source: Jira},
				},
			},
		},
	},
	{
		input: `
# Just a spec, multiple requirements
Trace: Jira:MYPROJECT-13, GitHub:myorg/myRepo#1
        `,
		expectedResult: []TestBacklog{
			{
				Test: Test{
					ClassName: "Just a spec, multiple requirements",
					FileURL:   "testFile.spec",
					Method:    "",
				},
				BacklogItem: []BacklogItem{
					{ID: "MYPROJECT-13", Source: Jira},
					{ID: "myorg/myRepo#1", Source: Github},
				},
			},
		},
	},
	{
		input: `
# Spec and Scenarios

Trace: Jira:MYPROJECT-561

## Scenario 1

Trace: Jira:MYPROJECT-321

* do something
* do more

## Scenario 2

Trace: Jira:MYPROJECT-2567

* test step
        `,
		expectedResult: []TestBacklog{
			{
				Test: Test{
					ClassName: "Spec and Scenarios",
					FileURL:   "testFile.spec",
					Method:    "",
				},
				BacklogItem: []BacklogItem{
					{ID: "MYPROJECT-561", Source: Jira},
				},
			},
			{
				Test: Test{
					ClassName: "Spec and Scenarios",
					FileURL:   "testFile.spec",
					Method:    "Scenario 1",
				},
				BacklogItem: []BacklogItem{
					{ID: "MYPROJECT-321", Source: Jira},
				},
			},
			{
				Test: Test{
					ClassName: "Spec and Scenarios",
					FileURL:   "testFile.spec",
					Method:    "Scenario 2",
				},
				BacklogItem: []BacklogItem{
					{ID: "MYPROJECT-2567", Source: Jira},
				},
			},
		},
	},
}

func TestGaugeSpecParsing(t *testing.T) {
	uut := GaugeSpecParser{}

	cfg := new(utils.Config)
	cfg.Mapping.Local = "NonPersistedMappingFileForTesting"
	cfg.Github.BaseURL = "https://github.com"

	var sc = utils.Sourcecode{Git: utils.Git{Branch: "master", Organization: "testOrg", Repository: "testRepo"}, Language: "gaugespec", Local: "./"}
	var file = os.NewFile(0, "testFile.spec")

	for i, mapping := range testGaugeSpecs {
		tb := uut.ParseContent(strings.NewReader(mapping.input), *cfg, sc, file)
		if !compareTestBacklog(tb, mapping.expectedResult) {
			t.Errorf("Comparism of Gauge Spec (No. %d):\n%s\nwith expected result failed.", i, mapping.input)
		}
	}
}

/*
type TestCase struct {
    ReportFileName, // Test report file (e.g. Surefire XML)
    ClassName, // Test class
    MethodName string // Test method
    Result int // Test result
}

type Test struct {
    FileURL   string
    ClassName string
    Method    string
}
*/

func TestGaugeTestCaseMatcher(t *testing.T) {
	type testSample struct {
		Description    string
		TestCase       testreport.TestCase
		TestBacklog    TestBacklog
		ExpectedResult bool
	}

	uut := &GaugeTestCaseMatcher{}

	samples := []testSample{
		testSample{
			Description:    "ClassNames are different",
			TestCase:       testreport.TestCase{ClassName: "spec title", MethodName: ""},
			TestBacklog:    TestBacklog{Test: Test{ClassName: "another spec title", Method: ""}, TestCaseMatcher: uut},
			ExpectedResult: false,
		},
		testSample{
			Description:    "Class assignments should be assigned to methods too",
			TestCase:       testreport.TestCase{ClassName: "spec title", MethodName: "scenario title"},
			TestBacklog:    TestBacklog{Test: Test{ClassName: "spec title", Method: ""}, TestCaseMatcher: uut},
			ExpectedResult: true,
		},
		testSample{
			Description:    "Full match",
			TestCase:       testreport.TestCase{ClassName: "spec title", MethodName: "scenario title"},
			TestBacklog:    TestBacklog{Test: Test{ClassName: "spec title", Method: "scenario title"}, TestCaseMatcher: uut},
			ExpectedResult: true,
		},
		testSample{
			Description:    "Gauge specs with parameters",
			TestCase:       testreport.TestCase{ClassName: "spec title", MethodName: "scenario title 1"},
			TestBacklog:    TestBacklog{Test: Test{ClassName: "spec title", Method: "scenario title"}, TestCaseMatcher: uut},
			ExpectedResult: true,
		},
	}

	for i, sample := range samples {
		actual := sample.TestBacklog.Matches(&sample.TestCase)
		expected := sample.ExpectedResult

		if actual != expected {
			t.Errorf("Test of GaugeTestCaseMatcher (No. %d) failed: \n%s\nActual: %v\nExpected: %v\n", i, sample.Description, actual, expected)
		}
	}
}
