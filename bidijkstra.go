package main

import (
	"math"
	"time"
)

type Metrics struct {
	Relaxations int     // 緩和回数
	Expansions  int     // ノード展開数
	PushCount   int     // PQ push数
	MaxHeapSize int     // 最大ヒープサイズ
	ElapsedNs   int64   // 実行時間(ns)
	Dist        float64 // 始点から終点までの最短経路長（実装の正しさ検証用）
}

// 関数：優先度キューを用いた双方向Dijkstra法で最短経路を求める
func BiDijkstraPQ(startID int, goalID int, graph Graph) *Metrics {
	// metricsを初期化
	metrics := Metrics{}
	startTime := time.Now()

	// dist[], prev[]を初期化
	size := len(graph.Nodes)

	dist1 := make([]float64, size) // start -> goal
	dist2 := make([]float64, size) // goal -> start
	prev1 := make([]int, size)     // start -> goal
	prev2 := make([]int, size)     // goal -> start

	for i := 0; i < size; i++ {
		dist1[i], dist2[i] = math.Inf(1), math.Inf(1)
		prev1[i], prev2[i] = -1, -1
	}
	dist1[startID], dist2[goalID] = 0, 0 // 開始地点は距離0
	pqueue1, pqueue2 := Heap{[]Item{{startID, 0}}}, Heap{[]Item{{goalID, 0}}}

	mu := math.Inf(1)
	// 未訪問ノードのうち最小のものを選んで訪問する(未訪問がなくなるまで)
	for len(pqueue1.items) > 0 && len(pqueue2.items) > 0 {
		if pqueue1.top().dist+pqueue2.top().dist >= mu { // 停止条件
			break
		}

		// ===== pqueue1 =====
		selectedNode := pqueue1.pop() // 最小のノード
		if selectedNode.dist <= dist1[selectedNode.id] {
			metrics.Expansions++

			// 見つかったノードのすべての隣接ノードについて、コストを更新 & PQに追加
			for _, dest := range graph.Adj[selectedNode.id] {
				cost := dist1[selectedNode.id] + dest.Weight
				metrics.Relaxations++
				if cost < dist1[dest.To] {
					dist1[dest.To] = cost
					prev1[dest.To] = selectedNode.id

					metrics.PushCount++
					pqueue1.push(Item{dest.To, cost})
					metrics.MaxHeapSize = max(metrics.MaxHeapSize, len(pqueue1.items))

					// muを更新
					if dist1[dest.To]+dist2[dest.To] < mu {
						mu = dist1[dest.To] + dist2[dest.To]
					}
				}
			}
		}

		// ===== pqueue2 =====
		selectedNode = pqueue2.pop() // 最小のノード
		if selectedNode.dist <= dist2[selectedNode.id] {
			metrics.Expansions++

			// 見つかったノードのすべての隣接ノードについて、コストを更新 & PQに追加
			for _, dest := range graph.Adj[selectedNode.id] {
				cost := dist2[selectedNode.id] + dest.Weight
				metrics.Relaxations++
				if cost < dist2[dest.To] {
					dist2[dest.To] = cost
					prev2[dest.To] = selectedNode.id
					metrics.PushCount++
					pqueue2.push(Item{dest.To, cost})
					metrics.MaxHeapSize = max(metrics.MaxHeapSize, len(pqueue2.items))

					if dist1[dest.To]+dist2[dest.To] < mu {
						mu = dist1[dest.To] + dist2[dest.To]
					}
				}
			}
		}
	}

	elapsed := time.Since(startTime)
	metrics.ElapsedNs = elapsed.Nanoseconds()
	metrics.Dist = mu // 最短経路長（両方向の確定距離の和）

	return &metrics
}
