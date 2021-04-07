package mapping

import (
	"bufio"
	"github.com/SAP/quality-continuous-traceability-monitor/utils"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// JSParser implements the mapping.Parser interface for JavaScript sourcecode
type JSParser struct {
}

// Parse JavaScript sourcecode to seek for traceability comments
func (jp JSParser) Parse(cfg utils.Config, sc utils.Sourcecode) []TestBacklog {

	var scName string
	if sc.Git.Organization != "" {
		scName = sc.Git.Organization + "/" + sc.Git.Repository
	} else {
		scName = sc.Local
	}

	defer utils.TimeTrack(time.Now(), "Parse JavaScript sourcecode ("+scName+")")

	var tb = []TestBacklog{}

	filepath.Walk(sc.Local, func(path string, fi os.FileInfo, err error) error {

		if fi.IsDir() || strings.Contains(path, "node_modules") {
			return nil
		}

		if filepath.Ext(path) == ".js" || filepath.Ext(path) == ".ts" {

			file, err := os.Open(path)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			reader := bufio.NewReader(file)

			var bli []BacklogItem

			var line,
				cn string
			var classIndentation = 0
			var mn []string
			for {
				line, err = reader.ReadString('\n')

				if err == io.EOF {
					break
				}

				// Does the line contain our marker with the backlog item?
				bi := reTraceMarker.FindAllString(line, -1)
				if len(bi) > 0 {
					bli = GetBacklogItem(line)
					continue
				}

				// Check whether this line contains the class name
				// Reminder: cn will always hold the last found class name
				// in case there a multiple classes in one file
				ce := strings.LastIndex(line, "describe('")
				if ce != -1 {
					// fmt.Println("Found class in ", file.Name())

					// Ensure we're really having the class definition string not a line containing a class cast
					// Check that the char before the word "class" is a blank
					endIndex := strings.Index(line[ce+10:], "',")
					if endIndex != -1 {

						curIndentation := len(line) - len(strings.TrimLeft(line, " "))
						if curIndentation > classIndentation {
							cn = cn + " " + line[ce+10:ce+10+endIndex]
						} else {
							cn = line[ce+10 : ce+10+endIndex]
						}
						classIndentation = curIndentation

						// We found a marker and no methods yet...this marker marks the whole class as relevant
						if bli != nil && len(mn) == 0 {
							// Create and append test backlog item (for complete class)
							t := &Test{getSourcecodeURL(cfg, sc, file), cn, ""}
							tbi := TestBacklog{Test: *t, BacklogItem: bli}
							tb = append(tb, tbi)

							// Point backlog item to nil to ensure we're not processing this class any further
							// Still be need to continue reading this file line by line as there might be another
							// class
							// bli = nil
						}
					}

					continue
				}

				// Check whether the line contains a test method
				if len(cn) > 0 && bli != nil { // We're inside a class and we've recently found a marker
					me := strings.Index(line, "it('") // Might also be an enum or something else inside a class
					if me != -1 {
						mne := strings.Index(line[me+4:], "',") // Start of method parameters is end of method name
						if mne != -1 {                          // Might be -1 in case of enums etc.
							methodName := line[me+4 : me+4+mne]
							methodName = strings.Trim(methodName, " ")

							//fmt.Println("method name is %v", method_name)

							// Create and append test backlog item (for this method)
							t := &Test{getSourcecodeURL(cfg, sc, file), cn, methodName}
							tbi := TestBacklog{Test: *t, BacklogItem: bli}
							tb = append(tb, tbi)

							mn = append(mn, methodName)
						}
					}
				}

			}

		}

		return nil

	})

	return tb

}
