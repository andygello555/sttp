package main

import (
	"bufio"
	"fmt"
	"github.com/andygello555/gotils/files"
	"os"
)

func main() {
	parser := BuildParser(&Statement{})
	memory := make(Memory)

	if len(os.Args) > 1 {
		sourceFile := os.Args[1]
		if files.IsFile(sourceFile) {
			stmt := &Statement{}
			if err := parser.ParseString(sourceFile, "", stmt); err != nil {
				fmt.Println(fmt.Sprintf("Error occurred whilst executing \"%s\": %v", sourceFile, err))
				os.Exit(1)
			}
			result, _ := stmt.Eval(memory)
			fmt.Println(result)
		}
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("  Interactive mode   ")
		fmt.Println("---------------------")

		for {
			fmt.Print("> ")
			text, _ := reader.ReadString('\n')

			if text == "quit" {
				break
			}

			stmt := &Statement{}
			if err := parser.ParseString("", text, stmt); err != nil {
				fmt.Println(fmt.Sprintf("Error occurred whilst executing input: %v", err))
			}
			result, _ := stmt.Eval(memory)
			fmt.Println(result)
		}
	}
	os.Exit(0)
}
