// グラフの構造体を定義する。
package main

import "math"

const earthRadiusM = 6371000.0

type Node struct {
	ID  int
	Lat float64 // 緯度
	Lon float64 // 経度
}

type Edge struct {
	To     int
	Weight float64
}

type Graph struct {
	Nodes []Node
	Adj   [][]Edge
}

// 2ノードの座標から直線距離を求める
func (g *Graph) distance(id1 int, id2 int) float64 {
	// ある２点が向かい合うように作った長方形の緯線方向の長さをdx, 経線方向の長さをdyとする
	lat1, lat2 := g.Nodes[id1].Lat, g.Nodes[id2].Lat
	latDiff := math.Abs(lat1 - lat2)
	latAvg := (lat1 + lat2) / 2

	lon1, lon2 := g.Nodes[id1].Lon, g.Nodes[id2].Lon
	lonDiff := math.Abs(lon1 - lon2)

	// 地球の断面を切った扇形における弧の長さを計算
	dy := 2 * math.Pi * earthRadiusM * latDiff / 360
	dx := 2 * math.Pi * earthRadiusM * lonDiff / 360 * math.Cos(latAvg*math.Pi/180)

	return math.Sqrt(dx*dx + dy*dy)
}
