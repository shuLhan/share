// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ints64

import (
	"fmt"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

var (
	d = [][]int64{
		{},
		{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		{0, 1, 0, 1, 0, 1, 0, 1, 0},
		{1, 1, 2, 2, 3, 1, 2},
	}
	dSorted = [][]int64{
		{},
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		{0, 0, 0, 0, 0, 1, 1, 1, 1},
		{1, 1, 1, 2, 2, 2, 3},
	}
	dSortedDesc = [][]int64{
		{},
		{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		{1, 1, 1, 1, 0, 0, 0, 0, 0},
		{3, 2, 2, 2, 1, 1, 1},
	}
)

func TestMaxEmpty(t *testing.T) {
	gotv, goti, gotok := Max(d[0])

	test.Assert(t, "", int64(0), gotv)
	test.Assert(t, "", 0, goti)
	test.Assert(t, "", false, gotok)
}

func TestMax(t *testing.T) {
	gotv, goti, gotok := Max(d[1])

	test.Assert(t, "", int64(9), gotv)
	test.Assert(t, "", 4, goti)
	test.Assert(t, "", true, gotok)
}

func TestMinEmpty(t *testing.T) {
	gotv, goti, gotok := Min(d[0])

	test.Assert(t, "", int64(0), gotv)
	test.Assert(t, "", 0, goti)
	test.Assert(t, "", false, gotok)
}

func TestMin(t *testing.T) {
	gotv, goti, gotok := Min(d[1])

	test.Assert(t, "", int64(0), gotv)
	test.Assert(t, "", 5, goti)
	test.Assert(t, "", true, gotok)
}

func TestSum(t *testing.T) {
	got := Sum(d[1])

	test.Assert(t, "", int64(45), got)
}

func TestCount(t *testing.T) {
	got := Count(d[0], 0)

	test.Assert(t, "", 0, got)

	got = Count(d[1], 1)

	test.Assert(t, "", 1, got)

	got = Count(d[2], 1)

	test.Assert(t, "", 4, got)

	got = Count(d[3], 0)

	test.Assert(t, "", 0, got)

	got = Count(d[3], 3)

	test.Assert(t, "", 1, got)
}

func TestCountsEmpty(t *testing.T) {
	classes := []int64{1, 2, 3}
	exp := []int{0, 0, 0}

	got := Counts(d[0], classes)

	test.Assert(t, "", exp, got)
}

func TestCountsEmptyClasses(t *testing.T) {
	classes := []int64{}
	var exp []int

	got := Counts(d[1], classes)

	test.Assert(t, "", exp, got)
}

func TestCounts(t *testing.T) {
	classes := []int64{1, 2, 3}
	exp := []int{3, 3, 1}

	got := Counts(d[3], classes)

	test.Assert(t, "", exp, got)
}

func Test64MaxCountOf(t *testing.T) {
	classes := []int64{0, 1}
	exp := int64(0)
	got, _ := MaxCountOf(d[2], classes)

	test.Assert(t, "", exp, got)

	// Swap the class values.
	classes = []int64{1, 0}
	got, _ = MaxCountOf(d[2], classes)

	test.Assert(t, "", exp, got)
}

func TestSwapEmpty(t *testing.T) {
	exp := []int64{}

	Swap(d[0], 1, 6)

	test.Assert(t, "", exp, d[0])
}

func TestSwapEqual(t *testing.T) {
	in := make([]int64, len(d[1]))
	copy(in, d[1])

	exp := make([]int64, len(in))
	copy(exp, in)

	Swap(in, 1, 1)

	test.Assert(t, "", exp, in)
}

func TestSwapOutOfRange(t *testing.T) {
	in := make([]int64, len(d[1]))
	copy(in, d[1])

	exp := make([]int64, len(in))
	copy(exp, in)

	Swap(in, 1, 100)

	test.Assert(t, "", exp, in)
}

func TestSwap(t *testing.T) {
	in := make([]int64, len(d[1]))
	copy(in, d[1])

	exp := make([]int64, len(in))
	copy(exp, in)

	Swap(in, 0, len(in)-1)

	tmp := exp[0]
	exp[0] = exp[len(exp)-1]
	exp[len(exp)-1] = tmp

	test.Assert(t, "", exp, in)
}

func TestIsExist(t *testing.T) {
	var s bool

	// True positive.
	for _, d := range d {
		for _, v := range d {
			s = IsExist(d, v)

			test.Assert(t, "", true, s)
		}
	}

	// False positive.
	for _, d := range d {
		s = IsExist(d, -1)
		test.Assert(t, "", false, s)
		s = IsExist(d, 10)
		test.Assert(t, "", false, s)
	}
}

func TestInsertionSortWithIndices(t *testing.T) {
	for x := range d {
		data := make([]int64, len(d[x]))

		copy(data, d[x])

		ids := make([]int, len(data))
		for x := range ids {
			ids[x] = x
		}

		InsertionSortWithIndices(data, ids, 0, len(ids), true)

		test.Assert(t, "", dSorted[x], data)
	}
}

func TestInsertionSortWithIndicesDesc(t *testing.T) {
	for x := range d {
		data := make([]int64, len(d[x]))

		copy(data, d[x])

		ids := make([]int, len(data))
		for x := range ids {
			ids[x] = x
		}

		InsertionSortWithIndices(data, ids, 0, len(ids), false)

		test.Assert(t, "", dSortedDesc[x], data)
	}
}

func TestSortByIndex(t *testing.T) {
	ids := [][]int{
		{},
		{5, 6, 7, 8, 9, 0, 1, 2, 3, 4},
		{0, 2, 4, 6, 8, 1, 3, 5, 7},
		{0, 1, 5, 6, 2, 3, 4},
	}

	for x := range d {
		data := make([]int64, len(d[x]))

		copy(data, d[x])

		SortByIndex(&data, ids[x])

		test.Assert(t, "", dSorted[x], data)
	}
}

var ints64InSorts = [][]int64{
	{9, 8, 7, 6, 5, 4, 3},
	{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	{0, 6, 7, 8, 5, 1, 2, 3, 4, 9},
	{9, 8, 7, 6, 5, 4, 3, 2, 1},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{51, 50, 56, 55, 55, 58, 55, 55, 58, 56,
		57, 50, 56, 59, 62, 60, 49, 63, 61, 56,
		58, 67, 61, 59, 60, 49, 56, 52, 61, 64,
		70, 57, 65, 69, 57, 64, 62, 66, 63, 62,
		54, 67, 61, 57, 55, 60, 30, 66, 57, 60,
		68, 60, 61, 63, 58, 58, 56, 57, 60, 69,
		69, 64, 63, 63, 67, 65, 58, 63, 64, 67,
		59, 72, 63, 63, 65, 71, 67, 76, 73, 64,
		67, 74, 60, 68, 65, 64, 67, 64, 65, 69,
		77, 67, 72, 77, 72, 77, 61, 79, 77, 68,
		62},
}

var ints64ExpSorts = [][]int64{
	{3, 4, 5, 6, 7, 8, 9},
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
	{1, 2, 3, 4, 5, 6, 7, 8, 9},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{30, 49, 49, 50, 50, 51, 52, 54, 55, 55,
		55, 55, 55, 56, 56, 56, 56, 56, 56, 57,
		57, 57, 57, 57, 57, 58, 58, 58, 58, 58,
		58, 59, 59, 59, 60, 60, 60, 60, 60, 60,
		60, 61, 61, 61, 61, 61, 61, 62, 62, 62,
		62, 63, 63, 63, 63, 63, 63, 63, 63, 64,
		64, 64, 64, 64, 64, 64, 65, 65, 65, 65,
		65, 66, 66, 67, 67, 67, 67, 67, 67, 67,
		67, 68, 68, 68, 69, 69, 69, 69, 70, 71,
		72, 72, 72, 73, 74, 76, 77, 77, 77, 77,
		79},
}

var ints64ExpSortsDesc = [][]int64{
	{9, 8, 7, 6, 5, 4, 3},
	{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
	{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
	{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
	{9, 8, 7, 6, 5, 4, 3, 2, 1},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{79, 77, 77, 77, 77, 76, 74, 73, 72, 72,
		72, 71, 70, 69, 69, 69, 69, 68, 68, 68,
		67, 67, 67, 67, 67, 67, 67, 67, 66, 66,
		65, 65, 65, 65, 65, 64, 64, 64, 64, 64,
		64, 64, 63, 63, 63, 63, 63, 63, 63, 63,
		62, 62, 62, 62, 61, 61, 61, 61, 61, 61,
		60, 60, 60, 60, 60, 60, 60, 59, 59, 59,
		58, 58, 58, 58, 58, 58, 57, 57, 57, 57,
		57, 57, 56, 56, 56, 56, 56, 56, 55, 55,
		55, 55, 55, 54, 52, 51, 50, 50, 49, 49,
		30},
}

func TestIndirectSort(t *testing.T) {
	var res, exp string

	for i := range ints64InSorts {
		IndirectSort(ints64InSorts[i], true)

		res = fmt.Sprint(ints64InSorts[i])
		exp = fmt.Sprint(ints64ExpSorts[i])

		test.Assert(t, "", exp, res)
	}
}

func TestIndirectSortDesc(t *testing.T) {
	var res, exp string

	for i := range ints64InSorts {
		IndirectSort(ints64InSorts[i], false)

		res = fmt.Sprint(ints64InSorts[i])
		exp = fmt.Sprint(ints64ExpSortsDesc[i])

		test.Assert(t, "", exp, res)
	}
}

func TestIndirectSort_Stability(t *testing.T) {
	exp := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	got := IndirectSort(ints64InSorts[5], true)

	test.Assert(t, "", exp, got)
}

func TestIndirectSortDesc_Stability(t *testing.T) {
	exp := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	got := IndirectSort(ints64InSorts[5], false)

	test.Assert(t, "", exp, got)
}

func TestInplaceMergesort(t *testing.T) {
	size := len(ints64InSorts[6])
	idx := make([]int, size)

	InplaceMergesort(ints64InSorts[6], idx, 0, size, true)

	test.Assert(t, "", ints64ExpSorts[6], ints64InSorts[6])
}

func TestIndirectSort_SortByIndex(t *testing.T) {
	expListID := []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	in1 := []int64{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	in2 := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	exp := fmt.Sprint(in1)

	sortedListID := IndirectSort(in1, true)

	test.Assert(t, "", expListID, sortedListID)

	// Reverse the sort.
	SortByIndex(&in2, sortedListID)

	got := fmt.Sprint(in2)

	test.Assert(t, "", exp, got)
}
