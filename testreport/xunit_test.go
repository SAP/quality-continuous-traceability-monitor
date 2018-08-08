package testreport

import (
	"testing"
)

func checkParsedTestSuite(t *testing.T, ts []TestSuite) {
	if len(ts) == 0 {
		t.Fatal("Result should not be empty")
	}

	if len(ts) != 1 {
		t.Error("Should parse only one test suite")
	}

	if len(ts[0].TestCase) != 1 {
		t.Fatal("Should parse exactly one test case in testsuite")
	}

	if ts[0].TestCase[0].ClassName != "XUNIT.Test" {
		t.Error("Invalid class name was parsed")
	}
}

func TestParseSingleTestsuite(t *testing.T) {
	fp := "test_path.xml"

	x := []byte(`
		<testsuite name="PhantomJS 2.1.1 (Mac OS X 0.0.0)" package="WTM" timestamp="2017-11-09T13:47:34" id="0" hostname="XXPM32390060A" tests="14" errors="0" failures="0" time="6.932">
			<properties>
				<property name="browser.fullName" value="Mozilla/5.0 (Macintosh; Intel Mac OS X) AppleWebKit/538.1 (KHTML, like Gecko) PhantomJS/2.1.1 Safari/538.1"/>
			</properties>
			<testcase name="This is simple test case" time="0.263" classname="XUNIT.Test"/>
		</testsuite>
  `)

	var ts = []TestSuite{}
	ts = parseXunitFile(fp, x, ts)
	checkParsedTestSuite(t, ts)
}

func TestParseMultipleTestsuites(t *testing.T) {
	fp := "test_path.xml"

	x := []byte(`
		<testsuites>
			<testsuite name="PhantomJS 2.1.1 (Mac OS X 0.0.0)" package="WTM" timestamp="2017-11-09T13:47:34" id="0" hostname="XXPM32390060A" tests="14" errors="0" failures="0" time="6.932">
				<properties>
					<property name="browser.fullName" value="Mozilla/5.0 (Macintosh; Intel Mac OS X) AppleWebKit/538.1 (KHTML, like Gecko) PhantomJS/2.1.1 Safari/538.1"/>
				</properties>
				<testcase name="This is simple test case" time="0.263" classname="XUNIT.Test"/>
			</testsuite>
		</testsuites>
  `)

	var ts = []TestSuite{}
	ts = parseXunitFile(fp, x, ts)
	checkParsedTestSuite(t, ts)
}

func TestParseInvalid(t *testing.T) {
	fp := "test_path.xml"

	x := []byte(`
		<testsuite name="PhantomJS 2.1.1 (Mac OS X 0.0.0)" package="WTM" timestamp="2017-11-09T13:47:34" id="0" hostname="XXPM32390060A" tests="14" errors="0" failures="0" time="6.932">
			<not-a-valid-testcase name="This is simple test case" time="0.263" classname="XUNIT.Test"/>
		</testsuite>
	`)

	var ts = []TestSuite{}
	ts = parseXunitFile(fp, x, ts)

	if len(ts) != 0 {
		t.Error("Should not parse a testsuite from invalid XML")
	}
}
