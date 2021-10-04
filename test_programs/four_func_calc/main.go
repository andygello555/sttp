package main

import (
	"fmt"
	"github.com/andygello555/gotils/files"
	"os"
)

func main() {
	parser := BuildParser(&Statement{})
	if len(os.Args) > 1 {
		sourceFile := os.Args[1]
		if files.IsFile(sourceFile) {
			stmt := &Statement{}
			err := parser.ParseString(sourceFile, "", stmt)
			fmt.Println(fmt.Sprintf("Error occurred whilst executing \"%s\": %v", sourceFile, err))
		}
	}
	// Interactive mode if no source file was specified?
	os.Exit(1)
}
