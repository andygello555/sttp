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
	ExamplePath = "_examples"
	ExamplePrefix = "example_"
)

var examples []string

func init() {
	examples = make([]string, 0)
	if files, err := ioutil.ReadDir(ExamplePath); err != nil {
		panic(err)
	} else {
		for _, file := range files {
			if strings.HasPrefix(file.Name(), ExamplePrefix) {
				fileBytes, _ := ioutil.ReadFile(filepath.Join(ExamplePath, file.Name()))
				examples = append(examples, string(fileBytes))
			}
		}
	}	
}

func TestParse(t *testing.T) {
	for testNo, example := range examples {
		err, p := parser.Parse("", example)
		if err != nil {
			t.Error("error:", err.Error())
		}

		actual := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(p.String(0), " ", ""), "\t", ""), "\n", ""), ";", "")
		expected := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(example, " ", ""), "\t", ""), "\n", ""), ";", "")
		if actual != expected {
			t.Errorf("%d: parsed output does not match input script", testNo)
		}
	}
}

func TestVM_Eval(t *testing.T) {
	skip := []int{0}
	skipPtr := 0
	for testNo, example := range examples {
		if skipPtr == len(skip) || testNo != skip[skipPtr] {
			vm := New()
			err, result := vm.Eval("", example)
			fmt.Println(err, result)
			if err != nil {
				t.Error(err.Error())
			}
		} else {
			skipPtr ++
		}
	}
}
