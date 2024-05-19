package myersdiff

import (
	"strconv"
)

type Myers interface {
	GenerateDiffScript(src, dst []string) Diff
}

type myers struct {
}

func NewMyersDiffCalculator() Myers {
	return myers{}
}

func (myers myers) GenerateDiffScript(src, dst []string) Diff {
	script := myers.shortestEditScript(src, dst)
	srcIndex, dstIndex := 0, 0

	var editString string
	insertedLineCount := 0
	deltaScript := make([]string, 0)

	for _, op := range script {
		switch op {
		case INSERT:
			editString += "i" + strconv.Itoa(dstIndex) + "-" + strconv.Itoa(insertedLineCount) + "$"
			deltaScript = append(deltaScript, dst[dstIndex])
			dstIndex += 1
			insertedLineCount++

		case MOVE:
			srcIndex += 1
			dstIndex += 1

		case DELETE:
			editString += "d" + strconv.Itoa(srcIndex) + "$"
			srcIndex += 1
		}
	}

	return Diff{Commands: editString, Data: deltaScript}
}

func (myers myers) shortestEditScript(src, dst []string) []operation {
	n := len(src)
	m := len(dst)
	max := n + m
	var trace []map[int]int
	var x, y int

loop:
	for d := 0; d <= max; d++ {
		v := make(map[int]int, d+2)
		trace = append(trace, v)

		if d == 0 {
			t := 0
			for len(src) > t && len(dst) > t && src[t] == dst[t] {
				t++
			}
			v[0] = t
			if t == len(src) && t == len(dst) {
				break loop
			}
			continue
		}

		lastV := trace[d-1]

		for k := -d; k <= d; k += 2 {
			if k == -d || (k != d && lastV[k-1] < lastV[k+1]) {
				x = lastV[k+1]
			} else {
				x = lastV[k-1] + 1
			}

			y = x - k

			for x < n && y < m && src[x] == dst[y] {
				x, y = x+1, y+1
			}

			v[k] = x

			if x == n && y == m {
				break loop
			}
		}
	}

	script := myers.backtrace(n, m, trace)
	return myers.reverse(script)
}

func (myers myers) backtrace(srcLen int, destLen int, trace []map[int]int) []operation {
	var script []operation

	x := srcLen
	y := destLen
	var k, prevK, prevX, prevY int

	for d := len(trace) - 1; d > 0; d-- {
		k = x - y
		lastV := trace[d-1]

		if k == -d || (k != d && lastV[k-1] < lastV[k+1]) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}

		prevX = lastV[prevK]
		prevY = prevX - prevK

		for x > prevX && y > prevY {
			script = append(script, MOVE)
			x -= 1
			y -= 1
		}

		if x == prevX {
			script = append(script, INSERT)
		} else {
			script = append(script, DELETE)
		}

		x, y = prevX, prevY
	}

	if trace[0][0] != 0 {
		for i := 0; i < trace[0][0]; i++ {
			script = append(script, MOVE)
		}
	}
	return script
}

func (myers myers) reverse(s []operation) []operation {
	result := make([]operation, len(s))

	for i, v := range s {
		result[len(s)-1-i] = v
	}

	return result
}
