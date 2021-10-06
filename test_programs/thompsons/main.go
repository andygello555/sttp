package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		// Thompson's construction
		graph, start, _, err := Thompson(os.Args[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Save the Thompson's construction graph
		if err = graph.Visualise(start, "thompsons"); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Epsilon transitions:", graph.EpsilonTransitions)
		fmt.Println("States:", graph.StateCount)
		os.Exit(0)
	}
	fmt.Println("No regular expression given")
	os.Exit(1)
}
