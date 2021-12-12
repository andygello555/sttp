package main

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/parser"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
)

// passFail is a lookup of the string to use for test outputs.
var passFail = map[bool]string {
	true: "PASS",
	false: "FAIL",
}

// TestResult is a single test result. There is a pointer to a TestStatement within an AST as well as whether the test 
// has passed.
type TestResult struct {
	Node   *parser.TestStatement
	Config *TestConfig
	Passed bool
}

// TestResults contains an array of pointers to TestResult, as well as maybe containing a list of inner suites. This 
// represents all the test results within a given sttp script. Also contains a pointer back to the parent TestSuite's 
// TestConfig.
type TestResults struct {
	Results []*TestResult
	Config  *TestConfig
}

// CheckPassed will check if all test results have their Passed field set.
func (t *TestResults) CheckPassed() bool {
	passed := true
	for _, result := range t.Results {
		if !result.Passed {
			passed = false
			break
		}
	}
	return passed
}

// AddTest adds a test onto the Results.
func (t *TestResults) AddTest(node *parser.TestStatement, passed bool) {
	t.Results = append(t.Results, &TestResult{
		Node:   node,
		Passed: passed,
	})
}

// GetConfig returns the reference to the TestConfig.
func (t *TestResults) GetConfig() parser.Config {
	return t.Config
}

func (t *TestResults) String(indent int) string {
	var b strings.Builder
	tabs := strings.Repeat("\t", indent)

	// We first iterate over the tests in the current file
	if len(t.Results) > 0 {
		for _, test := range t.Results {
			b.WriteString(fmt.Sprintf("%s%s - \"%s\" (%s)\n", tabs, test.Node.Pos.String(), test.Node.String(0), passFail[test.Passed]))
		}
	} else {
		b.WriteString(fmt.Sprintf("%sNO TEST RESULTS (PASS)\n", tabs))
	}

	return b.String()
}

// TestSuite is a map of directory paths to TestResults pointers. This can be used to construct a recursive test suite.
// A TestSuite represents a directory within the test structure.
type TestSuite struct {
	// There are TestResults for each sttp script within the current test suite.
	Suite          map[string]*TestResults
	// There can also be zero or many nested TestSuites (directories) within the current test suite.
	InnerSuites    map[string]*TestSuite
	Config         *TestConfig
	NestLevel      int
	Path           string
}

// NewSuite will create a new TestSuite.
func NewSuite(path string, breakOnFailure bool, nestLevel int) *TestSuite {
	return &TestSuite{
		Suite:          make(map[string]*TestResults),
		InnerSuites:    make(map[string]*TestSuite),
		Config:         &TestConfig{
			BreakOnFailure: breakOnFailure,
		},
		NestLevel:      nestLevel,
		Path:           path,
	}
}

// CheckPass will recursively check if each contained script and inner test suite has passed (or not passed) all their 
// tests.
func (ts *TestSuite) CheckPass() bool {
	passed := true
	for _, results := range ts.Suite {
		if !results.CheckPassed() {
			passed = false
			break
		}
	}

	if passed {
		for _, suite := range ts.InnerSuites {
			if !suite.CheckPass() {
				passed = false
			}
		}
	}
	return passed
}

func (ts *TestSuite) String() string {
	var b strings.Builder
	tabs := strings.Repeat("\t", ts.NestLevel)
	prefix := "SUB"
	if ts.NestLevel == 0 {
		prefix = "PENTHOUSE"
	}

	suffix := ""
	if len(ts.Suite) == 0 {
		suffix = "NO SCRIPTS"
	}
	b.WriteString(fmt.Sprintf("%s%s SUITE: %s %s (%s)\n", tabs, prefix, ts.Path, suffix, passFail[ts.CheckPass()]))

	// First we iterate over all the sttp scripts within the directory
	for _, results := range ts.Suite {
		b.WriteString(results.String(ts.NestLevel + 1))
	}

	// Then for each nested subdirectory we append the test suite output.
	for _, suite := range ts.InnerSuites {
		b.WriteString(suite.String())
	}
	return b.String()
}

// Run will create a new VM for each test script in the current and any sub-directories and will also construct a new 
// TestSuite for any sub-directories. The results of the test suite will be output at the end of the procedure.
func (ts *TestSuite) Run() error {
	if files, err := ioutil.ReadDir(ts.Path); err != nil {
		return err
	} else {
		for _, file := range files {
			path := filepath.Join(ts.Path, file.Name())
			if file.IsDir() {
				// Create a new test suite
				newSuite := NewSuite(path, ts.Config.BreakOnFailure, ts.NestLevel + 1)
				if err = newSuite.Run(); err != nil {
					return err
				}
				// Merge the test suite into the InnerSuites
				ts.InnerSuites[path] = newSuite
			} else {
				if filepath.Ext(path) == ".sttp" {
					// Create an entry in the Suite for the sttp script
					ts.Suite[path] = &TestResults{
						Results: make([]*TestResult, 0),
						Config:  ts.Config,
					}

					// Create a new VM and run the script
					vm := New(ts.Suite[path])
					fileBytes, _ := ioutil.ReadFile(path)
					if err, _ = vm.Eval(path, string(fileBytes)); err != nil && ts.Config.BreakOnFailure {
						break
					}
				}
			}
		}
		return nil
	}
}

type TestConfig struct {
	BreakOnFailure bool
}

// Get uses reflection to get the given TestConfig field by name. Will return nil if there is no such field.
func (tc *TestConfig) Get(name string) interface{} {
	field := reflect.ValueOf(tc).Elem().FieldByName(name)
	if !field.IsValid() {
		return nil
	}
	return field.Interface()
}
