package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const (
	ExamplePath = "_examples"
	ExamplePrefix = "example_"
)

func TestEval(t *testing.T) {
	type Example struct {
		Script string
		Results []float64
		Errors []string
	}

	if files, err := ioutil.ReadDir(ExamplePath); err != nil {
		panic(err)
	} else {
		//exampleTable := make([]Example, len(files))
		for _, file := range files {
			if strings.HasPrefix(file.Name(), ExamplePrefix) {
				fileBytes, err := ioutil.ReadFile(filepath.Join(ExamplePath, file.Name()))
				example := &Example{}
				if err = json.Unmarshal(fileBytes, example); err != nil {
					panic(err)
				}
				errors, results := Eval("", example.Script)
				if errors != nil && len(example.Errors) == 0 {
					t.Errorf("Example %s returned error: %v, when none were expected", file.Name(), errors)
				} else {
					for _, e := range example.Errors {
						if !strings.Contains(errors.Error(), e) {
							t.Errorf("Example %s returned error: %v, which does not contain the substring: %s", file.Name(), errors, e)
						}
					}
					continue
				}
				if !reflect.DeepEqual(results, example.Results) {
					t.Errorf("Example %s returns: %v, and not %v", file.Name(), results, example.Results)
				}
			}
		}
	}
}
