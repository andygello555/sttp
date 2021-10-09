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

	changed := true
	for _, currentMergedStateSet := range *ms {
		for _, newMergedStateSet := range *new {
			// We perform a difference on the currentMergedStateSet and the newMergedStateSet and vice versa to check if
			// the sets are equal
			if currentMergedStateSet.Equal(newMergedStateSet) {
				changed = false
				break
			}
		}
		if !changed {
			break
		}
	}
	return changed
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
	StateCols       map[StateKeyString]int
	Language 		[]string
	AcceptingStates StateSetExistence
	MergedStates    MergedStates
	Verbose         bool
}

// InitTT constructs a TransitionTable from the given Graph instance.
func InitTT(graph *Graph, verbose bool) *TransitionTable {
	// Setup the basic metadata for the transition table
	tt := TransitionTable{}
	tt.Verbose = verbose
	tt.Language = graph.Language()
	tt.Rows = len(tt.Language)
	tt.Cols = len(graph.DFA) + 1
	tt.Table = make([][]StateKeyString, tt.Rows)
	tt.States = []StateKey{DeadState}
	for _, state := range graph.DFA.States() {
		tt.States = append(tt.States, state)
	}
	tt.AcceptingStates = make(StateSetExistence)
	tt.MergedStates = make(MergedStates)
	tt.StateCols = make(map[StateKeyString]int)

	// Then fill out the table itself
	for i := range tt.Table {
		tt.Table[i] = make([]StateKeyString, tt.Cols)
		// Set each value accordingly in relation to the DFAs adjacency list.
		for j := range tt.Table[i] {
			if j == 0 {
				// We insert the dead state at col 0 (the dead state)
				tt.Table[i][j] = DeadState
				tt.StateCols[DeadState] = 0
			} else {
				state := tt.States[j]
				tt.StateCols[StateKeyString(state.Key())] = j
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

// ColSimilar checks if two columns make the same transitions.
func (tt *TransitionTable) ColSimilar(col1, col2 int) bool {
	same := true
	for i := range tt.Language {
		if tt.Table[i][col1] != tt.Table[i][col2] {
			same = false
			break
		}
	}
	return same
}

// StateKeySimilar checks if the given two StateKeys make the same transitions.
func (tt *TransitionTable) StateKeySimilar(state1, state2 StateKey) bool {
	return tt.ColSimilar(tt.StateCols[StateKeyString(state1.Key())], tt.StateCols[StateKeyString(state2.Key())])
}

// FindMerges sets the MergedStates field with all the possible merged states.
func (tt *TransitionTable) FindMerges() {
	// Create the new merged states plus some helpers
	mergedStateNum := len(tt.MergedStates)

	// We iterate over the current merged states.
	for mergedState, states := range tt.MergedStates {
		// We check if any of the merged states need to be split
		for checkingState := range *states {
			toLookAt := *states.Difference(&StateSetExistence{
				checkingState: true,
			})
			for len(toLookAt) > 0 {
				lookingState := toLookAt.Choose()
				toLookAt.Unmark(lookingState)
				if !tt.StateKeySimilar(checkingState, lookingState) {
					// Then we see if there are any "like-minded" states
					likeMinders := StateSetExistence{
						StateKeyString(lookingState.Key()): true,
					}
					containsDeadSet := lookingState == DeadState
					for likeMindedState := range *states {
						if likeMindedState != lookingState && tt.StateKeySimilar(lookingState, likeMindedState) {
							likeMinders.Mark(likeMindedState)
							toLookAt.Unmark(likeMindedState)
							if likeMindedState == DeadState {
								containsDeadSet = true
							}
						}
					}

					key := StateKeyString(strconv.Itoa(mergedStateNum))
					newSet := make(StateSetExistence)
					tt.MergedStates[key] = &newSet
					mergedStateNum += 1
					// If this new set contains the dead set then we move the other sets instead to keep the DeadSet in
					// the same set throughout Dead Set minimisation.
					toAddRemove := &likeMinders
					if containsDeadSet {
						toAddRemove = states.Difference(toAddRemove)
					}
					// Add the like-minded states to the new set...
					newSet.Mark(*toAddRemove.Keys()...)
					// ...And remove them from the old
					states.Unmark(*toAddRemove.Keys()...)
					if tt.Verbose {
						fmt.Println("Splitting", mergedState, "into:")
						fmt.Println(mergedState, ":", states.String())
						fmt.Println(key, ":", newSet.String())
					}
				}
			}
		}
	}
}

// Markup the transition table with the current merged states.
func (tt *TransitionTable) Markup(original *TransitionTable) {
	for i := range tt.Table {
		for j := range tt.Table[i] {
			mergedState, _ := tt.MergedStates.Find(original.Table[i][j])
			tt.Table[i][j] = mergedState
		}
	}
}

// Clone the referred to TransitionTable. This will produce a deep copy.
func (tt *TransitionTable) Clone() *TransitionTable {
	newTT := TransitionTable{}
	newTT.Verbose = tt.Verbose
	newTT.Rows = tt.Rows
	newTT.Cols = tt.Cols
	newTT.States = make([]StateKey, len(tt.States))
	copy(newTT.States, tt.States)
	newTT.Language = make([]string, len(tt.Language))
	copy(newTT.Language, tt.Language)
	newTT.AcceptingStates = make(StateSetExistence)
	newTT.AcceptingStates = *tt.AcceptingStates.Difference(new(StateSetExistence))
	newTT.MergedStates = make(MergedStates)
	newTT.StateCols = make(map[StateKeyString]int)

	// Copy over the state to column mapping
	for i, state := range tt.States {
		newTT.StateCols[StateKeyString(state.Key())] = i
	}

	// Copy over the merged states
	for mergedState, set := range tt.MergedStates {
		newTT.MergedStates[mergedState] = new(StateSetExistence)
		for state := range *set {
			newTT.MergedStates[mergedState].Mark(state)
		}
	}

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
	tt.MergedStates = make(MergedStates)
	// Make a set with all states including the DeadState
	allStates := make(StateSetExistence)
	for _, state := range tt.States {
		allStates.Mark(state)
	}

	// We create a clone of the transition table that we can markup
	minimalTable := tt.Clone()

	// Make a set with the: All\AcceptingStates
	deadStates := *allStates.Difference(&tt.AcceptingStates)
	minimalTable.MergedStates["0"] = &deadStates
	// We clone the AcceptingStates set by differencing it with an empty set
	minimalTable.MergedStates["1"] = tt.AcceptingStates.Difference(new(StateSetExistence))
	if tt.Verbose {
		fmt.Println("Starting transition table:")
		fmt.Println(minimalTable.String())
		fmt.Println("Starting merged states:")
		fmt.Println(minimalTable.MergedStates.String())
	}
	minimalTable.Markup(tt)

	// We keep merging states until the mergedStates haven't changed from the previous iteration
	for tt.MergedStates.Changed(&minimalTable.MergedStates) {
		if tt.Verbose {
			fmt.Println()
			fmt.Println("Marked up transition table:")
			fmt.Println(minimalTable.String())
		}
		tt.MergedStates = make(MergedStates)
		for mergedStates, set := range minimalTable.MergedStates {
			tt.MergedStates[mergedStates] = set.Difference(new(StateSetExistence))
		}
		minimalTable.FindMerges()
		minimalTable.Markup(tt)
		if tt.Verbose {
			fmt.Println("New merged states:")
			fmt.Println(minimalTable.MergedStates.String())
		}
	}
	*tt = *minimalTable
	if tt.Verbose {
		fmt.Println("NO CHANGE")
		fmt.Println("\nFinal transition table:")
		fmt.Println(minimalTable.String())
	}
}
