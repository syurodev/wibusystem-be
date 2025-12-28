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

// UserWallet holds the schema definition for the UserWallet entity.
type UserWallet struct {
	ent.Schema
}

// Annotations of the UserWallet.
func (UserWallet) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user_wallets", Schema: "payment"},
	}
}

// Fields of the UserWallet.
func (UserWallet) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("user_id", uuid.UUID{}).
			Unique(),
		field.Other("coin_balance", decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "numeric(20,2)",
			}).
			Default(decimal.Zero),
		field.Other("total_deposited", decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "numeric(20,2)",
			}).
			Default(decimal.Zero),
		field.Other("total_spent", decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "numeric(20,2)",
			}).
			Default(decimal.Zero),
		field.Other("total_subscription_spent", decimal.Decimal{}).
			SchemaType(map[string]string{
				"postgres": "numeric(20,2)",
			}).
			Default(decimal.Zero),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the UserWallet.
func (UserWallet) Edges() []ent.Edge {
	return nil
}
