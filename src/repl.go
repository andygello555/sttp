package main

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/atomicgo/cursor"
	"github.com/pkg/term"
	"strconv"
	"strings"
)

// Returns either an ascii code, or (if input is an arrow) a Javascript key code. This is taken from: 
// https://github.com/paulrademacher/climenu/blob/master/getchar.go.
func getChar() (ascii int, keyCode int, err error) {
	t, _ := term.Open("/dev/tty")
	_ = term.RawMode(t)
	bytes := make([]byte, 3)

	var numRead int
	numRead, err = t.Read(bytes)
	if err != nil {
		return
	}
	if numRead == 3 && bytes[0] == 27 && bytes[1] == 91 {
		// Three-character control sequence, beginning with "ESC-[".

		// Since there are no ASCII codes for arrow keys, we use
		// Javascript key codes.
		if bytes[2] == 65 {
			// Up
			keyCode = 38
		} else if bytes[2] == 66 {
			// Down
			keyCode = 40
		} else if bytes[2] == 67 {
			// Right
			keyCode = 39
		} else if bytes[2] == 68 {
			// Left
			keyCode = 37
		}
	} else if numRead == 1 {
		ascii = int(bytes[0])
	} else {
		// Two characters read??
	}
	_ = t.Restore()
	_ = t.Close()
	return
}

func REPL() {
	fmt.Println("-------------")
	fmt.Println("| sttp REPL |")
	fmt.Println("-------------")

	// Start REPL mode
	vm := New(true, nil, nil, nil, nil)
	lineNo := 0

	// The queue of history items
	const historyLength = 3
	history := make([]string, historyLength)
	// The pointer to the position in history with the next free history space.
	newHistoryPointer := 0
	// The pointer to the currently searched history item. This gets reset each line entered to the most.
	searchHistoryPointer := 0
	// Add new history item
	addHistory := func(text string) {
		text = strings.TrimSuffix(text, "\n")
		// If the previous history is the same as the text to add, then we will skip this action
		if newHistoryPointer > 0 && history[newHistoryPointer - 1] == text {
			return
		}

		history[newHistoryPointer] = text

		// If we have run out of space in history we will rotate the entire history
		if newHistoryPointer == historyLength - 1 {
			copy(history, history[1:])
			history[historyLength - 1] = ""
		} else {
			newHistoryPointer ++
		}
	}
	previousHistory := func() string {
		if searchHistoryPointer > 0 {
			searchHistoryPointer --
		}
		return history[searchHistoryPointer]
	}
	futureHistory := func() string {
		if searchHistoryPointer < newHistoryPointer {
			searchHistoryPointer ++
		}
		return history[searchHistoryPointer]
	}

	for {
		// Reset the searchHistoryPointer
		searchHistoryPointer = newHistoryPointer
		fmt.Print("> ")
		cursorPos := 0

		// Scan until CR, LF, or EOF
		var err error
		text := ""
		var ascii int
		for ascii != 10 && ascii != 13 && ascii != 4 {
			var char rune
			var keyCode int
			ascii, keyCode, err = getChar()
			//fmt.Println("", ascii, cursorPos, len(text), searchHistoryPointer, newHistoryPointer)
			if err != nil {
				break
			}

			switch keyCode {
			case 37:
				// Left: move the cursor left
				if cursorPos > 0 {
					cursorPos --
				}
			case 38:
				// Up: scroll upwards through history
				text = previousHistory()
				cursorPos = len(text)
			case 39:
				// Right: move the cursor right
				if cursorPos < len(text) {
					cursorPos ++
				}
			case 40:
				// Down: scroll downwards through history
				text = futureHistory()
				cursorPos = len(text)
			default:
				switch ascii {
				case 127:
					// Backspace
					if cursorPos != 0 {
						if cursorPos == len(text) {
							text = text[:cursorPos - 1]
						} else {
							text = text[:cursorPos - 1] + text[cursorPos:]
						}
						cursorPos--
					}
				default:
					char = rune(ascii)
					if strconv.IsPrint(char) {
						text = text[:cursorPos] + string(char) + text[cursorPos:]
						cursorPos ++
					}
				}
			}
			cursor.ClearLine()
			cursor.StartOfLine()
			fmt.Printf("> %s", text)
			cursor.StartOfLine()
			cursor.Move(cursorPos + 2, 0)
		}
		//text, err := reader.ReadString('\n')

		if ascii == 4 {
			fmt.Println("\nCTRL-D pressed. Quitting...")
			break
		}

		// Continue if the line is blank
		if strings.TrimSpace(text) == "" {
			continue
		}

		addHistory(text)
		fmt.Println()

		// We set stdout and stderr to temporary buffers so that we can check if anything was written to them and 
		// display them in a nice way
		var stdout strings.Builder
		var stderr strings.Builder
		vm.SetStdout(&stdout)
		vm.SetStderr(&stderr)

		var results *data.Value
		if err, results = vm.Eval(fmt.Sprintf("REPL:%d", lineNo), text); err != nil {
			fmt.Println(fmt.Sprintf("Error occurred whilst executing input: %v", err))
		}

		if vm.GetStdout().(*strings.Builder).Len() != 0 {
			fmt.Println("--- STDOUT ---")
			fmt.Print(vm.GetStdout().(*strings.Builder).String())
			fmt.Println("--------------")
		}

		if vm.GetStderr().(*strings.Builder).Len() != 0 {
			fmt.Println("--- STDERR ---")
			fmt.Print(vm.GetStderr().(*strings.Builder).String())
			fmt.Println("--------------")
		}

		// Print results if they are not null
		if results != nil {
			fmt.Println("Results:", results)
		}
		// Print the call stack if an error has not occurred
		if err == nil {
			fmt.Printf("Current heap: %v\n", *vm.GetCallStack().Current().GetHeap())
		}
		lineNo ++
	}
}
