package main

import (
	"fmt"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"strconv"
)

func (g *Graph) Visualise(start State, filePrefix string) error {
	viz := graphviz.New()
	graph, err := viz.Graph()
	graph.SetForceLabels(true)
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

	queue := []State{start}
	marked := make(map[State]struct{})
	for len(queue) > 0 {
		currentState := queue[0]
		queue = queue[1:]
		// Mark the startNode
		marked[currentState] = struct{}{}
		// Create/fetch the startNode
		outgoingName := strconv.Itoa(int(currentState))
		var startNode *cgraph.Node
		if startNode, _ = graph.Node(outgoingName); startNode == nil {
			startNode, err = graph.CreateNode(outgoingName)
			if err != nil {
				return err
			}
		}

		// Set the style of the node
		if g.AcceptingStates[currentState] {
			// If the node is an accepting state then we will set it to be a double circle
			startNode.SetShape(cgraph.DoubleCircleShape)
		} else if currentState == g.Start {
			startNode.SetShape(cgraph.PointShape)
		} else {
			startNode.SetShape(cgraph.CircleShape)
		}

		// Iterate over all adjacent nodes
		for _, edge := range g.Graph[currentState] {
			// If the startNode is not marked then we will add it to the queue
			if _, ok := marked[edge.Ingoing]; !ok {
				queue = append(queue, edge.Ingoing)
			}
			// We still need to draw the edge
			ingoingName := strconv.Itoa(int(edge.Ingoing))
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
			if edge.Read != EPSILON {
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
