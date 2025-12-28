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

// TopupOrder holds the schema definition for the TopupOrder entity.
type TopupOrder struct {
	ent.Schema
}

// Annotations of the TopupOrder.
func (TopupOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "topup_orders", Schema: "payment"},
	}
}

// Fields of the TopupOrder.
func (TopupOrder) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("package_id", uuid.UUID{}),
		field.String("order_code").
			Unique().
			NotEmpty(),
		field.Other("coin_amount", decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "numeric(20,2)",
			}),
		field.Other("base_coin_amount", decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "numeric(20,2)",
			}),
		field.Other("bonus_coin_amount", decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "numeric(20,2)",
			}),
		field.Other("vnd_amount", decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "numeric(20,2)",
			}),
		field.Enum("status").
			Values("pending", "success", "expired", "cancelled", "failed").
			Default("pending"),
		field.String("sepay_transaction_id").
			Optional().
			Nillable(),
		field.String("sepay_content").
			Optional().
			Nillable(),
		field.String("bank_name").
			Optional().
			Nillable(),
		field.String("bank_account").
			Optional().
			Nillable(),
		field.String("account_name").
			Optional().
			Nillable(),
		field.Time("completed_at").
			Optional().
			Nillable(),
		field.Time("expired_at"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the TopupOrder.
func (TopupOrder) Edges() []ent.Edge {
	return nil
}
