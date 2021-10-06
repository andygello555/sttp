package main

func (g *Graph) EClosure(stack... State) (closure map[State]bool) {
	closure = make(map[State]bool)
	// Keep popping from the stack until we have visited everything.
	for len(stack) > 0 {
		index := len(stack) - 1
		t := stack[index]
		stack = stack[:index]
		for _, edge := range g.Graph[t] {
			// For each state with an edge from t to u labeled epsilon and NOT Ingoing IN closure
			if edge.Read == EPSILON && !closure[edge.Ingoing] {
				stack = append(stack, edge.Ingoing)
				closure[edge.Ingoing] = true
			}
		}
	}
	return closure
}

func (g *Graph) Subset() {
	marked := make(map[State]bool)
	// Find the epsilon closure of the start state.
	dStates := g.EClosure(g.Start)
	for state := range dStates {
		// We only consider unmarked states
		if !marked[state] {
			marked[state] = true
		}
	}
}