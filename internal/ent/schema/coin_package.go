package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
	"github.com/shopspring/decimal"
)

// CoinPackage holds the schema definition for the CoinPackage entity.
type CoinPackage struct {
	ent.Schema
}

// Annotations of the CoinPackage.
func (CoinPackage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "coin_packages", Schema: "payment"},
	}
}

// Fields of the CoinPackage.
func (CoinPackage) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.String("name").
			NotEmpty(),
		field.String("slug").
			Unique().
			NotEmpty(),
		field.Other("coin_amount", decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "numeric(20,2)",
			}),
		field.Other("price_vnd", decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "numeric(20,2)",
			}),
		field.Int("bonus_percent").
			Default(0),
		field.Bool("is_popular").
			Default(false),
		field.Bool("is_active").
			Default(true),
		field.Int("display_order").
			Default(0),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the CoinPackage.
func (CoinPackage) Edges() []ent.Edge {
	return nil
}
