package main

import (
	"fmt"
	"github.com/andygello555/gotils/files"
	"io/ioutil"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		sourceFileOrScript := os.Args[1]
		var filename, s string

		if files.IsFile(sourceFileOrScript) && !files.IsDir(sourceFileOrScript) {
			// If the input is a file then we parse the file
			filename = sourceFileOrScript
			sByte, _ := ioutil.ReadFile(sourceFileOrScript)
			s = string(sByte)
		} else if !files.IsFile(sourceFileOrScript) && files.IsDir(sourceFileOrScript) {
			// If the input is a directory then we will assume it is a directory of tests, so we will construct a 
			// TestSuite and run it, then exit.
			suite := NewSuite(sourceFileOrScript, true, 0, nil)
			if err := suite.Run(os.Stdout, os.Stderr, nil); err != nil {
				fmt.Println(fmt.Sprintf("Error occurred whilst executing \"%s\": %v", sourceFileOrScript, err))
				os.Exit(1)
			}
			fmt.Println(suite.String(0))
			os.Exit(0)
		} else {
			// Otherwise, we parse the command line arg
			filename = "stdin"
			s = sourceFileOrScript
		}

		vm := New(nil, nil, nil, nil)
		if err, _ := vm.Eval(filename, s); err != nil {
			fmt.Println(fmt.Sprintf("Error occurred whilst executing \"%s\": %v", sourceFileOrScript, err))
			os.Exit(1)
		}
	}
	os.Exit(0)
}
