// Package dda provides the implementation of Degree Decomposition Algorithm
// for finding maximal degree-based quasi-clique (QC)
package dda

import (
	"errors"
	"log"
	"math"

	"gonum.org/v1/gonum/graph"
)

// GammaQCkSolver describes the function for solving gamma-QC(k) mixed-integer problem
type GammaQCkSolver func(
	in graph.Undirected,
	nodes []graph.Node,
	gamma float64,
	k int64,
	allSolutions bool,
) (qcNodes []graph.Nodes, qcSize int64, err error)

// SolveMode describes how many solutions will be found by DDA
type SolveMode bool

const (
	OneSolution  SolveMode = false
	AllSolutions SolveMode = true
)

type Opts struct {
	InputGraph graph.Undirected
	Gamma      float64
	SolveMode  SolveMode
	YQCKSolver GammaQCkSolver
}

func (o Opts) validate() error {
	if o.YQCKSolver == nil {
		return errors.New("dda: YQCSolver must be provided")
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
func DDA(opts Opts) ([]graph.Nodes, int64, error) {
	var currentMax int64
	if err := opts.validate(); err != nil {
		return []graph.Nodes{}, currentMax, err
	}

	degeneracy := graphDegeneracy(opts.InputGraph)

	var currentSolution []graph.Nodes

	k := int64(degeneracy) + 1
	for currentMax < int64(math.Floor(float64(k)/opts.Gamma))+1 {
		k--
		qcNodes, qcSize, err := opts.YQCKSolver(
			opts.InputGraph,
			graph.NodesOf(opts.InputGraph.Nodes()),
			opts.Gamma,
			k,
			bool(opts.SolveMode),
		)
		if err != nil {
			log.Printf("Error while solving y-QC(k) problem: %s\n", err.Error())
		}
		if qcNodes != nil {
			if qcSize > currentMax {
				currentMax = qcSize
				currentSolution = make([]graph.Nodes, len(qcNodes))
				for i, ns := range qcNodes {
					currentSolution[i] = ns
				}
			}
		}
	}
	return currentSolution, currentMax, nil
}
