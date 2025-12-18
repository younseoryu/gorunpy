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
	if len(os.Args) < 2 {
		fmt.Println("Usage: docling-example <pdf-file>")
		os.Exit(1)
	}

	pdfPath := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	client := NewPylibClient()

	fmt.Printf("Converting %s to Markdown...\n\n", pdfPath)

	markdown, err := client.PdfToMarkdown(ctx, pdfPath)
	if err != nil {
		log.Fatalf("Failed to convert: %v", err)
	}

	fmt.Println("=== Markdown Output ===")
	fmt.Println(markdown)
}

