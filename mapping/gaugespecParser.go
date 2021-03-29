package mapping

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/SAP/quality-continuous-traceability-monitor/testreport"
	"github.com/SAP/quality-continuous-traceability-monitor/utils"
	"github.com/golang/glog"
)

var (
	reHeadline *regexp.Regexp = regexp.MustCompile(`^(?P<level>#+)(?:\s*)(?P<title>.+)$`)
	reTags     *regexp.Regexp = regexp.MustCompile(`(?:^Trace:|,)\s*([^,\s]+)`)
)

type GaugeTestCaseMatcher struct{}

func (gtcm GaugeTestCaseMatcher) Matches(tb *TestBacklog, tc *testreport.TestCase) bool {
	if tb.Test.ClassName != tc.ClassName {
		return false
	}

	if tb.Test.Method == "" || tb.Test.Method == tc.MethodName {
		return true
	}

	if strings.HasPrefix(tc.MethodName, tb.Test.Method) {
		postfix := strings.TrimSpace(tc.MethodName[len(tb.Test.Method):len(tc.MethodName)])
		_, err := strconv.Atoi(postfix)

		return err == nil
	} else {
		return false
	}
}

/*
 * gaugeSpecHandler:
 */
type gaugeSpecHandler struct {
	cfg   utils.Config
	sc    utils.Sourcecode
	file  *os.File
	items *[]TestBacklog

	lastSeenSpec *string
}

func (gsh gaugeSpecHandler) Spec(specTitle string) {
	gsh.addNewItem(specTitle, "")

	*gsh.lastSeenSpec = specTitle
}

func (gsh gaugeSpecHandler) Scenario(scenarioTitle string) {
	gsh.addNewItem(*gsh.lastSeenSpec, scenarioTitle)
}

func (gsh gaugeSpecHandler) Requirements(requirements []string) {
	if len(*gsh.items) == 0 {
		return // ignore
	}

	bli := []BacklogItem{}

	for _, requirement := range requirements {
		bli = append(bli, GetBacklogItem(requirement)...)
	}

	(*gsh.items)[len(*gsh.items)-1].BacklogItem = bli
}

func (gsh gaugeSpecHandler) addNewItem(classname string, method string) {
	item := TestBacklog{Test: Test{getSourcecodeURL(gsh.cfg, gsh.sc, gsh.file), classname, method}, BacklogItem: []BacklogItem{}, TestCaseMatcher: &GaugeTestCaseMatcher{}}

	*gsh.items = append(*gsh.items, item)
}

func (gsh gaugeSpecHandler) GetTestBacklog() []TestBacklog {
	filteredItems := []TestBacklog{}

	for _, item := range *gsh.items {
		if len(item.BacklogItem) > 0 {
			filteredItems = append(filteredItems, item)
		}
	}

	return filteredItems
}

func newGaugeSpecHandler(cfg utils.Config, sc utils.Sourcecode, file *os.File) gaugeSpecHandler {
	lastSeenItem := ""
	return gaugeSpecHandler{cfg, sc, file, &[]TestBacklog{}, &lastSeenItem}
}

/*
 * GaugeSpecParser:
 */
type GaugeSpecParser struct {
	handler gaugeSpecHandler
}

func (gsp GaugeSpecParser) Parse(cfg utils.Config, sc utils.Sourcecode) []TestBacklog {
	testBacklog := &[]TestBacklog{}

	filepath.Walk(sc.Local, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() || strings.Contains(path, "node_modules") || filepath.Ext(path) != ".spec" {
			return nil
		}

		glog.Infof("Parsing %s\n", path)

		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		reader := bufio.NewReader(file)

		*testBacklog = append(*testBacklog, gsp.ParseContent(reader, cfg, sc, file)...)

		return nil
	})

	return *testBacklog
}

func (gsp GaugeSpecParser) ParseContent(spec io.Reader, cfg utils.Config, sc utils.Sourcecode, file *os.File) []TestBacklog {
	handler := newGaugeSpecHandler(cfg, sc, file)
	scanner := bufio.NewScanner(spec)

	for scanner.Scan() {
		line := scanner.Text()

		if gsp.isHeader(line) {
			title, level := gsp.parseHeader(line)

			switch level {
			case 1:
				handler.Spec(title)
				break
			case 2:
				handler.Scenario(title)
				break
			}
		} else if gsp.isRequirementsMapping(line) {
			handler.Requirements(gsp.parseRequirementsMapping(line))
		}
	}

	return handler.GetTestBacklog()
}

func (gsp GaugeSpecParser) isHeader(line string) bool {
	return strings.HasPrefix(line, "#")
}

func (gsp GaugeSpecParser) parseHeader(line string) (string, int) {
	level := 0
	title := ""

	matches := reHeadline.FindStringSubmatch(line)

	for i, group := range reHeadline.SubexpNames() {
		switch group {
		case "level":
			level = len(matches[i])
			break
		case "title":
			title = matches[i]
			break
		}
	}

	return title, level
}

func (gsp GaugeSpecParser) isRequirementsMapping(line string) bool {
	return strings.HasPrefix(line, "Trace:")
}

func (gsp GaugeSpecParser) parseRequirementsMapping(line string) []string {
	matches := reTags.FindAllStringSubmatch(line, -1)
	tags := []string{}

	for _, match := range matches {
		if len(match) > 1 {
			tags = append(tags, match[1])
		}
	}

	return tags
}
