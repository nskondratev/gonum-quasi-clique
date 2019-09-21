package glpk

import (
	"fmt"
	"log"
	"math"
	"runtime"

	"github.com/lukpank/go-glpk/glpk"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/iterator"
)

const (
	eps = 1e-6
)

func Solve(in graph.Undirected, gamma float64, k int64, allSolutions bool) ([]graph.Nodes, int64, error) {
	var quasiClique graph.Nodes
	var solutionNodesCount int64

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	lp := glpk.New()
	defer lp.Delete()

	lp.SetProbName("Gamma quasi clique problem")
	lp.SetObjName("Number of vertices")
	lp.SetObjDir(glpk.MAX)

	nodes := in.Nodes()
	nodesCount := nodes.Len()

	lp.AddCols(nodesCount)
	ind := make([]int32, nodesCount+1)
	ones := make([]float64, nodesCount+1)
	for i := 0; i < nodesCount; i++ {
		lp.SetColName(i+1, fmt.Sprintf("x%d", i+1))
		lp.SetColKind(i+1, glpk.BV)
		lp.SetObjCoef(i+1, 1)
		ind[i+1] = int32(i + 1)
		ones[i+1] = 1.0
	}

	lp.AddRows(nodesCount + 1)

	// Add constraints for each vertex
	for i := 0; i < nodesCount; i++ {
		coefs := make([]float64, nodesCount+1)
		lp.SetRowBnds(i+1, glpk.LO, 0.0, 0.0)
		lp.SetRowName(i+1, fmt.Sprintf("x%d has more than %d neighbours", i+1, k))
		for j := 0; j < nodesCount; j++ {
			var matVal float64
			switch {
			case i == j:
				matVal = float64(-k)
			case in.HasEdgeBetween(int64(i), int64(j)):
				matVal = 1.0
			}
			coefs[j+1] = matVal
		}
		lp.SetMatRow(i+1, ind, coefs)
	}

	lp.SetRowBnds(nodesCount+1, glpk.UP, 0.0, math.Floor(float64(k)/gamma)+1.0)
	lp.SetRowName(nodesCount+1, "qc is k-core")
	lp.SetMatRow(nodesCount+1, ind, ones)

	iocp := glpk.NewIocp()
	iocp.SetPresolve(true)
	iocp.SetMsgLev(glpk.MSG_OFF)

	if err := lp.Intopt(iocp); err != nil {
		log.Printf("Mip error: %v", err)
		return nil, 0, nil
	}

	var quasiCliques []graph.Nodes
	var prevSolution []int
	var prevSolutionSize int64
	if lp.MipStatus() == glpk.OPT {
		solutionNodesCount = int64(lp.MipObjVal())

		prevSolutionSize = solutionNodesCount
		//fmt.Printf("%s = %d", lp.ObjName(), solutionNodesCount)
		qcI := 0
		qcNodes := make([]graph.Node, solutionNodesCount)
		for i := 0; i < lp.NumCols(); i++ {
			//fmt.Printf("; %s = %g", lp.ColName(i+1), lp.MipColVal(i+1))
			if isOne(lp.MipColVal(i + 1)) {
				qcNodes[qcI] = in.Node(int64(i))
				prevSolution = append(prevSolution, i)
				qcI++
			}
		}
		//fmt.Println()
		quasiClique = iterator.NewOrderedNodes(qcNodes)
		quasiCliques = append(quasiCliques, quasiClique)

		// Find other solutions if necessary
		if allSolutions {
			//log.Printf("Find all solutions. Prev solution size: %d, prev solution: %v\n", prevSolutionSize, prevSolution)
			for prevSolutionSize == solutionNodesCount && lp.MipStatus() == glpk.OPT {
				// Exclude previous solution
				lp.AddRows(1)
				coefs := make([]float64, nodesCount+1)
				//log.Printf("coefs: %v\n", coefs)
				for _, i := range prevSolution {
					coefs[i+1] = 1.0
				}
				//log.Printf("coefs: %v\n", coefs)
				lp.SetRowBnds(lp.NumRows(), glpk.UP, 0.0, float64(prevSolutionSize)-0.5)
				lp.SetMatRow(lp.NumRows(), ind, coefs)
				lp.SetRowName(lp.NumRows(), fmt.Sprintf("exclude_prev_solution_%d", lp.NumRows()))

				//_ = lp.WriteLP(nil, "lp.txt")
				//_ = lp.WriteMPS(glpk.MPS_FILE, nil, "mps.txt")

				// Solve again
				if err := lp.Intopt(iocp); err != nil {
					log.Printf("Mip error: %v", err)
					return nil, 0, nil
				}

				//log.Printf("New solutuon is found. Status: %s\n", solStatToString(lp.MipStatus()))

				// If solution found and its size is equal to max solution, store it
				if lp.MipStatus() == glpk.OPT {
					prevSolutionSize = int64(lp.MipObjVal())
					//log.Printf("New solution size: %d\n", prevSolutionSize)
					prevSolution = make([]int, 0)
					if prevSolutionSize == solutionNodesCount {
						qcI := 0
						qcNodes := make([]graph.Node, solutionNodesCount)
						for i := 0; i < lp.NumCols(); i++ {
							//fmt.Printf("; %s = %g", lp.ColName(i+1), lp.MipColVal(i+1))
							if isOne(lp.MipColVal(i + 1)) {
								qcNodes[qcI] = in.Node(int64(i))
								prevSolution = append(prevSolution, i)
								qcI++
							}
						}
						//fmt.Println()
						quasiClique = iterator.NewOrderedNodes(qcNodes)
						quasiCliques = append(quasiCliques, quasiClique)
					}
				}
			}
		}
	}

	return quasiCliques, solutionNodesCount, nil
}

func isOne(val float64) bool {
	return 1.0-eps <= val && val <= 1.0+eps
}

func solStatToString(solStat glpk.SolStat) string {
	switch solStat {
	case glpk.UNDEF:
		return "UNDEF"
	case glpk.UNBND:
		return "UNBND"
	case glpk.OPT:
		return "OPT"
	case glpk.FEAS:
		return "FEAS"
	case glpk.INFEAS:
		return "INFEAS"
	case glpk.NOFEAS:
		return "NOFEAS"
	default:
		return "UNKNOWN"
	}
}
