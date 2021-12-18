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
	heap   *data.Heap
}

var examples []*example

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

		actual := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(p.String(0), " ", ""), "\t", ""), "\n", ""), ";", "")
		expected := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(e.script, " ", ""), "\t", ""), "\n", ""), ";", "")
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

func TestVM_Eval(t *testing.T) {
	skip := []int{0}
	skipPtr := 0

	// Start the echo chamber web server
	echoChamber := exec.Command(EchoChamberCmd, EchoChamberSource)
	if err := echoChamber.Start(); err != nil {
		t.Error(fmt.Errorf("could not start echo chamber: \"%s\"", err.Error()))
	}
	time.Sleep(150 * time.Millisecond)

	for testNo, e := range examples {
		if skipPtr == len(skip) || testNo != skip[skipPtr] {
			var stdout, stderr strings.Builder
			vm := New(nil, &stdout, &stderr, os.Stdout)
			err, result := vm.Eval(e.name, e.script)

			if testing.Verbose() {
				fmt.Println(testNo + 1, "vm Eval:", err, result)
			}

			// If the example's stdout field is not empty then we'll check if it matches the actual stdout
			if e.stdout != "" {
				if stdout.String() != e.stdout {
					if testing.Verbose() {
						fmt.Println()
						fmt.Println()
						fmt.Println(">>>>>>>", testNo+1, e.name)
						fmt.Println(e.stdout)
						fmt.Println("=========================== VS ===============================")
						fmt.Println(stdout.String())
						fmt.Println()
					}
					t.Errorf("example %d's stdout does not match the expected stdout", testNo + 1)
				}
			}

			// Same for stderr
			if e.stderr != "" {
				if stderr.String() != e.stderr {
					t.Errorf("example %d's stderr does not match the expected stderr", testNo+1)
				}
			}

			if err != nil && t != nil {
				t.Error(err.Error())
			}
		} else {
			skipPtr ++
		}
	}

	// Kill the echo chamber
	if err := echoChamber.Process.Kill(); err != nil {
		t.Error("failed to kill echo chamber")
	}
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
	echoChamber := exec.Command(EchoChamberCmd, EchoChamberSource)
	if err := echoChamber.Start(); err != nil {
		t.Error(fmt.Errorf("could not start echo chamber: \"%s\"", err.Error()))
	}
	time.Sleep(150 * time.Millisecond)

	for testNo, test := range []struct{
		items []*BatchItem
		expected BatchResults
	}{
		{
			items: []*BatchItem{
				{
					Method: eval.GET,
					Args:   []*data.Value{
						{
							Value: "http://127.0.0.1:3000/a",
							Type: data.String,
						},
					},
					Id:     0,
				},
				{
					Method: eval.GET,
					Args:   []*data.Value{
						{
							Value: "http://127.0.0.1:3000/b",
							Type: data.String,
						},
					},
					Id:     1,
				},
				{
					Method: eval.GET,
					Args:   []*data.Value{
						{
							Value: "http://127.0.0.1:3000/c",
							Type: data.String,
						},
					},
					Id:     2,
				},
				{
					Method: eval.GET,
					Args:   []*data.Value{
						{
							Value: "http://127.0.0.1:3000/d",
							Type: data.String,
						},
					},
					Id:     3,
				},
				{
					Method: eval.GET,
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
							"url":     "http://127.0.0.1:3000/d",
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

		if results.Len() != len(test.expected) {
			t.Errorf("test no. %d has %d results, expected %d results", testNo + 1, results.Len(), len(test.expected))
		} else {
			i := 0
			for results.Len() > 0 {
				r := heap.Pop(results).(parser.Result)
				e := test.expected.Pop().(parser.Result)
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
	if err := echoChamber.Process.Kill(); err != nil {
		t.Error("failed to kill echo chamber")
	}
}
