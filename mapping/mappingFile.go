package mapping

import (
	"ctm/utils"
	"encoding/json"
	"io/ioutil"
	"strings"
	"time"

	"github.com/golang/glog"
)

// JSONMappingFile is a parser for JSON mapping file (which map Backlog items to automated test classes). Pls. note, that this class does NOT implement the mapping.Parser interface
type JSONMappingFile struct {
}

type mappingFileContent []struct {
	SourceReference string `json:"source_reference"`
	Filelocation    struct {
		Git struct {
			Organization string `json:"organization"`
			Repository   string `json:"repository"`
			Branch       string `json:"branch"`
		} `json:"git"`
		RelativePath string `json:"relativePath"`
	} `json:"filelocation,omitempty"`
	JiraKeys   []string `json:"jira_keys,omitempty"`
	GithubKeys []string `json:"github_keys,omitempty"`
}

// Parse JSON mapping files of format:
// [
//   {
//     "source_reference": "com.myCompany.myApp.myJavaTest",
//     "jira_keys": [
//       "MYJIRAPROJECT-3"
//     ]
//   },
//   {
//     "source_reference": "com.myCompany.myApp.myJavaTest.myMethod()",
//     "github_keys": [
//       "myOrg/mySourcecodeRepo#1"
//     ]
//   },
//   {
//     "source_reference": "com.myCompany.myApp.myJavaTest.myOtherMethod()",
//     "filelocation": {
//         "git": {
//           "organization": "myOrg",
//           "repository": "mySourcecodeRepo",
//           "branch": "master"
//         },
//         "relativePath": "./src/test/java/com/myCompany/myApp/myJavaTest.java"
//     },
//     "jira_keys": [
//       "MYJIRAPROJECT-1",
//       "MYJIRAPROJECT-2"
//     ],
//     "github_keys": [
//       "myOrg/mySourcecodeRepo#2"
//     ]
//   }
// ]
func (mf JSONMappingFile) Parse(cfg utils.Config) []TestBacklog {

	defer utils.TimeTrack(time.Now(), "Read mapping file")

	f, err := ioutil.ReadFile(cfg.Mapping.Local)
	if err != nil {
		panic(err)
	}

	return parseJSON(f, cfg)

}

func parseJSON(mappingFile []byte, cfg utils.Config) []TestBacklog {

	var mfc *mappingFileContent
	// Get struct data type from json file
	err := json.Unmarshal(mappingFile, &mfc)
	if err != nil {
		glog.Fatal("Unable to read local mapping file (", cfg.Mapping.Local, "). Are you sure this is a valid JSON mapping file? Nested error was: ", err)
	}

	var tb = []TestBacklog{}

	for _, entry := range *mfc { // Converting each file entry into our struct
		// Create Test struct
		var methodname, classname string
		if strings.HasSuffix(entry.SourceReference, "()") { // Its a method
			methodname = entry.SourceReference[strings.LastIndex(entry.SourceReference, ".")+1:]
			methodname = strings.TrimSuffix(methodname, "()") // Remove () from methodname
			classname = entry.SourceReference[0:strings.LastIndex(entry.SourceReference, ".")]
		} else {
			methodname = ""
			classname = entry.SourceReference
		}

		var fileurl string
		if entry.Filelocation.RelativePath != "" {
			relPath := entry.Filelocation.RelativePath
			relPath = strings.Trim(relPath, ".")
			relPath = strings.Trim(relPath, "/")
			fileurl = cfg.Github.BaseURL + "/" + entry.Filelocation.Git.Organization + "/" + entry.Filelocation.Git.Repository + "/blob/" + entry.Filelocation.Git.Branch + "/" + relPath
		}
		test := Test{ClassName: classname, Method: methodname, FileURL: fileurl}

		// Create all assigned backlog items
		var bli []BacklogItem
		for _, jk := range entry.JiraKeys {
			bli = append(bli, BacklogItem{Source: Jira, ID: jk})
		}
		for _, ghk := range entry.GithubKeys {
			bli = append(bli, BacklogItem{Source: Github, ID: ghk})
		}

		// Create TestBacklog struct
		tb = append(tb, TestBacklog{BacklogItem: bli, Test: test})

	}

	return tb

}
