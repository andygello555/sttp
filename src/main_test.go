package main

import (
	"container/heap"
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/eval"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/parser"
	"github.com/andygello555/gotils/slices"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

const (
	ExamplePath          = "_examples"
	ExamplePrefix        = "example_"
	ExampleTestSuitePath = "_examples/test_suites"
	EchoChamberCmd       = "node"
	EchoChamberSource    = "_examples/echo_chamber/main.js"
)

type example struct {
	name   string
	script string
	stdout string
	stderr string
	tests  string
	err	   string
	heap   *data.Heap
}

var examples []*example

func startServer() *exec.Cmd {
	echoChamber := exec.Command(EchoChamberCmd, EchoChamberSource)
	echoChamber.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := echoChamber.Start(); err != nil {
		panic(fmt.Errorf("could not start echo chamber: \"%s\"", err.Error()))
	}
	time.Sleep(150 * time.Millisecond)
	return echoChamber
}

func killServer(echoChamber *exec.Cmd) {
	// Kill the echo chamber
	pgid, err := syscall.Getpgid(echoChamber.Process.Pid)
	if err == nil {
		if err = syscall.Kill(-pgid, 15); err != nil {
			panic("failed to kill echo chamber")
		}
	}
	_ = echoChamber.Wait()
}

func init() {
	examples = make([]*example, 0)
	if files, err := ioutil.ReadDir(ExamplePath); err != nil {
		panic(err)
	} else {
		for _, file := range files {
			if file.IsDir() && strings.HasPrefix(file.Name(), ExamplePrefix) {
				var exampleFiles []fs.FileInfo
				if exampleFiles, err = ioutil.ReadDir(filepath.Join(ExamplePath, file.Name())); err != nil {
					panic(err)
				} else {
					e := example{}
					e.name = file.Name()
					for _, exampleFile := range exampleFiles {
						if !exampleFile.IsDir() && strings.HasPrefix(exampleFile.Name(), ExamplePrefix) {
							fileBytes, _ := ioutil.ReadFile(filepath.Join(ExamplePath, file.Name(), exampleFile.Name()))
							switch filepath.Ext(exampleFile.Name()) {
							case ".sttp":
								e.script = string(fileBytes)
							case ".stdout":
								e.stdout = string(fileBytes)
							case ".stderr":
								e.stderr = string(fileBytes)
							case ".tests":
								e.tests = string(fileBytes)
							case ".err":
								e.err = string(fileBytes)
							}
						}
					}
					examples = append(examples, &e)
				}
			}
		}
	}
}

func TestParse(t *testing.T) {
	for testNo, e := range examples {
		err, p := parser.Parse(e.name, e.script)
		if err != nil && t != nil {
			t.Error(testNo + 1, "error:", err.Error())
		}

		//fmt.Println(testNo, ">>>>>>>>")
		//fmt.Println(p.String(0))
		expectedLines := strings.Split(e.script, "\n")
		for i, line := range expectedLines {
			expectedLines[i] = strings.TrimLeft(line, "\n\t ")
			if strings.HasPrefix(expectedLines[i], "//") {
				expectedLines[i] = ""
			}
		}

		actual := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(p.String(0)                  , " ", ""), "\t", ""), "\n", ""), ";", "")
		expected := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(                 strings.Join(expectedLines, ""), " ", ""), "\t", ""), ";", "")
		if actual != expected && t != nil {
			t.Errorf("%d: parsed output does not match input script", testNo+1)
			//fmt.Println("-------")
			//fmt.Println(actual)
			//fmt.Println(">>>>>>>>>>>")
			//fmt.Println(expected)
			//fmt.Println("-------")
		}
	}
}

func BenchmarkParse(b *testing.B) {
	for n := 0; n < b.N; n++ {
		TestParse(nil)
	}
}

func checkOut(key string, e *example, vm *VM, err error) (ok bool, actual string, expected string) {
	switch key {
	case "stdout":
		actual = vm.Stdout.(*strings.Builder).String()
		if e.stdout == "" { break }
		return e.stdout == actual, actual, e.stdout
	case "stderr":
		actual = vm.Stderr.(*strings.Builder).String()
		if e.stderr == "" { break }
		return e.stderr == actual, actual, e.stderr
	case "tests":
		if !vm.CheckTestResults() { break }
		actual = vm.TestResults.String(0)
		if e.tests == "" { break }
		return e.tests == actual, actual, e.tests
	case "err":
		if err == nil && e.err == "" { break }
		if err != nil {
			actual = err.Error()
		} else {
			actual = ""
		}
		return actual == e.err, actual, e.err
	default:
		return false, "", ""
	}
	return true, actual, expected
}

func TestVM_Eval(t *testing.T) {
	skip := []int{}
	skipPtr := 0

	// Start the echo chamber web server
	echoChamber := startServer()
	time.Sleep(150 * time.Millisecond)

	for testNo, e := range examples {
		if skipPtr == len(skip) || testNo != skip[skipPtr] {
			var stdout, stderr strings.Builder
			vm := New(nil, &stdout, &stderr, os.Stdout)
			err, result := vm.Eval(e.name, e.script)

			if testing.Verbose() {
				fmt.Println(testNo + 1, "vm Eval:", err, result)
			}

			// For each output that a test we can produce, we check if there is an expected output for it. If there is,
			// then we check if the actual output of that kind matches the expected output for that kind. If not, we 
			// error.
			for _, output := range []string{"stdout", "stderr", "tests", "err"} {
				ok, actual, expected := checkOut(output, e, vm, err)
				if !ok {
					if testing.Verbose() {
						fmt.Println()
						fmt.Println()
						fmt.Println(">>>>>>>", testNo+1, e.name)
						fmt.Println(expected)
						fmt.Println("=========================== VS ===============================")
						fmt.Println(actual)
						fmt.Println()
					}
					if t != nil {
						t.Errorf("example %d's %s does not match the expected %s", testNo+1, output, output)
					}
				}
			}
		} else {
			skipPtr ++
		}
	}

	// Kill the echo chamber
	killServer(echoChamber)
}

func BenchmarkVM_Eval(b *testing.B) {
	for n := 0; n < b.N; n++ {
		TestVM_Eval(nil)
	}
}

func TestTestSuite_Run(t *testing.T) {
	expected := [][]string{
		{
			`PENTHOUSE SUITE: _examples/test_suites/example_01  (PASS)`,
			`	_examples/test_suites/example_01/check_a.sttp:1:1 - "test 1 + 1 == 2" (PASS)`,
			`	_examples/test_suites/example_01/check_b.sttp:1:1 - "test 2 * 2 == 4" (PASS)`,
			`	_examples/test_suites/example_01/check_c.sttp:1:1 - "test 4 % 2 == 0" (PASS)`,
			`	SUB SUITE: _examples/test_suites/example_01/get_facebook  (PASS)`,
			`		_examples/test_suites/example_01/get_facebook/facebook.sttp:2:1 - "test "" + a == "true"" (PASS)`,
			`		_examples/test_suites/example_01/get_facebook/facebook.sttp:3:1 - "test a" (PASS)`,
			`	SUB SUITE: _examples/test_suites/example_01/get_google  (PASS)`,
			`		_examples/test_suites/example_01/get_google/google.sttp:2:1 - "test a" (PASS)`,
			`	SUB SUITE: _examples/test_suites/example_01/get_twitter  (PASS)`,
			`		_examples/test_suites/example_01/get_twitter/twitter.sttp:2:1 - "test a" (PASS)`,
			``,
		},
	}

	if files, err := ioutil.ReadDir(ExampleTestSuitePath); err != nil {
		panic(err)
	} else {
		for testNo, file := range files {
			if file.IsDir() && strings.HasPrefix(file.Name(), ExamplePrefix) {
				suite := NewSuite(filepath.Join(ExampleTestSuitePath, file.Name()), true, 0)
				if err = suite.Run(nil, nil, os.Stdout); err != nil {
					panic(err)
				}

				if testing.Verbose() {
					fmt.Println("TEST SUITE", testNo + 1, ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
					fmt.Println(suite.String())
					fmt.Println("===================================================")
				}

				ifSlice := func(lines []string) []interface{} {
					s := make([]interface{}, len(lines))
					for i, line := range lines {
						s[i] = line
					}
					return s
				}

				if !slices.SameElements(
					ifSlice(strings.Split(suite.String(), "\n")),
					ifSlice(expected[testNo]),
				) {
					t.Errorf("test no. %d suite's string output does not match expected output", testNo + 1)
				}
			}
		}
	}
}

func TestBatchSuite_Execute(t *testing.T) {
	// Start the echo chamber web server
	echoChamber := startServer()

	for testNo, test := range []struct{
		items []*BatchItem
		expected BatchResults
	}{
		{
			items: []*BatchItem{
				{
					Method: &parser.MethodCall{
						Method: eval.GET,
					},
					Args:   []*data.Value{
						{
							Value: "http://127.0.0.1:3000/a",
							Type: data.String,
						},
					},
					Id:     0,
				},
				{
					Method: &parser.MethodCall{
						Method: eval.GET,
					},
					Args:   []*data.Value{
						{
							Value: "http://127.0.0.1:3000/b",
							Type: data.String,
						},
					},
					Id:     1,
				},
				{
					Method: &parser.MethodCall{
						Method: eval.GET,
					},
					Args:   []*data.Value{
						{
							Value: "http://127.0.0.1:3000/c",
							Type: data.String,
						},
					},
					Id:     2,
				},
				{
					Method: &parser.MethodCall{
						Method: eval.GET,
					},
					Args:   []*data.Value{
						{
							Value: "http://127.0.0.1:3000/d",
							Type: data.String,
						},
					},
					Id:     3,
				},
				{
					Method: &parser.MethodCall{
						Method: eval.GET,
					},
					Args:   []*data.Value{
						{
							Value: "http://127.0.0.1:3000/e",
							Type: data.String,
						},
					},
					Id:     4,
				},
			},
			expected: BatchResults{
				&BatchResult{
					Id:  0,
					Err: nil,
					Value: &data.Value{
						Value: map[string]interface{}{
							"code": nil,
							"headers": map[string]interface{}{
								"accept-encoding": "gzip",
								"host":            "127.0.0.1:3000",
								"user-agent":      "go-resty/2.7.0 (https://github.com/go-resty/resty)",
							},
							"method": "GET",
							"query_params": map[string]interface{} {},
							"url":     "http://127.0.0.1:3000/a",
							"version": "1.1",
						},
						Type: data.Object,
					},
				},
				&BatchResult{
					Id:  1,
					Err: nil,
					Value: &data.Value{
						Value: map[string]interface{}{
							"code": nil,
							"headers": map[string]interface{}{
								"accept-encoding": "gzip",
								"host":            "127.0.0.1:3000",
								"user-agent":      "go-resty/2.7.0 (https://github.com/go-resty/resty)",
							},
							"method": "GET",
							"query_params": map[string]interface{} {},
							"url":     "http://127.0.0.1:3000/b",
							"version": "1.1",
						},
						Type: data.Object,
					},
				},
				&BatchResult{
					Id:  2,
					Err: nil,
					Value: &data.Value{
						Value: map[string]interface{}{
							"code": nil,
							"headers": map[string]interface{}{
								"accept-encoding": "gzip",
								"host":            "127.0.0.1:3000",
								"user-agent":      "go-resty/2.7.0 (https://github.com/go-resty/resty)",
							},
							"method": "GET",
							"query_params": map[string]interface{} {},
							"url":     "http://127.0.0.1:3000/c",
							"version": "1.1",
						},
						Type: data.Object,
					},
				},
				&BatchResult{
					Id:  3,
					Err: nil,
					Value: &data.Value{
						Value: map[string]interface{}{
							"code": nil,
							"headers": map[string]interface{}{
								"accept-encoding": "gzip",
								"host":            "127.0.0.1:3000",
								"user-agent":      "go-resty/2.7.0 (https://github.com/go-resty/resty)",
							},
							"method": "GET",
							"query_params": map[string]interface{} {},
							"url":     "http://127.0.0.1:3000/d",
							"version": "1.1",
						},
						Type: data.Object,
					},
				},
				&BatchResult{
					Id:  4,
					Err: nil,
					Value: &data.Value{
						Value: map[string]interface{}{
							"code": nil,
							"headers": map[string]interface{}{
								"accept-encoding": "gzip",
								"host":            "127.0.0.1:3000",
								"user-agent":      "go-resty/2.7.0 (https://github.com/go-resty/resty)",
							},
							"method": "GET",
							"query_params": map[string]interface{} {},
							"url":     "http://127.0.0.1:3000/e",
							"version": "1.1",
						},
						Type: data.Object,
					},
				},
			},
		},
	}{
		batch := Batch(nil)
		batch.Batch = test.items
		results := batch.Execute(-1)
		heap.Init(&test.expected)

		if results.Len() != len(test.expected) {
			t.Errorf("test no. %d has %d results, expected %d results", testNo + 1, results.Len(), len(test.expected))
		} else {
			i := 0
			for results.Len() > 0 {
				r := heap.Pop(results).(parser.Result)
				e := heap.Pop(&test.expected).(parser.Result)
				err, same := eval.EqualInterface(r.GetValue().Value.(map[string]interface{})["content"], e.GetValue().Value)
				if r.GetErr() != e.GetErr() || !same || err != nil {
					if testing.Verbose() {
						fmt.Println(testNo+1, "result:", i, ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
						if err != nil {
							fmt.Println("Unexpected error:", err)
						}
						fmt.Println("Actual:  ", r.GetValue().Value.(map[string]interface{})["content"], r.GetErr())
						fmt.Println("Expected:", e.GetValue().Value, e.GetErr())
						fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
					}
					t.Errorf("test no. %d, result no. %d does not match expected, or error occurred", testNo+1, i)
				}
				i++
			}
		}
	}

	// Kill the echo chamber
	killServer(echoChamber)
}

// Benchmarking batches can be done in the following way.
//  go test -run=XXX -bench="Benchmark(No)?Batch" -benchtime=5x -count=3
// This will run a batch-less sttp script...
//  for i = 0; i < %d; i = i + 1 do
//      result.result = $GET("http://127.0.0.1:3000/" + i);
//      result.results[i] = result.result.content;
//      $print(result.results[i]);
//  end;
//  oops_forgot_this = $GET("http://127.0.0.1:3000/" + %d);
//  $print(oops_forgot_this.content);
// And a batched sttp script.
//  batch this
//      for i = 0; i < %d; i = i + 1 do
//          result.result = $GET("http://127.0.0.1:3000/" + i);
//          result.results[i] = result.result.content;
//          $print(result.results[i]);
//      end;
//      oops_forgot_this = $GET("http://127.0.0.1:3000/" + %d);
//      $print(oops_forgot_this.content);
//  end;
// Each script will iterate up to the given number of iterations (10, 20, 30, 50, 100, 200). Each benchmark will run 5 
// times. Anything more than 5 will cause problems with the limit of the number of open sockets. If you are getting 
// i/o timeouts or other socket issues, then try increasing ulimit -n.
func batchBenchmarkSetup(i int) []interface{} {
	var null strings.Builder
	var expectedStdout strings.Builder
	for l := 0; l <= i; l++ {
		expectedStdout.WriteString(fmt.Sprintf("{\"code\":null,\"headers\":{\"accept-encoding\":\"gzip\",\"host\":\"127.0.0.1:3000\",\"user-agent\":\"go-resty/2.7.0 (https://github.com/go-resty/resty)\"},\"method\":\"GET\",\"query_params\":{},\"url\":\"http://127.0.0.1:3000/%d\",\"version\":\"1.1\"}\n", l))
	}
	return []interface{}{i, expectedStdout.String(), New(nil, &null, &null, ioutil.Discard)}
}

func benchmarkNoBatch(args []interface{}, b *testing.B) {
	i := args[0].(int)
	expected := args[1].(string)
	vm := args[2].(*VM)
	s := fmt.Sprintf(`for i = 0; i < %d; i = i + 1 do
        result.result = $GET("http://127.0.0.1:3000/" + i);
        result.results[i] = result.result.content;
        $print(result.results[i]);
    end;
    oops_forgot_this = $GET("http://127.0.0.1:3000/" + %d);
    $print(oops_forgot_this.content);`, i, i)
	var stdout, stderr strings.Builder
	vm.SetStdout(&stdout); vm.SetStderr(&stderr)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err, _ := vm.Eval("batch example", s)
		if err != nil {
			b.Errorf("error should not have occurred \"%v\"", err)
		}

		if stdout.String() != expected {
			fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
			fmt.Println(stdout.String())
			fmt.Println("================================")
			fmt.Println(expected)
			b.Error("stdout does not match to expected")
		}
		stdout.Reset(); stderr.Reset()
	}
}

func BenchmarkNoBatch10(b *testing.B) { s := startServer(); benchmarkNoBatch(batchBenchmarkSetup(10), b); killServer(s) }
func BenchmarkNoBatch20(b *testing.B) { s := startServer(); benchmarkNoBatch(batchBenchmarkSetup(20), b); killServer(s) }
func BenchmarkNoBatch30(b *testing.B) { s := startServer(); benchmarkNoBatch(batchBenchmarkSetup(30), b); killServer(s) }
func BenchmarkNoBatch50(b *testing.B) { s := startServer(); benchmarkNoBatch(batchBenchmarkSetup(50), b); killServer(s) }
func BenchmarkNoBatch100(b *testing.B) { s := startServer(); benchmarkNoBatch(batchBenchmarkSetup(100), b); killServer(s) }
func BenchmarkNoBatch200(b *testing.B) { s := startServer(); benchmarkNoBatch(batchBenchmarkSetup(200), b); killServer(s) }

func benchmarkBatch(args []interface{}, b *testing.B) {
	i := args[0].(int)
	expected := args[1].(string)
	vm := args[2].(*VM)
	s := fmt.Sprintf(`batch this
    for i = 0; i < %d; i = i + 1 do
        result.result = $GET("http://127.0.0.1:3000/" + i);
        result.results[i] = result.result.content;
        $print(result.results[i]);
    end;
    oops_forgot_this = $GET("http://127.0.0.1:3000/" + %d);
    $print(oops_forgot_this.content);
end;`, i, i)
	var stdout, stderr strings.Builder
	vm.SetStdout(&stdout); vm.SetStderr(&stderr)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err, _ := vm.Eval("batch example", s)
		if err != nil {
			b.Errorf("error should not have occurred \"%v\"", err)
		}

		if stdout.String() != expected {
			fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
			fmt.Println(stdout.String())
			fmt.Println("================================")
			fmt.Println(expected)
			b.Error("stdout does not match to expected")
		}
		stdout.Reset(); stderr.Reset()
	}
}

func BenchmarkBatch10(b *testing.B) { s := startServer(); benchmarkBatch(batchBenchmarkSetup(10), b); killServer(s) }
func BenchmarkBatch20(b *testing.B) { s := startServer(); benchmarkBatch(batchBenchmarkSetup(20), b); killServer(s) }
func BenchmarkBatch30(b *testing.B) { s := startServer(); benchmarkBatch(batchBenchmarkSetup(30), b); killServer(s) }
func BenchmarkBatch50(b *testing.B) { s := startServer(); benchmarkBatch(batchBenchmarkSetup(50), b); killServer(s) }
func BenchmarkBatch100(b *testing.B) { s := startServer(); benchmarkBatch(batchBenchmarkSetup(100), b); killServer(s) }
func BenchmarkBatch200(b *testing.B) { s := startServer(); benchmarkBatch(batchBenchmarkSetup(200), b); killServer(s) }
