package main

import (
	"fmt"
	"github.com/bndr/gotabulate"
	"sort"
	"strings"
)

const DeadState = StateKeyString("dead")

// Language finds the set of all input strings in the DFA of a graph.
func (g *Graph) Language() (language []string) {
	languageSet := make(map[string]bool)
	for _, edges := range g.DFA {
		for _, edge := range edges {
			languageSet[edge.Read] = true
		}
	}

	language = make([]string, len(languageSet))
	i := 0
	for input := range languageSet {
		language[i] = input
		i++
	}
	// We'll also sort the strings for neatness
	sort.Strings(language)
	return language
}

// CheckIfAccepting checks if the given StateKey in the DFA is an accepted state.
func (g *Graph) CheckIfAccepting(state StateKey) bool {
	acceptingState := false
	for stateKey := range g.AcceptingStates {
		if strings.Contains(state.Key(), stateKey.Key()) {
			acceptingState = true
			break
		}
	}
	return acceptingState
}

// TransitionTable represents the table of all moves from all states with the given language, including the dead state.
// Rows represent the language input strings, and Cols represent the possible states from a DFA.
type TransitionTable struct {
	Table    		[][]StateKeyString
	Rows     		int
	Cols     		int
	States   		[]StateKey
	Language 		[]string
	AcceptingStates StateSetExistence
}

// InitTT constructs a TransitionTable from the given Graph instance.
func InitTT(graph *Graph) *TransitionTable {
	// Setup the basic metadata for the transition table
	tt := TransitionTable{}
	tt.Language = graph.Language()
	tt.Rows = len(tt.Language)
	tt.Cols = len(graph.DFA) + 1
	tt.Table = make([][]StateKeyString, tt.Rows)
	tt.States = []StateKey{DeadState}
	for _, state := range graph.DFA.States() {
		tt.States = append(tt.States, state)
	}
	tt.AcceptingStates = make(StateSetExistence)

	// Then fill out the table itself
	for i := range tt.Table {
		tt.Table[i] = make([]StateKeyString, tt.Cols)
		// Set each value accordingly in relation to the DFAs adjacency list.
		for j := range tt.Table[i] {
			if j == 0 {
				// We insert the dead state at col 0 (the dead state)
				tt.Table[i][j] = DeadState
			} else {
				state := tt.States[j]
				// Then we find out if the state is an accepting state by comparing the key to the NFA
				if graph.CheckIfAccepting(state) {
					tt.AcceptingStates.Mark(state)
				}

				// Find out which inputs the state can accept
				possibleInputs := make(map[string]*Edge)
				for _, edge := range graph.DFA.Get(state) {
					possibleInputs[edge.Read] = &edge
				}

				// We set the transition to a dead state by default
				tt.Table[i][j] = DeadState
				for _, edge := range graph.DFA.Get(state) {
					if edge.Read == tt.Language[i] {
						tt.Table[i][j] = StateKeyString(edge.Ingoing.Key())
						break
					}
				}
			}
		}
	}
	return &tt
}

func (tt *TransitionTable) String() string {
	headers := make([]string, len(tt.States) + 1)
	headers[0] = "input"
	i := 1
	for _, header := range tt.States {
		headers[i] = header.Key()
		i++
	}

	// We will also prepend the input language
	duplicate := make([][]string, tt.Rows)
	for i := range tt.Table {
		duplicate[i] = make([]string, tt.Cols + 1)
		for j := 0; j < tt.Cols + 1; j++ {
			if j == 0 {
				duplicate[i][j] = tt.Language[i]
			} else {
				duplicate[i][j] = tt.Table[i][j - 1].Key()
			}
		}
	}

	t := gotabulate.Create(duplicate)
	t.SetMaxCellSize(16)
	t.SetWrapStrings(true)
	t.SetHeaders(headers)
	return fmt.Sprintf("%s\nAccepting states = %s", t.Render("grid"), tt.AcceptingStates.String())
}

func (tt *TransitionTable) DeadStateMinimisation() {

}