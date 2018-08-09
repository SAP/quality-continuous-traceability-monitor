package projectmanagement

import (
	"bytes"
	"quality-continuous-traceability-monitor/mapping"
	"quality-continuous-traceability-monitor/utils"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
)

// jiraAPIPath contains the Jira API path (make sure this ends with a /)
const jiraAPIPath string = "/rest/api/2/issue/"

// CreateLinkInJiraBackLogItem creates a link in a Jira issue (Backlogitem) pointing to the corresponding entry for that backlog item
// in the traceability repository
func CreateLinkInJiraBackLogItem(cfg utils.Config, traces []Trace) {

	defer utils.TimeTrack(time.Now(), "Create links in Jira Backlog items")

	// Open cache file
	cacheFilepath := cfg.WorkDir + string(os.PathSeparator) + CacheFilename
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
		if trace.BacklogItem.Source == mapping.Jira {
			// Check if comment was already posted by checking cache file
			cacheEntry := strconv.FormatInt(int64(trace.BacklogItem.Source), 10) + "#" + trace.BacklogItem.ID
			if !strings.Contains(cache, cacheEntry) { // Create new entry
				// If a delivery version is set and we have a GitHub access token, than we can create GitHub releases
				branch := GetGHBranch(cfg)
				body := GetTestResultURL(cfg, trace.BacklogItem, branch)

				var jsonBody = []byte("{ \"body\": \"" + body + " \"}")
				var jiraBaseURL = cfg.Jira.BaseURL
				if jiraBaseURL[len(cfg.Jira.BaseURL)-1:] != "/" {
					jiraBaseURL = jiraBaseURL + "/"
				}
				var jiraURL = cfg.Jira.BaseURL + jiraAPIPath + trace.BacklogItem.ID + "/comment"

				req, err := http.NewRequest(http.MethodPost, jiraURL, bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req.SetBasicAuth(cfg.Jira.BasicAuth.User, cfg.Jira.BasicAuth.Password)

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					glog.Error("Unable to post traceability comment on Jira Backlog Item: "+trace.BacklogItem.ID+" Error was: ", err)
				} else {
					if resp.StatusCode != http.StatusCreated {
						glog.Error("Error posting traceability comment on Jira Backlog item: " + resp.Status)
					} else {
						cache = cache + "\n" + cacheEntry
					}
				}
				defer resp.Body.Close()
			}
		}
	}

	// Update cache file
	cf, err := os.Create(cacheFilepath)
	if err != nil {
		panic(err)
	}
	defer cf.Close()
	cf.WriteString(cache)
	cf.Sync()

}
