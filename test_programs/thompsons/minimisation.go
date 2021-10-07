package main

import (
	"fmt"
	"github.com/bndr/gotabulate"
	"sort"
	"strconv"
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

// MergedStates represents a mapping of merged states.
type MergedStates map[StateKeyString]*StateSetExistence

func (ms *MergedStates) String() string {
	out := make([]string, len(*ms))
	i := 0
	for key, set := range *ms {
		out[i] = fmt.Sprintf("%s: %s", key.Key(), set.String())
		i++
	}
	return strings.Join(out, "\n")
}

// Changed tests whether the given MergedStates have changed from the current referred instance.
func (ms *MergedStates) Changed(new *MergedStates) bool {
	if len(*ms) != len(*new) {
		return true
	}

	same := true
	for _, currentMergedStateSet := range *ms {
		for _, newMergedStateSet := range *new {
			// We perform a difference on the currentMergedStateSet and the newMergedStateSet and vice versa to check if
			// the sets are equal
			if len(*currentMergedStateSet.Difference(newMergedStateSet)) != len(*newMergedStateSet.Difference(currentMergedStateSet)) {
				same = false
				break
			}
		}
		if !same {
			break
		}
	}
	return same
}

// Find will find the given state within the set of merged states.
// Returns the StateKeyString of the merged state as well as the StateSetExistence of the merged state.
func (ms *MergedStates) Find(find StateKeyString) (StateKeyString, *StateSetExistence) {
	for key, mergedState := range *ms {
		if mergedState.Check(find) {
			return key, mergedState
		}
	}
	return DeadState, nil
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

// FindMerges returns a MergedStates instance with all the possible merges.
func (tt *TransitionTable) FindMerges(currentMergedStates *MergedStates) *MergedStates {
	// Create the new merged states plus some helpers
	mergedStateNum := 0
	mergedStateKey := func() StateKeyString {
		return StateKeyString(strconv.Itoa(mergedStateNum))
	}
	mergedStates := make(MergedStates)

	// Iterate over the columns and find patterns
	for checkingCol := 0; checkingCol < tt.Cols; checkingCol++ {
		// We add the column as a merged state if we could not yet find a place for it elsewhere. If we could find a
		// place for it then we skip checking it.
		_, found := mergedStates.Find(StateKeyString(tt.States[checkingCol].Key()))
		if found == nil {
			checkingColState := make(StateSetExistence)
			mergedStates[mergedStateKey()] = &checkingColState
			fmt.Println("checkingCol", tt.States[checkingCol])
			checkingColState.Mark(tt.States[checkingCol])
			mergedStateNum += 1
			// Look for similar cols to the current column.
			for lookingCol := 0; lookingCol < tt.Cols; lookingCol++ {
				// Try to find the looking for column state within the merged state structure. Only continue if the column
				// state has not yet been added to the merged states structure.
				_, found := mergedStates.Find(StateKeyString(tt.States[lookingCol].Key()))
				fmt.Println("\tlookingCol", tt.States[lookingCol], found)
				if checkingCol != lookingCol && found == nil {
					// Check each transition on each row for the columns
					match := true
					for row := 0; row < tt.Rows; row++ {
						fmt.Println("\t\trow =", row, "checking:", tt.Table[row][checkingCol], "looking:", tt.Table[row][lookingCol])
						if tt.Table[row][checkingCol] != tt.Table[row][lookingCol] {
							match = false
						}
					}
					if match {
						fmt.Println("\tFound right place for", tt.States[lookingCol], "in", checkingColState.String())
						// If there was a match then we add the checkingCol state to the current merged state.
						checkingColState.Mark(tt.States[lookingCol])
					}
				}
			}
		}
	}
	return &mergedStates
}

// Markup the transition table with the new merged states.
func (tt *TransitionTable) Markup(original *TransitionTable, mergedStates *MergedStates) {
	for i := range tt.Table {
		for j := range tt.Table[i] {
			mergedState, _ := mergedStates.Find(original.Table[i][j])
			tt.Table[i][j] = mergedState
		}
	}
}

// Clone the referred to TransitionTable. This will produce a deep copy.
func (tt *TransitionTable) Clone() *TransitionTable {
	newTT := TransitionTable{}
	newTT.Rows = tt.Rows
	newTT.Cols = tt.Cols
	newTT.States = make([]StateKey, len(tt.States))
	copy(newTT.States, tt.States)
	newTT.Language = make([]string, len(tt.Language))
	copy(newTT.Language, tt.Language)
	newTT.AcceptingStates = make(StateSetExistence)
	newTT.AcceptingStates = *tt.AcceptingStates.Difference(new(StateSetExistence))

	// Finally, we create the table rows and copy each column into the table.
	newTT.Table = make([][]StateKeyString, newTT.Rows)
	for i := range tt.Table {
		newTT.Table[i] = make([]StateKeyString, newTT.Cols)
		copy(newTT.Table[i], tt.Table[i])
	}
	return &newTT
}

func (tt *TransitionTable) DeadStateMinimisation() {
	// Create a map of the merged states
	mergedStates := make(MergedStates)
	previousMergedStates := make(MergedStates)
	// Make a set with all states including the DeadState
	allStates := make(StateSetExistence)
	for _, state := range tt.States {
		allStates.Mark(state)
	}

	// We create a clone of the transition table that we can markup
	minimalTable := tt.Clone()
	fmt.Println(minimalTable)

	// Make a set with the: All\AcceptingStates
	deadStates := *allStates.Difference(&tt.AcceptingStates)
	mergedStates["0"] = &deadStates
	// We clone the AcceptingStates set by differencing it with an empty set
	mergedStates["1"] = tt.AcceptingStates.Difference(new(StateSetExistence))
	minimalTable.Markup(tt, &mergedStates)

	// We keep merging states until the mergedStates haven't changed from the previous iteration
	for previousMergedStates.Changed(&mergedStates) {
		fmt.Println()
		fmt.Println("Marked up transition table:")
		fmt.Println(minimalTable.String())
		previousMergedStates, mergedStates = mergedStates, *minimalTable.FindMerges(&mergedStates)
		minimalTable.Markup(tt, &mergedStates)
		fmt.Println("New merged states:")
		fmt.Println(mergedStates.String())
	}
	*tt = *minimalTable
	fmt.Println("\nFinal transition table:")
	fmt.Println(minimalTable.String())
}
