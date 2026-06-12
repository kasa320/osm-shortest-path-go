package main

import (
	"math"
)

type Item struct {
	id   int
	dist float64
}

type Heap struct {
	items []Item
}

// 最小ヒープ構造を保ったまま、要素を追加する
func (h *Heap) push(item Item) {
	h.items = append(h.items, item)
	index := len(h.items) - 1
	parent := (index - 1) / 2

	// 親 > 子であればswapする→次の親を見る→根までいけば終了。
	for index > 0 {
		if h.items[parent].dist > h.items[index].dist {
			h.items[parent], h.items[index] = h.items[index], h.items[parent] // swap(parent, index)
			index = parent
			parent = (index - 1) / 2
		} else {
			break
		}
	}
}

// 最小ヒープ構造を保ったまま初めの要素(最小値)を取り出す
func (h *Heap) pop() Item {
	min := h.items[0]

	// 最小値と末尾を交換し、末尾を1縮める
	n := len(h.items)
	h.items[0], h.items[n-1] = h.items[n-1], h.items[0]
	h.items = h.items[0 : n-1]
	n-- // 縮めた分デクリメント

	// ヒープを再構成するために、rootを下降させる
	index := 0
	left, right := 1, 2

	// 親 > 子であればswapする→次の子を見る→末尾まで行けば終了
	for left < n { // 左の子が配列内であれば下降し続ける
		var leftNum, rightNum Item
		var toBeSwapped int

		if left < n {
			leftNum = h.items[left]
		} else {
			leftNum = Item{-1, math.Inf(1)}
		}
		if right < n {
			rightNum = h.items[right]
		} else {
			rightNum = Item{-1, math.Inf(1)}
		}
		if leftNum.dist < rightNum.dist {
			toBeSwapped = left
		} else {
			toBeSwapped = right
		}

		// 親の方が大きければ、親と小さい方の子をswap
		if h.items[index].dist > h.items[toBeSwapped].dist {
			h.items[index], h.items[toBeSwapped] = h.items[toBeSwapped], h.items[index]
			index = toBeSwapped
			left = 2*index + 1
			right = left + 1
		} else {
			break
		}
	}

	return min
}

func (h *Heap) top() Item {
	return h.items[0]
}
