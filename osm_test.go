package main

// グラフの構造体定義
//
// type Node struct {
// 	ID  int
// 	Lat float64 // 緯度
// 	Lon float64 // 経度
// }
//
// type Edge struct {
// 	To     int
// 	Weight float64
// }
//
// type Graph struct {
// 	Nodes []Node
// 	Adj   [][]Edge
// }

import (
	"fmt"
	"testing"
)

func TestOSM(t *testing.T) {
	var g *Graph = JsonToGraph("data/osm_yokohama.json")
	fmt.Println("Num of nodes: ", len(g.Nodes))

	total := 0
	for _, adj_i := range g.Adj {
		total += len(adj_i)
	}
	fmt.Println("Sum of edges: ", total)
}
