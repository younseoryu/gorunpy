//go:generate gorunpy build py -o .
//go:generate go run github.com/younseoryu/gorunpy/cmd/gorunpy-gen -binary py -package main -output client.go

package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/younseoryu/gorunpy/gorunpy"
)

func main() {
	client := NewClient("./py")
	ctx := context.Background()

	sum, _ := client.Sum(ctx, 1, 2)
	fmt.Printf("1 + 2 = %d\n", sum)

	greeting, _ := client.Greet(ctx, nil, "World")
	fmt.Println(greeting)

	stats, _ := client.GetStats(ctx, []float64{1, 2, 3, 4, 5})
	fmt.Printf("mean = %.1f\n", stats["mean"])

	_, err := client.Divide(ctx, 10, 0)
	if err != nil {
		var e *gorunpy.ErrInvalidInput
		if errors.As(err, &e) {
			fmt.Printf("Error: %s\n", e.Message)
		}
	}
}
