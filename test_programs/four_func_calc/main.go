package main

import (
	"bufio"
	"fmt"
	"github.com/andygello555/gotils/files"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		sourceFileOrScript := os.Args[1]
		emptyString := ""
		var filename, s *string

		if files.IsFile(sourceFileOrScript) {
			// If the input is a file then we parse the file
			filename = &sourceFileOrScript
			s = &emptyString
		} else {
			// Otherwise, we parse the command line arg
			filename = &emptyString
			s = &sourceFileOrScript
		}

		if err, results := Eval(*filename, *s); err != nil {
			fmt.Println(fmt.Sprintf("Error occurred whilst executing \"%s\": %v", sourceFileOrScript, err))
			os.Exit(1)
		} else {
			fmt.Println("Results:", results)
		}
	} else {
		memory := make(Memory)
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("  Interactive mode   ")
		fmt.Println("---------------------")

		for {
			fmt.Print("> ")
			text, _ := reader.ReadString('\n')

			if text == "quit\n" {
				break
			}

			if err, results := Eval("", text, &memory); err != nil {
				fmt.Println(fmt.Sprintf("Error occurred whilst executing input: %v", err))
			} else {
				fmt.Println("Results:", results)
			}
		}
	}
	os.Exit(0)
}
