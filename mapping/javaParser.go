package mapping

import (
	"bufio"
	"github.com/SAP/quality-continuous-traceability-monitor/utils"
	"github.com/golang/glog"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
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

		// Is this a test method annotation?
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

		// Simple hack to identify the closing of an (outer) class
		// Only works if the outerclass } is the first char in a line (will not work with "stange" coding formats)
		// This should help us using inner classes, where cn would be set and need to be extended by innerclass name
		classClose := strings.Index(line, "}")
		if classClose == 0 && len(cn) > 0 {
			cn = ""
			continue
		}

		// Check whether this line contains the class name
		// Reminder: cn will always hold the last found class name
		// in case there a multiple classes in one file
		ce := strings.LastIndex(line, "class ")
		if ce != -1 {
			// Ensure we're really having the class definition string not a line containing a class cast
			// Check that the char before the word "class" is a blank or the first line char
			if ce != 0 {
				ls := line[ce-1 : ce]
				ls = strings.TrimLeft(ls, " ")
				/*if ls != "" {
					continue
				}*/
			}
			tcn := line[ce+5:]
			// As t_cn could now still hold interface and parent class definitions we have
			// to slice it a bit more
			tcn = strings.TrimLeft(tcn, " ")
			cne := strings.Index(tcn, " ")
			if cne == -1 { // After the classname there was a space char. (could be omitted e.g. line end)
				cne = strings.Index(tcn, "\n")
			}
			if ce == -1 {
				glog.Fatalln("Couldn't find classname in: ", line)
				continue
			}
			tcn = tcn[:cne]

			// There could be a generic type at the end...trim this also
			cng := strings.LastIndex(tcn, "<")
			if cng != -1 {
				tcn = tcn[:cng]
			}

			// There could be a { right after the class name...trim this also
			cncr := strings.LastIndex(tcn, "{")
			if cncr != -1 {
				tcn = tcn[:cncr]
			}

			if len(cn) > 0 { // Should be an inner class
				// Maybe there is already an inner class in the class name. Cut it off, as a new inner classname will be attached
				// Only works with one inner class. Inner classes of inner classes are not supported
				// TODO: Make inner class detection more robust
				cninner := strings.LastIndex(cn, "$")
				if cninner != -1 {
					cn = cn[:cninner]
				}
				cn = cn + "$" + tcn
			} else { // Should be outer class
				// Add package name to classname
				if pn != "" {
					cn = pn + "." + tcn
				}
			}

			continue
		}

		// Check whether the line contains a test method
		// We're inside a class --> (len(cn) > 0)
		// and we've recently found a traceability annotation --> (m_bli != nil || c_bli != nil)
		// Testing on test annotation (tm) will be done later, as JUnit tests could also be indicated by method name starting with 'test...'
		if len(cn) > 0 && (mBli != nil || cBli != nil) {
			me := strings.LastIndex(line, "{") // Might also be an enum or something else inside a class
			if me != -1 {
				mne := strings.Index(line, "(") // Start of method parameters is end of method name
				if mne != -1 {                  // Might be -1 in case of enums etc.
					mnes := line[:mne] // Get line until end of method name
					mnes = strings.TrimLeft(mnes, " ")
					mns := strings.LastIndex(mnes, " ") // This must be where the method name starts
					m := mnes[mns+1:]

					// We didn't find a test annotation (@Test) yet. Check if method starts with test
					if tm == false && strings.HasPrefix(m, "test") {
						tm = true
					}

					if tm {

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

	}

	return tb

}
