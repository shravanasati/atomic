package internal

import "testing"

type addTest struct {
	args     []int
	expected int
}

func add(args ...int) int {
	sum := 0
	for _, a := range args {
		sum += a
	}
	return sum
}


func Test_Add(t *testing.T) {
	tests := []addTest{
		{[]int{1, 2, 3},             6},
		{[]int{1, 2, 3, 4},          10},
		{[]int{1, 2, 3, 4, 5},       15},
		{[]int{1, 2, 3, 4, 5, 6},    21},
		{[]int{1, 2, 3, 4, 5, 6, 7}, 28},
	}

	for _, test := range tests {
		if actual := add(test.args...); actual != test.expected {
			t.Errorf("Add(%v) == %v, expected %v", test.args, actual, test.expected)
		}
	}
}
