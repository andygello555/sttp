package main

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/eval"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/parser"
	"github.com/alecthomas/participle/v2/lexer"
	"io"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"
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

// Passed implementors must be able to check whether tests have passed. Implemented by TestResults, TestSuite, and 
// TestPath.
type Passed interface {
	parser.IndentString
	CheckPass() bool
	Run(stdout io.Writer, stderr io.Writer, debug io.Writer, mergedEnv *Env) error
}

// TestPath contains either a pointer to a TestResults instance, or a pointer to a TestSuite instance. This encapsulates
// the structures that can be contained within a TestSuite.
type TestPath struct {
	Path        string
	TestResults *TestResults
	TestSuite   *TestSuite
}

// GetPath will get the TestResults or the TestSuite, whichever is not nil.
func (tp *TestPath) GetPath() Passed {
	switch {
	case tp.TestResults != nil:
		return tp.TestResults
	case tp.TestSuite != nil:
		return tp.TestSuite
	default:
		panic(fmt.Errorf("TestPath cannot have both fields nil"))
	}
}

// CheckPass will check if this TestPath has passed.
func (tp *TestPath) CheckPass() bool {
	return tp.GetPath().CheckPass()
}

// String will call the String method of the TestResults or TestSuite instance (whatever is not nil).
func (tp *TestPath) String(indent int) string {
	return tp.GetPath().String(indent)
}

// Run will run the TestResults or TestSuite instance (whatever is not nil).
func (tp *TestPath) Run(stdout io.Writer, stderr io.Writer, debug io.Writer, mergedEnv *Env) error {
	return tp.GetPath().Run(stdout, stderr, debug, mergedEnv)
}

// TestPaths represents a sorted array structure where TestPath(s) are ordered by their Path field in ascending 
//lexicographical order. 
type TestPaths []*TestPath
func (tps TestPaths) Len() int { return len(tps) }
func (tps TestPaths) Less(i, j int) bool { return tps[i].Path < tps[j].Path }
func (tps TestPaths) Swap(i, j int) { tps[i], tps[j] = tps[j], tps[i] }

// TestResults contains an array of pointers to TestResult, as well as maybe containing a list of inner suites. This 
// represents all the test results within a given sttp script. Also contains a pointer back to the parent TestSuite's 
// TestConfig, as well as the Path of the script that needs to be run.
type TestResults struct {
	Path    string
	Results []*TestResult
	Config  *TestConfig
}

// Run will create and run a new VM for the script at the Path.
func (t *TestResults) Run(stdout io.Writer, stderr io.Writer, debug io.Writer, mergedEnv *Env) error {
	var err error
	vm := New(t, stdout, stderr, debug, mergedEnv)
	fileBytes, _ := ioutil.ReadFile(t.Path)
	err, _ = vm.Eval(t.Path, string(fileBytes))
	if err != nil {
		var pos lexer.Position
		failedTest := false
		switch err.(type) {
		case struct { errors.ProtoSttpError }:
			sttpErr := err.(struct { errors.ProtoSttpError })
			if !sttpErr.FromNullVM {
				pos = sttpErr.Pos
			}
		case errors.PurposefulError:
			// We filter out any FailedTest errors
			if err.(errors.PurposefulError) == errors.FailedTest {
				failedTest = true
			}
		default:
			pos = lexer.Position{
				Filename: t.Path,
			}
		}

		// We will not add any FailedTest errors
		if !failedTest {
			t.AddTest(&parser.TestStatement{
				Pos:        pos,
				Expression: nil,
			}, false)
		}
	}
	return err
}

// CheckPass will check if all test results have their Passed field set.
func (t *TestResults) CheckPass() bool {
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
			if test.Node.Expression != nil {
				b.WriteString(fmt.Sprintf(
					"%s%s - \"%s\" (%s)\n",
					tabs,
					test.Node.Pos.String(),
					test.Node.String(0),
					passFail[test.Passed],
				))
			} else if !test.Passed {
				b.WriteString(fmt.Sprintf("%s%s - error occurred (FAIL)\n", tabs, test.Node.Pos.String()))
			}
		}
	} else {
		b.WriteString(fmt.Sprintf("%sNO TEST RESULTS (PASS)\n", tabs))
	}

	return b.String()
}

// TestSuite is a map of directory paths to TestResults pointers. This can be used to construct a recursive test suite.
// A TestSuite represents a directory within the test structure.
type TestSuite struct {
	// There are TestResults for each sttp script within the current test suite, and TestSuite(s) for each directory 
	// within the current test suite. This is encapsulated within a TestPath instance.
	Paths     TestPaths
	Config    *TestConfig
	NestLevel int
	Path      string
}

// NewSuite will create a new TestSuite.
func NewSuite(path string, breakOnFailure bool, nestLevel int) *TestSuite {
	return &TestSuite{
		Paths:     make(TestPaths, 0),
		Config:    &TestConfig{
			BreakOnFailure: breakOnFailure,
		},
		NestLevel: nestLevel,
		Path:      path,
	}
}

// GetPaths will first check if the Paths are sorted, if not then they will be sorted. The sorted Paths will be 
// returned.
func (ts *TestSuite) GetPaths() *TestPaths {
	if !sort.IsSorted(ts.Paths) {
		sort.Sort(ts.Paths)
	}
	return &ts.Paths
}

// GetNoScripts will return the number of TestPath(s) within Paths that are TestResults.
func (ts *TestSuite) GetNoScripts() int {
	count := 0
	for _, path := range ts.Paths {
		if path.TestResults != nil {
			count ++
		}
	}
	return count
}

// CheckPass will recursively check if each contained script and inner test suite has passed (or not passed) all their 
// tests.
func (ts *TestSuite) CheckPass() bool {
	passed := true
	for _, path := range ts.Paths {
		if !path.CheckPass() {
			passed = false
			break
		}
	}
	return passed
}

// String will return an indented output representing the entire directory structure of the TestSuite. Should not be 
// called before Run has been called.
func (ts *TestSuite) String(indent int) string {
	var b strings.Builder
	tabs := strings.Repeat("\t", ts.NestLevel)
	prefix := "SUB"
	if ts.NestLevel == 0 {
		prefix = "PENTHOUSE"
	}

	suffix := ""
	if ts.GetNoScripts() == 0 {
		suffix = "NO SCRIPTS"
	}
	b.WriteString(fmt.Sprintf("%s%s SUITE: %s %s (%s)\n", tabs, prefix, ts.Path, suffix, passFail[ts.CheckPass()]))

	// We iterate over all the sttp scripts and nested subdirectories within the test suite.
	for _, path := range *ts.GetPaths() {
		// NOTE: We pass in (ts.NestLevel + 1) to TestSuite.String but this method will not use that indent level anyway
		b.WriteString(path.String(ts.NestLevel + 1))
	}
	return b.String()
}

// Run will create a new VM for each test script in the current and any subdirectories and will also construct a new 
// TestSuite for any subdirectories. The results of the test suite will be output at the end of the procedure. You can
// also specify the io.Writer for stdout, stderr, and debug, if these are nil then these will default to os.Stdout, 
// os.Stderr, and ioutil.Discard respectively. It also takes a mergedEnv which can either be nil, or an environment that
// has been passed down from a parent TestSuite.
func (ts *TestSuite) Run(stdout io.Writer, stderr io.Writer, debug io.Writer, mergedEnv *Env) error {
	if files, err := ioutil.ReadDir(ts.Path); err != nil {
		return err
	} else {
		// Create a temporary type to manage the paths we find when iterating over this directory.
		environments := make([]parser.Env, 0)

		// Gather all the paths for the files and directories within the directory
		for _, file := range files {
			path := filepath.Join(ts.Path, file.Name())
			if file.IsDir() {
				// Create a new test suite (don't run just yet)
				newSuite := NewSuite(path, ts.Config.BreakOnFailure, ts.NestLevel + 1)
				ts.Paths = append(ts.Paths, &TestPath{
					Path:      path,
					TestSuite: newSuite,
				})
			} else {
				switch filepath.Ext(path) {
				case ".sttp":
					// Create an entry in the Suite for the sttp script
					ts.Paths = append(ts.Paths, &TestPath{
						Path: path,
						TestResults: &TestResults{
							Results: make([]*TestResult, 0),
							Config:  ts.Config,
							Path:    path,
						},
					})
				case ".env":
					// For each environment file, we will parse it to an Env, then append it to an array of environments
					var env *Env
					if err, env = EnvFromFile(path); err != nil {
						return err
					}
					environments = append(environments, env)
				}
			}
		}

		// If mergedEnv is nil then we will use an empty environment, otherwise we will create a copy of mergedEnv.
		if mergedEnv == nil {
			mergedEnv = EmptyEnv()
		} else {
			mergedEnv = &Env{
				Paths: mergedEnv.Paths,
				Value: &data.Value{
					Value:    mergedEnv.Value.Value,
					Type:     mergedEnv.Value.Type,
					Global:   true,
					ReadOnly: true,
				},
			}
		}

		// If we have environments, we will sort them by their paths and merge them into a single environment
		if len(environments) > 0 {
			sort.Slice(environments, func(i, j int) bool {
				return environments[i].GetPaths()[0] < environments[j].GetPaths()[0]
			})

			// We use the environment stored within the Environment field to merge into
			if err = mergedEnv.MergeN(environments...); err != nil {
				return err
			}
		}

		// We first iterate over all the scripts and the directories in a lexicographical fashion and run each TestPath
		for _, path := range *ts.GetPaths() {
			if err = path.Run(stdout, stderr, debug, mergedEnv); err != nil && ts.Config.BreakOnFailure {
				break
			}
		}
		return err
	}
}

// TestConfig is passed to a TestSuite to describe which features are enabled within the TestSuite.
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

// defaultTestConfig should be used as the default config for TestSuites.
var defaultTestConfig = &TestConfig{
	BreakOnFailure: false,
}

// Env represents an environment that can be passed to a VM, and merged with another Env.
type Env struct {
	Paths []string
	Value *data.Value
}

// EmptyEnv returns an empty Env. This environment will have a Value of an empty data.Object, and a Paths that is an 
// empty array of strings.
func EmptyEnv() *Env {
	return &Env{
		Paths:  make([]string, 0),
		Value: &data.Value{
			Value:    map[string]interface{}{},
			Type:     data.Object,
			Global:   true,
			ReadOnly: true,
		},
	}
}

// EnvFromFile returns an Env from the given JSON formatted string within the file of the given path.
func EnvFromFile(path string) (err error, env *Env) {
	// An anonymous function to quickly construct an error to return from this function
	errFunc := func() error {
		return fmt.Errorf("cannot construct environment: %s", err.Error())
	}

	// We read in the file at the filepath
	var file []byte
	if file, err = ioutil.ReadFile(path); err != nil {
		return errFunc(), nil
	}

	// Construct the sttp value from the read in file
	var val *data.Value
	if err, val = data.ConstructSymbol(string(file), false); err != nil {
		return errFunc(), nil
	}

	// Return a pointer to a newly constructed environment
	return nil, &Env{
		Paths: []string{path},
		Value: val,
	}
}

// String method so that we can easily return errors with environments within them. This will return a string in the 
// format:
//  ("empty"|Paths[0]:Paths[1]:...:Paths[n]) = e.Value.String()
func (e *Env) String() string {
	paths := "empty"
	if len(e.Paths) > 0 {
		paths = strings.Join(e.Paths, ":")
	}
	return fmt.Sprintf("%s = %s", paths, e.Value.String())
}

// Merge the given environment into the referred to environment. This will first merge the sttp Value, and then 
// concatenate the Paths from the given environment to the referred to environments. The merging of sttp data.Object is
// handled using eval.Compute, so both the LHS and RHS values will be cast to data.Object and any similar keys within 
// LHS will be overridden. 
func (e *Env) Merge(env parser.Env) (err error) {
	// We will cast the LHS to an Object so that we can merge it.
	if e.Value.Type != data.Object {
		if err, e.Value = eval.Cast(env.GetValue(), data.Object); err != nil {
			return fmt.Errorf(
				"environment: %s, is a %s and cannot be cast to an object",
				e.String(),
				e.Value.Type.String(),
			)
		}
	}

	// Merge the Value properties
	if err, e.Value = eval.Compute(eval.Add, e.Value, env.GetValue()); err != nil {
		return fmt.Errorf("error occurred whilst merging into %s: %s", e.String(), err.Error())
	}

	// Add the paths from the RHS environment into the paths in the LHS environments
	e.Paths = append(e.Paths, env.GetPaths()...)
	return nil
}

// MergeN will apply the Merge method to each given environment.
func (e *Env) MergeN(envs... parser.Env) (err error) {
	for _, env := range envs {
		if err = e.Merge(env); err != nil {
			return err
		}
	}
	return nil
}

func (e *Env) GetPaths() []string {
	return e.Paths
}

func (e *Env) GetValue() *data.Value {
	return e.Value
}
