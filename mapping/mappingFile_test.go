package mapping

import (
	"ctm/utils"
	"strconv"
	"strings"
	"testing"
)

type testMapping struct {
	input          string
	expectedResult []TestBacklog
}

// TestBacklog mappings (correct)
var testJSONMappings = []testMapping{
	testMapping{input: `[
		{
			"source_reference": "com.myCompany.myApp.myJavaTest",
			"jira_keys": ["MYJIRAPROJECT-3"]
		}
	]`,
		expectedResult: []TestBacklog{{Test: Test{ClassName: "com.myCompany.myApp.myJavaTest", FileURL: "", Method: ""},
			BacklogItem: []BacklogItem{BacklogItem{ID: "MYJIRAPROJECT-3", Source: Jira}}}}},
	testMapping{input: `[
		{
			"source_reference": "com.myCompany.myApp.myJavaTest",
			"jira_keys": [
				"MYJIRAPROJECT-1"
			]
		},
		{
			"source_reference": "com.myCompany.myApp.myJavaTest.myMethod()",
			"github_keys": [
				"myOrg/mySourcecodeRepo#1"
			]
		},
		{
			"source_reference": "com.myCompany.myApp.myJavaTest.myOtherMethod()",
			"filelocation": {
					"git": {
						"organization": "myOrg",
						"repository": "mySourcecodeRepo",
						"branch": "master"
					},
					"relativePath": "./src/test/java/com/myCompany/myApp/myJavaTest.java"
			},
			"jira_keys": [
				"MYJIRAPROJECT-4",
				"MYJIRAPROJECT-5"
			],
			"github_keys": [
				"myOrg/mySourcecodeRepo#2"
			]
		}
	]`,
		expectedResult: []TestBacklog{
			{Test: Test{ClassName: "com.myCompany.myApp.myJavaTest", FileURL: "", Method: ""},
				BacklogItem: []BacklogItem{BacklogItem{ID: "MYJIRAPROJECT-1", Source: Jira}}},
			{Test: Test{ClassName: "com.myCompany.myApp.myJavaTest", FileURL: "", Method: "myMethod"},
				BacklogItem: []BacklogItem{BacklogItem{ID: "myOrg/mySourcecodeRepo#1", Source: Github}}},
			{Test: Test{ClassName: "com.myCompany.myApp.myJavaTest", FileURL: "https://github.com/myOrg/mySourcecodeRepo/blob/master/src/test/java/com/myCompany/myApp/myJavaTest.java", Method: "myOtherMethod"},
				BacklogItem: []BacklogItem{
					BacklogItem{ID: "MYJIRAPROJECT-4", Source: Jira},
					BacklogItem{ID: "MYJIRAPROJECT-5", Source: Jira},
					BacklogItem{ID: "myOrg/mySourcecodeRepo#2", Source: Github}}}}}}

func TestJSONParsing(t *testing.T) {

	cfg := new(utils.Config)
	cfg.Mapping.Local = "NonPersistedMappingFileForTesting"
	cfg.Github.BaseURL = "https://github.com"

	for i, mapping := range testJSONMappings {
		tb := parseJSON([]byte(mapping.input), *cfg)
		if !compareTestBacklog(tb, mapping.expectedResult) {
			t.Error("Comparism of JSON mapping (No. " + strconv.Itoa(i) + "): \n" + mapping.input + "\n with expected result failed.")
		}
	}

}

func compareTestBacklog(tb1, tb2 []TestBacklog) bool {

	// Simple quick check
	if len(tb1) != len(tb2) {
		return false
	}

	// Compare each TestBacklog from tb1 with all from tb2
	var tblCount = 0
	for _, currentTBL1 := range tb1 {
		for _, currentTBL2 := range tb2 {
			TBL1File := currentTBL1.Test.FileURL[strings.LastIndex(currentTBL1.Test.FileURL, "/")+1:]
			TBL2File := currentTBL2.Test.FileURL[strings.LastIndex(currentTBL2.Test.FileURL, "/")+1:]
			if currentTBL1.Test.ClassName == currentTBL2.Test.ClassName &&
				TBL1File == TBL2File &&
				currentTBL1.Test.Method == currentTBL2.Test.Method { // Test is equal

				// Simple (fast) precheck
				if len(currentTBL1.BacklogItem) != len(currentTBL2.BacklogItem) {
					return false
				}
				// Compare each Backlog item from currentTBL1 with all from currentTBL2
				var blCount = 0
				for _, currentTBL1bl := range currentTBL1.BacklogItem {
					for _, currentTBL2bl := range currentTBL2.BacklogItem {
						if currentTBL1bl.Source == currentTBL2bl.Source {
							if currentTBL1bl.ID == currentTBL2bl.ID { // Found equal Backlog
								blCount++
							}
						}
					}
				}
				if blCount == len(currentTBL1.BacklogItem) { // Found equal TestBacklog as equal Test and Backlogs are found in both
					tblCount++
				}
			}
		}
	}

	if tblCount == len(tb1) {
		return true
	}

	return false

}
