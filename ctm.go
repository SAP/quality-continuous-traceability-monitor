package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/SAP/quality-continuous-traceability-monitor/mapping"
	"github.com/SAP/quality-continuous-traceability-monitor/projectmanagement"
	"github.com/SAP/quality-continuous-traceability-monitor/testreport"
	"github.com/SAP/quality-continuous-traceability-monitor/utils"

	"github.com/golang/glog"
)

const reportBaseName = "ctm_report"

// To be replaced by Make (using LDFLAGS) setting the latest git commit ID
var ctmVersion = "1.0.6"

// Supported test result formats
// xunit-xml format = https://github.com/windyroad/JUnit-Schema/blob/master/JUnit.xsd
var supportedReporttypes = []string{"xunit-xml"}

// Checks if a test report type (like maven-surefire, etc.) is already supported
func reportTypeSupported(reportType string) bool {
	for _, current := range supportedReporttypes {
		if current == reportType {
			return true
		}
	}
	return false
}

func setupLogging(cfg utils.Config) {

	// Setup logging
	var loglevel = "INFO" // INFO as default log level
	if strings.ToLower(cfg.Log.Level) == "warning" {
		loglevel = "WARNING"
	} else if strings.ToLower(cfg.Log.Level) == "error" {
		loglevel = "ERROR"
	} else if strings.ToLower(cfg.Log.Level) == "fatal" {
		loglevel = "FATAL"
	}
	flag.Set("stderrthreshold", loglevel)
	flag.Parse()

}

func addTraceTest(traces []projectmanagement.Trace, backlogItems *[]mapping.BacklogItem, tt projectmanagement.TraceTest) []projectmanagement.Trace {

	for _, backlogItem := range *backlogItems {
		var found = false
		for i, trace := range traces {
			// Check if we already have a Trace for this backlog item
			if trace.BacklogItem == backlogItem {
				trace.TraceTests = append(trace.TraceTests, tt)
				traces[i] = trace
				found = true
				break
			}
		}
		if !found {
			// Create new Trace for this BacklogItem
			var tta = []projectmanagement.TraceTest{tt}
			t := projectmanagement.Trace{TraceTests: tta, BacklogItem: backlogItem}
			traces = append(traces, t)
		}
	}

	return traces

}

// Maps a backlog items to test results
func createTraces(testSuite []testreport.TestSuite, backlogItems []mapping.TestBacklog) []projectmanagement.Trace {

	defer utils.TimeTrack(time.Now(), "Create Traces")
	var traces []projectmanagement.Trace
	var tt projectmanagement.TraceTest
	for _, sourceCodeTest := range backlogItems { // Iterating over all marked traceability relevant tests from sourcecode

		// Check whether we find the corresponding test case
		for _, ts := range testSuite { // Checking in each test suite...
			for _, tc := range ts.TestCase { // ...to find the test case
				if sourceCodeTest.Matches(tc) {
					tt = projectmanagement.TraceTest{SourceFile: sourceCodeTest.Test.FileURL, ReportFile: tc.ReportFileName, ClassName: tc.ClassName, MethodName: tc.MethodName, TestResult: tc.Result}
					traces = addTraceTest(traces, &sourceCodeTest.BacklogItem, tt)
				}
			}
		}

	}

	sort.Slice(traces, func(i, j int) bool {

		if traces[i].BacklogItem.Source == traces[j].BacklogItem.Source {
			return traces[i].BacklogItem.ID < traces[j].BacklogItem.ID
		}
		return traces[i].BacklogItem.Source < traces[j].BacklogItem.Source

	})

	return traces

}

// Get all traces which are relevant for the given delivery
func getDeliveryTraces(traces []projectmanagement.Trace, cfg utils.Config) []projectmanagement.Trace {

	var deliveryTraces = []projectmanagement.Trace{}
	selBacklogitems := mapping.GetBacklogItem(cfg.Delivery.Backlogitems)
	for _, trace := range traces {
		// Iterate over all backlog items (from delivery) and check whether its assigned to the current trace
		for _, selBli := range selBacklogitems {
			if trace.BacklogItem == selBli {
				deliveryTraces = append(deliveryTraces, trace)
				continue
			}
		}
	}

	// Check if all backlog items (from delivery) are found. If a backlog item is missing add a dummy one to indicate missing test
	for _, dBli := range selBacklogitems {
		if !containsBacklogItem(deliveryTraces, dBli) {
			deliveryTraces = append(deliveryTraces, projectmanagement.Trace{TraceTests: nil, BacklogItem: dBli})
		}
	}

	return deliveryTraces

}

func containsBacklogItem(t []projectmanagement.Trace, bli mapping.BacklogItem) bool {
	for _, a := range t {
		if a.BacklogItem == bli {
			return true
		}
	}
	return false
}

func main() {

	// Command line arguments we're taking
	argConfigFile := flag.String("c", "./myConfig.json", "Configuration file for CTM")
	argCommandLineVersion := flag.String("sd", "", "Delivery version")
	argCommandLineProgram := flag.String("sp", "", "Delivery program")
	argSelectiveBacklogItems := flag.String("bi", "", "Comma separated list of delivery relevant backlog items")
	argDeliveryFile := flag.String("df", "", "Delivery file")
	argVersion := flag.Bool("version", false, "CTM Version")
	argExportRequirementsMapping := flag.Bool("erm", false, "Export an requirements mapping file")

	// Initialy set log level to INFO. (Once we've parsed the actual configuration, well set the desired loglevel)
	flag.Set("stderrthreshold", "INFO")

	// Get commandline arguments and read config
	flag.Parse()

	// Check whether we're only called to print version
	if *argVersion {
		fmt.Println(ctmVersion)
		os.Exit(0)
	}

	// Read config
	cfg := utils.Config{}
	cfg.ReadConfig(argConfigFile)

	// If Delivery file is given, read it and set it in cfg
	if *argDeliveryFile != "" {
		cfg.ReadDelivery(argDeliveryFile)
	}

	// Command line delivery arguments overwrite any file specified delivery parameters
	// Check if we need to overwrite the config file parameter with a command line parameter
	if *argCommandLineVersion != "" {
		cfg.Delivery.Version = *argCommandLineVersion
	}
	if *argCommandLineProgram != "" {
		cfg.Delivery.Program = *argCommandLineProgram
	}
	if *argSelectiveBacklogItems != "" {
		cfg.Delivery.Backlogitems = *argSelectiveBacklogItems
	}

	// Configure glog framework
	setupLogging(cfg)

	// Check mapping mode. Parse source code repositories or read mapping file?
	var biMapping []mapping.TestBacklog
	if cfg.Mapping.Local != "" {
		// Read the mapping.json file so we get the traceability relevant test classes and methods incl. their related
		// backlog items
		p := mapping.JSONMappingFile{}
		biMapping = p.Parse(cfg)
	} else {
		// Parse the source code so we get the traceability relevant test classes and methods incl. their related
		// backlog items
		var p mapping.Parser
		for _, sc := range cfg.Sourcecode {
			switch sc.Language {
			case "java":
				p = mapping.JavaParser{}
				break
			case "python":
				p = mapping.PythonParser{}
				break
			case "javascript":
				p = mapping.JSParser{}
				break
			case "gaugespec":
				p = mapping.GaugeSpecParser{}
				break
			default:
				glog.Fatal("Sourcecode language for parsing needs to be 'python', 'java', 'javascript' or 'gaugespec'")
			}

			biMapping = append(biMapping, p.Parse(cfg, sc)...)
		}
	}

	// Parse the test report to get test results
	var testSuite = []testreport.TestSuite{}
	for _, tr := range cfg.TestReport {
		if reportTypeSupported(tr.Type) {
			trXU := testreport.XUTestReport{}
			suites := trXU.Parse(tr.Local)
			// Ensure we don't collect doublicates
			for _, s := range suites {
				found := false
				if s.TestCase == nil { // A Test suite with not Testcases. Seen that with Mocca framework
					glog.Info("Empty Test Suite found")
					continue
				}
				for _, cts := range testSuite {
					if (cts.Name == s.Name) && (len(cts.TestCase) == len(s.TestCase)) && (s.TestCase[0].ReportFileName == cts.TestCase[0].ReportFileName) {
						found = true
						break
					}
				}
				if !found {
					testSuite = append(testSuite, s)
				}
			}
		} else {
			glog.Error("Unsupported test report format. Supported formats are: ", supportedReporttypes)
		}
	}

	// Map backlog items (from sourcecode) to test results
	traces := createTraces(testSuite, biMapping)

	if traces == nil {
		glog.Warning("++++ No issues traced to automated tests yet! Generated reports will be empty. (Just saying)")
	}

	// Get list of delivery relevant traces
	var deliveryTraces = []projectmanagement.Trace{}
	if cfg.Delivery.Backlogitems != "" {
		deliveryTraces = getDeliveryTraces(traces, cfg)

		// Delivery traces might contain backlog items which are NOT in the all traces array,
		// as new backlog items might not have tests yet, but the backlog item is already in the
		// delivery json for monitoring
		// In order to show missing tests for a backlog item also in the ALL report (as they already appear in
		// the delivery specific report), we must had those misisng backlog items to the traces array
		// See issue #9
		for _, dtrace := range deliveryTraces {
			var found = false
			for _, trace := range traces {
				if trace.BacklogItem == dtrace.BacklogItem {
					found = true
					continue
				}
			}
			if !found {
				traces = append(traces, dtrace)
			}
		}
	}

	// Update traceability repository (if given in config)
	if cfg.TraceabilityRepo.Git.Repository != "" {
		ghClient := projectmanagement.CreateGitHubClient(cfg)
		projectmanagement.UpdateTraceabilityRepository(traces, deliveryTraces, ghClient)

		if cfg.Github.CreateLinksInBacklogItems {
			// Create Links in GitHub Backlog items
			projectmanagement.CreateLinkInGHBackLogItem(ghClient, traces)
		}

		if cfg.Jira.CreateLinksInBacklogItems {
			// Create Links in Jira Backlog items
			projectmanagement.CreateLinkInJiraBackLogItem(cfg, traces)
		}
	}

	reportingStartTime := time.Now()
	// Create HTML and JSON containing ALL traces
	projectmanagement.CreateHTMLReport(cfg.OutputDir+string(os.PathSeparator)+reportBaseName+"_all.html", traces, cfg, true)
	projectmanagement.CreateJSONReport(cfg.OutputDir+string(os.PathSeparator)+reportBaseName+"_all.json", traces, cfg)
	if *argExportRequirementsMapping {
		projectmanagement.CreateRequirementsMappingReport(cfg.OutputDir+string(os.PathSeparator)+reportBaseName+"traceability-mapping.json", traces, cfg)
	}

	// Create HTML and JSON containing DELIVERY relevant traces
	if cfg.Delivery.Backlogitems != "" {
		projectmanagement.CreateHTMLReport(cfg.OutputDir+string(os.PathSeparator)+reportBaseName+"_"+cfg.Delivery.Version+".html", deliveryTraces, cfg, false)
		projectmanagement.CreateJSONReport(cfg.OutputDir+string(os.PathSeparator)+reportBaseName+"_"+cfg.Delivery.Version+".json", deliveryTraces, cfg)
		if *argExportRequirementsMapping {
			projectmanagement.CreateRequirementsMappingReport(cfg.OutputDir+string(os.PathSeparator)+reportBaseName+"_"+cfg.Delivery.Version+"-traceability-mapping.json", deliveryTraces, cfg)
		}
	}
	utils.TimeTrack(reportingStartTime, "Create HTML and JSON reports")

	glog.Info("Number of NOT successful tested requirements: ", len(traces)-projectmanagement.GetNumberOfSuccessfulTestedTraces(traces))

}
