package mapping

import (
	"quality-continuous-traceability-monitor/utils"
	"os"
	"strconv"
	"strings"
	"testing"
)

// TestBacklog mappings (correct)
// Ensure the a least one space char before the word "class"
var testPythonCode = []testMapping{
	testMapping{input: `
	import unittest

	 # Trace(Jira:MYPROJECT-1)
	 class TestStringMethods(unittest.TestCase):
	
		def test_upper(self):
			self.assertEqual('foo'.upper(), 'FOO')
	
		# Trace(GitHub:myorg/myRepo#1)
		def test_isupper(self):
			self.assertTrue('FOO'.isupper())
			self.assertFalse('Foo'.isupper())
	
		def test_split(self):
			s = 'hello world'
			self.assertEqual(s.split(), ['hello', 'world'])
			# check that s.split fails when the separator is not a string
			with self.assertRaises(TypeError):
	 		s.split(2)
	`,
		expectedResult: []TestBacklog{{
			Test:        Test{ClassName: "testFile.TestStringMethods", FileURL: "/tmp/test/testFile.py", Method: "test_upper"},
			BacklogItem: []BacklogItem{BacklogItem{ID: "MYPROJECT-1", Source: Jira}}},
			{Test: Test{ClassName: "testFile.TestStringMethods", FileURL: "/tmp/test/testFile.py", Method: "test_isupper"},
				BacklogItem: []BacklogItem{BacklogItem{ID: "myorg/myRepo#1", Source: Github}}},
			{Test: Test{ClassName: "testFile.TestStringMethods", FileURL: "/tmp/test/testFile.py", Method: "test_isupper"},
				BacklogItem: []BacklogItem{BacklogItem{ID: "MYPROJECT-1", Source: Jira}}},
			{Test: Test{ClassName: "testFile.TestStringMethods", FileURL: "/tmp/test/testFile.py", Method: "test_split"},
				BacklogItem: []BacklogItem{BacklogItem{ID: "MYPROJECT-1", Source: Jira}}},
		}},
	testMapping{input: `
	import unittest

	 # Trace(Jira:MYPROJECT-1, GitHub:myOrg/myRepo#2)
	 class TestStringMethods(unittest.TestCase):
	
		def test_upper(self):
			self.assertEqual('foo'.upper(), 'FOO')
	
		def test_isupper(self):
			self.assertTrue('FOO'.isupper())
			self.assertFalse('Foo'.isupper())
	`,
		expectedResult: []TestBacklog{{Test: Test{ClassName: "testFile.TestStringMethods", FileURL: "/tmp/test/testFile.py", Method: "test_upper"},
			BacklogItem: []BacklogItem{
				BacklogItem{ID: "MYPROJECT-1", Source: Jira},
				BacklogItem{ID: "myOrg/myRepo#2", Source: Github},
			}},
			{Test: Test{ClassName: "testFile.TestStringMethods", FileURL: "/tmp/test/testFile.py", Method: "test_isupper"},
				BacklogItem: []BacklogItem{
					BacklogItem{ID: "MYPROJECT-1", Source: Jira},
					BacklogItem{ID: "myOrg/myRepo#2", Source: Github},
				}},
		}}}

func TestPythonParsing(t *testing.T) {

	cfg := new(utils.Config)
	cfg.Mapping.Local = "NonPersistedMappingFileForTesting"
	cfg.Github.BaseURL = "https://github.com"

	var sc = utils.Sourcecode{Git: utils.Git{Branch: "master", Organization: "testOrg", Repository: "testRepo"}, Language: "python", Local: "/tmp/test/"}
	var file = os.NewFile(0, "/tmp/test/testFile.py")

	for i, mapping := range testPythonCode {
		tb := parsePython(strings.NewReader(mapping.input), *cfg, sc, file)
		if !compareTestBacklog(tb, mapping.expectedResult) {
			t.Error("Comparism of Python Code (No. " + strconv.Itoa(i) + "): \n" + mapping.input + "\n with expected result failed.")
		}
	}

}
