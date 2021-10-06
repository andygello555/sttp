package main

const EPSILON = "e"

// State represents a state within the constructed Thompson's construction.
type State int

// AdjacencyList represents a graph as an adjacency list.
type AdjacencyList map[State][]Edge

// Edge represents an edge within the AdjacencyList.
type Edge struct {
	Read     string
	Outgoing State
	Ingoing  State
}

// Graph is a wrapper for all graph methods.
type Graph struct {
	// A counter for assigning numbers to States.
	StateCount         State
	// The starting state. Defaults to 0.
	Start              State
	// A set of accepting states.
	AcceptingStates    map[State]bool
	// The adjacency list used to store edges.
	Graph              AdjacencyList
	EpsilonTransitions int
}

func InitGraph() *Graph	{
	return &Graph{
		0,
		0,
		make(map[State]bool),
		make(AdjacencyList),
		0,
	}
}

func Thompson(regexString string) (graph *Graph, start State, end State, err error) {
	graph = InitGraph()
	err, regex := Parse(regexString)
	if err != nil {
		return graph, start, end, err
	}
	start, end = regex.Thompson(graph)
	// We set graph.Start to be the start state returned by Thompson's construction as well as adding the end state to
	// the set of accepting states.
	graph.Start = start
	graph.AcceptingStates[end] = true
	//b, err := json.MarshalIndent(graph, "", "  ")
	//if err != nil {
	//	fmt.Println("error:", err)
	//}
	//fmt.Println("\nAdjacency Map:")
	//fmt.Println(string(b))
	//fmt.Println("Start, end =", start, end)
	return graph, start, end, nil
}

// AddEdge adds an edge from the outgoing State into the ingoing State.
func (g *Graph) AddEdge(outgoing State, ingoing State, read string) {
	if _, ok := g.Graph[outgoing]; !ok {
		// If the outgoing state does not yet exist in the map then we will construct the array
		g.Graph[outgoing] = make([]Edge, 0)
	}
	if read == EPSILON {
		g.EpsilonTransitions += 1
	}
	g.Graph[outgoing] = append(g.Graph[outgoing], Edge{
		Read:     read,
		Outgoing: outgoing,
		Ingoing:  ingoing,
	})
}

func (g *Graph) Union(start1 State, end1 State, start2 State, end2 State) (start State, end State) {
	// We create two new states
	start = g.StateCount
	end = g.StateCount + 1
	g.StateCount += 2

	// Connect them in the Thompson's construction union format
	g.AddEdge(start, start1, EPSILON)
	g.AddEdge(start, start2, EPSILON)
	g.AddEdge(end1, end, EPSILON)
	g.AddEdge(end2, end, EPSILON)
	return start, end
}

func (g *Graph) Concatenation(start1 State, end1 State, start2 State, end2 State) (start State, end State) {
	g.AddEdge(end1, start2, EPSILON)
	return start1, end2
}

func (g *Graph) Closure(start1 State, end1 State) (start State, end State) {
	start = g.StateCount
	end = g.StateCount + 1
	g.StateCount += 2

	// Connect them in the Thompson's construction closure format
	g.AddEdge(start, end, EPSILON)  // Skip to end if no input is matched
	g.AddEdge(start, start1, EPSILON)  // Skip to start from start
	g.AddEdge(end1, start1, EPSILON)  // Loop back to start to match another input
	g.AddEdge(end1, end, EPSILON)  // Skip to end from end
	return start, end
}

func (b *Base) Thompson(graph *Graph) (start State, end State) {
	if b.Char != nil {
		// Construct the two states with the edge reading the Char
		start = graph.StateCount
		end = graph.StateCount + 1
		graph.StateCount += 2
		if *(b.Char) != EPSILON {
			// If b.Char is "e" we will treat it as EPSILON
			graph.AddEdge(start, end, *(b.Char))
		} else {
			graph.AddEdge(start, end, EPSILON)
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
