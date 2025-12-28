//go:build ignore

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func main() {
	config := &gen.Config{
		// Package path cho generated code
		Package: "system/internal/ent/generated",

		// Target directory - generated code sẽ được đặt ở đây
		Target: "./generated",

		// Bật các features hữu ích
		Features: []gen.Feature{
			gen.FeaturePrivacy,      // Privacy layer cho access control
			gen.FeatureEntQL,        // EntQL filtering
			gen.FeatureSnapshot,     // Schema snapshots cho migrations
			gen.FeatureSchemaConfig, // Per-schema configuration
			gen.FeatureIntercept,    // Query interceptors
		},
	}

	if err := entc.Generate("./schema", config); err != nil {
		log.Fatal("running ent codegen:", err)
	}
}
