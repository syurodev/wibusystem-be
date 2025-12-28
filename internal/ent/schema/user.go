package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Annotations of the User.
func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "users", Schema: "identify"},
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7).
			Comment("Primary key using UUID v7"),
		field.String("email").
			Unique().
			NotEmpty(),
		field.Bool("email_verified").
			Default(false),
		field.String("password_hash").
			Sensitive(),
		field.String("full_name").
			Optional().
			Nillable(),
		field.String("avatar_url").
			Optional().
			Nillable(),
		field.String("phone").
			Optional().
			Nillable(),
		field.Enum("status").
			Values("active", "suspended", "deleted").
			Default("active"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("last_login_at").
			Optional().
			Nillable(),
		field.JSON("settings", map[string]any{}).
			Optional(),
		field.String("display_name").
			Optional().
			Nillable(),
		field.String("username").
			Optional().
			Nillable().
			Unique(),
		field.JSON("bio", []any{}).
			Optional(),
		field.Bool("is_verified").
			Default(false),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("novels", Novel.Type),
	}
}
