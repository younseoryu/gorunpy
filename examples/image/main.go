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
		fmt.Println("Usage: image-example <image-file>")
		os.Exit(1)
	}

	imagePath := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := NewPylibClient()

	// Get image info
	info, err := client.GetInfo(ctx, imagePath)
	if err != nil {
		log.Fatalf("Failed to get info: %v", err)
	}

	fmt.Println("=== Image Info ===")
	fmt.Printf("Format: %v\n", info["format"])
	fmt.Printf("Size: %v x %v\n", info["width"], info["height"])
	fmt.Printf("Mode: %v\n", info["mode"])

	// Create thumbnail
	thumbPath := "thumbnail.png"
	_, err = client.Thumbnail(ctx, imagePath, 200, thumbPath)
	if err != nil {
		log.Fatalf("Failed to create thumbnail: %v", err)
	}
	fmt.Printf("\nCreated thumbnail: %s\n", thumbPath)
}

