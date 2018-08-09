package projectmanagement   

import (
	"context"
	"github.com/SAP/quality-continuous-traceability-monitor/mapping"
	"github.com/SAP/quality-continuous-traceability-monitor/testreport"
	"github.com/SAP/quality-continuous-traceability-monitor/utils"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// gitHubEnterpriseAPIPath contains the GitHub API path for the GitHub Enterprise version
const gitHubEnterpriseAPIPath string = "api/v3/"

// gitHubAPIURL contains the GitHub API URL for the public GitHub version
const gitHubAPIURL string = "https://api.github.com/"

// gitHubBaseURL contains the URL for the public GitHub version
const gitHubBaseURL string = "https://github.com"

// GitHubClient contains a GitHubClient to access GitHub as well as a CTM configuration
type GitHubClient struct {
	gh     *github.Client
	config utils.Config
}

// CreateGitHubClient creates a GitHub client which uses an access token for authentication
func CreateGitHubClient(cfg utils.Config) *GitHubClient {

	ctx := context.Background()

	var ghAPI string
	if strings.Contains(cfg.Github.BaseURL, "github.com") { // Set public API
		ghAPI = gitHubAPIURL
	} else { // Set enterprise API
		ghAPI = cfg.Github.BaseURL
		if ghAPI[len(ghAPI)-1:] != "/" {
			ghAPI = ghAPI + "/"
		}
		ghAPI = ghAPI + gitHubEnterpriseAPIPath
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Github.AccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	ghAPIURL, err := url.Parse(ghAPI)
	if err != nil {
		panic(err)
	}
	client.BaseURL = ghAPIURL

	return &GitHubClient{client, cfg}

}

// UpdateTraceabilityRepository updates the traceability repository with the latest traceability reports (test result, ...)
func UpdateTraceabilityRepository(traces, deliveryTraces []Trace, client *GitHubClient) {

	defer utils.TimeTrack(time.Now(), "Create GitHub report")

	// Clone traceability repository
	repoPath := client.config.WorkDir + string(os.PathSeparator) + client.config.TraceabilityRepo.Git.Repository

	utils.CloneOrPullRepo(utils.GetRepositorySSHUrl(client.config, client.config.TraceabilityRepo.Git), repoPath, true)

	repoPath = repoPath + string(os.PathSeparator)

	createBacklogFolder(repoPath, traces, client.config)

	var readme *os.File
	if client.config.Delivery.Version != "" {
		readme = createDeliveryFolder(repoPath, deliveryTraces, client.config)

		deliveryPath := repoPath + string(os.PathSeparator) + "Deliveries" + string(os.PathSeparator) + client.config.Delivery.Version
		CreateHTMLReport(deliveryPath+string(os.PathSeparator)+client.config.Delivery.Program+"_"+client.config.Delivery.Version+".html", deliveryTraces, client.config, false)
		CreateJSONReport(deliveryPath+string(os.PathSeparator)+client.config.Delivery.Program+"_"+client.config.Delivery.Version+".json", deliveryTraces, client.config)
	}

	// General README.md file (for complete set/repo)
	createReadme(repoPath, traces, false, client.config)

	utils.CommitAndPush(repoPath)

	// Create the release after we did the commit and push to ensure the release tag
	// is pointing to our commit (which has all the updated README files)
	if client.config.Delivery.Version != "" {
		createRelease(client, readme)
	}

}

// Create a Markdown README file
// path - the dirpath where to create the file
// traces - list of all traces from the sourcecode
// verbose - Should we add the test class name to the report
// cfg - An ctm config struct
func createReadme(path string, traces []Trace, verbose bool, cfg utils.Config) *os.File {

	f, err := os.Create(path + "/README.md")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Header
	f.WriteString("# Traceability Summary Report\n")
	if cfg.Delivery.Version != "" {
		f.WriteString("## Delivery Version: " + cfg.Delivery.Version + "\n")
	}
	f.WriteString("  \n")
	f.WriteString("  \n")

	// Table
	f.WriteString("Backlog Item | Test result ")
	if verbose {
		f.WriteString(" | Test classes")
	}
	f.WriteString("\n------------ | ----------- ")

	if verbose {
		f.WriteString("| -----------\n")
	} else {
		f.WriteString("\n")
	}

	var biSource string
	var testResult string
	var testClass string
	for _, trace := range traces {

		testResult = ""
		testClass = ""
		if trace.BacklogItem.Source == mapping.Github {
			biSource = "[GitHub:" + trace.BacklogItem.ID + "](" + trace.BacklogItem.GetIssueURL(cfg) + ")"
		} else {
			biSource = "[Jira:" + trace.BacklogItem.ID + "](" + trace.BacklogItem.GetIssueURL(cfg) + ")"
		}
		if trace.TraceTests == nil {
			testResult = ":heavy_exclamation_mark:"
			testClass = "Missing"
		} else {
			for _, tt := range trace.TraceTests {
				// If a delivery version is set and we have a GitHub access token, than we can create GitHub releases
				branch := GetGHBranch(cfg)
				if tt.TestResult == testreport.SUCCESS && !strings.Contains(testResult, ":x:") { // One failing test, fails the complete backlog item
					testResult = "[:heavy_check_mark:](" + GetTestResultURL(cfg, trace.BacklogItem, branch) + ")"
				} else if tt.TestResult == testreport.FAILURE {
					testResult = "[:x:](" + GetTestResultURL(cfg, trace.BacklogItem, branch) + ")"
				}
				if verbose {
					var classAndMethod string
					if tt.MethodName != "" {
						classAndMethod = tt.ClassName + " - " + tt.MethodName
					} else {
						classAndMethod = tt.ClassName
					}
					if tt.SourceFile != "" {
						testClass = testClass + " * [" + classAndMethod + "](" + tt.SourceFile + ") => "
					} else {
						testClass = testClass + " * " + classAndMethod + " => "
					}

					if tt.TestResult == testreport.SUCCESS {
						testClass = testClass + ":heavy_check_mark:<br>"
					} else if tt.TestResult == testreport.FAILURE {
						testClass = testClass + ":x:<br>"
					}
				}
			}
		}
		f.WriteString(biSource + " | " + testResult)
		if verbose {
			f.WriteString(" | " + testClass)
		}
		f.WriteString("  \n")
	}

	// Footer
	f.WriteString("  \n")
	f.WriteString("  \n")

	if traces == nil {
		f.WriteString("### No issues traced to automated tests yet.")

		f.WriteString("  \n")
		f.WriteString("  \n")
	}

	f.WriteString("##### _Report generated " + time.Now().UTC().Format(time.RFC1123) + "_  \n")
	f.WriteString("-----\n")
	f.WriteString("made with &#10084; by SAP")

	f.Sync()

	return f

}

func createBacklogFolder(tmpRepoPath string, traces []Trace, cfg utils.Config) {

	for _, trace := range traces {

		// Create backlog item path (if it doesn't exist)
		var pmPath = tmpRepoPath + string(os.PathSeparator)
		if trace.BacklogItem.Source == mapping.Github {
			pmPath = pmPath + pmGitHubPath
		} else if trace.BacklogItem.Source == mapping.Jira {
			pmPath = pmPath + pmJiraPath
		}
		var biPath = pmPath + string(os.PathSeparator) + trace.BacklogItem.GetTraceabilityRepoPath()
		os.MkdirAll(biPath, os.FileMode(0755))

		for _, test := range trace.TraceTests {
			var rfPath = biPath + string(os.PathSeparator) + path.Base(test.ReportFile)
			Copy(test.ReportFile, rfPath)
		}

		createReadme(biPath, []Trace{trace}, true, cfg)

	}

}

func createDeliveryFolder(tmpRepoPath string, deliveryTraces []Trace, cfg utils.Config) *os.File {

	delPath := tmpRepoPath + string(os.PathSeparator) + "Deliveries" + string(os.PathSeparator) + cfg.Delivery.Version
	os.MkdirAll(delPath, os.FileMode(0755))

	return createReadme(delPath, deliveryTraces, true, cfg)

}

// CreateLinkInGHBackLogItem creates a link in a GitHub issue (Backlogitem) pointing to the corresponding entry for that backlog item
// in the traceability repository
func CreateLinkInGHBackLogItem(client *GitHubClient, traces []Trace) {

	defer utils.TimeTrack(time.Now(), "Create links in GitHub Backlog items")

	// Open cache file
	cacheFilepath := client.config.WorkDir + string(os.PathSeparator) + CacheFilename
	f, err := ioutil.ReadFile(cacheFilepath)
	if err != nil {
		if os.IsNotExist(err) { // Cache does not exist
			glog.Warning("Cache file not found. Create new one at " + cacheFilepath)
		} else { // Something else happened
			panic(err)
		}
	}
	var cache = ""
	if f != nil {
		cache = string(f)
	}

	for _, trace := range traces {
		if trace.BacklogItem.Source == mapping.Github {
			// Get GitHub organization and repo name
			orgName, _ := trace.BacklogItem.GetGitHubOrganization()
			repoName, _ := trace.BacklogItem.GetGitHubRepository()
			// Check if comment was already posted by checking cache file
			cacheEntry := strconv.FormatInt(int64(trace.BacklogItem.Source), 10) + "#" + trace.BacklogItem.ID
			if !strings.Contains(cache, cacheEntry) { // Create new entry
				// If a delivery version is set and we have a GitHub access token, than we can create GitHub releases
				branch := GetGHBranch(client.config)
				body := GetTestResultURL(client.config, trace.BacklogItem, branch)
				ic := github.IssueComment{Body: &body}
				iid, _ := trace.BacklogItem.GetGitHubIssue()
				idNumber, _ := strconv.ParseInt(iid, 10, 32)
				_, _, err := client.gh.Issues.CreateComment(context.Background(), orgName, repoName, int(idNumber), &ic)
				if err != nil {
					glog.Error("Unable to post traceability comment on GitHub Backlog Item: "+trace.BacklogItem.ID+" Error was: ", err)
				} else {
					cache = cache + "\n" + cacheEntry
				}
			}
		}
	}

	// Update cache file
	cf, err := os.Create(cacheFilepath)
	if err != nil {
		glog.Error("Unable to write comment cache file: " + cacheFilepath + " (This might result in posting doublicate comments at your backlog items!)")
	}
	defer cf.Close()
	cf.WriteString(cache)
	cf.Sync()

}

func createRelease(client *GitHubClient, file *os.File) {

	c, _ := ioutil.ReadFile(file.Name())
	content := string(c)

	release := github.RepositoryRelease{Name: &client.config.Delivery.Version, Body: &content, TagName: &client.config.Delivery.Version}

	_, _, err := client.gh.Repositories.CreateRelease(context.Background(), client.config.TraceabilityRepo.Git.Organization, client.config.TraceabilityRepo.Git.Repository, &release)
	if err != nil {
		glog.Error("Cannot create GitHub release: ", err)
	}

}
