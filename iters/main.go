package main

import (
	"fmt"
	"iter"
)

type fibonacci struct {
	a   int
	b   int
	Max int
}

func (f *fibonacci) Sequence() iter.Seq[int] {
	return func(yield func(int) bool) {
		ok := true
		for f.a < f.Max && ok {
			ok = yield(f.a)
			f.a, f.b = f.b, f.a+f.b
		}
	}
}

func NewFibonacci(max int) *fibonacci {
	return &fibonacci{
		a:   0,
		b:   1,
		Max: max,
	}
}

func main() {
	fib := NewFibonacci(500)
	// simple case
	for v := range fib.Sequence() {
		fmt.Println(v)
	}

	// resuming iteration
	// for v := range fib.Sequence() {
	// 	fmt.Println(v)
	// 	if v > 35 {
	// 		break
	// 	}
	// }
	// fmt.Printf("\nbreak\n\n")
	// for v := range fib.Sequence() {
	// 	fmt.Println(v)
	// }
}
