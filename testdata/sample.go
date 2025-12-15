package sample

func example(x int) int {
	if x > 0 {
		x = x * 2
	} else {
		x = x + 1
	}
	return x
}

func loop(n int) int {
	sum := 0
	for i := 0; i < n; i++ {
		sum += i
	}
	return sum
}
