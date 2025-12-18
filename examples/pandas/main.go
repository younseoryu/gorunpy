//go:generate ../../.gorunpy/venv/bin/gorunpy

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	csvPath := "data.csv"
	if len(os.Args) >= 2 {
		csvPath = os.Args[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := NewPylibClient()

	fmt.Printf("Analyzing %s...\n\n", csvPath)

	// Properly typed: returns map[string]float64
	avgSalary, err := client.AnalyzeWithAggregation(ctx, csvPath, "department", "salary")
	if err != nil {
		log.Fatalf("Failed to analyze: %v", err)
	}

	fmt.Println("=== Average Salary by Department ===")
	for dept, avg := range avgSalary {
		fmt.Printf("  %s: $%.0f\n", dept, avg)
	}
}

