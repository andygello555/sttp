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

func TestParse(t *testing.T) {
	type Example struct {
		Script string
		Results []float64
		Errors []string
	}

	if files, err := ioutil.ReadDir(ExamplePath); err != nil {
		panic(err)
	} else {
		for _, file := range files {
			if strings.HasPrefix(file.Name(), ExamplePrefix) {
				fileBytes, _ := ioutil.ReadFile(filepath.Join(ExamplePath, file.Name()))
				err, p := parser.Parse("", string(fileBytes))
				//fmt.Println(p.Tokens)
				if err != nil {
					fmt.Println(err.Error())
				}
				fmt.Println(p.String())
			}
		}
	}
}
