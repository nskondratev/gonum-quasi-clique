package dda

import (
	"flag"
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

var allSolutions bool

func init() {
	flag.BoolVar(&allSolutions, "all-solutions", false, "Set to find all maximal quasi-cliques")
}

func TestDDA(t *testing.T) {
	g, _ := getGraphFromFile(path.Join("..", "testdata", "small.dot"))()

	cases := []struct {
		gamma          float64
		solveMode      SolveMode
		expectedQC     [][]int
		expectedQCSize int64
	}{
		{1.0, OneSolution, [][]int{{0, 1, 3, 4}}, 4},
		{0.5, OneSolution, [][]int{}, 5},
		{0.5, AllSolutions, [][]int{{0, 1, 3, 4, 5}, {0, 1, 2, 3, 4}}, 5},
	}

	for i, c := range cases {
		_, err := osutil.CaptureWithCGo(func() {
			ddaOpts := Opts{
				InputGraph: g,
				Gamma:      c.gamma,
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

			if len(c.expectedQC) > 0 && len(quasiCliques) != len(c.expectedQC) {
				t.Errorf("[%d] Quasi-cliques count mismatch. Expected %d, got %d", i, len(c.expectedQC), len(quasiCliques))
			}

			if len(c.expectedQC) > 0 && !quasiCliquesAreEqual(quasiCliques, c.expectedQC) {
				t.Errorf("[%d] quasi-cliqes mismatch.", i)
			}
		})
		if err != nil {
			t.Errorf("[%d] Error while capturing stdout: %s", i, err.Error())
		}
	}
}

func BenchmarkDDA_glpk(b *testing.B) {
	flag.Parse()
	fmt.Printf("Find all solutions: %v\n", allSolutions)

	gammas := []float64{1.0, 0.9, 0.8, 0.65}

	solveMode := OneSolution
	if allSolutions {
		solveMode = AllSolutions
	}

	cases := []struct {
		name     string
		getGraph graphGetter
	}{
		// Small graph: 6 nodes
		// {"small", getGraphFromFile(path.Join("..", "testdata", "small.dot"))},
		// Social network
		// {"social_210", getGraphFromFile(filepath.Join("..", "testdata", "social.dot"))},
		// GNP 50 nodes
		// {"gnp_50_p_0.1", getGraphFromFile(filepath.Join("..", "testdata", "gnp_50_0.1.dot"))},
		// {"gnp_50_p_0.3", getGraphFromFile(filepath.Join("..", "testdata", "gnp_50_0.3.dot"))},
		// GNP 100 nodes
		// {"gnp_100_p_0.1", getGraphFromFile(filepath.Join("..", "testdata", "gnp_100_0.1.dot"))},
		// GNP 500 nodes
		{"gnp_500_p_0.01", getGraphFromFile(filepath.Join("..", "testdata", "gnp_500_0.01.dot"))},
	}

	for i, c := range cases {
		for _, gamma := range gammas {
			b.Run(fmt.Sprintf("%s_y_%g", c.name, gamma), func(b *testing.B) {
				g, err := c.getGraph()
				if err != nil {
					b.Fatalf("[%d] Failed to get graph: %s", i, err)
				}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					_, err := osutil.CaptureWithCGo(func() {
						b.StartTimer()
						_, _, err := DDA(Opts{
							InputGraph: g,
							Gamma:      gamma,
							SolveMode:  solveMode,
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

func quasiCliquesAreEqual(received []graph.Nodes, expected [][]int) bool {
OUTER:
	for _, it := range received {
		var nodes []int
		for it.Next() {
			nodes = append(nodes, int(it.Node().ID()))
		}
		sort.Ints(nodes)
		for _, e := range expected {
			if reflect.DeepEqual(nodes, e) {
				continue OUTER
			}
		}
		return false
	}
	return true
}
