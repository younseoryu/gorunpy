package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/younseoryu/gorunpy/example/go/mathlib"
	"github.com/younseoryu/gorunpy/gorunpy"
)

func main() {
	client := mathlib.NewClient("../dist/mathlib")
	ctx := context.Background()

	sum, _ := client.Sum(ctx, 1, 2)
	fmt.Printf("1 + 2 = %d\n", sum)

	product, _ := client.Multiply(ctx, 3.5, 2.0)
	fmt.Printf("3.5 * 2.0 = %.1f\n", product)

	greeting, _ := client.Greet(ctx, nil, "World")
	fmt.Println(greeting)

	stats, _ := client.GetStats(ctx, []float64{1, 2, 3, 4, 5})
	fmt.Printf("mean([1,2,3,4,5]) = %.1f\n", stats["mean"])

	_, err := client.Divide(ctx, 10, 0)
	if err != nil {
		var e *gorunpy.ErrInvalidInput
		if errors.As(err, &e) {
			fmt.Printf("Error: %s\n", e.Message)
		} else {
			log.Fatal(err)
		}
	}
}
