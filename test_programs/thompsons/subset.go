package main

func (g *Graph) Move(read string, states...StateKey) *StateSetExistence {
	reachable := make(StateSetExistence)
	for _, state := range states {
		for _, edge := range g.NFA.Get(state) {
			if edge.Read == read {
				reachable.Mark(edge.Ingoing)
			}
		}
	}
	return &reachable
}

func (g *Graph) PossibleInputs(states...StateKey) []string {
	possibleInputs := make(StateSetExistence)
	for _, state := range states {
		for _, edge := range g.NFA.Get(state) {
			if edge.Read != Epsilon {
				possibleInputs.Mark(StateKeyString(edge.Read))
			}
		}
	}
	inputs := make([]string, len(possibleInputs))
	i := 0
	for key := range possibleInputs {
		inputs[i] = key.Key()
		i++
	}
	return inputs
}

func (g *Graph) EClosure(stack...StateKey) (closure *StateSet) {
	closureSet := make(StateSetExistence)
	// Keep popping from the stack until we have visited everything.
	for len(stack) > 0 {
		index := len(stack) - 1
		t := stack[index]
		// We add the state to the eclosure
		closureSet.Mark(t)
		stack = stack[:index]
		for _, edge := range g.NFA.Get(t) {
			// For each state with an edge from t to u labeled epsilon and NOT Ingoing IN closureSet
			if edge.Read == Epsilon && !closureSet.Check(edge.Ingoing) {
				stack = append(stack, edge.Ingoing)
				closureSet.Mark(edge.Ingoing)
			}
		}
	}

	// Retrieve set keys
	return closureSet.Keys()
}

func (g *Graph) Subset() StateKey {
	marked := make(StateSetExistence)

	// Find the epsilon closure of the start state.
	start := g.EClosure(g.Start)
	dStates := []StateSet{*start}

	// We create a set version of dStates in order to keep track of what exists within it.
	dStatesSet := make(StateSetExistence)
	for _, state := range dStates {
		dStatesSet.Mark(state)
	}

	// We treat dStates as a queue
	for len(dStates) > 0 {
		// Pop from both the queue and the set
		states := dStates[0]
		dStates = dStates[1:]
		dStatesSet.Unmark(states)

		// We only consider unmarked states
		if !marked.Check(states) {
			marked.Mark(states)
			// For each possible input find states that can be moved to then find the epsilon closure of these states.
			for _, input := range g.PossibleInputs(states...) {
				U := g.EClosure(*(g.Move(input, states...).Keys())...)
				// If U is not in dStates...
				if !dStatesSet.Check(U) {
					dStatesSet.Mark(U)
					dStates = append(dStates, *U)
				}
				// Then we add the arrow going from the current state to U reading in the current input
				g.AddEdge(states, U, input, true)
			}
		}
	}
	return start
}