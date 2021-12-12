package main

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/parser"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

const (
	ExamplePath          = "_examples"
	ExamplePrefix        = "example_"
	ExampleTestSuitePath = "_examples/test_suites"
)

var examples []string

func init() {
	examples = make([]string, 0)
	if files, err := ioutil.ReadDir(ExamplePath); err != nil {
		panic(err)
	} else {
		for _, file := range files {
			if !file.IsDir() {
				if strings.HasPrefix(file.Name(), ExamplePrefix) {
					fileBytes, _ := ioutil.ReadFile(filepath.Join(ExamplePath, file.Name()))
					examples = append(examples, string(fileBytes))
				}
			}
		}
	}	
}

func TestParse(t *testing.T) {
	for testNo, example := range examples {
		err, p := parser.Parse("", example)
		if err != nil && t != nil {
			t.Error("error:", err.Error())
		}

		//fmt.Println(testNo, ">>>>>>>>")
		//fmt.Println(p.String(0))

		actual := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(p.String(0), " ", ""), "\t", ""), "\n", ""), ";", "")
		expected := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(example, " ", ""), "\t", ""), "\n", ""), ";", "")
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
	for testNo, example := range examples {
		if skipPtr == len(skip) || testNo != skip[skipPtr] {
			vm := New(nil)
			err, result := vm.Eval("", example)

			if testing.Verbose() {
				fmt.Println("vm Eval:", err, result)
			}

			if err != nil && t != nil {
				t.Error(err.Error())
			}
		} else {
			skipPtr ++
		}
	}
}

func BenchmarkVM_Eval(b *testing.B) {
	for n := 0; n < b.N; n++ {
		TestVM_Eval(nil)
	}
}

func TestTestSuite_Run(t *testing.T) {
	if files, err := ioutil.ReadDir(ExampleTestSuitePath); err != nil {
		panic(err)
	} else {
		for _, file := range files {
			if file.IsDir() && strings.HasPrefix(file.Name(), ExamplePrefix) {
				suite := NewSuite(filepath.Join(ExampleTestSuitePath, file.Name()), true, 0)
				if err = suite.Run(); err != nil {
					panic(err)
				}
				fmt.Println(suite.String())
			}
		}
	}
}
