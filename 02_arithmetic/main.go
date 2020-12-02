package main

import (
	"fmt"
	"math"
)

func main() {
	// a := 15
	// b := 25
	var a int
	var b int
	fmt.Println("Enter two Integers separted by space as the input: ")
	_, err := fmt.Scanf("%d %d", &a, &b)

	if err != nil {
		fmt.Println(err)
	} else {
		sum := a + b
		sub := a - b
		mul := a * b
		div := float64(a / b)
		squareRoot := math.Sqrt(float64(a * b))

		fmt.Println(sum)
		fmt.Println(sub)
		fmt.Println(mul)
		fmt.Println(div)
		fmt.Printf("Square Root of %d * %d = %f", a, b, squareRoot)
	}
}
