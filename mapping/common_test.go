package mapping

import (
	"os"
	"testing"

	"github.com/SAP/quality-continuous-traceability-monitor/utils"
)

func TestGetSourcecodeURL(t *testing.T) {
	type testSample struct {
		Description    string
		GithubBaseUrl  string
		Git            utils.Git
		Local          string
		FilePath       string
		ExpectedResult string
	}

	testSamples := []testSample{
		testSample{
			GithubBaseUrl:  "https://github.com/",
			Git:            utils.Git{},
			FilePath:       "testFile.spec",
			ExpectedResult: "",
		},
		testSample{
			Description:   "GithubBaseUrl with slash at the end",
			GithubBaseUrl: "https://github.com/",
			Git: utils.Git{
				Organization: "myorg",
				Repository:   "myrepo",
				Branch:       "master",
			},
			FilePath:       "testFile.spec",
			ExpectedResult: "https://github.com/myorg/myrepo/blob/master/testFile.spec",
		},
		testSample{
			Description:   "GithubBaseUrl without slash at the end",
			GithubBaseUrl: "https://github.com",
			Git: utils.Git{
				Organization: "myorg",
				Repository:   "myrepo",
				Branch:       "master",
			},
			FilePath:       "testFile.spec",
			ExpectedResult: "https://github.com/myorg/myrepo/blob/master/testFile.spec",
		},
		testSample{
			Description:   "Local set, current dir",
			GithubBaseUrl: "https://github.com",
			Git: utils.Git{
				Organization: "myorg",
				Repository:   "myrepo",
				Branch:       "master",
			},
			Local:          ".",
			FilePath:       "subdir/testFile.spec",
			ExpectedResult: "https://github.com/myorg/myrepo/blob/master/subdir/testFile.spec",
		},
		testSample{
			Description:   "Local set, existing subdir",
			GithubBaseUrl: "https://github.com",
			Git: utils.Git{
				Organization: "myorg",
				Repository:   "myrepo",
				Branch:       "master",
			},
			Local:          "subdir/",
			FilePath:       "subdir/testFile.spec",
			ExpectedResult: "https://github.com/myorg/myrepo/blob/master/testFile.spec",
		},
		testSample{
			Description:   "Private github",
			GithubBaseUrl: "https://github.mycompany.local",
			Git: utils.Git{
				Organization: "myorg",
				Repository:   "myrepo",
				Branch:       "master",
			},
			FilePath:       "subdir/testFile.spec",
			ExpectedResult: "https://github.mycompany.local/myorg/myrepo/blob/master/subdir/testFile.spec",
		},
	}

	for i, ts := range testSamples {
		cfg := &utils.Config{}
		cfg.Github.BaseURL = ts.GithubBaseUrl

		sc := utils.Sourcecode{Git: ts.Git, Local: ts.Local}
		file := os.NewFile(0, ts.FilePath)

		expected := ts.ExpectedResult
		actual := getSourcecodeURL(*cfg, sc, file)

		if expected != actual {
			t.Errorf("Test of getSourcecodeURL (No. %d) failed: \nDescription: %s\nExpected: %s\nActual: %s", i, ts.Description, expected, actual)
		}
	}
}
