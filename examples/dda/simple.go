package main

import (
	"fmt"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"

	"github.com/nskondratev/gonum-quasi-clique/dda"
	"github.com/nskondratev/gonum-quasi-clique/dda/solvers/glpk"
)

func main() {
	g := simple.NewUndirectedGraph()

	// Add nodes
	for i := 0; i < 6; i++ {
		g.AddNode(simple.Node(i))
	}

	// Add edges
	g.SetEdge(simple.Edge{F: simple.Node(0), T: simple.Node(1)})
	g.SetEdge(simple.Edge{F: simple.Node(0), T: simple.Node(3)})
	g.SetEdge(simple.Edge{F: simple.Node(0), T: simple.Node(4)})
	g.SetEdge(simple.Edge{F: simple.Node(0), T: simple.Node(5)})
	g.SetEdge(simple.Edge{F: simple.Node(1), T: simple.Node(2)})
	g.SetEdge(simple.Edge{F: simple.Node(1), T: simple.Node(3)})
	g.SetEdge(simple.Edge{F: simple.Node(1), T: simple.Node(4)})
	g.SetEdge(simple.Edge{F: simple.Node(2), T: simple.Node(3)})
	g.SetEdge(simple.Edge{F: simple.Node(3), T: simple.Node(4)})
	g.SetEdge(simple.Edge{F: simple.Node(4), T: simple.Node(5)})

	// Set up DDA options
	ddaOpts := dda.DDAOpts{
		InputGraph: g,
		Gamma:      0.5,
		GraphConstructor: func() dda.GraphBuilder {
			return simple.NewUndirectedGraph()
		},
		EdgeConstructor: func(n1, n2 graph.Node) graph.Edge {
			return simple.Edge{F: n1, T: n2}
		},
		SolveMode:  dda.OneSolution, // You can also specify dda.AllSolutions to get all maximal cliques. But it will be slower
		YQCKSolver: glpk.Solve,
	}

	// Find solutions
	_, quasiCliqueSize, err := dda.DDA(ddaOpts)
	if err != nil {
		panic(err) // Just for example
	}
	fmt.Printf("Quasi-clique size: %d\n", quasiCliqueSize) // Output: Quasi-clique size: 5
}
