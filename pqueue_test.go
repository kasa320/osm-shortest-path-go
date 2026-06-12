package main

import (
	"fmt"
	"testing"
)

func TestHeap(t *testing.T) {
	array := []Item{
		{7, 0},
		{9, 0},
		{12, 0},
		{7, 0},
		{15, 0},
		{20, 0},
		{17, 0},
	}
	heap := Heap{array}
	heap.push(Item{5, 0})
	fmt.Println("items:", heap.items)
}
