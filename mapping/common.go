package mapping

import (
	"errors"
	"fmt"
	"github.com/SAP/quality-continuous-traceability-monitor/testreport"
	"github.com/SAP/quality-continuous-traceability-monitor/utils"
	"github.com/golang/glog"
	"os"
	"regexp"
	"strings"
)

// Used to search for our backlog item markers in sourcecode files (could be comma separated list of multiple)
var reTraceMarker = regexp.MustCompile(`Trace\(((GitHub|Jira):([a-zA-Z0-9\-\/#\_])+\s*,*\s*)+\)`)

// A const for backlog item sources
const (
	Github int = 0
	Jira   int = 1
)

type TestCaseMatcher interface {
	Matches(*TestBacklog, *testreport.TestCase) bool
}

// Parser interface for programming language dependend parsers. cfg is complete configuration. sc is sourcecode repo/local path of current run
type Parser interface {
	Parse(cfg utils.Config, sc utils.Sourcecode) []TestBacklog
}

// Test is an automated test defined by a file, classname and method (which could be empty if complete class should be tested)
type Test struct {
	FileURL   string
	ClassName string
	Method    string
}

// BacklogItem is a requirement definition from a project management system
type BacklogItem struct {
	Source int
	ID     string
}

// TestBacklog maps an automated test to one or more backlogitems
type TestBacklog struct {
	Test            Test
	BacklogItem     []BacklogItem
	TestCaseMatcher TestCaseMatcher
}

func (tb *TestBacklog) Matches(tc *testreport.TestCase) bool {
	if tb.TestCaseMatcher == nil {
		return tc.ClassName == tb.Test.ClassName && (tc.MethodName == tb.Test.Method || tb.Test.Method == "")
	}

	return tb.TestCaseMatcher.Matches(tb, tc)
}

// GetGitHubOrganization returns the GitHub organization from a backlog item
func (bi BacklogItem) GetGitHubOrganization() (string, error) {

	if bi.Source != Github {
		return "", errors.New("Not an GitHub item")
	}

	return bi.ID[:strings.Index(bi.ID, "/")], nil

}

// GetGitHubRepository retruns  the GitHub repository from a backlog item
func (bi BacklogItem) GetGitHubRepository() (string, error) {

	if bi.Source != Github {
		return "", errors.New("Not an GitHub item")
	}

	return bi.ID[strings.Index(bi.ID, "/")+1 : strings.Index(bi.ID, "#")], nil

}

// GetGitHubIssue returns the GitHub issue from a backlog item
func (bi BacklogItem) GetGitHubIssue() (string, error) {

	if bi.Source != Github {
		return "", errors.New("Not an GitHub item")
	}

	return bi.ID[strings.Index(bi.ID, "#")+1:], nil

}

// GetTraceabilityRepoPath retruns the traceability repository path from a backlog item
func (bi BacklogItem) GetTraceabilityRepoPath() string {

	if bi.Source == Jira {
		return strings.Replace(bi.ID, "-", "/", -1)
	} else if bi.Source == Github {
		return strings.Replace(bi.ID, "#", "/", -1)
	}

	return bi.ID

}

// GetIssueURL returns the GitHub issue or Jira backlogitem URL
func (bi BacklogItem) GetIssueURL(cfg utils.Config) string {

	if bi.Source == Github {
		org, _ := bi.GetGitHubOrganization()
		repo, _ := bi.GetGitHubRepository()
		issue, _ := bi.GetGitHubIssue()
		return cfg.Github.BaseURL + "/" + org + "/" + repo + "/issues/" + issue
	} else if bi.Source == Jira {
		return cfg.Jira.BaseURL + "/browse/" + bi.ID
	} else {
		return "No link available"
	}

}

// GetBacklogItem constructs one or more BacklogItems from a traceability sourcecode comment
func GetBacklogItem(m string) []BacklogItem {

	// Check if there are multiple backlog items
	var mi []string
	if strings.Contains(m, ",") {
		mi = strings.Split(m, ",")
	} else {
		mi = append(mi, m)
	}

	var bi []BacklogItem
	for _, i := range mi {

		if !strings.Contains(i, ":") { // Try to check on valid trace entry
			glog.Warning("Found suspicious traceability comment in code: " + i + " in " + m)
			continue
		}

		// Get Backlog item Id (e.g. d036774/bulletinboard-ads#5 or JENKINSBCKLG-3)
		biID := i[strings.Index(i, ":")+1:]
		biID = strings.Replace(biID, " ", "", -1)
		// Get Backlog system
		biSource := i[:strings.Index(i, ":")]
		if strings.Contains(biID, ")") { // Last entry in list comes with a closing bracket...
			biID = biID[:strings.LastIndex(biID, ")")] // ... if so, cut if off
		}
		cbi := BacklogItem{}
		if strings.Contains(strings.ToLower(biSource), "github") {
			cbi = BacklogItem{Github, biID}
		} else if strings.Contains(strings.ToLower(biSource), "jira") {
			cbi = BacklogItem{Jira, biID}
		} else {
			// Report that we've found something strange here
			glog.Warningln("Found a backlog item from an unknown source:", m)
			// Create the backlog item. However we'll not contact any system to update the backlog item
			cbi = BacklogItem{-1, biID}
		}

		bi = append(bi, cbi)

	}

	return bi

}

func getSourcecodeURL(cfg utils.Config, sc utils.Sourcecode, file *os.File) string {
	// No Github information given -> we cannot create the sourcecode link
	if cfg.Github.BaseURL == "" || sc.Git.Organization == "" || sc.Git.Repository == "" || sc.Git.Branch == "" {
		return ""
	}

	ghBaseURL := cfg.Github.BaseURL

	if strings.HasSuffix(ghBaseURL, "/") {
		ghBaseURL = ghBaseURL[0 : len(ghBaseURL)-1]
	}

	fileName := file.Name()

	if strings.HasPrefix(fileName, sc.Local) {
		fileName = fileName[len(sc.Local):len(fileName)]
	}

	if strings.HasPrefix(fileName, "/") {
		fileName = fileName[1:len(fileName)]
	}

	templateParams := map[string]interface{}{"base": ghBaseURL, "git.org": sc.Git.Organization, "git.repository": sc.Git.Repository, "git.branch": sc.Git.Branch, "fileName": fileName}
	template := "%{base}/%{git.org}/%{git.repository}/blob/%{git.branch}/%{fileName}"

	if sc.URLTemplate != "" {
		template = sc.URLTemplate
	}

	return tPrintf(template, templateParams)
}

func tPrintf(template string, params map[string]interface{}) string {
	for key, val := range params {
		template = strings.ReplaceAll(template, "%{"+key+"}", fmt.Sprintf("%s", val))
	}

	return template
}
