package projectmanagement

import (
	"github.com/SAP/quality-continuous-traceability-monitor/mapping"
	"github.com/go-test/deep"
	"testing"
)

func TestCreateRequirementsMapping(t *testing.T) {

	var traces = []Trace{
		{
			TraceTests: []TraceTest{
				{
					ClassName:  "Class1",
					MethodName: "WithMultipleRequirements",
				},
			},
			BacklogItem: mapping.BacklogItem{
				Source: 1,
				ID:     "JIRA-1",
			},
		},
		{
			TraceTests: []TraceTest{
				{
					ClassName:  "Class1",
					MethodName: "WithMultipleRequirements",
				},
			},
			BacklogItem: mapping.BacklogItem{
				Source: 1,
				ID:     "JIRA-2",
			},
		},
		{
			TraceTests: []TraceTest{
				{
					ClassName:  "Class1",
					MethodName: "SingleRequirement",
				},
			},
			BacklogItem: mapping.BacklogItem{
				Source: 1,
				ID:     "JIRA-3",
			},
		},
		{
			TraceTests: []TraceTest{
				{
					ClassName:  "Class1",
					MethodName: "GitHubRequirement",
				},
			},
			BacklogItem: mapping.BacklogItem{
				Source: 0,
				ID:     "GITHUB-4711",
			},
		},
	}

	var expectedMapping = []Mapping{
		{
			SourceReference: "Class1.WithMultipleRequirements()",
			JiraKeys:        []string{"JIRA-1", "JIRA-2"},
			GithubKeys:      []string{},
		},
		{
			SourceReference: "Class1.SingleRequirement()",
			JiraKeys:        []string{"JIRA-3"},
			GithubKeys:      []string{},
		},
		{
			SourceReference: "Class1.GitHubRequirement()",
			GithubKeys:      []string{"GITHUB-4711"},
			JiraKeys:        []string{},
		},
	}

	var generatedMapping = CreateRequirementsMapping(traces)

	if diff := deep.Equal(generatedMapping, expectedMapping); diff != nil {
		t.Error(diff)
	}

}
