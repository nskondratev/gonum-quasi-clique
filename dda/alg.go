// Package dda provides the implementation of Degree Decomposition Algorithm
// for finding maximal degree-based quasi-clique (QC)
package dda

import (
	"errors"
	"log"
	"math"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

// GammaQCkSolver describes the function for solving gamma-QC(k) mixed-integer problem
type GammaQCkSolver func(in graph.Undirected, gamma float64, k int64, allSolutions bool) (qcNodes []graph.Nodes, qcSize int64, err error)

type GraphBuilder interface {
	graph.NodeAdder
	graph.EdgeAdder
	graph.Graph
}

// SolveMode describes how many solutions will be found by DDA
type SolveMode bool

const (
	OneSolution  SolveMode = false
	AllSolutions SolveMode = true
)

type DDAOpts struct {
	InputGraph       graph.Undirected
	Gamma            float64
	GraphConstructor func() GraphBuilder
	EdgeConstructor  func(n1, n2 graph.Node) graph.Edge
	SolveMode        SolveMode
	YQCKSolver       GammaQCkSolver
}

func (o DDAOpts) validate() error {
	if o.YQCKSolver == nil {
		return errors.New("dda: YQCSolver must be provided")
	}
	if o.GraphConstructor == nil {
		return errors.New("dda: GraphConstructor must be provided")
	}
	if o.EdgeConstructor == nil {
		return errors.New("dda: EdgeConstructor must be provided")
	}
	if o.InputGraph == nil {
		return errors.New("dda: InputGraph must be provided")
	}
	if o.Gamma < 0.0 || o.Gamma > 1.0 {
		return errors.New("dda: Gamma must be between 0 and 1")
	}
	return nil
}

// Implementation of DDA
// Returns three values: the array of maximal QC, the size of the max QC and the error
func DDA(opts DDAOpts) ([]graph.Undirected, int64, error) {
	var currentMax int64
	var res []graph.Undirected
	if err := opts.validate(); err != nil {
		return res, currentMax, err
	}

	degeneracy := graphDegeneracy(opts.InputGraph)

	var currentSolution []graph.Nodes

	k := int64(degeneracy) + 1
	for currentMax < int64(math.Floor(float64(k)/opts.Gamma))+1 {
		k--
		//log.Printf("\n==================\nSolve y-QC(k) for k = %d\n==================\n", k)
		qcNodes, qcSize, err := opts.YQCKSolver(opts.InputGraph, opts.Gamma, k, bool(opts.SolveMode))
		if err != nil {
			//return nil, errors.Wrap(err, "failed to solve gamma quasi clique problem")
			log.Printf("Error while solving y-QC(k) problem: %s\n", err.Error())
		}
		if qcNodes != nil {
			//log.Printf("Solution is found. Num of vertices: %d\n", qcSize)
			if qcSize > currentMax {
				currentMax = qcSize
				currentSolution = make([]graph.Nodes, len(qcNodes))
				for i, ns := range qcNodes {
					currentSolution[i] = ns
				}
			}
		}
	}
	// Write result to dst
	if len(currentSolution) > 0 {
		res = make([]graph.Undirected, len(currentSolution))

		for i, nodes := range currentSolution {
			dst := opts.GraphConstructor()
			// Add nodes
			for nodes.Next() {
				n := nodes.Node()
				dst.AddNode(n)
			}
			// Add edges
			outerIt := dst.Nodes()
			for outerIt.Next() {
				outerNode := outerIt.Node()
				innerIt := dst.Nodes()
				for innerIt.Next() {
					innerNode := innerIt.Node()
					if opts.InputGraph.HasEdgeBetween(innerNode.ID(), outerNode.ID()) {
						dst.SetEdge(simple.Edge{F: innerNode, T: outerNode})
					}
				}
			}
			res[i] = dst.(graph.Undirected)
		}
	}

	return res, currentMax, nil
}
