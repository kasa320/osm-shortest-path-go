package main

import (
	"fmt"
	"math"
	"time"
)

// 関数：隣接リストを隣接行列に変換し、戻す
func ToMatrix(g Graph) [][]float64 {
	V := len(g.Adj)
	matrix := make([][]float64, len(g.Adj))
	for i := 0; i < V; i++ {
		matrix[i] = make([]float64, V)
		for _, edge := range g.Adj[i] { // 隣接リストでノードi を検索
			matrix[i][edge.To] = edge.Weight // 隣接行列に重みを書き込む
		}
	}
	return matrix
}

// 関数：<隣接行列>を用いてDijkstra法で最短経路を求める
func DijkstraMatrix(startID int, goalID int, matrix [][]float64) *Metrics {
	metrics := Metrics{}
	startTime := time.Now()

	// dist[], prev[], visited[] を初期化
	size := len(matrix)
	dist := make([]float64, size)
	prev := make([]int, size)
	visited := make([]bool, size)

	for i := 0; i < size; i++ {
		dist[i] = math.Inf(1)
		prev[i] = -1
		visited[i] = false
	}
	dist[startID] = 0 // 開始地点は距離0

	// 未訪問ノードのうち最小のものを選んで訪問する(未訪問がなくなるまで)
	for {
		selectedNode := -1
		minCost := math.Inf(1)

		for i := 0; i < size; i++ {
			if !visited[i] && (dist[i] < minCost) {
				selectedNode = i
				minCost = dist[i]
			}
		}
		if selectedNode == -1 { // 未訪問ノードがない場合
			break
		}
		if selectedNode == goalID { // ゴールに辿り着いたら終了
			break
		}
		visited[selectedNode] = true
		metrics.Expansions++

		// 見つかったノードのすべての隣接ノードについて、コストを更新
		for i := 0; i < size; i++ {
			if matrix[selectedNode][i] > 0 {
				cost := dist[selectedNode] + matrix[selectedNode][i]
				metrics.Relaxations++
				if cost < dist[i] {
					dist[i] = cost
					prev[i] = selectedNode
				}
			}
		}
	}

	elapsed := time.Since(startTime)
	metrics.ElapsedNs = elapsed.Nanoseconds()
	metrics.Dist = dist[goalID]

	return &metrics
}

// 関数：<優先度キュー>を用いたDijkstra法で最短経路を求める
func DijkstraPQ(startID int, goalID int, graph Graph) *Metrics {
	metrics := Metrics{}
	startTime := time.Now()

	// dist[], prev[]を初期化
	size := len(graph.Nodes)
	dist := make([]float64, size)
	prev := make([]int, size)

	for i := 0; i < size; i++ {
		dist[i] = math.Inf(1)
		prev[i] = -1
	}
	dist[startID] = 0                    // 開始地点は距離0
	pqueue := Heap{[]Item{{startID, 0}}} // 優先度付きキューを初期化（全地点ではなく開始地点のみで初期化している）

	// 未訪問ノードのうち最小のものを選んで訪問する(未訪問がなくなるまで)
	for len(pqueue.items) > 0 {
		selectedNode := pqueue.pop() // 最小のノード

		if selectedNode.dist > dist[selectedNode.id] { // 講義とは異なり重複pushなため古いエントリを捨てる
			continue
		}
		metrics.Expansions++

		if selectedNode.id == goalID { // ゴールに辿り着いたら終了
			break
		}

		// 見つかったノードのすべての隣接ノードについて、コストを更新 & PQに追加
		for _, dest := range graph.Adj[selectedNode.id] {
			cost := dist[selectedNode.id] + dest.Weight
			metrics.Relaxations++
			if cost < dist[dest.To] {
				dist[dest.To] = cost
				prev[dest.To] = selectedNode.id
				metrics.PushCount++
				pqueue.push(Item{dest.To, cost})
				metrics.MaxHeapSize = max(metrics.MaxHeapSize, len(pqueue.items))
			}
		}
	}

	elapsed := time.Since(startTime)
	metrics.ElapsedNs = elapsed.Nanoseconds()
	metrics.Dist = dist[goalID]

	return &metrics
}

// 戻り値のpathはゴールから遡っていることに注意
func ReconstructPath(dist []float64, prev []int, startID int, goalID int) ([]int, float64) {
	path := []int{}
	if dist[goalID] == math.Inf(1) {
		return []int{}, -1
	}
	for current := goalID; current != -1; current = prev[current] {
		path = append(path, current)
	}

	return path, dist[goalID]
}

func PrintPath(path []int, dist float64) {
	if len(path) <= 0 {
		fmt.Printf("最短経路が存在しません。")
	}

	fmt.Printf("最短経路（道のりコスト %.2f)\n", dist)

	for i := len(path) - 1; i >= 0; i-- {
		fmt.Printf("%d ", path[i])
		if i > 0 {
			fmt.Print("→")
		}
		if (i+1)%20 == 0 {
			fmt.Println("")
		}
	}

	fmt.Print("\n\n")
}
