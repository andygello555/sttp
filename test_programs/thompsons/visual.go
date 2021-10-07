package main

import (
	"fmt"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

func (g *Graph) Visualise(start StateKey, filePrefix string, dfa bool) error {
	viz := graphviz.New()
	graph, err := viz.Graph()
	if err != nil {
		return err
	}

	defer func() {
		if err := graph.Close(); err != nil {
			panic(err)
		}
		err := viz.Close()
		if err != nil {
			panic(err)
		}
	}()

	queue := []StateKey{start}
	marked := make(StateSetExistence)
	for len(queue) > 0 {
		currentState := queue[0]
		queue = queue[1:]
		// Mark the startNode
		marked.Mark(currentState)
		// Create/fetch the startNode
		outgoingName := currentState.Key()
		var startNode *cgraph.Node
		if startNode, _ = graph.Node(outgoingName); startNode == nil {
			startNode, err = graph.CreateNode(outgoingName)
			if err != nil {
				return err
			}
		}

		// Set the style of the node
		// This depends on the graph type (NFA vs DFA)
		if !dfa {
			if g.AcceptingStates.Check(currentState) {
				// If the node is an accepting state then we will set it to be a double circle
				startNode.SetShape(cgraph.DoubleCircleShape)
			} else if currentState == g.Start {
				startNode.SetShape(cgraph.PointShape)
			} else {
				startNode.SetShape(cgraph.CircleShape)
			}
		} else {
			// Check if the state contains any of the accepting states in the NFA
			if g.CheckIfAccepting(currentState) {
				startNode.SetShape(cgraph.DoubleOctagonShape)
			} else if currentState == start {
				startNode.SetShape(cgraph.PointShape)
			} else {
				startNode.SetShape(cgraph.OctagonShape)
			}
		}

		// Iterate over all adjacent nodes
		adjacencyList := &g.NFA
		if dfa {
			adjacencyList = &g.DFA
		}
		for _, edge := range adjacencyList.Get(currentState) {
			// If the startNode is not marked then we will add it to the queue
			if !marked.Check(edge.Ingoing) {
				queue = append(queue, edge.Ingoing)
			}
			// We still need to draw the edge
			ingoingName := edge.Ingoing.Key()
			var endNode *cgraph.Node
			if endNode, _ = graph.Node(ingoingName); endNode == nil {
				endNode, err = graph.CreateNode(ingoingName)
				if err != nil {
					return err
				}
			}
			fmt.Println(edge.Read, edge.Outgoing, edge.Ingoing, startNode, endNode)
			var graphEdge *cgraph.Edge
			if graphEdge, err = graph.CreateEdge(edge.Read, startNode, endNode); err != nil {
				return err
			}
			if edge.Read != Epsilon {
				graphEdge.SetLabel(edge.Read)
			} else {
				graphEdge.SetLabel("ε")
			}
		}
	}

	// Render visualisations to both png and dot
	if err := viz.RenderFilename(graph, graphviz.PNG, fmt.Sprintf("%s.png", filePrefix)); err != nil {
		return err
	}
	if err := viz.RenderFilename(graph, graphviz.XDOT, fmt.Sprintf("%s.dot", filePrefix)); err != nil {
		return err
	}
	return nil
}

func (tt *TransitionTable) Visualisation() {

}

