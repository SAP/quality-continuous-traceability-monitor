package projectmanagement

import (
	"fmt"
	"github.com/SAP/quality-continuous-traceability-monitor/mapping"
	"github.com/SAP/quality-continuous-traceability-monitor/testreport"
	"github.com/SAP/quality-continuous-traceability-monitor/utils"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const pmGitHubPath string = "GitHub"
const pmJiraPath string = "Jira"

// CacheFilename specifies the name of the CTM cachefile
const CacheFilename string = ".commentCache"

// TraceTest maps an automated test (ClassName/MethodName) to it's result (TestResult), with the test result report (e.g. XUNIT) and the test sourcecode file
type TraceTest struct {
	SourceFile string
	ReportFile string
	ClassName  string
	MethodName string
	TestResult int
}

// Trace maps a TraceTest (automated test and result) to a BacklogItem
type Trace struct {
	TraceTests  []TraceTest
	BacklogItem mapping.BacklogItem
}

// Copy a file on OS level
func Copy(src, dst string) (int64, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	srcFileStat, err := srcFile.Stat()
	if err != nil {
		return 0, err
	}

	if !srcFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer dstFile.Close()
	return io.Copy(dstFile, srcFile)
}

// GetTestResultURL return a link to the test result in the traceability repository
// branch can be used to overwrite the default branch from traceability repo config
// e.g. in case of a release tag, you want to overwrite the default branch (e.g. master with 1.5.0)
// so that link will point to a git tag
func GetTestResultURL(cfg utils.Config, item mapping.BacklogItem, branch string) string {

	httpsURL := utils.GetRepositoryHTTPSUrl(cfg, cfg.TraceabilityRepo.Git)
	httpsURL = httpsURL + "/tree/" + branch + "/"
	if item.Source == 0 {
		httpsURL = httpsURL + pmGitHubPath
	} else if item.Source == 1 {
		httpsURL = httpsURL + pmJiraPath
	}

	return httpsURL + "/" + item.GetTraceabilityRepoPath()

}

// GetGHBranch calculates the correct git branch for linking into the traceability repository
func GetGHBranch(cfg utils.Config) string {

	branch := cfg.TraceabilityRepo.Git.Branch
	// If a delivery version is set and we have a GitHub access token, than we can create GitHub releases
	if cfg.Delivery.Version != "" && cfg.Github.AccessToken != "" {
		branch = cfg.Delivery.Version
	}

	return branch

}

// GetNumberOfSuccessfulTestedTraces returns the number of successfully tested requirements
func GetNumberOfSuccessfulTestedTraces(traces []Trace) int {

	var successfulReq int
	for _, trace := range traces {
		if trace.TraceTests != nil { // Trace has tests
			var allSuccessfull = true
			for _, test := range trace.TraceTests {
				if test.TestResult != testreport.SUCCESS {
					allSuccessfull = false
					break
				}
			}
			if allSuccessfull {
				successfulReq++
			}
		}
	}

	return successfulReq

}

// CreateHTMLReport creates a HTML report file (and returns the file reference)
// filepath - the dirpath where to create the file
// traces - list of all traces from the sourcecode
// cfg - An ctm config struct
func CreateHTMLReport(filepath string, traces []Trace, cfg utils.Config, fullReport bool) *os.File {

	f, err := os.Create(filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	const header string = `<head>
                             <title> Full Software Requirement Test Report</title>
                             <meta name="author" content="SAP Continuous Traceability Monitor">
                             <style>
							   body{
								 font-family: Arial, Verdana;
							   }
							   table{
								 border-collapse: collapse;
							   }
	                           h2{
	                             color: #666666;
	                           }
	                           h3{
	                             color: #666666;
	                           }
							   div.code{
								 font-family: "Courier New", "Lucida Console";
							   }
							   th{
								 border-top: 1px solid #ddd;
							   }
							   th, td{
								 padding: 12px;
								 text-align: left;
								 border-bottom: 1px solid #ddd;
								 border-right: 1px solid #ddd;
							   }
							   tr:nth-child(even) {
								 background-color: #f2f2f2;
							   }
                               .nobullets {
                                 list-style-type:none;
                                 padding-left: 0;
                                 padding-bottom: 0;
                                 margin: 0;
                               }
                               .notok {
                                 background-color: #ffe5e5;
                                 padding: 5px
                               }
                               .ok {
                                 background-color: #e1f5a9;
                                 padding: 5px
                               }
							   .green{
								 color: #4FB810;
							   }
							   .red{
								 color: #E35500;
							   }
                             </style>
                           </head >`

	const body string = `<body>
                           <h1>Full Software Requirement Test Report</h1>
						   %programAndVersion%
	                       <div><h3>Total number of requirements: %totalNumberOfRequirements%<br/>
	                            Total number of successful requirements: %totalNumberOfSuccessfulRequirements%</h3></div>
	                       <p><div style="color:#666666"><i>Snapshot taken: %timestamp%</i></div></p>
                           <hr/>
                           <table>
                             <tr>
                               <th>#</th>
                               <th>Backlog ID</th>
                               <th>Test Mapping</th>
                             </tr>
	                         %tabledata%
                           </table>
	                     </body>
	`

	f.WriteString("<html>")

	f.WriteString(header)

	var table string
	var successfulReq int
	for i, trace := range traces {
		table = table + "<tr><td>" + strconv.Itoa(i+1) + "</td>"
		table = table + "<td><a href=\"" + trace.BacklogItem.GetIssueURL(cfg) + "\" target=\"_blank\">%backlogItem%</a></td>"
		var tests = "<td><div><ul class=\"nobullets\">"
		var allSuccessful = true
		if trace.TraceTests == nil { // We have traces, but no test results
			allSuccessful = false
			tests = tests + "<li class=\"notok\"><b>Missing</b>"
		} else {
			for _, test := range trace.TraceTests {
				if test.TestResult == testreport.SUCCESS {
					tests = tests + "<li class=\"ok\"><b>OK&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;</b>"
				} else if test.TestResult == testreport.SKIPPED {
					allSuccessful = false
					tests = tests + "<li class=\"notok\"><b>not OK</b>"
				} else {
					allSuccessful = false
					tests = tests + "<li class=\"notok\"><b>not OK</b>"
				}
				var sourceCodeLink string
				if test.SourceFile != "" {
					sourceCodeLink = "<a href=\"" + test.SourceFile + "\" target=\"_blank\">"
				}
				tests = tests + ": " + sourceCodeLink + test.ClassName
				if test.MethodName != "" {
					tests = tests + "." + test.MethodName
				}
				if sourceCodeLink != "" {
					tests = tests + "</a>"
				}
				tests = tests + "</li>"
			}
		}
		if allSuccessful {
			successfulReq++
			table = strings.Replace(table, "%backlogItem%", trace.BacklogItem.ID, 1)
		} else {
			table = strings.Replace(table, "%backlogItem%", "<span class=\"notok\">"+trace.BacklogItem.ID+"</span>", 1)
		}
		table = table + tests + "</div></ul></tr>"
	}

	data := strings.Replace(body, "%timestamp%", time.Now().UTC().Format(time.RFC1123), 1)
	var version string
	if !fullReport && cfg.Delivery.Version != "" {
		version = "<h2>Program: <i>" + cfg.Delivery.Program + "</i>&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Delivery: <i>" + cfg.Delivery.Version + "</i></h2>"
	}
	data = strings.Replace(data, "%programAndVersion%", version, 1)
	data = strings.Replace(data, "%tabledata%", table, 1)
	data = strings.Replace(data, "%totalNumberOfRequirements%", strconv.FormatInt(int64(len(traces)), 10), 1)
	if successfulReq == len(traces) {
		data = strings.Replace(data, "%totalNumberOfSuccessfulRequirements%", "<span class=\"green\">"+strconv.FormatInt(int64(successfulReq), 10)+"</span>", 1)
	} else {
		data = strings.Replace(data, "%totalNumberOfSuccessfulRequirements%", "<span class=\"red\">"+strconv.FormatInt(int64(successfulReq), 10)+"</span>", 1)
	}

	f.WriteString(data)

	if traces == nil {
		f.WriteString("<div><h2><center><span class=\"red\"><b>No issues traced to automated tests yet.</b></span></center></h2></div>")
	}

	f.WriteString("</html>")

	f.Sync()

	return f

}

// CreateJSONReport creates a JSON report file (for potential further electronical processing of traceability result)
// filepath - the dirpath where to create the file
// traces - list of all traces from the sourcecode
// cfg - An ctm config struct
func CreateJSONReport(filepath string, traces []Trace, cfg utils.Config) *os.File {

	f, err := os.Create(filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// We serialize the JSON by ourself. Using the json.Marshal function would require us to create a new
	// go struct for each backlog item, as the report JSON has each backlog item as an own object (with the backlog
	// item ID as name)
	const INTENT = "    "
	f.WriteString("{\n")

	for i, trace := range traces {
		f.WriteString(INTENT + "\"" + trace.BacklogItem.ID + "\": {\n")
		f.WriteString(INTENT + INTENT + "\"link\": \"" + trace.BacklogItem.GetIssueURL(cfg) + "\",\n")
		f.WriteString(INTENT + INTENT + "\"test_cases\": [\n")
		for j, testCase := range trace.TraceTests {
			f.WriteString(INTENT + INTENT + INTENT + "{\n")
			var fullname string
			if testCase.MethodName != "" {
				fullname = testCase.ClassName + "." + testCase.MethodName
			} else {
				fullname = testCase.ClassName
			}
			f.WriteString(INTENT + INTENT + INTENT + INTENT + "\"test_fullname\": \"" + fullname + "\",\n")
			f.WriteString(INTENT + INTENT + INTENT + INTENT + "\"test_name\": \"" + testCase.MethodName + "\",\n")
			f.WriteString(INTENT + INTENT + INTENT + INTENT + "\"test_class\": \"" + testCase.ClassName + "\",\n")
			if testCase.SourceFile != "" {
				f.WriteString(INTENT + INTENT + INTENT + INTENT + "\"test_source\": \"" + testCase.SourceFile + "\",\n")
			}
			if testCase.TestResult == testreport.SUCCESS {
				f.WriteString(INTENT + INTENT + INTENT + INTENT + "\"passed\": true,\n")
			} else {
				f.WriteString(INTENT + INTENT + INTENT + INTENT + "\"passed\": false,\n")
			}
			if testCase.TestResult == testreport.SKIPPED {
				f.WriteString(INTENT + INTENT + INTENT + INTENT + "\"skipped\": true\n")
			} else {
				f.WriteString(INTENT + INTENT + INTENT + INTENT + "\"skipped\": false\n")
			}
			if j != len(trace.TraceTests)-1 {
				f.WriteString(INTENT + INTENT + INTENT + "},\n")
			} else {
				f.WriteString(INTENT + INTENT + INTENT + "}\n")
			}
		}
		f.WriteString(INTENT + INTENT + "]\n")
		if i != len(traces)-1 {
			f.WriteString(INTENT + "},\n")
		} else {
			f.WriteString(INTENT + "}\n")
		}
	}
	f.WriteString("}\n")

	f.Sync()

	return f

}
