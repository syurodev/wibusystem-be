package main

import (
	"fmt"

	"github.com/gosimple/slug"
)

func main() {
	titles := []string{
		"Tiếng Việt có dấu",
		"日本語のタイトル",
		"English Title",
		"Mixed English and 日本語",
	}

	for _, t := range titles {
		fmt.Printf("Original: %s -> Slug: %s\n", t, slug.Make(t))
	}
}
