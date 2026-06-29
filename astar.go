package main

import (
	"math"
	"time"
)

// 関数：A*法で最短経路を求める
func Astar(startID int, goalID int, graph Graph) *Metrics {
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
	dist[startID] = 0 // 開始地点は距離0

	// 優先度付きキューにいれる値はdist + 推定距離h（今回は直線距離）となる。
	pqueue := Heap{[]Item{{startID, graph.distance(startID, goalID)}}}

	// 未訪問ノードのうち最小のものを選んで訪問する(未訪問がなくなるまで)
	for len(pqueue.items) > 0 {
		selectedNode := pqueue.pop() // 最小のノード
		h := graph.distance(selectedNode.id, goalID)
		if selectedNode.dist > dist[selectedNode.id]+h {
			continue
		}
		metrics.Expansions++
		if selectedNode.id == goalID {
			break
		}

		// 見つかったノードのすべての隣接ノードについて、コストを更新 & PQに追加
		for _, dest := range graph.Adj[selectedNode.id] {
			cost := dist[selectedNode.id] + dest.Weight
			metrics.Relaxations++
			if cost < dist[dest.To] {
				dist[dest.To] = cost
				prev[dest.To] = selectedNode.id
				h = graph.distance(dest.To, goalID)
				metrics.PushCount++
				pqueue.push(Item{dest.To, cost + h})
				metrics.MaxHeapSize = max(metrics.MaxHeapSize, len(pqueue.items))
			}
		}
	}

	elapsed := time.Since(startTime)
	metrics.ElapsedNs = elapsed.Nanoseconds()
	metrics.Dist = dist[goalID]

	return &metrics
}
