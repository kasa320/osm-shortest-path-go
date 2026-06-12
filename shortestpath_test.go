package main

import (
	"math"
	"testing"
)

// floatスライスがほぼ等しいか（浮動小数の誤差を許容）
func distEqual(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.IsInf(a[i], 1) && math.IsInf(b[i], 1) {
			continue
		}
		if math.Abs(a[i]-b[i]) > 1e-9 {
			return false
		}
	}
	return true
}

// sample_graph10 は手計算しやすい10ノードの無向重み付きグラフ。
// 座標は東京近辺に散らし、A*のヒューリスティックも自然に働くようにしてある。
//
// 辺（無向）:
//
//	0-1=4, 0-2=1, 1-2=2, 1-3=1, 2-3=5, 3-4=3,
//	4-5=2, 3-5=6, 5-6=4, 5-7=3, 6-7=1, 7-8=2, 6-9=5, 8-9=3
//
// 始点0からの最短距離（手計算）:
//
//	0:0  2:1  1:3  3:4  4:7  5:9  6:13  7:12  8:14  9:17
func sample_graph10() Graph {
	Nodes := []Node{
		{ID: 0, Lat: 35.6812, Lon: 139.7671}, // 東京駅
		{ID: 1, Lat: 35.6895, Lon: 139.6917}, // 新宿
		{ID: 2, Lat: 35.6586, Lon: 139.7454}, // 東京タワー
		{ID: 3, Lat: 35.7100, Lon: 139.8107}, // スカイツリー
		{ID: 4, Lat: 35.7295, Lon: 139.7109}, // 池袋
		{ID: 5, Lat: 35.6284, Lon: 139.7367}, // 品川
		{ID: 6, Lat: 35.6580, Lon: 139.7016}, // 渋谷
		{ID: 7, Lat: 35.6654, Lon: 139.7707}, // 銀座
		{ID: 8, Lat: 35.7148, Lon: 139.7967}, // 浅草
		{ID: 9, Lat: 35.7056, Lon: 139.7519}, // 上野
	}
	// 無向グラフ：各辺を両方向に登録する
	undirected := [][3]float64{
		{0, 1, 4}, {0, 2, 1}, {1, 2, 2}, {1, 3, 1}, {2, 3, 5}, {3, 4, 3},
		{4, 5, 2}, {3, 5, 6}, {5, 6, 4}, {5, 7, 3}, {6, 7, 1}, {7, 8, 2},
		{6, 9, 5}, {8, 9, 3},
	}
	Adj := make([][]Edge, len(Nodes))
	for _, e := range undirected {
		u, v, w := int(e[0]), int(e[1]), e[2]
		Adj[u] = append(Adj[u], Edge{To: v, Weight: w})
		Adj[v] = append(Adj[v], Edge{To: u, Weight: w})
	}
	return Graph{Nodes: Nodes, Adj: Adj}
}

// 10ノードグラフで5アルゴリズムがpanicせず完走し、メトリクスが妥当な値を返すことを検証する。
// （各アルゴリズムは経路と距離を内部でPrintPath表示し、戻り値は*Metricsのみ）
func TestShortestPath10(t *testing.T) {
	g := sample_graph10()
	start, goal := 0, 9

	matrix := ToMatrix(g)
	mM := DijkstraMatrix(start, goal, matrix)
	mPQ := DijkstraPQ(start, goal, g)
	mBF := BellmanFord(start, goal, g)
	mA := Astar(start, goal, g)
	mBi := BiDijkstraPQ(start, goal, g)

	// 全アルゴリズムで緩和が1回以上行われていること（探索が走った証拠）
	for name, m := range map[string]*Metrics{
		"DijkstraMatrix": mM,
		"DijkstraPQ":     mPQ,
		"BellmanFord":    mBF,
		"Astar":          mA,
		"BiDijkstraPQ":   mBi,
	} {
		if m.Relaxations <= 0 {
			t.Errorf("%s: Relaxations=%d, want > 0", name, m.Relaxations)
		}
	}

	// PQ系（PQ/A*/双方向）はpush・ヒープサイズが正、行列/BFは0であること
	if mPQ.PushCount <= 0 || mA.PushCount <= 0 || mBi.PushCount <= 0 {
		t.Errorf("PQ系のPushCountが0以下: PQ=%d A*=%d Bi=%d",
			mPQ.PushCount, mA.PushCount, mBi.PushCount)
	}
	if mM.PushCount != 0 || mBF.PushCount != 0 {
		t.Errorf("非PQ系のPushCountが0でない: Matrix=%d BF=%d", mM.PushCount, mBF.PushCount)
	}
	// Bellman-Fordはノード展開の概念を持たない（Expansions=0）
	if mBF.Expansions != 0 {
		t.Errorf("BellmanFord.Expansions=%d, want 0", mBF.Expansions)
	}
}
