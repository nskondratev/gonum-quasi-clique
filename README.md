# gonum-quasi-clique
Implementation of algorithms for finding a.k.a. "quasi-cliques" (relaxation of clique) for [Gonum](https://www.gonum.org/) graph representation.

## Implemented algorithms
* Degree Decomposition Algorithm (DDA, degree-based quasi-cluqe) - [Pastukhov, G., Veremyev, A., Boginski, V., & Prokopyev, O. A. (2018). On maximum degree‐based‐quasi‐clique problem: Complexity and exact approaches. Networks, 71(2), 136-152.](https://doi.org/10.1002/net.21791)

## Prerequisites
* [Go v1.13](https://golang.org/dl/) (tested only with this version)

Some algorithms require external solver to be installed (e.g. DDA), currently only the following solvers are supported: 
* [GLPK (GNU Linear Programming Kit)](https://www.gnu.org/software/glpk/)

## Example
### DDA
```go
package main

import (
	"fmt"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	
	"github.com/nskondratev/gonum-quasi-clique/dda"
	"github.com/nskondratev/gonum-quasi-clique/dda/solvers/glpk"
)

func main() {
	// Create simple graph with 6 vertices and 10 edges
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
	ddaOpts := dda.Opts{
		InputGraph: g,
		Gamma:      0.5,
		// You can also specify dda.AllSolutions to get all maximal quasi-cliques
		SolveMode:  dda.OneSolution,
		YQCKSolver: glpk.Solve,
	}

	// Find solutions
	_, quasiCliqueSize, err := dda.DDA(ddaOpts)
	if err != nil {
		panic(err) // Just for example
	}
	fmt.Printf("Quasi-clique size: %d\n", quasiCliqueSize) // Output: Quasi-clique size: 5
}
```
