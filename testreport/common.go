package testreport

const (
	// SUCCESS result of automated test
	SUCCESS int = 0
	// FAILURE result of automated test
	FAILURE int = 1
	// ERROR result of automated test
	ERROR int = 2
	// SKIPPED result of automated test
	SKIPPED int = 3
)

// TestReport interface implements the parse method which parses test reports
type TestReport interface {
	Parse(reportRootPath string) []TestSuite
}

// TestCase represents on automated test case incl. ReportFileName (the automated test result report file), ClassName (Test class), MethodName (Test method/procedure/function), Result (Test result)
type TestCase struct {
	ReportFileName, // Test report file (e.g. Surefire XML)
	ClassName, // Test class
	MethodName string // Test method
	Result int // Test result
}

// TestSuite is a collection of TestCase
type TestSuite struct {
	Name     string      // Name of Testsuite
	TestCase []*TestCase // Array of Testcases
}
