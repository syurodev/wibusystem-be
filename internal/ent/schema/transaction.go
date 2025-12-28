package schema

import (
	"encoding/json"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
	"github.com/shopspring/decimal"
)

// Transaction holds the schema definition for the Transaction entity.
type Transaction struct {
	ent.Schema
}

// Annotations of the Transaction.
func (Transaction) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "transactions", Schema: "payment"},
	}
}

// Fields of the Transaction.
func (Transaction) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("user_id", uuid.UUID{}),
		field.Enum("type").
			Values("topup", "purchase_chapter", "purchase_series", "rental", "subscription", "refund", "admin_adjustment"),
		field.Other("coin_amount", decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "numeric(20,2)",
			}),
		field.Other("vnd_amount", decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "numeric(20,2)",
			}).
			Optional().
			Nillable(),
		field.Other("balance_after", decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "numeric(20,2)",
			}),
		field.String("reference_type").
			Optional().
			Nillable(),
		field.UUID("reference_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("creator_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Other("creator_revenue_vnd", decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "numeric(20,2)",
			}).
			Optional().
			Nillable(),
		field.Other("platform_revenue_vnd", decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "numeric(20,2)",
			}).
			Optional().
			Nillable(),
		field.String("description").
			Optional().
			Nillable(),
		field.JSON("metadata", json.RawMessage{}).
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Transaction.
func (Transaction) Edges() []ent.Edge {
	return nil
}
