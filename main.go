package main

import (
	"fmt"
	"math"
	"sort"
)

const benchN = 100  // 実行時間計測の反復回数
const numPairs = 10 // 評価する始点-終点ペア数

const DATA_PATH = "data/osm_data.json"

func main() {
	g := JsonToGraph(DATA_PATH)
	fmt.Printf("グラフ: V=%d, E(無向)=%d\n\n", len(g.Nodes), countUndirectedEdges(g))

	// 始点を固定し、そこから到達可能なノードを道路距離で層化して終点を選ぶ。
	// （横浜データは橋などで分断された複数の連結成分を含むため、到達可能性で絞る）
	start := 5630
	dist := fullDijkstra(start, *g) // 始点からの全ノード道路距離
	goals := pickGoals(start, dist, numPairs)

	matrix := ToMatrix(*g)

	// 各ペアを評価し、迂回率と各アルゴリズムのメトリクスを出力する
	for i, goal := range goals {
		straight := g.distance(start, goal)
		road := dist[goal]
		detour := straight / road // 直線/道路 ≤ 1（A*ヒューリスティックの効きの目安）

		fmt.Printf("=== ペア%d: 始点=%d 終点=%d 直線=%.0fm 道路=%.0fm 迂回率=%.3f ===\n",
			i+1, start, goal, straight, road, detour)
		benchPair(start, goal, *g, matrix)
		fmt.Println()
	}

	// レポート本文の代表ペア（5630→4371）。図表との整合のため固定計測する。
	rep := 4371
	fmt.Printf("=== 代表ペア: 始点=%d 終点=%d 直線=%.0fm 道路=%.0fm 迂回率=%.3f ===\n",
		start, rep, g.distance(start, rep), dist[rep], g.distance(start, rep)/dist[rep])
	benchPair(start, rep, *g, matrix)
	fmt.Println()

	// 全アルゴリズムが同じ最短経路長を返すことを確認する（実装の正しさ検証）
	verifyPathLength(start, rep, *g, matrix)
}

// verifyPathLength は全アルゴリズムが同一の最短経路長を返すことを確認する。
// 等長の別経路がありうるため経路のノード列ではなく経路長で照合する。
func verifyPathLength(start, goal int, g Graph, matrix [][]float64) {
	type result struct {
		name string
		dist float64
	}
	results := []result{
		{"Dijkstra(行列)", DijkstraMatrix(start, goal, matrix).Dist},
		{"Dijkstra(PQ)", DijkstraPQ(start, goal, g).Dist},
		{"Bellman-Ford", BellmanFord(start, goal, g).Dist},
		{"A*", Astar(start, goal, g).Dist},
		{"双方向Dijkstra", BiDijkstraPQ(start, goal, g).Dist},
	}

	fmt.Printf("=== 最短経路長の一致確認（始点=%d 終点=%d）===\n", start, goal)
	base := results[0].dist
	allMatch := true
	const eps = 1e-6 // 浮動小数の演算順序による微小差を許容する
	for _, r := range results {
		match := math.Abs(r.dist-base) < eps
		if !match {
			allMatch = false
		}
		fmt.Printf("%-16s 経路長=%.4f m  一致=%v\n", r.name, r.dist, match)
	}
	fmt.Printf("→ 全アルゴリズム一致: %v\n", allMatch)
}

// benchPair は1ペアについて5アルゴリズムを評価して表を出力する
func benchPair(start, goal int, g Graph, matrix [][]float64) {
	type algo struct {
		name string
		run  func() *Metrics
	}
	algos := []algo{
		{"Dijkstra(行列)", func() *Metrics { return DijkstraMatrix(start, goal, matrix) }},
		{"Dijkstra(PQ)", func() *Metrics { return DijkstraPQ(start, goal, g) }},
		{"Bellman-Ford", func() *Metrics { return BellmanFord(start, goal, g) }},
		{"A*", func() *Metrics { return Astar(start, goal, g) }},
		{"双方向Dijkstra", func() *Metrics { return BiDijkstraPQ(start, goal, g) }},
	}

	fmt.Printf("%-16s %12s %12s %12s %12s %14s %12s\n",
		"アルゴリズム", "緩和回数", "展開数", "push数", "最大heap", "中央値[μs]", "標準偏差")
	fmt.Println("------------------------------------------------------------------------------------------------------")
	for _, a := range algos {
		m := a.run() // 構造的指標は決定的なので1回

		// 実行時間はN回集めて中央値・標準偏差を求める（ns計測→µsに換算）
		us := make([]float64, benchN)
		for i := range us {
			us[i] = float64(a.run().ElapsedNs) / 1000.0
		}
		med, sd := medianStdev(us)

		fmt.Printf("%-16s %12d %12d %12d %12d %14.1f %12.1f\n",
			a.name, m.Relaxations, m.Expansions, m.PushCount, m.MaxHeapSize, med, sd)
	}
}

// fullDijkstra は始点から全ノードへの道路距離を返す（到達不能は +Inf）。
// 終点選定の到達可能性判定に使うため打ち切らず全探索する。
func fullDijkstra(start int, g Graph) []float64 {
	size := len(g.Nodes)
	dist := make([]float64, size)
	for i := range dist {
		dist[i] = math.Inf(1)
	}
	dist[start] = 0
	pq := Heap{[]Item{{start, 0}}}

	for len(pq.items) > 0 {
		cur := pq.pop()
		if cur.dist > dist[cur.id] {
			continue
		}
		for _, e := range g.Adj[cur.id] {
			if c := dist[cur.id] + e.Weight; c < dist[e.To] {
				dist[e.To] = c
				pq.push(Item{e.To, c})
			}
		}
	}
	return dist
}

// pickGoals は到達可能ノードを道路距離で昇順に並べ、k等分した各帯から1つずつ終点を選ぶ。
// 近〜遠が均等に散らばるため距離依存性の分布が作りやすい。
func pickGoals(start int, dist []float64, k int) []int {
	type nd struct {
		id   int
		dist float64
	}
	var reachable []nd
	for id, d := range dist {
		if id != start && !math.IsInf(d, 1) {
			reachable = append(reachable, nd{id, d})
		}
	}
	sort.Slice(reachable, func(i, j int) bool { return reachable[i].dist < reachable[j].dist })

	goals := make([]int, 0, k)
	n := len(reachable)
	for b := 0; b < k; b++ {
		// 帯bの代表として各区間の中央付近のノードを選ぶ
		idx := (2*b + 1) * n / (2 * k)
		goals = append(goals, reachable[idx].id)
	}
	return goals
}

// medianStdev はサンプルの中央値と標準偏差を返す
func medianStdev(xs []float64) (median, stdev float64) {
	n := len(xs)
	sort.Float64s(xs)
	if n%2 == 1 {
		median = xs[n/2]
	} else {
		median = (xs[n/2-1] + xs[n/2]) / 2
	}

	var sum float64
	for _, x := range xs {
		sum += x
	}
	mean := sum / float64(n)
	var sq float64
	for _, x := range xs {
		sq += (x - mean) * (x - mean)
	}
	stdev = math.Sqrt(sq / float64(n))
	return median, stdev
}

// countUndirectedEdges は無向エッジ数（有向の合計の半分）を返す
func countUndirectedEdges(g *Graph) int {
	total := 0
	for _, adj := range g.Adj {
		total += len(adj)
	}
	return total / 2
}
