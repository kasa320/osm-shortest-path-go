// OSMのJSONデータをパースするプログラム
//
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

package main

import (
	"encoding/json"
	"os"
)

type OSMData struct {
	Elements []Element `json:"elements"`
}

// node と wayが混在。wayはnodeのつながり。nodesフィールドが道上のノードを表す。
type Element struct {
	Type  string  `json:"type"` // "node" or "way"
	ID    int64   `json:"id"`
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	Nodes []int64 `json:"nodes"`
}

// パースのメイン。JSONをグラフに変換する一連の流れ。
func JsonToGraph(data_path string) *Graph {
	f, _ := os.Open(data_path)
	defer f.Close()

	var data OSMData
	json.NewDecoder(f).Decode(&data)

	g := Graph{[]Node{}, [][]Edge{}}
	nodeIDs := map[int64]int{} // map[OSMID] = 内部インデックス

	// すべてのelementsを走査してノードをグラフに追加。
	for _, e := range data.Elements {
		if e.Type == "node" {
			addNode(&g, e, nodeIDs)
		}
	}

	// 同様にelementsを操作して辺を隣接リストとしてグラフに追加。
	for _, e := range data.Elements {
		if e.Type == "way" {
			addEdge(&g, e, nodeIDs)
		}
	}
	return &g
}

func addNode(g *Graph, e Element, nodeIDs map[int64]int) {
	idx := len(g.Nodes) // 内部ID
	nodeIDs[e.ID] = idx

	node := Node{
		idx, // e.IDではなく、内部IDを用いる(0, ... V-1)
		e.Lat,
		e.Lon,
	}

	g.Nodes = append(g.Nodes, node)
	g.Adj = append(g.Adj, []Edge{})
}

func addEdge(g *Graph, e Element, nodeIDs map[int64]int) {
	for i := 0; i < len(e.Nodes)-1; i++ {
		id1_osm, id2_osm := e.Nodes[i], e.Nodes[i+1]
		id1, id2 := nodeIDs[id1_osm], nodeIDs[id2_osm]

		weight := g.distance(id1, id2)
		g.Adj[id1] = append(g.Adj[id1], Edge{id2, weight})
		g.Adj[id2] = append(g.Adj[id2], Edge{id1, weight})
	}
}
