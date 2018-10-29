package mapping

import (
	"github.com/SAP/quality-continuous-traceability-monitor/utils"
	"os"
	"strconv"
	"strings"
	"testing"
)

// TestBacklog mappings (correct)
var testJavaCode = []testMapping{
	{input: `
		package com.sap.ctm.testing;

		import org.junit.*;

		// Trace(Jira:MYJIRAPROJECT-3)
		public class MyTest {

			@Test
			public void someTest() {
				// Do something meaningful
			}

		}
	`,
		expectedResult: []TestBacklog{{Test: Test{ClassName: "com.sap.ctm.testing.MyTest", FileURL: "testFile.java", Method: "someTest"},
			BacklogItem: []BacklogItem{{ID: "MYJIRAPROJECT-3", Source: Jira}}}}},
	{input: `
		package com.sap.ctm.testing;

		import org.junit.*;

		// Trace(Jira:MYJIRAPROJECT-3, )    This one should not fail the parser
		public class MyTest {

			@Test
			public void someTest() {
				// Do something meaningful
			}

		}
	`,
		expectedResult: []TestBacklog{{Test: Test{ClassName: "com.sap.ctm.testing.MyTest", FileURL: "testFile.java", Method: "someTest"},
			BacklogItem: []BacklogItem{{ID: "MYJIRAPROJECT-3", Source: Jira}}}}},
	{input: `
		package com.sap.ctm.testing;

		import org.junit.*;
		
		public class MyTest {

			// Trace(Jira:MYJIRAPROJECT-2)
			@Ignore @Test
			public void someTest() {
				// Do something meaningful
			}

		}
	`,
		expectedResult: []TestBacklog{{Test: Test{ClassName: "com.sap.ctm.testing.MyTest", FileURL: "testFile.java", Method: "someTest"},
			BacklogItem: []BacklogItem{{ID: "MYJIRAPROJECT-2", Source: Jira}}}}},
	{input: `
		package com.sap.ctm.testing;

		import org.junit.*;

		// Trace(Jira:MYJIRAPROJECT-1, GitHub:myOrg/myRepo#42)
		public class MyTest {

			@Test
			public void someTest() {
				// Do something meaningful
			}

		}
	`,
		expectedResult: []TestBacklog{{Test: Test{ClassName: "com.sap.ctm.testing.MyTest", FileURL: "testFile.java", Method: "someTest"},
			BacklogItem: []BacklogItem{
				{ID: "MYJIRAPROJECT-1", Source: Jira},
				{ID: "myOrg/myRepo#42", Source: Github}}}}},
	{input: `
			package com.sap.ctm.testing;
	
			import org.junit.*;
	
			// This is not a Trace parameter
			public class SomeTestClass {

				// Trace(Jira:MYJIRAPROJECT-12, GitHub:myOrg/myRepo#52, GitHub:myOrg/myRepo#62)
				@Test
				public void myTestMethod(String someParameter) {

				}
				
			}
		`,
		expectedResult: []TestBacklog{{Test: Test{ClassName: "com.sap.ctm.testing.SomeTestClass", FileURL: "testFile.java", Method: "myTestMethod"},
			BacklogItem: []BacklogItem{
				{ID: "MYJIRAPROJECT-12", Source: Jira},
				{ID: "myOrg/myRepo#52", Source: Github},
				{ID: "myOrg/myRepo#62", Source: Github}}}}},
	{input: `	
				import org.junit.Test;
		
				public class SomeTestClass
				{
					// Trace(Jira:CLOUDECOSYSTEM-6381)				
					@Test
					public void myTestMethod() {
					}					
				}
			`,
		expectedResult: []TestBacklog{{Test: Test{ClassName: "SomeTestClass", FileURL: "testFile.java", Method: "myTestMethod"},
			BacklogItem: []BacklogItem{
				{ID: "CLOUDECOSYSTEM-6381", Source: Jira}}}}},
	{input: `	
				package com.mycorp;

				import junit.framework.Test;
				import junit.framework.TestCase;
				import junit.framework.TestSuite;
				
				/**
				* Unit test for simple App.
				*/
				public class AppTest extends TestCase {
				
					/** 
					* Create the test case
					*
					* @param testName name of the test case
					*/
					public AppTest(String testName) {
						super(testName);
					}   
				
					/** 
					* @return the suite of tests being tested
					*/
					public static Test suite() {
						return new TestSuite(AppTest.class);
					}   
				
					/** 
					* Overwhelming complex test, which covers my
					* product requirement #1
					*/
					// Trace(GitHub:doergn/sourcecodeRepo#1)
					public void testApp() {
						assertTrue(true);
					}   
				
				}	
			`,
		expectedResult: []TestBacklog{{Test: Test{ClassName: "com.mycorp.AppTest", FileURL: "testFile.java", Method: "testApp"},
			BacklogItem: []BacklogItem{
				{ID: "doergn/sourcecodeRepo#1", Source: Github}}}}},
	{input: `
		package com.sap.ctm.testing;

		import org.junit.*;
		import com.more.imports.*;

		// This is not a Trace parameter
		public class SomeTestClass {

			// Trace(Jira:MYJIRAPROJECT-12, GitHub:myOrg/myRepo#52, GitHub:myOrg/myRepo#62)
			@Test
			public void myTestMethod(String someParameter) {

			}
			
			@Test
			public void notTracedTest() {

			}

			// Trace(Jira:MYJIRAPROJECT-100)
			@Test
			public boolean anotherTestMethod() {

			}

		}
		`,
		expectedResult: []TestBacklog{{
			Test: Test{ClassName: "com.sap.ctm.testing.SomeTestClass", FileURL: "testFile.java", Method: "myTestMethod"},
			BacklogItem: []BacklogItem{
				{ID: "MYJIRAPROJECT-12", Source: Jira},
				{ID: "myOrg/myRepo#52", Source: Github},
				{ID: "myOrg/myRepo#62", Source: Github}}},
			{Test: Test{ClassName: "com.sap.ctm.testing.SomeTestClass", FileURL: "testFile.java", Method: "anotherTestMethod"},
				BacklogItem: []BacklogItem{
					{ID: "MYJIRAPROJECT-100", Source: Jira}}}}}}

func TestJavaParsing(t *testing.T) {

	cfg := new(utils.Config)
	cfg.Mapping.Local = "NonPersistedMappingFileForTesting"
	cfg.Github.BaseURL = "https://github.com"

	var sc = utils.Sourcecode{Git: utils.Git{Branch: "master", Organization: "testOrg", Repository: "testRepo"}, Language: "java", Local: "./"}
	var file = os.NewFile(0, "testFile.java")

	for i, mapping := range testJavaCode {
		tb := parseJava(strings.NewReader(mapping.input), *cfg, sc, file)
		if !compareTestBacklog(tb, mapping.expectedResult) {
			t.Error("Comparism of Java Code (No. " + strconv.Itoa(i) + "): \n" + mapping.input + "\n with expected result failed.")
		}
	}

}
