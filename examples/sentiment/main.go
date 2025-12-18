//go:generate ../../.gorunpy/venv/bin/gorunpy

package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := NewPylibClient()

	texts := []string{
		"I love this product! It's amazing!",
		"This is terrible, worst purchase ever.",
		"It's okay, nothing special.",
	}

	fmt.Println("=== Sentiment Analysis ===\n")

	for _, text := range texts {
		result, err := client.Analyze(ctx, text)
		if err != nil {
			log.Fatalf("Failed to analyze: %v", err)
		}
		fmt.Printf("Text: %q\n", text)
		fmt.Printf("  → %s (%.2f%%)\n\n", result["label"], result["score"].(float64)*100)
	}
}

