package mapping

import (
	"bufio"
	"quality-continuous-traceability-monitor/utils"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/glog"
)

// JavaParser implements the mapping.Parser interface for Java sourcecode
// Known limitations: Multiple test classes in one Java file won't get processed right (not sure if that use case makes sense at all)
type JavaParser struct {
}

// Parse java sourcecode to seek for traceability comments
func (jp JavaParser) Parse(cfg utils.Config, sc utils.Sourcecode) []TestBacklog {

	var scName string
	if sc.Git.Organization != "" {
		scName = sc.Git.Organization + "/" + sc.Git.Repository
	} else {
		scName = sc.Local
	}

	defer utils.TimeTrack(time.Now(), "Parse java sourcecode ("+scName+")")

	var tb = []TestBacklog{}

	filepath.Walk(sc.Local, func(path string, fi os.FileInfo, err error) error {

		if fi.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".java" {

			file, err := os.Open(path)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			tb = append(tb, parseJava(file, cfg, sc, file)...)

		}

		return nil
	})

	return tb

}

func parseJava(coding io.Reader, cfg utils.Config, sc utils.Sourcecode, file *os.File) []TestBacklog {

	var tb = []TestBacklog{}
	var cBli []BacklogItem // Traceability annotation for class
	var mBli []BacklogItem // Traceability annotation for method
	var line,
		pn,
		cn string
	var err error
	var tm bool // Indicates we've found a @Test annotated method

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

		// Is this a test method marker?
		if strings.Contains(line, "@Test") {
			tm = true
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

		// Check whether this line contains the package name
		if len(pn) == 0 {
			pe := strings.LastIndex(line, "package ")
			if pe != -1 {
				pee := strings.LastIndex(line, ";")
				pn = line[pe+8 : pee]
				continue
			}
		}

		// Check whether this line contains the class name
		// Reminder: cn will always hold the last found class name
		// in case there a multiple classes in one file
		ce := strings.LastIndex(line, "class ")
		if ce != -1 {
			// Ensure we're really having the class definition string not a line containing a class cast
			// Check that the char before the word "class" is a blank
			ls := line[ce-1 : ce]
			ls = strings.TrimLeft(ls, " ")
			if ls != "" {
				continue
			}
			cn = line[ce+5 : len(line)]
			// As cn could now still hold interface and parent class definitions we have
			// to slice it a bit more
			cn = strings.TrimLeft(cn, " ")
			cne := strings.Index(cn, " ")
			if cne == -1 { // After the classname there was a space char. (could be omited e.g. line end)
				cne = strings.Index(cn, "\n")
			}
			if ce == -1 {
				glog.Fatalln("Couldn't find classname in: ", line)
				continue
			}
			cn = cn[:cne]

			// There could be a generic type at the end...trim this also
			cng := strings.LastIndex(cn, "<")
			if cng != -1 {
				cn = cn[:cng]
			}

			// Add package name to classname
			if pn != "" {
				cn = pn + "." + cn
			}

			continue
		}

		// Check whether the line contains a test method
		// We're inside a class --> (len(cn) > 0)
		// and we've recently found a traceability annotation --> (m_bli != nil || c_bli != nil)
		// and we found a @Test annotation --> (tm)
		if len(cn) > 0 && (mBli != nil || cBli != nil) && tm {
			me := strings.LastIndex(line, "{") // Might also be an enum or something else inside a class
			if me != -1 {
				mne := strings.Index(line, "(") // Start of method parameters is end of method name
				if mne != -1 {                  // Might be -1 in case of enums etc.
					mnes := line[:mne] // Get line until end of method name
					mnes = strings.TrimLeft(mnes, " ")
					mns := strings.LastIndex(mnes, " ") // This must be where the method name starts
					m := mnes[mns+1:]

					// Create and append test backlog item (for this method)
					t := &Test{getSourcecodeURL(cfg, sc, file), cn, m}
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

					// We handled this test method. Reset @Test annotation marker
					tm = false
				}
			}
		}

	}

	return tb

}
