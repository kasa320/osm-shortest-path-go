package main

import (
	"math"
	"time"
)

// Bellman-Ford法を用いて最短経路を求める
func BellmanFord(startID int, goalID int, graph Graph) *Metrics {
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

	for {
		updated := false
		// すべての辺(edge)について最小コストを更新できるなら更新する
		for i := 0; i < size; i++ {
			for _, edge := range graph.Adj[i] {
				cost := dist[i] + edge.Weight
				metrics.Relaxations++
				if cost < dist[edge.To] {
					dist[edge.To] = cost
					prev[edge.To] = i
					updated = true
				}
			}
		}
		if !updated {
			break
		}
	}
	elapsed := time.Since(startTime)
	metrics.ElapsedNs = elapsed.Nanoseconds()
	metrics.Dist = dist[goalID]

	return &metrics
}
