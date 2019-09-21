package dda

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/zimmski/osutil"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/simple"

	"github.com/nskondratev/gonum-quasi-clique/dda/solvers/glpk"
)

func TestDDA(t *testing.T) {
	g, _ := getGraphFromFile(path.Join("..", "testdata", "small.dot"))()

	cases := []struct {
		gamma          float64
		solveMode      SolveMode
		expectedQC     [][]int
		expectedQCSize int64
	}{
		{1.0, OneSolution, [][]int{{0, 1, 3, 4}}, 4},
		{0.5, OneSolution, [][]int{{0, 1, 3, 4, 5}}, 5},
		{0.5, AllSolutions, [][]int{{0, 1, 3, 4, 5}, {0, 1, 2, 3, 4}}, 5},
	}

	for i, c := range cases {
		_, err := osutil.CaptureWithCGo(func() {
			ddaOpts := DDAOpts{
				InputGraph: g,
				Gamma:      c.gamma,
				GraphConstructor: func() GraphBuilder {
					return simple.NewUndirectedGraph()
				},
				EdgeConstructor: func(n1, n2 graph.Node) graph.Edge {
					return simple.Edge{F: n1, T: n2}
				},
				SolveMode:  c.solveMode,
				YQCKSolver: glpk.Solve,
			}
			quasiCliques, qcSize, err := DDA(ddaOpts)
			if err != nil {
				t.Errorf("[%d] Unexpected error: %s", i, err.Error())
			}
			if qcSize != c.expectedQCSize {
				t.Errorf("[%d] Quasi-clique size mismatch. Expected %d, got %d", i, c.expectedQCSize, qcSize)
			}

			if len(quasiCliques) != len(c.expectedQC) {
				t.Errorf("[%d] Quasi-cliques count mismatch. Expected %d, got %d", i, len(c.expectedQC), len(quasiCliques))
			}

			for i := 0; i < len(quasiCliques); i++ {
				if !nodesEqual(quasiCliques[i], c.expectedQC[i]) {
					t.Errorf("[%d] quasi-cliqes mismatch", i)
				}
			}
		})
		if err != nil {
			t.Errorf("[%d] Error while capturing stdout: %s", i, err.Error())
		}
	}
}

func BenchmarkDDA_glpk(b *testing.B) {
	gammas := []float64{1.0, 0.9, 0.8, 0.6, 0.5}

	cases := []struct {
		name      string
		solveMode SolveMode
		getGraph  graphGetter
	}{
		// Small graph: 6 nodes
		{"small", OneSolution, getGraphFromFile(path.Join("..", "testdata", "small.dot"))},
		{"small", AllSolutions, getGraphFromFile(path.Join("..", "testdata", "small.dot"))},
		// Social network
		{"social_210", OneSolution, getGraphFromFile(filepath.Join("..", "testdata", "social.dot"))},
		{"social_210", AllSolutions, getGraphFromFile(filepath.Join("..", "testdata", "social.dot"))},
		// GNP 50 nodes
		{"gnp_50_p_0.1", OneSolution, getGraphFromFile(filepath.Join("..", "testdata", "gnp_50_0.1.dot"))},
		{"gnp_50_p_0.1", AllSolutions, getGraphFromFile(filepath.Join("..", "testdata", "gnp_50_0.1.dot"))},
		{"gnp_50_p_0.3", OneSolution, getGraphFromFile(filepath.Join("..", "testdata", "gnp_50_0.3.dot"))},
		{"gnp_50_p_0.5", OneSolution, getGraphFromFile(filepath.Join("..", "testdata", "gnp_50_0.5.dot"))},
		// GNP 100 nodes
		{"gnp_100_p_0.1", OneSolution, getGraphFromFile(filepath.Join("..", "testdata", "gnp_100_0.1.dot"))},
		{"gnp_100_p_0.3", OneSolution, getGraphFromFile(filepath.Join("..", "testdata", "gnp_100_0.3.dot"))},
		// GNP 1000 nodes
		{"gnp_1000_p_0.1", OneSolution, getGraphFromFile(filepath.Join("..", "testdata", "gnp_1000_0.1.dot"))},
	}

	for i, c := range cases {
		for _, gamma := range gammas {
			b.Run(fmt.Sprintf("%s_y_%g_all_solutions_%v", c.name, gamma, c.solveMode), func(b *testing.B) {
				g, err := c.getGraph()
				if err != nil {
					b.Fatalf("[%d] Failed to get graph: %s", i, err)
				}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					_, err := osutil.CaptureWithCGo(func() {
						b.StartTimer()
						_, _, err := DDA(DDAOpts{
							InputGraph: g,
							Gamma:      gamma,
							GraphConstructor: func() GraphBuilder {
								return simple.NewUndirectedGraph()
							},
							EdgeConstructor: func(n1, n2 graph.Node) graph.Edge {
								return simple.Edge{F: n1, T: n2}
							},
							SolveMode:  c.solveMode,
							YQCKSolver: glpk.Solve,
						})
						if err != nil {
							b.Fatalf("[%d] Unexpected error: %s", i, err.Error())
						}
						b.StopTimer()
					})
					if err != nil {
						b.Fatalf("[%d] Unexpected error: %s", i, err.Error())
					}
				}
			})
		}
	}
}

type graphGetter func() (graph.Undirected, error)

func getGraphFromFile(filename string) func() (graph.Undirected, error) {
	return func() (graph.Undirected, error) {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		g := simple.NewUndirectedGraph()
		err = dot.Unmarshal(b, g)
		return g, err
	}
}

func nodesEqual(g graph.Graph, expected []int) bool {
	var nodes []int
	it := g.Nodes()
	for it.Next() {
		nodes = append(nodes, int(it.Node().ID()))
	}
	sort.Ints(nodes)
	return reflect.DeepEqual(nodes, expected)
}