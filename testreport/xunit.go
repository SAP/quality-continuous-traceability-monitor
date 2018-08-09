package testreport

import (
	"quality-continuous-traceability-monitor/utils"
	"encoding/xml"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/golang/glog"
)

// XUFailure xunit failure struct
type XUFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Text    string `xml:",chardata"`
}

// XUError xunit error struct
type XUError struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Text    string `xml:",chardata"`
}

// XUSkipped xunit skipped struct
type XUSkipped struct {
	Message string `xml:"message,attr,omitempty"`
}

// XUProperties xunit properties struct
type XUProperties struct {
	Property []*XUProperty `xml:"property,omitempty"`
}

// XUProperty xunit property struct
type XUProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// XUSystemOut xunit system out structure
type XUSystemOut struct {
	Text string `xml:",chardata" json:",omitempty"`
}

// XUSystemErr xunit system err structure
type XUSystemErr struct {
	Text string `xml:",chardata" json:",omitempty"`
}

// XUTestcase xunit test case structure
type XUTestcase struct {
	Classname    string       `xml:"classname,attr,omitempty"`
	Group        string       `xml:"group,attr,omitempty"`
	Name         string       `xml:"name,attr"`
	Time         string       `xml:"time,attr,omitempty"`
	File         string       `xml:"file,attr,omitempty"` // Python specific
	Line         string       `xml:"line,attr,omitempty"` // Python specific
	Error        *XUError     `xml:"error,omitempty"`
	Skipped      *XUSkipped   `xml:"skipped,omitempty"`
	Failure      *XUFailure   `xml:"failure,omitempty"`
	ReRunFailure *XUFailure   `xml:"rerunFailure,omitempty"`
	SystemOut    *XUSystemOut `xml:"system-out,omitempty"`
	SystemErr    *XUSystemErr `xml:"system-err,omitempty"`
}

// XUTestsuite xunit test suite structure
type XUTestsuite struct {
	Errors                 string        `xml:"errors,attr"`
	Failures               string        `xml:"failures,attr"`
	Name                   string        `xml:"name,attr"`
	Skipped                string        `xml:"skipped,attr,omitempty"`
	Group                  string        `xml:"group,attr,omitempty"` // omit empty to support Python
	Skips                  string        `xml:"skips,attr,omitempty"` // Python specific
	Tests                  string        `xml:"tests,attr"`
	Time                   string        `xml:"time,attr,omitempty"`
	Properties             *XUProperties `xml:"properties,omitempty"`
	Testcase               []*XUTestcase `xml:"testcase,omitempty"`
	XmlnsXsi               string        `xml:"xmlns xsi,attr,omitempty"`
	XsiSpaceSchemaLocation string        `xml:"http://www.w3.org/2001/XMLSchema-instance schemaLocation,attr,omitempty"`
}

// XUTestsuites xunit test suites structure
type XUTestsuites struct {
	Testsuite []*XUTestsuite `xml:"testsuite,omitempty"`
}

// XUTestReport xunit test report structure
type XUTestReport struct {
}

// Convenience method as we support skipped (Xunit) and skips (phyton xunit) in the testsuite xml element
// Should always be used when checking on skipped test cases
func (xuts *XUTestsuite) getSkipped() int {

	if xuts.Skipped != "" {
		skip, _ := strconv.ParseInt(xuts.Skipped, 10, 32)
		return int(skip)
	}

	if xuts.Skips != "" {
		skips, _ := strconv.ParseInt(xuts.Skips, 10, 32)
		return int(skips)
	}

	return 0

}

// Parse a xunit XML test result report
func (xutr *XUTestReport) Parse(reportRootPath string) []TestSuite {
	defer utils.TimeTrack(time.Now(), "Scan xunit XML test reports")

	var ts = []TestSuite{}

	filepath.Walk(reportRootPath, func(path string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return nil
		}

		// We're only interssted in xml files
		if filepath.Ext(path) == ".xml" {
			glog.Info("Parsing ", path)
			xmlFile, err := os.Open(path)
			if err != nil {
				glog.Error("Unable to open file: ", err)
			}
			defer xmlFile.Close()

			xfb, _ := ioutil.ReadAll(xmlFile)
			ts = parseXunitFile(path, xfb, ts)
		}

		return nil
	})

	return ts
}

func parseXunitFile(xmlFilePath string, xfb []byte, ts []TestSuite) []TestSuite {
	// try parsing with single root testsuite
	var xuTestsuite XUTestsuite
	xml.Unmarshal(xfb, &xuTestsuite)
	if xuTestsuite.Testcase != nil {
		return addXUTestSuiteToTestResult(xmlFilePath, ts, xuTestsuite)
	}

	// try parsing with root testsuites
	var xuTestsuites XUTestsuites
	xml.Unmarshal(xfb, &xuTestsuites)
	if xuTestsuites.Testsuite != nil {
		for _, xuTestsuite := range xuTestsuites.Testsuite {
			ts = addXUTestSuiteToTestResult(xmlFilePath, ts, *xuTestsuite)
		}
		return ts
	}

	glog.Info("No test cases found in ", xmlFilePath)
	return ts
}

// Map the xunit specific test suite (and within that test cases) to the common (generalized)
// TestSuite struct
func addXUTestSuiteToTestResult(xmlFile string, ts []TestSuite, xuts XUTestsuite) []TestSuite {
	// Iterate through all xunit test cases, convert them and add them to an array
	var testcases []*TestCase
	for _, xutestcase := range xuts.Testcase {
		var result = FAILURE
		if xutestcase.Failure == nil && xutestcase.Error == nil && xutestcase.Skipped == nil {
			result = SUCCESS
		}
		testcase := &TestCase{xmlFile, xutestcase.Classname, xutestcase.Name, result}
		testcases = append(testcases, testcase)
	}

	// Create test suite and add it to test report
	return append(ts, TestSuite{xuts.Name, testcases})
}
