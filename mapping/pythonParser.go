package mapping

import (
	"bufio"
	"quality-continuous-traceability-monitor/utils"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PythonParser implements the mapping.Parser interface for Python sourcecode
type PythonParser struct {
}

// Parse python sourcecode to seek for traceability comments
func (jp PythonParser) Parse(cfg utils.Config, sc utils.Sourcecode) []TestBacklog {

	var scName string
	if sc.Git.Organization != "" {
		scName = sc.Git.Organization + "/" + sc.Git.Repository
	} else {
		scName = sc.Local
	}

	defer utils.TimeTrack(time.Now(), "Parse python sourcecode ("+scName+")")

	var tb = []TestBacklog{}

	filepath.Walk(sc.Local, func(path string, fi os.FileInfo, err error) error {

		if fi.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".py" {

			file, err := os.Open(path)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			tb = append(tb, parsePython(file, cfg, sc, file)...)

		}

		return nil

	})

	return tb

}

func parsePython(coding io.Reader, cfg utils.Config, sc utils.Sourcecode, file *os.File) []TestBacklog {

	var tb = []TestBacklog{}
	var line,
		cn string
	var cBli []BacklogItem // Traceability annotation for class
	var mBli []BacklogItem // Traceability annotation for method
	var err error

	// Getting "package" name
	if sc.Local[len(sc.Local)-1] == '/' {
		sc.Local = sc.Local[:len(sc.Local)-1]
	}
	var pkgPath = file.Name()[len(sc.Local)+1:]
	var packageName = strings.Replace(strings.Replace(pkgPath, "/", ".", -1), ".py", "", -1)

	reader := bufio.NewReader(coding)
	for {
		line, err = reader.ReadString('\n')

		if err == io.EOF {
			break
		}

		// Empty line
		if line == "" || line == "\n" {
			continue
		}

		// Does the line contain our marker with the backlog item?
		bi := reTraceMarker.FindAllString(line, -1)

		if len(bi) > 0 {
			if cn != "" { // Traceability annotation for a class or a method
				mBli = GetBacklogItem(line) // We're inside a class...must belong to a test method
			} else {
				cBli = GetBacklogItem(line)
			}
			continue
		}

		// Check whether this line contains the class name
		// Reminder: cn will always hold the last found class name
		// in case there a multiple classes in one file
		ce := strings.LastIndex(line, "class ")
		if ce != -1 {
			// Ensure we're really having the class definition string not a line containing a class cast
			// Check that the char before the word "class" is a blank
			if ce > 0 {
				ls := line[ce-1 : ce]
				ls = strings.TrimLeft(ls, " ")
				if ls != "" {
					continue
				}
			}
			endIndex := strings.Index(line[ce+5:], "(")
			if endIndex != -1 {
				cn = line[ce+5 : ce+5+endIndex]
				// As cn could now still hold interface and parent class definitions we have
				// to slice it a bit more
				cn = strings.Trim(cn, " ")
				// fmt.Println("class name is ", cn)

				// Add package name to classname
				cn = packageName + "." + cn
			}

			continue
		}

		// Check whether the line contains a test method
		if len(cn) > 0 && (mBli != nil || cBli != nil) { // We're inside a class and we've recently found a marker
			me := strings.Index(line, "def ") // Might also be an enum or something else inside a class
			if me != -1 {
				mne := strings.Index(line[me+4:], "(") // Start of method parameters is end of method name
				if mne != -1 {                         // Might be -1 in case of enums etc.
					methodName := line[me+4 : me+4+mne]
					methodName = strings.Trim(methodName, " ")

					if !strings.HasPrefix(methodName, "test_") { // Not a test method
						continue
					}

					// Create and append test backlog item (for this method)
					t := &Test{getSourcecodeURL(cfg, sc, file), cn, methodName}
					var tbi TestBacklog
					if cBli != nil {
						tbi = TestBacklog{*t, cBli}
						tb = append(tb, tbi)
					}
					if mBli != nil {
						tbi = TestBacklog{*t, mBli}
						tb = append(tb, tbi)
					}

					// We handled this traceability relevant test method. Reset traceability method annotation
					mBli = nil
				}
			}
		}

	}

	return tb

}
