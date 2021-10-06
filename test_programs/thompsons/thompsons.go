package main

import (
	"fmt"
	"sort"
	"strconv"
)

const Epsilon = "e"

// StateKey represents the key used within AdjacencyLists.
// It wraps both State and StateSet to provide a string key usable in maps.
type StateKey interface {
	// Key generates a key from a State or StateSet
	Key() string
}

// State represents a state within the constructed Thompson's construction.
type State int

func (s State) Key() string {
	return strconv.Itoa(int(s))
}

// StateSet represents a collection of States. Used for Subset Construction.
type StateSet []StateKey

func (ss StateSet) Key() string {
	ints := make([]int, len(ss))
	for i := range ss {
		ints[i], _ = strconv.Atoi(ss[i].Key())
	}
	sort.Ints(ints)
	return fmt.Sprintf("%v", ints)
}

// StateKeyString is a string type which implements the StateKey interface. This is done as maps require hashable types
// for their keys. Thus, StateSetExistence uses StateKeyString keys.
type StateKeyString string

func (sks StateKeyString) Key() string {
	return string(sks)
}

// StateSetExistence represents a unique set of States.
type StateSetExistence map[StateKeyString]bool

func (sse *StateSetExistence) Keys() *StateSet {
	out := make(StateSet, len(*sse))
	i := 0
	for key := range *sse {
		out[i] = key
		i++
	}
	return &out
}

func (sse *StateSetExistence) String() string {
	out := "{"
	keys := sse.Keys()
	for i, state := range *keys {
		out += state.Key()
		if i != len(*keys) - 1 {
			out += ", "
		}
	}
	return out + "}"
}

// Mark a StateKey in the set. The given StateKey will be converted into a string and then cast into the StateKeyString
// type.
func (sse *StateSetExistence) Mark(state StateKey) {
	(*sse)[StateKeyString(state.Key())] = true
}

// Unmark a StateKey in the set.
func (sse *StateSetExistence) Unmark(state StateKey) {
	delete(*sse, StateKeyString(state.Key()))
}

func (sse *StateSetExistence) Check(state StateKey) bool {
	return (*sse)[StateKeyString(state.Key())]
}

// AdjacencyList represents a graph with collections of states as nodes as an adjacency list.
type AdjacencyList map[StateKey][]Edge

func (al *AdjacencyList) Get(state StateKey) []Edge {
	return (*al)[StateKeyString(state.Key())]
}

func (al *AdjacencyList) AddEdge(edge *Edge) {
	if _, ok := (*al)[StateKeyString(edge.Outgoing.Key())]; !ok {
		// If the outgoing state does not yet exist in the map then we will construct the array
		(*al)[StateKeyString(edge.Outgoing.Key())] = make([]Edge, 0)
	}
	(*al)[StateKeyString(edge.Outgoing.Key())] = append((*al)[StateKeyString(edge.Outgoing.Key())], *edge)
}

func (al *AdjacencyList) States() (states []StateKey) {
	states = make([]StateKey, len(*al))
	i := 0
	for key := range *al {
		states[i] = key
		i++
	}
	return states
}

// Edge represents an edge within the StateAdjacencyList.
type Edge struct {
	Read     string
	Outgoing StateKey
	Ingoing  StateKey
}

// Graph is a wrapper for all graph methods.
type Graph struct {
	// A counter for assigning numbers to States.
	StateCount         State
	// The starting state. Defaults to 0.
	Start              StateKey
	// A set of accepting states.
	AcceptingStates    StateSetExistence
	// The adjacency list used to store edges for our NFA (After Thompson's construction).
	NFA                AdjacencyList
	// The adjacency list used to store edges for our DFA (After Subset Construction).
	DFA                AdjacencyList
	// The number of epsilon transitions in the NFA. Displayed after Thompson's construction.
	EpsilonTransitions int
}

func InitGraph() *Graph	{
	return &Graph{
		0,
		State(0),
		make(StateSetExistence),
		make(AdjacencyList),
		make(AdjacencyList),
		0,
	}
}

func Thompson(regexString string) (graph *Graph, start StateKey, end StateKey, err error) {
	graph = InitGraph()
	err, regex := Parse(regexString)
	if err != nil {
		return graph, start, end, err
	}
	start, end = regex.Thompson(graph)
	// We set graph.Start to be the start state returned by Thompson's construction as well as adding the end state to
	// the set of accepting states.
	graph.Start = start
	graph.AcceptingStates.Mark(end)
	//b, err := json.MarshalIndent(graph, "", "  ")
	//if err != nil {
	//	fmt.Println("error:", err)
	//}
	//fmt.Println("\nAdjacency Map:")
	//fmt.Println(string(b))
	//fmt.Println("Start, end =", start, end)
	return graph, start, end, nil
}

// AddEdge adds an edge from the outgoing StateKey into the ingoing StateKey, reading the given input.
func (g *Graph) AddEdge(outgoing StateKey, ingoing StateKey, read string, dfa bool) {
	if read == Epsilon {
		g.EpsilonTransitions += 1
	}

	adjacencyList := &g.NFA
	if dfa {
		adjacencyList = &g.DFA
	}

	edge := Edge{
		Read:     read,
		Outgoing: StateKeyString(outgoing.Key()),
		Ingoing:  StateKeyString(ingoing.Key()),
	}
	fmt.Println("Making edge from", outgoing.Key(), "to", ingoing.Key(), "which reads in:", read)
	adjacencyList.AddEdge(&edge)
}

func (g *Graph) Union(start1 State, end1 State, start2 State, end2 State) (start State, end State) {
	// We create two new states
	start = g.StateCount
	end = g.StateCount + 1
	g.StateCount += 2

	// Connect them in the Thompson's construction union format
	g.AddEdge(start, start1, Epsilon, false)
	g.AddEdge(start, start2, Epsilon, false)
	g.AddEdge(end1, end, Epsilon, false)
	g.AddEdge(end2, end, Epsilon, false)
	return start, end
}

func (g *Graph) Concatenation(start1 State, end1 State, start2 State, end2 State) (start State, end State) {
	g.AddEdge(end1, start2, Epsilon, false)
	return start1, end2
}

func (g *Graph) Closure(start1 State, end1 State) (start State, end State) {
	start = g.StateCount
	end = g.StateCount + 1
	g.StateCount += 2

	// Connect them in the Thompson's construction closure format
	g.AddEdge(start, end, Epsilon, false)    // Skip to end if no input is matched
	g.AddEdge(start, start1, Epsilon, false) // Skip to start from start
	g.AddEdge(end1, start1, Epsilon, false)  // Loop back to start to match another input
	g.AddEdge(end1, end, Epsilon, false)     // Skip to end from end
	return start, end
}

func (b *Base) Thompson(graph *Graph) (start State, end State) {
	if b.Char != nil {
		// Construct the two states with the edge reading the Char
		start = graph.StateCount
		end = graph.StateCount + 1
		graph.StateCount += 2
		if *(b.Char) != Epsilon {
			// If b.Char is "e" we will treat it as Epsilon
			graph.AddEdge(start, end, *(b.Char), false)
		} else {
			graph.AddEdge(start, end, Epsilon, false)
		}
		return start, end
	}
	return b.Regex.Thompson(graph)
}

func (f *Factor) Thompson(graph *Graph) (start State, end State) {
	start, end = f.Base.Thompson(graph)
	if f.Closure {
		// If there is a Kleene Closure we will wrap the Factor
		start, end = graph.Closure(start, end)
	}
	return start, end
}

func (t *Term) Thompson(graph *Graph) (start State, end State) {
	start, end = t.Factors[0].Thompson(graph)
	if len(t.Factors) > 1 {
		// If we have more than one factor we concatenate each factor together
		for i := 0; i < len(t.Factors) - 1; i++ {
			tempStart, tempEnd := t.Factors[i + 1].Thompson(graph)
			start, end = graph.Concatenation(start, end, tempStart, tempEnd)
		}
	}
	return start, end
}

// Thompson construction for a Regex symbol.
func (r *Regex) Thompson(graph *Graph) (start State, end State) {
	start, end = r.Term.Thompson(graph)
	if r.Regex != nil {
		// If the Regex symbol exists then we construct the union between the Term and the Regex symbol.
		start2, end2 := r.Regex.Thompson(graph)
		start, end = graph.Union(start, end, start2, end2)
	}
	return start, end
}
