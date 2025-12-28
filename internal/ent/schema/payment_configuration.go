package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gofrs/uuid/v5"
)

// PaymentConfiguration holds the schema definition for the PaymentConfiguration entity.
type PaymentConfiguration struct {
	ent.Schema
}

// Annotations of the PaymentConfiguration.
func (PaymentConfiguration) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "payment_configurations", Schema: "payment"},
	}
}

// Fields of the PaymentConfiguration.
func (PaymentConfiguration) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.String("key").
			Unique().
			NotEmpty(),
		field.String("value").
			Default(""),
		field.Enum("value_type").
			Values("string", "number", "boolean", "json").
			Default("string"),
		field.String("description").
			Optional().
			Nillable(),
		field.Bool("is_sensitive").
			Default(false),
		field.UUID("updated_by", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Indexes of the PaymentConfiguration.
func (PaymentConfiguration) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("key"),
	}
}

// Edges of the PaymentConfiguration.
func (PaymentConfiguration) Edges() []ent.Edge {
	return nil
}
