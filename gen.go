package main

import (
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/graphs/gen"
	"gonum.org/v1/gonum/graph/simple"
	"io/ioutil"
	"log"
)

type ToGen struct {
	N         int
	P         float64
	GraphName string
	Filename  string
}

func main() {
	toGen := []ToGen{
		{500, 0.01, "nodes_500_p_0.01", "./testdata/gnp_500_0.01.dot"},
		{1000, 0.01, "nodes_1000_p_0.01", "./testdata/gnp_1000_0.01.dot"},
	}
	log.Printf("Start generating graphs. Number: %d\n", len(toGen))
	for _, gn := range toGen {
		log.Printf("Generate graph %s\n", gn.GraphName)
		g := simple.NewUndirectedGraph()
		if err := gen.Gnp(g, gn.N, gn.P, nil); err != nil {
			log.Fatal(err)
		}
		bs, err := dot.Marshal(g, gn.GraphName, "", " ")
		if err != nil {
			log.Fatal(err)
		}
		if err := ioutil.WriteFile(gn.Filename, bs, 0644); err != nil {
			log.Fatal(err)
		}
	}
	log.Println("Done.")
}
