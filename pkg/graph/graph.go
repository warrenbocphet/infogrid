package graph

type Node struct {
	ID        int
	Neighbors map[int]float64
	Value     float64
}

type Graph struct {
	Nodes []Node
}

func (g *Graph) AddNode(id int, value float64, neighbors ...int) {
	neighborsMap := make(map[int]float64)
	for _, n := range neighbors {
		if n == id { // The neighbor cannot be itself
			continue
		}
		neighborsMap[n] = -1
	}
	newNode := Node{ID: id,
		Neighbors: neighborsMap,
		Value:     value}

	if (id >= len(g.Nodes)) || (id < 0) {
		// by default, append the node to the end
		g.Nodes = append(g.Nodes, newNode)

	} else if id < len(g.Nodes) {
		// replace previous node with current node
		g.Nodes[id] = newNode

	}

}
