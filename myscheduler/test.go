package main

import "sort"

const mod int = 1e9 + 7

func nthMagicalNumber(n int, a int, b int) int {
	right := n * (a + b)
	lcm := a * b / gcd(a, b)
	return sort.Search(right, func(i int) bool {
		return i/a+i/b-i/lcm > n
	})
}

func gcd(a, b int) int {
	if b == 0 {
		return a
	} else {
		return gcd(b, a%b)
	}
}
